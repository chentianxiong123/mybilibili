package favorite

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	switch r.Method {
	case "GET":
		// 支持 ?uid=X 查看他人收藏夹；不传则用当前登录用户
		targetUID := httputil.GetUserIDFromHeader(r)
		if uidStr := r.URL.Query().Get("uid"); uidStr != "" {
			if uid, err := strconv.ParseInt(uidStr, 10, 64); err == nil && uid > 0 {
				targetUID = uid
			}
		}
		if targetUID == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			return
		}
		h.listFolders(w, r, targetUID)
	case "POST":
		userID := httputil.GetUserIDFromHeader(r)
		if userID == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			return
		}
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
		UserID     int64  `json:"user_id"`
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
		f.UserID = userID
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

	// 验证 folder 存在
	var exists int
	h.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM favorite_folders WHERE id = $1`, folderID).Scan(&exists)
	if exists == 0 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "收藏夹不存在", "data": nil})
		return
	}

	// PUT/DELETE 操作需要鉴权且 folder 属于当前用户
	if r.Method != "GET" {
		if userID == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
			return
		}
		var cnt int
		h.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM favorite_folders WHERE id = $1 AND user_id = $2`,
			folderID, userID).Scan(&cnt)
		if cnt == 0 {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "收藏夹不存在或无权操作", "data": nil})
			return
		}
	}

	switch r.Method {
	case "GET":
		page, size := httputil.ParsePageParams(r)
		offset := (int(page) - 1) * int(size)
		rule := r.URL.Query().Get("rule")
		orderBy := "ffv.created_at DESC"
		switch rule {
		case "2":
			orderBy = "m.view_count DESC"
		case "3":
			orderBy = "m.like_count DESC"
		}
		rows, err := h.db.QueryContext(r.Context(),
			fmt.Sprintf(`SELECT ffv.manuscript_id, ffv.created_at,
			        m.title, m.description, m.cover_url, m.status, m.review_status,
			        m.duration, m.duration_seconds, m.view_count, m.like_count,
			        m.coin_count, m.collect_count, m.comment_count, m.share_count,
			        m.category_id, m.upload_time,
			        u.id, u.nickname, u.avatar
			 FROM favorite_folder_videos ffv
			 JOIN manuscripts m ON m.id = ffv.manuscript_id
			 LEFT JOIN users u ON u.id = m.user_id
			 WHERE ffv.folder_id = $1
			 ORDER BY %s
			 LIMIT $2 OFFSET $3`, orderBy), folderID, size, offset)
		if err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "查询失败", "data": nil})
			return
		}
		defer rows.Close()
		type videoItem struct {
			Info struct {
				VID  string `json:"vid"`
				FID  int64  `json:"fid"`
				Time string `json:"time"`
			} `json:"info"`
			Video struct {
				VID        string `json:"vid"`
				Title      string `json:"title"`
				CoverUrl   string `json:"coverUrl"`
				Descr      string `json:"descr"`
				Duration   any    `json:"duration"`
				Status     int    `json:"status"`
				UploadDate string `json:"uploadDate"`
			} `json:"video"`
			Stats struct {
				Play     int `json:"play"`
				Collect  int `json:"collect"`
				Like     int `json:"like"`
				Coin     int `json:"coin"`
				Share    int `json:"share"`
				Comment  int `json:"comment"`
				Danmaku  int `json:"danmaku"`
			} `json:"stats"`
			User struct {
				UID      int64  `json:"uid"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
			} `json:"user"`
		}
		list := []videoItem{}
		for rows.Next() {
			var it videoItem
			var manuscriptID int64
			var collectTime time.Time
			var uploadTime time.Time
			var uid int64
			var coverURL, descr, durationStr, nickname, avatar sql.NullString
			var status, reviewStatus, durSec, viewCount, likeCount, coinCount, collectCount, commentCount, shareCount, categoryID int
			if err := rows.Scan(
				&manuscriptID, &collectTime,
				&it.Video.Title, &descr, &coverURL, &status, &reviewStatus,
				&durationStr, &durSec, &viewCount, &likeCount,
				&coinCount, &collectCount, &commentCount, &shareCount,
				&categoryID, &uploadTime,
				&uid, &nickname, &avatar,
			); err != nil {
				continue
			}
			it.Info.VID = strconv.FormatInt(manuscriptID, 10)
			it.Info.FID = folderID
			it.Info.Time = collectTime.Format("2006-01-02T15:04:05Z")
			it.Video.VID = it.Info.VID
			if coverURL.Valid {
				it.Video.CoverUrl = coverURL.String
			}
			if descr.Valid {
				it.Video.Descr = descr.String
			}
			it.Video.Duration = durSec
			// 映射状态：status=3+review_status=1 → 1(已发布)，-1 → 3(已删除)，其他 → 0
			switch {
			case status == 3 && reviewStatus == 1:
				it.Video.Status = 1
			case status == -1:
				it.Video.Status = 3
			default:
				it.Video.Status = 0
			}
			it.Video.UploadDate = uploadTime.Format("2006-01-02T15:04:05Z")
			it.Stats.Play = viewCount
			it.Stats.Collect = collectCount
			it.Stats.Like = likeCount
			it.Stats.Coin = coinCount
			it.Stats.Share = shareCount
			it.Stats.Comment = commentCount
			it.User.UID = uid
			if nickname.Valid {
				it.User.Nickname = nickname.String
			}
			if avatar.Valid {
				it.User.Avatar = avatar.String
			}
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
// handleCheck 检查当前用户是否收藏了某个稿件（接受 manuscript_id / manuscriptId / vid 三种参数名）。
//
// @Summary      检查是否已收藏
// @Description  按稿件 id 或视频 id 判断当前用户是否已收藏；兼容 manuscript_id / manuscriptId / vid 三种参数名
// @Tags         favorite
// @Produce      json
// @Security     BearerAuth
// @Param        manuscript_id  query     int  false  "稿件 id"
// @Param        manuscriptId   query     int  false  "稿件 id（驼峰写法）"
// @Param        vid            query     int  false  "视频 id"
// @Success      200            {object}  string  "Favorite check result"
// @Failure      401            {string}  string  "未登录"
// @Router       /favorites/check [get]
func (h *FavoriteHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	// 兼容 manuscript_id / manuscriptId / vid 三种参数名
	msIDStr := r.URL.Query().Get("manuscript_id")
	if msIDStr == "" {
		msIDStr = r.URL.Query().Get("manuscriptId")
	}
	if msIDStr == "" {
		msIDStr = r.URL.Query().Get("vid")
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

// handleFlatList 当前用户的全部收藏稿件（平铺，不分组）。
//
// @Summary      我的收藏列表（平铺）
// @Description  返回当前登录用户收藏过的全部稿件，按收藏时间倒序
// @Tags         favorite
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "页码（默认 1）"
// @Param        page_size  query     int  false  "每页条数（默认 30）"
// @Success      200        {object}  string  "Favorite list"
// @Failure      401        {string}  string  "未登录"
// @Router       /favorites/list [get]
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
// handleManuscriptFolders 返回某稿件所在的收藏夹列表（用于稿件页展示）。
//
// @Summary      稿件所属收藏夹列表
// @Description  按稿件 id 取该稿件被加入的全部收藏夹
// @Tags         favorite
// @Produce      json
// @Param        id  path      int  true  "稿件 id"
// @Success      200 {object}  string  "Manuscript folders"
// @Failure      400 {string}  string  "invalid manuscript id"
// @Router       /favorites/manuscript/{id} [get]
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


