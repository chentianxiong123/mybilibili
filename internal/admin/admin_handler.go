package admin

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
	mux.HandleFunc("/api/v1/admin/login", h.handleLogin)
	mux.HandleFunc("/api/v1/admin/register", h.handleRegister)
	mux.HandleFunc("/api/v1/admin/list", h.handleListAdmins)
	mux.HandleFunc("/api/v1/admin/roles", h.handleRoles)
	mux.HandleFunc("/api/v1/admin/roles/", h.handleRolesByID)
	mux.HandleFunc("/api/v1/admin/permissions", h.handlePermissions)
	mux.HandleFunc("/api/v1/admin/audit-logs", h.handleAuditLogs)
	mux.HandleFunc("/api/v1/admin/login-logs", h.handleLoginLogs)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	admin, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	json.NewEncoder(w).Encode(admin)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Level    int32  `json:"level"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.svc.CreateAdmin(r.Context(), req.Username, req.Password, req.Nickname, req.Level); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListAdmins(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListRoles(r.Context())
		json.NewEncoder(w).Encode(list)
	case "POST":
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.CreateRole(r.Context(), req.Name, req.Description)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleRolesByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/roles/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)

	if len(parts) >= 2 && parts[1] == "permissions" {
		if r.Method == "GET" {
			ids, _ := h.svc.GetRolePermissions(r.Context(), id)
			json.NewEncoder(w).Encode(ids)
		} else if r.Method == "PUT" {
			var req struct {
				PermissionIDs []int64 `json:"permission_ids"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			h.svc.SetRolePermissions(r.Context(), id, req.PermissionIDs)
			w.Write([]byte(`{"status":"ok"}`))
		}
		return
	}

	switch r.Method {
	case "PUT":
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateRole(r.Context(), id, req.Name, req.Description)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteRole(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handlePermissions(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListPermissions(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, size := parsePage(r)
	list, _ := h.svc.ListAuditLogs(r.Context(), page, size)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleLoginLogs(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	page, size := parsePage(r)
	list, _ := h.svc.ListLoginLogs(r.Context(), userID, page, size)
	json.NewEncoder(w).Encode(list)
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
