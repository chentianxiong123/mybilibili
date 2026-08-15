package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AdminDataHandler struct {
	db *sql.DB
}

func NewAdminDataHandler(db *sql.DB) *AdminDataHandler {
	return &AdminDataHandler{db: db}
}

func (h *AdminDataHandler) Register(mux *http.ServeMux) {
	// /api/v1/manuscript/admin/ is registered in admin_manuscript_handler.go
	mux.HandleFunc("/api/v1/video/admin/", h.handleVideoAdmin)
	mux.HandleFunc("/api/v1/admin/comment/", h.handleCommentAdmin)
	mux.HandleFunc("/api/v1/admin/live/", h.handleLiveAdmin)
	mux.HandleFunc("/api/v1/admin/meeting/", h.handleMeetingAdmin)
	// /api/v1/admin/security-settings is registered in admin_handler.go
	mux.HandleFunc("/api/v1/admin/content-review/", h.handleContentReview)
}

func (h *AdminDataHandler) handleManuscriptAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/admin/")
	parts := strings.Split(path, "/")
	switch {
	case parts[0] == "pending" && r.Method == "GET":
		h.listManuscripts(w, r, `review_status = 0`)
	case parts[0] == "processing" && r.Method == "GET":
		h.listManuscripts(w, r, `status = 2 AND process_status = 1`)
	case parts[0] == "all" && r.Method == "GET":
		h.listManuscripts(w, r, `1 = 1`)
	case parts[0] == "statistics" && r.Method == "GET":
		json.NewEncoder(w).Encode(h.manuscriptStats(r))
	case len(parts) == 1 && parts[0] != "" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		row := h.db.QueryRowContext(r.Context(),
			`SELECT m.id, m.user_id, m.title, m.description, m.cover_url, m.category_id,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.coin_count,0), COALESCE(m.collect_count,0),
			        m.review_status, m.status, m.process_status, m.created_at, m.updated_at
			 FROM manuscripts m WHERE m.id = $1`, id)
		var mid, uid, catID, views, likes, coins, collects int64
		var title, desc, cover, created, updated string
		var reviewStatus, status, processStatus int32
		if err := row.Scan(&mid, &uid, &title, &desc, &cover, &catID, &views, &likes, &coins, &collects,
			&reviewStatus, &status, &processStatus, &created, &updated); err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": mid, "user_id": uid, "title": title, "description": desc, "cover_url": cover,
			"category_id": catID, "view_count": views, "like_count": likes, "coin_count": coins,
			"collect_count": collects, "review_status": reviewStatus, "status": status,
			"process_status": processStatus, "created_at": created, "updated_at": updated,
		})
	case len(parts) >= 2 && parts[1] == "videos" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT v.id, v.video_order, v.title, v.play_url_hd, v.play_url_sd, v.play_url_ld,
			        COALESCE(v.duration_seconds,0), v.process_status, v.has_subtitle, v.has_summary
			 FROM videos v WHERE v.manuscript_id = $1 ORDER BY v.video_order`, id)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var vid, order int64
			var title, hd, sd, ld string
			var dur, processStatus int64
			var hasSub, hasSum bool
			rows.Scan(&vid, &order, &title, &hd, &sd, &ld, &dur, &processStatus, &hasSub, &hasSum)
			list = append(list, map[string]interface{}{
				"id": vid, "video_order": order, "title": title,
				"play_url_hd": hd, "play_url_sd": sd, "play_url_ld": ld,
				"duration_seconds": dur, "process_status": processStatus,
				"has_subtitle": hasSub, "has_summary": hasSum,
			})
		}
		json.NewEncoder(w).Encode(list)
	case len(parts) >= 2 && parts[0] == "approve" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET review_status=1, status=2, review_time=NOW() WHERE id=$1`, id)
	case len(parts) >= 2 && parts[1] == "approve-with-process" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET review_status=1, status=2, review_time=NOW() WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "reject" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET review_status=2, status=4, review_time=NOW() WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "publish" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=3 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "unpublish" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=-1 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "retry" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE manuscripts SET status=2 WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) listManuscripts(w http.ResponseWriter, r *http.Request, where string) {
	keyword := r.URL.Query().Get("keyword")
	page, size := parsePage(r)
	query := `SELECT m.id, m.user_id, m.title, m.cover_url, m.review_status, m.status, m.process_status, m.created_at
	          FROM manuscripts m WHERE ` + where
	args := []interface{}{}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		query += ` AND (m.title ILIKE $` + itoa(len(args)) + ` OR m.description ILIKE $` + itoa(len(args)) + `)`
	}
	args = append(args, size, (page-1)*size)
	query += ` ORDER BY m.id DESC LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))
	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, uid int64
		var title, cover, created string
		var reviewStatus, status, processStatus int32
		rows.Scan(&id, &uid, &title, &cover, &reviewStatus, &status, &processStatus, &created)
		list = append(list, map[string]interface{}{
			"id": id, "user_id": uid, "title": title, "cover_url": cover,
			"review_status": reviewStatus, "status": status, "process_status": processStatus, "created_at": created,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *AdminDataHandler) manuscriptStats(r *http.Request) map[string]interface{} {
	total := "0"
	pending := "0"
	published := "0"
	rejected := "0"
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts`).Scan(&total)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 0`).Scan(&pending)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE status = 3`).Scan(&published)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE status = 4`).Scan(&rejected)
	return map[string]interface{}{
		"total": total, "pending": pending, "published": published, "rejected": rejected,
	}
}

var itoa = func(n int) string { return strconv.Itoa(n) }

func (h *AdminDataHandler) handleVideoAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/video/admin/"), "/")
	if parts[0] == "list" && r.Method == "GET" {
		keyword := r.URL.Query().Get("keyword")
		status := r.URL.Query().Get("status")
		page, size := parsePage(r)
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
			`SELECT id, manuscript_id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld,
			        duration_seconds, process_progress, process_stage, has_subtitle, has_summary
			 FROM videos WHERE id=$1`, id)
		var vid int64
		var msID int64
		var order int32
		var title, desc, hd, sd, ld string
		var dur int32
		var progress int32
		var stage string
		var hasSub, hasSum bool
		if err := row.Scan(&vid, &msID, &order, &title, &desc, &hd, &sd, &ld,
			&dur, &progress, &stage, &hasSub, &hasSum); err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": vid, "manuscript_id": msID, "video_order": order, "title": title,
			"description": desc, "play_url_hd": hd, "play_url_sd": sd, "play_url_ld": ld,
			"duration_seconds": dur, "process_progress": progress, "process_stage": stage,
			"has_subtitle": hasSub, "has_summary": hasSum,
		})
	} else if len(parts) >= 1 && r.Method == "DELETE" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `DELETE FROM videos WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleCommentAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/comment/"), "/")
	switch {
	case parts[0] == "list" && r.Method == "GET":
		status := r.URL.Query().Get("status")
		keyword := r.URL.Query().Get("keyword")
		page, size := parsePage(r)
		offset := (page - 1) * size
		conds := ""
		args := []interface{}{}
		if status != "" {
			args = append(args, status)
			conds += fmt.Sprintf(" AND c.status = $%d", len(args))
		}
		if keyword != "" {
			args = append(args, "%"+keyword+"%")
			conds += fmt.Sprintf(" AND c.content LIKE $%d", len(args))
		}
		args = append(args, size, offset)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT c.id, c.manuscript_id, c.user_id, c.content, c.like_count, c.reply_count, c.status, c.created_at
			 FROM comments c WHERE 1=1`+conds+` ORDER BY c.id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"list": []map[string]interface{}{}, "total": 0})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id, msID, userID, likeCount, replyCount, st int64
			var content string
			var createdAt time.Time
			rows.Scan(&id, &msID, &userID, &content, &likeCount, &replyCount, &st, &createdAt)
			list = append(list, map[string]interface{}{
				"id": id, "manuscript_id": msID, "user_id": userID, "content": content,
				"like_count": likeCount, "reply_count": replyCount, "status": st,
				"created_at": createdAt.Format("2006-01-02T15:04:05Z"),
			})
		}
		var total int64
		countConds := ""
		countArgs := []interface{}{}
		if status != "" {
			countArgs = append(countArgs, status)
			countConds += fmt.Sprintf(" AND c.status = $%d", len(countArgs))
		}
		if keyword != "" {
			countArgs = append(countArgs, "%"+keyword+"%")
			countConds += fmt.Sprintf(" AND c.content LIKE $%d", len(countArgs))
		}
		h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM comments c WHERE 1=1`+countConds, countArgs...).Scan(&total)
		json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total})
	case len(parts) >= 1 && parts[0] != "" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var msID, userID, likeCount, replyCount, st int64
		var content string
		var createdAt time.Time
		err := h.db.QueryRowContext(r.Context(),
			`SELECT id, manuscript_id, user_id, content, like_count, reply_count, status, created_at
			 FROM comments WHERE id = $1`, id).Scan(&id, &msID, &userID, &content, &likeCount, &replyCount, &st, &createdAt)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": id, "manuscript_id": msID, "user_id": userID, "content": content,
			"like_count": likeCount, "reply_count": replyCount, "status": st,
			"created_at": createdAt.Format("2006-01-02T15:04:05Z"),
		})
	case len(parts) >= 2 && parts[1] == "status" && r.Method == "PUT":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Status int `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.exec(w, r, `UPDATE comments SET status=$1, updated_at=NOW() WHERE id=$2`, req.Status, id)
	case len(parts) >= 2 && parts[1] == "delete" && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `DELETE FROM comments WHERE id=$1`, id)
	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.exec(w, r, `UPDATE comments SET status=1, updated_at=NOW() WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleLiveAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/live/"), "/")
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

func (h *AdminDataHandler) handleMeetingAdmin(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/meeting/"), "/")
	switch {
	case parts[0] == "rooms" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, creator_id, room_name, status, create_time FROM meeting_room ORDER BY create_time DESC LIMIT 50`)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var id, uid, st int64
			var title, t string
			rows.Scan(&id, &uid, &title, &st, &t)
			list = append(list, map[string]interface{}{
				"id": id, "creator_id": uid, "room_name": title, "status": st, "create_time": t,
			})
		}
		json.NewEncoder(w).Encode(list)
	case parts[0] == "pending" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, creator_id, room_name, status, create_time FROM meeting_room WHERE status = 0 ORDER BY create_time DESC LIMIT 50`)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		var list []map[string]interface{}
		for rows.Next() {
			var id, uid, st int64
			var title, t string
			rows.Scan(&id, &uid, &title, &st, &t)
			list = append(list, map[string]interface{}{
				"id": id, "user_id": uid, "title": title, "status": st, "created_at": t,
			})
		}
		json.NewEncoder(w).Encode(list)
	case len(parts) >= 2 && parts[0] == "approve" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE meeting_room SET status=0 WHERE id=$1`, id)
	case len(parts) >= 2 && parts[0] == "reject" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.exec(w, r, `UPDATE meeting_room SET status=4 WHERE id=$1`, id)
	}
}

