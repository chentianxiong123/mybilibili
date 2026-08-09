package moderation

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/admin/prohibited-words", h.handleWords)
	mux.HandleFunc("/api/v1/admin/prohibited-words/", h.handleWordByID)
	mux.HandleFunc("/api/v1/report/submit", h.handleSubmitReport)
	mux.HandleFunc("/api/v1/admin/report/", h.handleReports)
}

func (h *Handler) handleWords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		page, size := parsePage(r)
		list, _ := h.svc.ListWords(r.Context(), page, size)
		json.NewEncoder(w).Encode(list)
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
	id, _ := strconv.ParseInt(r.URL.Path[len("/api/v1/admin/prohibited-words/"):], 10, 64)
	if r.Method == "DELETE" {
		h.svc.DeleteWord(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleSubmitReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserID(r)
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
	path := r.URL.Path[len("/api/v1/admin/report/"):]
	if path == "list" && r.Method == "GET" {
		page, size := parsePage(r)
		status := r.URL.Query().Get("status")
		list, _ := h.svc.ListReports(r.Context(), page, size, status)
		json.NewEncoder(w).Encode(list)
		return
	}
	if len(path) > 0 && r.Method == "PUT" {
		id, _ := strconv.ParseInt(path, 10, 64)
		var req struct {
			Action string `json:"action"`
			Remark string `json:"admin_remark"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.ProcessReport(r.Context(), id, req.Action, req.Remark)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func getUserID(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}

func parsePage(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return int32(page), int32(size)
}
