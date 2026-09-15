package favorite

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
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
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}

	switch r.Method {
	case "GET":
		h.listFolders(w, r, userID)
	case "POST":
		h.createFolder(w, r, userID)
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
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
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
		return
	}
	defer rows.Close()

	type folder struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		VideoCount int64  `json:"video_count"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
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
	httputil.WriteOK(w, list)
}

func (h *FavoriteHandler) createFolder(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "name required", "data": nil})
		return
	}
	var id int64
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO favorite_folders (user_id, name) VALUES ($1, $2) RETURNING id`,
		userID, req.Name).Scan(&id)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "创建失败", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{"id": id, "name": req.Name})
}

// PUT/DELETE /api/v1/favorites/{id}
func (h *FavoriteHandler) handleByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/")
	if path == "" {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}

	// /api/v1/favorites/{id}/manuscripts — POST 添加, DELETE 移除
	if strings.Contains(path, "/manuscripts") {
		h.handleFolderManuscripts(w, r, path)
		return
	}

	// /api/v1/favorites/{id}/videos — GET 分页查询, PUT 批量更新
	if strings.Contains(path, "/videos") {
		h.handleFolderVideos(w, r, path)
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid id", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}

	switch r.Method {
	case "PUT":
		var req struct {
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if strings.TrimSpace(req.Name) == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "name required", "data": nil})
			return
		}
		_, err := h.db.ExecContext(r.Context(),
			`UPDATE favorite_folders SET name = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`,
			req.Name, id, userID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "更新失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case "DELETE":
		_, err := h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folders WHERE id = $1 AND user_id = $2`, id, userID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "删除失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// POST/DELETE /api/v1/favorites/{id}/manuscripts — 添加/移除稿件到收藏夹
func (h *FavoriteHandler) handleFolderManuscripts(w http.ResponseWriter, r *http.Request, path string) {
	// path = "{folderId}/manuscripts" 或 "{folderId}/manuscripts/{manuscriptId}"
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	folderID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid folder id", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}

	// 确认 folder 属于当前用户
	var cnt int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folders WHERE id = $1 AND user_id = $2`,
		folderID, userID).Scan(&cnt)
	if cnt == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "收藏夹不存在或无权操作", "data": nil})
		return
	}

	// DELETE /api/v1/favorites/{folderId}/manuscripts/{manuscriptId}
	if r.Method == http.MethodDelete && len(parts) >= 3 {
		msID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid manuscript id", "data": nil})
			return
		}
		h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folder_videos WHERE folder_id = $1 AND manuscript_id = $2`,
			folderID, msID)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
		return
	}

	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}

	var req struct {
		ManuscriptID int64 `json:"manuscript_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.ManuscriptID == 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "manuscript_id required", "data": nil})
		return
	}
	_, err = h.db.ExecContext(r.Context(),
		`INSERT INTO favorite_folder_videos (folder_id, manuscript_id)
		 VALUES ($1, $2) ON CONFLICT (folder_id, manuscript_id) DO NOTHING`,
		folderID, req.ManuscriptID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "添加失败", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}