func (h *AdminDataHandler) handleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{"rate_limit": 100, "captcha_enabled": false})
	} else {
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *AdminDataHandler) handleContentReview(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/content-review/"), "/")
	contentType := r.URL.Query().Get("contentType")
	status := r.URL.Query().Get("status")
	page, size := parsePage(r)
	conds := ""
	args := []interface{}{}
	if contentType != "" {
		args = append(args, contentType)
		conds += fmt.Sprintf(" AND type = $%d", len(args))
	}
	args = append(args, status)
	conds += fmt.Sprintf(" AND ($%d::text = '' OR status = $%d::text)", len(args), len(args))
	args = append(args, page, size)
	switch {
	case parts[0] == "pending" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, type, user_id, content, status, reviewed_at FROM content_reviews
			 WHERE status = 'pending'`+conds+` ORDER BY id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id int64
			var typ, content string
			var userID sql.NullInt64
			var st sql.NullString
			var reviewedAt sql.NullTime
			rows.Scan(&id, &typ, &userID, &content, &st, &reviewedAt)
			list = append(list, map[string]interface{}{
				"id": id, "type": typ, "user_id": userID.Int64, "content": content,
				"status": st.String, "reviewed_at": reviewedAt,
			})
		}
		json.NewEncoder(w).Encode(list)
	case parts[0] == "all" && r.Method == "GET":
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT id, type, user_id, content, status, reviewed_at FROM content_reviews
			 WHERE 1=1`+conds+` ORDER BY id DESC LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
		if err != nil {
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id int64
			var typ, content string
			var userID sql.NullInt64
			var st sql.NullString
			var reviewedAt sql.NullTime
			rows.Scan(&id, &typ, &userID, &content, &st, &reviewedAt)
			list = append(list, map[string]interface{}{
				"id": id, "type": typ, "user_id": userID.Int64, "content": content,
				"status": st.String, "reviewed_at": reviewedAt,
			})
		}
		json.NewEncoder(w).Encode(list)
	case parts[0] == "restore" && len(parts) >= 3 && r.Method == "PUT":
		h.exec(w, r, `UPDATE content_reviews SET status = 'active', reviewed_at = NOW() WHERE type = $1 AND id = $2`, parts[1], parts[2])
	case len(parts) >= 2 && r.Method == "DELETE":
		h.exec(w, r, `UPDATE content_reviews SET status = 'deleted', reviewed_at = NOW() WHERE type = $1 AND id = $2`, parts[0], parts[1])
	case parts[0] == "batch" && r.Method == "POST":
		var req struct {
			Action string          `json:"action"`
			Items  []map[string]any `json:"items"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		for _, item := range req.Items {
			typ, _ := item["type"].(string)
			id, _ := item["id"].(float64)
			if req.Action == "restore" {
				h.exec(w, r, `UPDATE content_reviews SET status = 'active', reviewed_at = NOW() WHERE type = $1 AND id = $2`, typ, int64(id))
			} else {
				h.exec(w, r, `UPDATE content_reviews SET status = 'deleted', reviewed_at = NOW() WHERE type = $1 AND id = $2`, typ, int64(id))
			}
		}
		w.Write([]byte(`{"status":"ok"}`))
	default:
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *AdminDataHandler) exec(w http.ResponseWriter, r *http.Request, query string, args ...interface{}) {
	_, err := h.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}
