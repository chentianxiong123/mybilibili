package support

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
	mux.HandleFunc("/api/v1/operation/tickets", h.handleCreate)
	mux.HandleFunc("/api/v1/support/admin/tickets", h.handleList)
	mux.HandleFunc("/api/v1/support/admin/tickets/", h.handleTicketByID)
	mux.HandleFunc("/api/v1/operation/internal/tickets/customer-session", h.handleCustomerSession)
	mux.HandleFunc("/api/v1/operation/internal/tickets/session/", h.handleSessionProcess)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
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
	page, size := httputil.ParsePageParams(r)
	list, total, _ := h.svc.List(r.Context(), status, page, size)
	if list == nil {
		list = []*Ticket{}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size,
	})
}

func (h *Handler) handleTicketByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/support/admin/tickets/")
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
		adminID := httputil.GetAdminIDFromHeader(r)
		h.svc.Process(r.Context(), id, adminID, req.AdminReply)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE":
		h.svc.repo.db.ExecContext(r.Context(), `DELETE FROM support_tickets WHERE id = $1`, id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleCustomerSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		UserID   int64  `json:"userId"`
		TicketNo string `json:"ticketNo"`
		Title    string `json:"title"`
		Content  string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.UserID == 0 {
		req.UserID = httputil.GetUserIDFromHeader(r)
	}
	t, err := h.svc.Create(r.Context(), req.UserID, req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ticketId": t.ID, "ticketNo": t.TicketNo, "sessionId": t.SessionID,
	})
}

func (h *Handler) handleSessionProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	sessionID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/operation/internal/tickets/session/"), 10, 64)
	var req struct {
		AdminReply string `json:"adminReply"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_, err := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE support_tickets SET admin_reply = $2, status = 'processed', processed_at = NOW() WHERE session_id = $1`,
		sessionID, req.AdminReply)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
