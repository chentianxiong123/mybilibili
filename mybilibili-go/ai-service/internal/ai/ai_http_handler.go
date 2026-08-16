package ai

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
		resp = map[string]interface{}{"passed": true, "reason": ""}
	}
	passed, _ := resp["passed"].(bool)
	reason, _ := resp["reason"].(string)
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
	userID := getUserID(r)
	var req struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	reply, err := h.customer.Chat(r.Context(), userID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"reply": reply})
}

func (h *AIChatHandler) handleCustomerHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/ai/customer/history/"), 10, 64)
	if userID == 0 {
		userID = getUserID(r)
	}
	history, err := h.customer.History(r.Context(), userID)
	if err != nil {
		history = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(history)
}

func (h *AIChatHandler) handleCustomerTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserID(r)
	if err := h.customer.Transfer(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func getUserID(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}
