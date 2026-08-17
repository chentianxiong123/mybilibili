package coreapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type LiveAdminHandler struct {
	db *sql.DB
}

func NewLiveAdminHandler(db *sql.DB) *LiveAdminHandler {
	return &LiveAdminHandler{db: db}
}

func (h *LiveAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/live/admin/", h.handleRoute)
}

func (h *LiveAdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/live/admin/"), "/")
	if parts[0] == "rooms" && r.Method == "GET" {
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, user_id, title, status, viewer_count, created_at
			 FROM live_rooms ORDER BY created_at DESC LIMIT 50`)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var id, uid, st, vc int64
			var title, t string
			rows.Scan(&id, &uid, &title, &st, &vc, &t)
			list = append(list, map[string]interface{}{
				"id": id, "user_id": uid, "title": title, "status": st, "viewer_count": vc, "created_at": t,
			})
		}
		json.NewEncoder(w).Encode(list)
	} else if parts[0] == "stats" && r.Method == "GET" {
		var liveCount, totalViewers int64
		_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM live_rooms WHERE status = 1`).Scan(&liveCount)
		_ = h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(viewer_count), 0) FROM live_rooms WHERE status = 1`).Scan(&totalViewers)
		json.NewEncoder(w).Encode(map[string]interface{}{"live_count": liveCount, "total_viewers": totalViewers})
	}
}
