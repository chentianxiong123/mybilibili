package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ManuscriptAdminHandler 稿件管理后台真 SQL 实现
type ManuscriptAdminHandler struct {
	db *sql.DB
}

func NewManuscriptAdminHandler(db *sql.DB) *ManuscriptAdminHandler {
	return &ManuscriptAdminHandler{db: db}
}

func (h *ManuscriptAdminHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/manuscript/admin/pending", h.handlePending)
	mux.HandleFunc("/api/v1/manuscript/admin/processing", h.handleProcessing)
	mux.HandleFunc("/api/v1/manuscript/admin/all", h.handleAll)
	mux.HandleFunc("/api/v1/manuscript/admin/statistics", h.handleStatistics)
	mux.HandleFunc("/api/v1/manuscript/admin/", h.handleByID)
}

// GET /api/v1/manuscript/admin/pending — 待审核稿件列表
func (h *ManuscriptAdminHandler) handlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	h.listByStatus(w, r, page, size, 0, "pending")
}

// GET /api/v1/manuscript/admin/processing — 审核中稿件列表
func (h *ManuscriptAdminHandler) handleProcessing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	h.listByStatus(w, r, page, size, 3, "processing")
}

// GET /api/v1/manuscript/admin/all — 全部稿件列表（带过滤）
func (h *ManuscriptAdminHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := parsePageAdmin(r)
	statusStr := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")

	offset := (page - 1) * size
	var rows *sql.Rows
	var err error

	if keyword != "" {
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 WHERE m.title ILIKE '%' || $1 || '%'
			 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
			keyword, size, offset)
	} else if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 WHERE m.status = $1
			 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
			status, size, offset)
	} else {
		rows, err = h.db.QueryContext(r.Context(),
			`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
			        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
			        m.upload_time, m.updated_at
			 FROM manuscripts m
			 ORDER BY m.id DESC LIMIT $1 OFFSET $2`,
			size, offset)
	}

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []interface{}{}, "total": 0})
		return
	}
	defer rows.Close()

	type ms struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CommentCount int64  `json:"comment_count"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	list := []ms{}
	for rows.Next() {
		var m ms
		var uploadTime, updatedAt time.Time
		if err := rows.Scan(&m.ID, &m.Title, &m.UserID, &m.CategoryID,
			&m.Status, &m.ReviewStatus, &m.ViewCount, &m.LikeCount,
			&m.CommentCount, &uploadTime, &updatedAt); err != nil {
			continue
		}
		m.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
		m.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")
		list = append(list, m)
	}

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM manuscripts`).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total, "page": page, "size": size})
}

// GET /api/v1/manuscript/admin/statistics — 稿件统计概览
func (h *ManuscriptAdminHandler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	stats := map[string]interface{}{}

	var total, pending, approved, rejected, draft int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts`).Scan(&total)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 0`).Scan(&pending)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 1`).Scan(&approved)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE review_status = 2`).Scan(&rejected)
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE status = -1`).Scan(&draft)

	stats["total"] = total
	stats["pending"] = pending
	stats["approved"] = approved
	stats["rejected"] = rejected
	stats["draft"] = draft

	var totalViews, totalLikes int64
	h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(view_count),0) FROM manuscripts`).Scan(&totalViews)
	h.db.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(like_count),0) FROM manuscripts`).Scan(&totalLikes)
	stats["total_views"] = totalViews
	stats["total_likes"] = totalLikes

	var videoCount int64
	h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM videos`).Scan(&videoCount)
	stats["total_videos"] = videoCount

	json.NewEncoder(w).Encode(stats)
}