// GET|PUT /api/v1/favorites/{folderId}/videos — 分页查询/批量更新收藏夹视频
func (h *FavoriteHandler) handleFolderVideos(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	folderID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid folder id", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}

	// 确认 folder 属于当前用户
	var cnt int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folders WHERE id = $1 AND user_id = $2`,
		folderID, userID).Scan(&cnt)
	if cnt == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "收藏夹不存在或无权操作", "data": nil})
		return
	}

	switch r.Method {
	case "GET":
		page, size := httputil.ParsePageParams(r)
		offset := (int(page) - 1) * int(size)
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT ffv.manuscript_id, m.title, ffv.created_at
			 FROM favorite_folder_videos ffv
			 JOIN manuscripts m ON m.id = ffv.manuscript_id
			 WHERE ffv.folder_id = $1
			 ORDER BY ffv.created_at DESC
			 LIMIT $2 OFFSET $3`, folderID, size, offset)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
			return
		}
		defer rows.Close()
		type item struct {
			ManuscriptID int64  `json:"manuscript_id"`
			Title        string `json:"title"`
			CreatedAt    string `json:"created_at"`
		}
		list := []item{}
		for rows.Next() {
			var it item
			var t time.Time
			if err := rows.Scan(&it.ManuscriptID, &it.Title, &t); err != nil {
				continue
			}
			it.CreatedAt = t.Format("2006-01-02T15:04:05Z")
			list = append(list, it)
		}
		httputil.WriteOK(w, map[string]any{"list": list, "page": page, "size": size})

	case "PUT":
		var req struct {
			ManuscriptIDs []int64 `json:"manuscript_ids"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		tx, err := h.db.BeginTx(r.Context(), nil)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "事务开启失败", "data": nil})
			return
		}
		defer tx.Rollback()
		_, err = tx.ExecContext(r.Context(), `DELETE FROM favorite_folder_videos WHERE folder_id = $1`, folderID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "更新失败", "data": nil})
			return
		}
		for _, msID := range req.ManuscriptIDs {
			_, err = tx.ExecContext(r.Context(),
				`INSERT INTO favorite_folder_videos (folder_id, manuscript_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				folderID, msID)
			if err != nil {
				httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "更新失败", "data": nil})
				return
			}
		}
		tx.Commit()
		httputil.WriteOK(w, map[string]any{"status": "ok"})

	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// GET /api/v1/favorites/check?manuscript_id=xxx — 检查稿件是否已收藏
func (h *FavoriteHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	// 兼容 manuscript_id 和 manuscriptId 两种参数名
	msIDStr := r.URL.Query().Get("manuscript_id")
	if msIDStr == "" {
		msIDStr = r.URL.Query().Get("manuscriptId")
	}
	msID, _ := strconv.ParseInt(msIDStr, 10, 64)
	if msID == 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "manuscript_id required", "data": nil})
		return
	}
	var cnt int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folder_videos ffv
		 JOIN favorite_folders ff ON ff.id = ffv.folder_id
		 WHERE ff.user_id = $1 AND ffv.manuscript_id = $2`,
		userID, msID).Scan(&cnt)
	httputil.WriteOK(w, map[string]any{"favorited": cnt > 0, "count": cnt})
}

// GET /api/v1/favorites/list — 用户全部收藏（平铺，不分文件夹）
func (h *FavoriteHandler) handleFlatList(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	page, size := httputil.ParsePageParams(r)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT ffv.manuscript_id, ff.name, ffv.created_at
		 FROM favorite_folder_videos ffv
		 JOIN favorite_folders ff ON ff.id = ffv.folder_id
		 WHERE ff.user_id = $1
		 ORDER BY ffv.created_at DESC
		 LIMIT $2 OFFSET $3`, userID, size, (page-1)*size)
	if err != nil {
		httputil.WriteOK(w, []any{})
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
	httputil.WriteOK(w, map[string]any{"list": list, "page": page, "size": size})
}

// GET /api/v1/favorites/manuscript/{manuscriptId} — 稿件所在收藏夹列表
func (h *FavoriteHandler) handleManuscriptFolders(w http.ResponseWriter, r *http.Request) {
	msIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/favorites/manuscript/")
	msID, err := strconv.ParseInt(msIDStr, 10, 64)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid manuscript id", "data": nil})
		return
	}
	userID := httputil.GetUserIDFromHeader(r)

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT ff.id, ff.name FROM favorite_folders ff
		 JOIN favorite_folder_videos ffv ON ffv.folder_id = ff.id
		 WHERE ff.user_id = $1 AND ffv.manuscript_id = $2`,
		userID, msID)
	if err != nil {
		httputil.WriteOK(w, []any{})
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
	httputil.WriteOK(w, list)
}


