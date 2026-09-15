package video

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/video/admin/", h.handleRoute)
}

func (h *AdminHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/video/admin/"), "/")
	if parts[0] == "list" && r.Method == "GET" {
		keyword := r.URL.Query().Get("keyword")
		status := r.URL.Query().Get("status")
		page, size := httputil.ParsePageParams(r)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT v.id, v.manuscript_id, v.title, v.process_status, m.user_id, m.title
			 FROM videos v LEFT JOIN manuscripts m ON v.manuscript_id = m.id
			 WHERE ($1 = '' OR v.title ILIKE '%'||$1||'%' OR m.title ILIKE '%'||$1||'%')
			   AND ($2 = '' OR v.process_status::text = $2)
			 ORDER BY v.id DESC LIMIT $3 OFFSET $4`, keyword, status, size, (page-1)*size)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id, msID int64
			var title string
			var processStatus int32
			var userID sql.NullInt64
			var msTitle sql.NullString
			rows.Scan(&id, &msID, &title, &processStatus, &userID, &msTitle)
			list = append(list, map[string]interface{}{
				"id": id, "manuscript_id": msID, "title": title,
				"process_status": processStatus, "user_id": userID.Int64, "manuscript_title": msTitle.String,
			})
		}
		json.NewEncoder(w).Encode(list)
	} else if parts[0] == "batch" && r.Method == "DELETE" {
		var ids []int64
		json.NewDecoder(r.Body).Decode(&ids)
		for _, id := range ids {
			h.exec(w, r, `DELETE FROM videos WHERE id=$1`, id)
		}
		w.Write([]byte(`{"status":"ok"}`))
	} else if len(parts) >= 1 && r.Method == "GET" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		row := h.db.QueryRowContext(r.Context(),
			`SELECT id, manuscript_id, video_order, title, play_url_hd, play_url_sd, play_url_ld,
			        duration_seconds, process_progress, COALESCE(process_stage,''), has_subtitle, has_summary
			 FROM videos WHERE id=$1`, id)
		var vid int64
		var msID int64
		var order int32
		var title, hd, sd, ld string
		var dur int32
		var progress int32
		var stage string
		var hasSub, hasSum bool
		if err := row.Scan(&vid, &msID, &order, &title, &hd, &sd, &ld,
			&dur, &progress, &stage, &hasSub, &hasSum); err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": vid, "manuscript_id": msID, "video_order": order, "title": title,
			"play_url_hd": hd, "play_url_sd": sd, "play_url_ld": ld,
			"duration_seconds": dur, "process_progress": progress, "process_stage": stage,
			"has_subtitle": hasSub, "has_summary": hasSum,
		})
	} else if len(parts) >= 1 && r.Method == "DELETE" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `DELETE FROM videos WHERE id=$1`, id)
	}
}

func (h *AdminHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	_, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
