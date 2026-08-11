package core

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FavoriteHandler 收藏夹/收藏管理 HTTP 接口
type FavoriteHandler struct {
	db *sql.DB
}

func NewFavoriteHandler(db *sql.DB) *FavoriteHandler {
	return &FavoriteHandler{db: db}
}

func (h *FavoriteHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/favorites", h.handleFavorites)
	mux.HandleFunc("/api/v1/favorites/check", h.handleCheck)
	mux.HandleFunc("/api/v1/favorites/list", h.handleFlatList)
	mux.HandleFunc("/api/v1/favorites/manuscript/", h.handleManuscriptFolders)
	mux.HandleFunc("/api/v1/favorites/", h.handleByID)
}

// GET /api/v1/favorites — 用户收藏夹列表
func (h *FavoriteHandler) handleFavorites(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromReq(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}

	switch r.Method {
	case "GET":
		h.listFolders(w, r, userID)
	case "POST":
		h.createFolder(w, r, userID)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *FavoriteHandler) listFolders(w http.ResponseWriter, r *http.Request, userID int64) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT f.id, f.name, f.created_at, f.updated_at,
		        COALESCE(fc.cnt, 0) AS video_count
		 FROM favorite_folders f
		 LEFT JOIN (
		     SELECT folder_id, COUNT(*) AS cnt
		     FROM favorite_folder_videos
		     GROUP BY folder_id
		 ) fc ON fc.folder_id = f.id
		 WHERE f.user_id = $1
		 ORDER BY f.created_at DESC`, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	type folder struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		VideoCount int64 `json:"video_count"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	list := []folder{}
	for rows.Next() {
		var f folder
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&f.ID, &f.Name, &createdAt, &updatedAt, &f.VideoCount); err != nil {
			continue
		}
		f.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z")
		f.UpdatedAt = updatedAt.Format("2006-01-02T15:04:05Z")
		list = append(list, f)
	}
	json.NewEncoder(w).Encode(list)
}

func (h *FavoriteHandler) createFolder(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name required", 400)
		return
	}
	var id int64
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO favorite_folders (user_id, name) VALUES ($1, $2) RETURNING id`,
		userID, req.Name).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   id,
		"name": req.Name,
	})
}

// PUT/DELETE /api/v1/favorites/{id}
func (h *FavoriteHandler) handleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/")
	if path == "" {
		http.Error(w, "not found", 404)
		return
	}

	// /api/v1/favorites/{id}/manuscripts — POST 添加, DELETE 移除
	if strings.Contains(path, "/manuscripts") {
		h.handleFolderManuscripts(w, r, path)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}
	userID := getUserIDFromReq(r)

	switch r.Method {
	case "PUT":
		var req struct {
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if strings.TrimSpace(req.Name) == "" {
			http.Error(w, "name required", 400)
			return
		}
		_, err := h.db.ExecContext(r.Context(),
			`UPDATE favorite_folders SET name = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`,
			req.Name, id, userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		_, err := h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folders WHERE id = $1 AND user_id = $2`, id, userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// POST/DELETE /api/v1/favorites/{id}/manuscripts — 添加/移除稿件到收藏夹
func (h *FavoriteHandler) handleFolderManuscripts(w http.ResponseWriter, r *http.Request, path string) {
	// path = "{folderId}/manuscripts" 或 "{folderId}/manuscripts/{manuscriptId}"
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "not found", 404)
		return
	}
	folderID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid folder id", 400)
		return
	}
	userID := getUserIDFromReq(r)

	// 确认 folder 属于当前用户
	var cnt int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folders WHERE id = $1 AND user_id = $2`,
		folderID, userID).Scan(&cnt)
	if cnt == 0 {
		http.Error(w, "folder not found", 404)
		return
	}

	// DELETE /api/v1/favorites/{folderId}/manuscripts/{manuscriptId}
	if r.Method == "DELETE" && len(parts) >= 3 {
		msID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			http.Error(w, "invalid manuscript id", 400)
			return
		}
		h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folder_videos WHERE folder_id = $1 AND manuscript_id = $2`,
			folderID, msID)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	var req struct {
		ManuscriptID int64 `json:"manuscript_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.ManuscriptID == 0 {
		http.Error(w, "manuscript_id required", 400)
		return
	}
	_, err = h.db.ExecContext(r.Context(),
		`INSERT INTO favorite_folder_videos (folder_id, manuscript_id)
		 VALUES ($1, $2) ON CONFLICT (folder_id, manuscript_id) DO NOTHING`,
		folderID, req.ManuscriptID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(`{"status":"ok"}`))
}

// GET /api/v1/favorites/check?manuscript_id=xxx — 检查稿件是否已收藏
func (h *FavoriteHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromReq(r)
	msID, _ := strconv.ParseInt(r.URL.Query().Get("manuscript_id"), 10, 64)
	if msID == 0 {
		http.Error(w, "manuscript_id required", 400)
		return
	}
	var cnt int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folder_videos ffv
		 JOIN favorite_folders ff ON ff.id = ffv.folder_id
		 WHERE ff.user_id = $1 AND ffv.manuscript_id = $2`,
		userID, msID).Scan(&cnt)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"favorited": cnt > 0,
		"count":     cnt,
	})
}

// GET /api/v1/favorites/list — 用户全部收藏（平铺，不分文件夹）
func (h *FavoriteHandler) handleFlatList(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromReq(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}
	page, size := parsePageFromReq(r)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT ffv.manuscript_id, ff.name, ffv.created_at
		 FROM favorite_folder_videos ffv
		 JOIN favorite_folders ff ON ff.id = ffv.folder_id
		 WHERE ff.user_id = $1
		 ORDER BY ffv.created_at DESC
		 LIMIT $2 OFFSET $3`, userID, size, (page-1)*size)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type fav struct {
		ManuscriptID int64  `json:"manuscript_id"`
		FolderName   string `json:"folder_name"`
		CreatedAt    string `json:"created_at"`
	}
	list := []fav{}
	for rows.Next() {
		var f fav
		var t time.Time
		if err := rows.Scan(&f.ManuscriptID, &f.FolderName, &t); err != nil {
			continue
		}
		f.CreatedAt = t.Format("2006-01-02T15:04:05Z")
		list = append(list, f)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "page": page, "size": size})
}

// GET /api/v1/favorites/manuscript/{manuscriptId} — 稿件所在收藏夹列表
func (h *FavoriteHandler) handleManuscriptFolders(w http.ResponseWriter, r *http.Request) {
	msIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/manuscript/")
	msID, err := strconv.ParseInt(msIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid manuscript id", 400)
		return
	}
	userID := getUserIDFromReq(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT ff.id, ff.name FROM favorite_folders ff
		 JOIN favorite_folder_videos ffv ON ffv.folder_id = ff.id
		 WHERE ff.user_id = $1 AND ffv.manuscript_id = $2`,
		userID, msID)
	if err != nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type folder struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	list := []folder{}
	for rows.Next() {
		var f folder
		rows.Scan(&f.ID, &f.Name)
		list = append(list, f)
	}
	json.NewEncoder(w).Encode(list)
}

func getUserIDFromReq(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}

func parsePageFromReq(r *http.Request) (int32, int32) {
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
