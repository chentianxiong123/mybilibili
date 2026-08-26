package live

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

// AdminHandler 直播管理后台接口（打散自 core-service 原 admin 包）。
// 鉴权由 Traefik forwardAuth（jwt-auth + X-Admin-Id）统一处理。
type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/live/admin/", h.handleRoute)
}

func (h *AdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/live/admin/"), "/")
	if parts[0] == "rooms" && r.Method == "GET" && len(parts) == 1 {
		h.handleRooms(w, r)
		return
	}
	if parts[0] == "stats" && r.Method == "GET" && len(parts) == 1 {
		h.handleStats(w, r)
		return
	}
	if len(parts) >= 2 {
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid room id", "data": nil})
			return
		}
		switch {
		case len(parts) == 2 && parts[0] == "rooms" && r.Method == "GET":
			h.handleRoomByID(w, r, id)
			return
		case len(parts) == 3 && parts[0] == "rooms" && parts[2] == "status" && r.Method == "PUT":
			h.handleRoomStatus(w, r, id)
			return
		}
	}
	httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
}

func (h *AdminHandler) handleRooms(w http.ResponseWriter, r *http.Request) {
	page, size := httputil.ParsePageParams(r)
	status := r.URL.Query().Get("status")
	offset := (page - 1) * size
	conds := ""
	args := []interface{}{}
	if status != "" {
		args = append(args, status)
		conds += " AND status = $1"
	}
	args = append(args, size, offset)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, user_id, title, status, viewer_count, created_at
		 FROM live_rooms WHERE 1=1`+conds+` ORDER BY created_at DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		httputil.WriteOK(w, []map[string]any{})
		return
	}
	defer rows.Close()
	var list []map[string]any
	for rows.Next() {
		var id, vc int64
		var uid, st, title, t string
		rows.Scan(&id, &uid, &title, &st, &vc, &t)
		list = append(list, map[string]any{
			"id": id, "user_id": uid, "title": title, "status": st, "viewer_count": vc, "created_at": t,
		})
	}
	httputil.WriteOK(w, list)
}

func (h *AdminHandler) handleRoomByID(w http.ResponseWriter, r *http.Request, id int64) {
	var rid, vc int64
	var uid, st, title, t string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, user_id, title, status, viewer_count, created_at
		 FROM live_rooms WHERE id = $1`, id).Scan(&rid, &uid, &title, &st, &vc, &t)
	if err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "room not found", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{
		"id": rid, "user_id": uid, "title": title, "status": st, "viewer_count": vc, "created_at": t,
	})
}

func (h *AdminHandler) handleRoomStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if _, err := h.db.ExecContext(r.Context(),
		`UPDATE live_rooms SET status = $1, updated_at = NOW() WHERE id = $2`, req.Status, id); err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]string{"status": "ok"})
}

func (h *AdminHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	var liveCount, totalViewers int64
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM live_rooms WHERE status = 1`).Scan(&liveCount)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(viewer_count), 0) FROM live_rooms WHERE status = 1`).Scan(&totalViewers)
	httputil.WriteOK(w, map[string]any{"live_count": liveCount, "total_viewers": totalViewers})
}
