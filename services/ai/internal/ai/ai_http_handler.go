package ai

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type AIChatHandler struct {
	review   *ReviewService
	customer *CustomerService
}

func NewAIChatHandler(review *ReviewService, customer *CustomerService) *AIChatHandler {
	return &AIChatHandler{review: review, customer: customer}
}

func (h *AIChatHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/ai/review/content", h.handleReviewContent)
	mux.HandleFunc("/api/v1/ai/review/comment", h.handleReviewComment)
	mux.HandleFunc("/api/v1/ai/review/reply", h.handleReviewReply)
	mux.HandleFunc("/api/v1/ai/review/report", h.handleReviewReport)
	mux.HandleFunc("/api/v1/ai/customer/chat", h.handleCustomerChat)
	mux.HandleFunc("/api/v1/ai/customer/history/", h.handleCustomerHistory)
	mux.HandleFunc("/api/v1/ai/customer/transfer", h.handleCustomerTransfer)
}

func (h *AIChatHandler) handleReviewContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	content := r.URL.Query().Get("content")
	scene := r.URL.Query().Get("scene")
	if scene == "" {
		scene = "COMMENT"
	}
	var tags []string
	resp, err := h.review.Moderate(r.Context(), content, scene)
	if err == nil {
		passed, _ := resp["passed"].(bool)
		if passed {
			tags = []string{}
		} else if tagsRaw, ok := resp["tags"].([]interface{}); ok {
			for _, t := range tagsRaw {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *AIChatHandler) handleReviewComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	h.moderateWith(w, r, "COMMENT")
}

func (h *AIChatHandler) handleReviewReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	h.moderateWith(w, r, "REPLY")
}

func (h *AIChatHandler) handleReviewReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	h.moderateWith(w, r, "REPORT")
}

func (h *AIChatHandler) moderateWith(w http.ResponseWriter, r *http.Request, scene string) {
	content := r.URL.Query().Get("content")
	resp, err := h.review.Moderate(r.Context(), content, scene)
	if err != nil {
		// FINDING-050 修复：fail-closed——审核 provider 故障不得自动批准（此前
		// passed:true 使不当内容在审核服务不可用时全部放行）。
		log.Printf("ai review %s failed: %v — fail closed", scene, err)
		resp = map[string]interface{}{"passed": false, "reason": "审核服务不可用，请稍后重试"}
	}
	passed, _ := resp["passed"].(bool)
	reason, _ := resp["reason"].(string)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"passed": passed, "reason": reason,
		"provider": r.URL.Query().Get("provider"),
	})
}

func (h *AIChatHandler) handleCustomerChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	var req struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	reply, err := h.customer.Chat(r.Context(), userID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"reply": reply})
}

func (h *AIChatHandler) handleCustomerHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	// FINDING-038 修复：聊天历史需认证；非 admin 时强制用认证 uid（忽略路径 id，
	// 防 IDOR 泄露任意用户历史）。
	if httputil.GetAdminIDFromHeader(r) == 0 {
		uid, ok := httputil.RequireUser(w, r)
		if !ok {
			return
		}
		r.Header.Set("X-User-Id", strconv.FormatInt(uid, 10))
	}
	userID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/ai/customer/history/"), 10, 64)
	if userID == 0 {
		userID = httputil.GetUserIDFromHeader(r)
	}
	history, err := h.customer.History(r.Context(), userID)
	if err != nil {
		history = []map[string]interface{}{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 200, "data": history,
	})
}

func (h *AIChatHandler) handleCustomerTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if err := h.customer.Transfer(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}


