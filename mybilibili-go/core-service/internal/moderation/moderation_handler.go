package moderation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/moderation/admin/prohibited-words/batch-import", h.handleBatchImport)
	mux.HandleFunc("/api/v1/moderation/admin/prohibited-words", h.handleWords)
	mux.HandleFunc("/api/v1/moderation/admin/prohibited-words/", h.handleWordByID)
	mux.HandleFunc("/api/v1/report/submit", h.handleSubmitReport)
	mux.HandleFunc("/api/v1/moderation/admin/report/", h.handleReports)
}

func (h *Handler) handleBatchImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Words []*ProhibitedWord `json:"words"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	imported, err := h.svc.BatchImportWords(r.Context(), req.Words)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"imported": imported})
}

func (h *Handler) handleWords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		page, size := httputil.ParsePageParams(r)
		list, _ := h.svc.ListWords(r.Context(), page, size)
		if list == nil {
			list = []*ProhibitedWord{}
		}
		var total int64
		h.svc.repo.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM prohibited_words`).Scan(&total)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"list": list, "total": total, "page": page, "size": size,
		})
	case "POST":
		var req struct {
			Word      string `json:"word"`
			MatchType string `json:"match_type"`
			Category  string `json:"category"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.CreateWord(r.Context(), req.Word, req.MatchType, req.Category)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleWordByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/admin/prohibited-words/")
	id, _ := strconv.ParseInt(path, 10, 64)
	switch r.Method {
	case "GET":
		word, err := h.svc.GetWord(r.Context(), id)
		if err != nil {
			http.Error(w, "word not found", 404)
			return
		}
		json.NewEncoder(w).Encode(word)
	case "PUT":
		var req struct {
			Word      string `json:"word"`
			MatchType string `json:"match_type"`
			Category  string `json:"category"`
			IsEnabled int32  `json:"is_enabled"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.IsEnabled == 0 {
			req.IsEnabled = 1
		}
		h.svc.UpdateWord(r.Context(), id, req.Word, req.MatchType, req.Category, req.IsEnabled)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteWord(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleSubmitReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	var req struct {
		TargetType   string `json:"target_type"`
		TargetID     int64  `json:"target_id"`
		ManuscriptID int64  `json:"manuscript_id"`
		Reason       string `json:"reason"`
		Description  string `json:"description"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.svc.SubmitReport(r.Context(), userID, req.TargetType, req.TargetID, req.ManuscriptID, req.Reason, req.Description); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleReports(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/admin/report/")
	if path == "list" && r.Method == "GET" {
		page, size := httputil.ParsePageParams(r)
		status := r.URL.Query().Get("status")
		targetType := r.URL.Query().Get("target_type")
		_ = targetType
		list, _ := h.svc.ListReports(r.Context(), page, size, status)
		json.NewEncoder(w).Encode(list)
		return
	}
	if path == "ai-review-result" && r.Method == "PUT" {
		var req struct {
			ReportID  int64  `json:"reportId"`
			Verdict   string `json:"verdict"`
			RiskLevel string `json:"riskLevel"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateAIRegReview(r.Context(), req.ReportID, req.Verdict, req.RiskLevel)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) >= 2 && parts[0] == "process" && r.Method == "PUT" {
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		var req struct {
			Action string `json:"action"`
			Remark string `json:"admin_remark"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.ProcessReport(r.Context(), id, req.Action, req.Remark)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	http.Error(w, "not found", 404)
}
