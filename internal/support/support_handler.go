package support

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/operation/tickets", h.handleCreate)
	mux.HandleFunc("/api/v1/operation/admin/tickets", h.handleList)
	mux.HandleFunc("/api/v1/operation/admin/tickets/", h.handleTicketByID)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserID(r)
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Title == "" {
		http.Error(w, "title required", 400)
		return
	}
	t, err := h.svc.Create(r.Context(), userID, req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, size := parsePage(r)
	list, _ := h.svc.List(r.Context(), status, page, size)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleTicketByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/operation/admin/tickets/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	switch {
	case r.Method == "GET":
		t, err := h.svc.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(t)
	case len(parts) >= 2 && parts[1] == "process" && r.Method == "PUT":
		var req struct {
			AdminReply string `json:"adminReply"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		adminID := getAdminID(r)
		h.svc.Process(r.Context(), id, adminID, req.AdminReply)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE":
		h.svc.repo.db.ExecContext(r.Context(), `DELETE FROM support_tickets WHERE id = $1`, id)
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

func getAdminID(r *http.Request) int64 {
	idStr := r.Header.Get("X-Admin-Id")
	if idStr == "" {
		return getUserID(r)
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
	if size < 1 || size > 50 {
		size = 20
	}
	return int32(page), int32(size)
}