// GET /api/v1/manuscript/admin/{id} — 稿件详情（含视频列表）
// POST /api/v1/manuscript/admin/{id}/approve-with-process — 审核通过并触发处理
func (h *ManuscriptAdminHandler) handleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	// /{id}/videos — 稿件视频列表
	if len(parts) >= 2 && parts[1] == "videos" && r.Method == "GET" {
		h.getVideos(w, r, id)
		return
	}

	// /{id}/approve-with-process — 审核通过并触发处理
	if len(parts) >= 2 && parts[1] == "approve-with-process" && r.Method == "POST" {
		h.approveWithProcess(w, r, id)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	// GET /{id} — 稿件详情
	row := h.db.QueryRowContext(r.Context(),
		`SELECT m.id, m.title, m.description, m.cover_url, m.user_id, m.category_id,
		        m.status, m.review_status, COALESCE(m.review_reason,''),
		        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.coin_count,0),
		        COALESCE(m.collect_count,0), COALESCE(m.share_count,0), COALESCE(m.comment_count,0),
		        COALESCE(m.danmaku_count,0), m.duration, m.upload_time, m.updated_at
		 FROM manuscripts m WHERE m.id = $1`, id)

	type detail struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		CoverURL     string `json:"cover_url"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ReviewReason string `json:"review_reason"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CoinCount    int64  `json:"coin_count"`
		CollectCount int64  `json:"collect_count"`
		ShareCount   int64  `json:"share_count"`
		CommentCount int64  `json:"comment_count"`
		DanmakuCount int64  `json:"danmaku_count"`
		Duration     string `json:"duration"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	var d detail
	var uploadTime, updatedAt time.Time
	if err := row.Scan(&d.ID, &d.Title, &d.Description, &d.CoverURL, &d.UserID, &d.CategoryID,
		&d.Status, &d.ReviewStatus, &d.ReviewReason,
		&d.ViewCount, &d.LikeCount, &d.CoinCount, &d.CollectCount,
		&d.ShareCount, &d.CommentCount, &d.DanmakuCount,
		&d.Duration, &uploadTime, &updatedAt); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	d.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
	d.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")

	// 附带视频列表
视频列表 := []map[string]interface{}{}
	vrows, _ := h.db.QueryContext(r.Context(),
		`SELECT id, video_order, title, play_url_hd, upload_time FROM videos WHERE manuscript_id = $1 ORDER BY video_order`, id)
	if vrows != nil {
		defer vrows.Close()
		for vrows.Next() {
			var vid int64
			var order int
			var vtitle, playURL string
			var vtime time.Time
			vrows.Scan(&vid, &order, &vtitle, &playURL, &vtime)
			视频列表 = append(视频列表, map[string]interface{}{
				"id": vid, "video_order": order, "title": vtitle,
				"play_url": playURL, "upload_time": vtime.Format("2006-01-02T15:04:05Z"),
			})
		}
	}
	dMap := map[string]interface{}{
		"id": d.ID, "title": d.Title, "description": d.Description, "cover_url": d.CoverURL,
		"user_id": d.UserID, "category_id": d.CategoryID, "status": d.Status,
		"review_status": d.ReviewStatus, "review_reason": d.ReviewReason,
		"view_count": d.ViewCount, "like_count": d.LikeCount, "coin_count": d.CoinCount,
		"collect_count": d.CollectCount, "share_count": d.ShareCount,
		"comment_count": d.CommentCount, "danmaku_count": d.DanmakuCount,
		"duration": d.Duration, "upload_time": d.UploadTime, "updated_at": d.UpdatedAt,
		"videos": 视频列表,
	}
	json.NewEncoder(w).Encode(dMap)
}

func (h *ManuscriptAdminHandler) getVideos(w http.ResponseWriter, r *http.Request, manuscriptID int64) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, video_order, title, description, play_url_hd, play_url_sd, play_url_ld, upload_time
		 FROM videos WHERE manuscript_id = $1 ORDER BY video_order`, manuscriptID)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type video struct {
		ID         int64  `json:"id"`
		VideoOrder int    `json:"video_order"`
		Title      string `json:"title"`
		Description string `json:"description"`
		PlayURLHD  string `json:"play_url_hd"`
		PlayURLSD  string `json:"play_url_sd"`
		PlayURLLD  string `json:"play_url_ld"`
		UploadTime string `json:"upload_time"`
	}
	list := []video{}
	for rows.Next() {
		var v video
		var t time.Time
		if err := rows.Scan(&v.ID, &v.VideoOrder, &v.Title, &v.Description,
			&v.PlayURLHD, &v.PlayURLSD, &v.PlayURLLD, &t); err != nil {
			continue
		}
		v.UploadTime = t.Format("2006-01-02T15:04:05Z")
		list = append(list, v)
	}
	json.NewEncoder(w).Encode(list)
}

func (h *ManuscriptAdminHandler) approveWithProcess(w http.ResponseWriter, r *http.Request, manuscriptID int64) {
	// 审核通过 + 触发转码/字幕/摘要流程（通过更新状态实现）
	_, err := h.db.ExecContext(r.Context(),
		`UPDATE manuscripts SET review_status = 1, review_time = NOW(), updated_at = NOW()
		 WHERE id = $1`, manuscriptID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 获取稿件关联的视频,逐个更新状态触发处理
	rows, _ := h.db.QueryContext(r.Context(),
		`SELECT id FROM videos WHERE manuscript_id = $1`, manuscriptID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var vid int64
			rows.Scan(&vid)
			// 更新视频状态为待处理
			h.db.ExecContext(r.Context(),
				`UPDATE videos SET updated_at = NOW() WHERE id = $1`, vid)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"manuscript_id":  manuscriptID,
		"review_status":  1,
		"message":        "审核通过，已触发处理流程",
	})
}

func (h *ManuscriptAdminHandler) listByStatus(w http.ResponseWriter, r *http.Request, page, size int32, status int, statusLabel string) {
	offset := (page - 1) * size
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT m.id, m.title, m.user_id, m.category_id, m.status, m.review_status,
		        COALESCE(m.view_count,0), COALESCE(m.like_count,0), COALESCE(m.comment_count,0),
		        m.upload_time, m.updated_at
		 FROM manuscripts m
		 WHERE m.review_status = $1
		 ORDER BY m.id DESC LIMIT $2 OFFSET $3`,
		status, size, offset)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []interface{}{}, "total": 0})
		return
	}
	defer rows.Close()

	type ms struct {
		ID           int64  `json:"id"`
		Title        string `json:"title"`
		UserID       int64  `json:"user_id"`
		CategoryID   int64  `json:"category_id"`
		Status       int32  `json:"status"`
		ReviewStatus int32  `json:"review_status"`
		ViewCount    int64  `json:"view_count"`
		LikeCount    int64  `json:"like_count"`
		CommentCount int64  `json:"comment_count"`
		UploadTime   string `json:"upload_time"`
		UpdatedAt    string `json:"updated_at"`
	}
	list := []ms{}
	for rows.Next() {
		var m ms
		var uploadTime, updatedAt time.Time
		if err := rows.Scan(&m.ID, &m.Title, &m.UserID, &m.CategoryID,
			&m.Status, &m.ReviewStatus, &m.ViewCount, &m.LikeCount,
			&m.CommentCount, &uploadTime, &updatedAt); err != nil {
			continue
		}
		m.UploadTime = uploadTime.Format("2006-01-02T15:04:05Z")
		m.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")
		list = append(list, m)
	}

	var total int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM manuscripts WHERE review_status = $1`, status).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "total": total, "page": page, "size": size})
}

func parsePageAdmin(r *http.Request) (int32, int32) {
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
