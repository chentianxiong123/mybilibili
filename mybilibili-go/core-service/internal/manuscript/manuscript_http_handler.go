package manuscript

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"mybilibili/core-service/internal/comment"
	"mybilibili/pkg/httputil"
	"mybilibili/core-service/internal/social"
	"mybilibili/pkg/errors"
	"mybilibili/pkg/imageutil"
	pb "mybilibili/pkg/pb"
)

type CommentService = comment.CommentService
type InteractionService = social.InteractionService

// ManuscriptHTTPHandler 提供稿件域的 HTTP JSON 端点（Flutter App 与 web-ts 直接消费），
// 覆盖公开稿件列表/详情、互动、以及上传会话等内部兜底路由。
// 全部走单一 /api/v1/manuscript/ 子树分发，避免 ServeMux 通配符 pattern 互相冲突。
type ManuscriptHTTPHandler struct {
	db             *sql.DB
	manuscriptSvc  *ManuscriptService
	commentSvc     *CommentService
	interactionSvc *InteractionService
}

func NewManuscriptHTTPHandler(db *sql.DB, manuscriptSvc *ManuscriptService, commentSvc *CommentService, interactionSvc *InteractionService) *ManuscriptHTTPHandler {
	return &ManuscriptHTTPHandler{
		db:             db,
		manuscriptSvc:  manuscriptSvc,
		commentSvc:     commentSvc,
		interactionSvc: interactionSvc,
	}
}

func (h *ManuscriptHTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/manuscript/", h.handleRouter)
}

// handleRouter 手动分派 /api/v1/manuscript/ 下的所有子路径。
func (h *ManuscriptHTTPHandler) handleRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/")
	if path == "" {
		http.Error(w, "not found", 404)
		return
	}
	h.handleManuscriptRoute(w, r, strings.Split(path, "/"))
}

func (h *ManuscriptHTTPHandler) handleManuscriptRoute(w http.ResponseWriter, r *http.Request, parts []string) {
	switch manuscriptRouteName(parts) {
	case "uploadSession":
		h.handleUploadSession(w, r)
	case "uploadSessionByID":
		h.handleUploadSessionByID(w, r)
	case "uploadSessionComplete":
		r.SetPathValue("id", parts[1])
		h.handleUploadComplete(w, r)
	case "uploadChunk":
		h.handleUploadChunk(w, r)
	case "uploadCompleteWeb":
		h.handleUploadCompleteWeb(w, r)
	case "fixDurations":
		h.handleFixDurations(w, r)
	case "internal":
		h.handleInternalByPath(w, r)
	case "recommended":
		h.handleRecommended(w, r)
	case "hot":
		h.handleHot(w, r)
	case "list", "meList":
		h.handleManuscriptList(w, r)
	case "meStats":
		h.handleMyManuscriptStats(w, r)
	case "category":
		r.SetPathValue("id", parts[1])
		h.handleCategory(w, r)
	case "userSearch":
		r.SetPathValue("id", parts[1])
		h.handleUserSearch(w, r)
	case "userStats":
		r.SetPathValue("id", parts[1])
		h.handleUserManuscriptStats(w, r)
	case "userManuscripts":
		r.SetPathValue("id", parts[1])
		h.handleUserManuscripts(w, r)
	case "userCollections":
		h.handleMyCollections(w, r)
	case "userLikes":
		h.handleMyLikes(w, r)
	case "favoriteFolders":
		h.handleFavoriteFolders(w, r)
	case "favoriteFolderByID":
		r.SetPathValue("id", parts[2])
		h.handleFavoriteFolderByID(w, r)
	case "favoriteFolderVideos":
		r.SetPathValue("id", parts[2])
		h.handleFavoriteFolderVideos(w, r)
	case "detail":
		r.SetPathValue("id", parts[0])
		h.handleManuscriptDetail(w, r)
	case "updateManuscript":
		r.SetPathValue("id", parts[0])
		h.handleUpdateManuscript(w, r)
	case "deleteManuscript":
		r.SetPathValue("id", parts[0])
		h.handleDeleteManuscript(w, r)
	case "status":
		r.SetPathValue("id", parts[0])
		h.handleInteractionStatus(w, r)
	case "like":
		r.SetPathValue("id", parts[0])
		h.handleLike(w, r)
	case "coin":
		r.SetPathValue("id", parts[0])
		h.handleCoin(w, r)
	case "collect":
		r.SetPathValue("id", parts[0])
		h.handleCollect(w, r)
	case "share":
		r.SetPathValue("id", parts[0])
		h.handleShare(w, r)
	case "commentCount":
		r.SetPathValue("id", parts[0])
		h.handleCommentCount(w, r)
	case "incrementComment":
		r.SetPathValue("id", parts[0])
		h.handleIncrementComment(w, r)
	case "decrementComment":
		r.SetPathValue("id", parts[0])
		h.handleDecrementComment(w, r)
	default:
		http.Error(w, "not found", 404)
	}
}

// manuscriptRouteName 将 /api/v1/manuscript/ 前缀后的路径段映射为路由名。
func manuscriptRouteName(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "upload-session":
		switch len(parts) {
		case 1:
			return "uploadSession"
		case 2:
			return "uploadSessionByID"
		case 3:
			if parts[2] == "complete" {
				return "uploadSessionComplete"
			}
		}
		return ""
	case "upload-chunk":
		if len(parts) == 1 {
			return "uploadChunk"
		}
		return ""
	case "upload-complete":
		if len(parts) == 1 {
			return "uploadCompleteWeb"
		}
		return ""
	case "fix-durations":
		if len(parts) == 1 {
			return "fixDurations"
		}
		return ""
	case "internal":
		return "internal"
	case "recommended":
		if len(parts) == 1 {
			return "recommended"
		}
		return ""
	case "hot":
		if len(parts) == 1 {
			return "hot"
		}
		return ""
	case "list":
		if len(parts) == 1 {
			return "list"
		}
		return ""
	case "me":
		if len(parts) == 2 && parts[1] == "list" {
			return "meList"
		}
		if len(parts) == 2 && parts[1] == "stats" {
			return "meStats"
		}
		return ""
	case "category":
		if len(parts) == 2 {
			return "category"
		}
		return ""
	case "user":
		if len(parts) == 3 && parts[2] == "search" {
			return "userSearch"
		}
		if len(parts) == 3 && parts[2] == "stats" {
			return "userStats"
		}
		if len(parts) == 2 && parts[1] == "collections" {
			return "userCollections"
		}
		if len(parts) == 2 && parts[1] == "likes" {
			return "userLikes"
		}
		if len(parts) == 2 {
			return "userManuscripts"
		}
		return ""
	case "favorite":
		if len(parts) == 2 && parts[1] == "folders" {
			return "favoriteFolders"
		}
		if len(parts) == 3 && parts[1] == "folders" {
			return "favoriteFolderByID"
		}
		if len(parts) == 4 && parts[1] == "folders" && parts[3] == "videos" {
			return "favoriteFolderVideos"
		}
		return ""
	default:
		switch len(parts) {
		case 1:
			return "detail"
		case 2:
			switch parts[1] {
			case "status":
				return "status"
			case "like":
				return "like"
			case "coin":
				return "coin"
			case "collect":
				return "collect"
			case "share":
				return "share"
			case "comment-count":
				return "commentCount"
			case "increment-comment":
				return "incrementComment"
			case "decrement-comment":
				return "decrementComment"
			case "publish":
				return "publishManuscript"
			case "unpublish":
				return "unpublishManuscript"
			}
		}
		return ""
	}
}

// ---- 上传会话与内部兜底 ----

func (h *ManuscriptHTTPHandler) handleUploadSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}
	var req struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		CategoryID  int64          `json:"category_id"`
		Tags        []string       `json:"tags"`
		Videos      []map[string]any `json:"videos"`
		TotalChunks *int           `json:"total_chunks"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Title == "" && req.CategoryID == 0 && req.TotalChunks == nil {
		http.Error(w, "invalid upload session request", 400)
		return
	}
	sum := md5.Sum([]byte(fmt.Sprintf("%d-%d-%s", userID, req.CategoryID, req.Title)))
	uploadID := hex.EncodeToString(sum[:])
	total := 0
	if req.TotalChunks != nil {
		total = *req.TotalChunks
	}
	for _, v := range req.Videos {
		if tc, ok := v["total_chunks"].(float64); ok {
			total += int(tc)
		}
	}
	tags, _ := json.Marshal(req.Tags)
	videos, _ := json.Marshal(req.Videos)
	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO upload_sessions (id, user_id, title, description, category_id, tags, videos, total_chunks, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending')
		 ON CONFLICT (id) DO UPDATE SET title=$3, updated_at=NOW()`,
		uploadID, userID, req.Title, req.Description, req.CategoryID, string(tags), string(videos), total)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"upload_id": uploadID, "status": "pending", "total_chunks": total})
}

func (h *ManuscriptHTTPHandler) handleUploadSessionByID(w http.ResponseWriter, r *http.Request) {
	uploadID := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/upload-session/")
	if uploadID == "" {
		http.Error(w, "not found", 404)
		return
	}
	userID := httputil.GetUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		row := h.db.QueryRowContext(r.Context(),
			`SELECT id, user_id, title, COALESCE(category_id,0), uploaded_chunks, total_chunks, status
			 FROM upload_sessions WHERE id = $1 AND user_id = $2`, uploadID, userID)
		var id string
		var uid int64
		var title string
		var catID int64
		var uploaded, total int
		var status string
		if err := row.Scan(&id, &uid, &title, &catID, &uploaded, &total, &status); err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"upload_id": id, "user_id": uid, "title": title, "category_id": catID,
			"uploaded_chunks": uploaded, "total_chunks": total, "status": status,
		})
	case "DELETE":
		_, err := h.db.ExecContext(r.Context(), `DELETE FROM upload_sessions WHERE id = $1 AND user_id = $2`, uploadID, userID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// handleUploadComplete 从会话创建稿件与视频（URL 导入）。
func (h *ManuscriptHTTPHandler) handleUploadComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	uploadID := httputil.PathValue(r, "id")
	var owner int64
	if err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM upload_sessions WHERE id=$1`, uploadID).Scan(&owner); err != nil {
		errors.WriteHTTPError(w, errors.ErrNotFound("upload session not found"))
		return
	}
	if owner != uid {
		errors.WriteHTTPError(w, errors.ErrPermissionDenied("forbidden"))
		return
	}
	var title, desc, tags, videos string
	var catID sql.NullInt64
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT title, description, category_id, tags, videos
		 FROM upload_sessions WHERE id = $1`, uploadID).Scan(&title, &desc, &catID, &tags, &videos); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("load upload session failed"))
		return
	}

	playURL := ""
	var vv []map[string]interface{}
	_ = json.Unmarshal([]byte(videos), &vv)
	if len(vv) > 0 {
		if u, ok := vv[0]["url"].(string); ok {
			playURL = u
		}
		if u, ok := vv[0]["video_url"].(string); ok {
			playURL = u
		}
	}
	if playURL == "" {
		errors.WriteHTTPError(w, errors.ErrInvalidArgument("no playable video source"))
		return
	}

	categoryID := int64(0)
	if catID.Valid {
		categoryID = catID.Int64
	}
	var msID int64
	if err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO manuscripts (title, description, cover_url, user_id, category_id, status, review_status, upload_time, updated_at)
		 VALUES ($1,$2,'',$3,$4,0,0,NOW(),NOW()) RETURNING id`,
		title, desc, uid, categoryID).Scan(&msID); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("create manuscript failed"))
		return
	}
	var vid int64
	if err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO videos (manuscript_id, video_order, title, play_url_hd, source_video_url, process_status, upload_time, updated_at)
		 VALUES ($1,0,$2,$3,$3,0,NOW(),NOW()) RETURNING id`,
		msID, title, playURL).Scan(&vid); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("create video failed"))
		return
	}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		var tid int64
		_ = h.db.QueryRowContext(r.Context(),
			`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`, t).Scan(&tid)
		_, _ = h.db.ExecContext(r.Context(),
			`INSERT INTO video_tags (video_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, vid, tid)
	}
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE upload_sessions SET status = 'uploaded', updated_at = NOW() WHERE id = $1`, uploadID)
	httputil.WriteOK(w, map[string]interface{}{"manuscript_id": msID, "status": "uploaded"})
}

func (h *ManuscriptHTTPHandler) handleFixDurations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE manuscripts m SET duration_seconds = COALESCE((
		   SELECT SUM(v.duration_seconds) FROM videos v WHERE v.manuscript_id = m.id
		 ), 0), updated_at = NOW()`)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ManuscriptHTTPHandler) handleInternalByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/manuscript/internal/")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 && parts[1] == "take-down" && r.Method == "PUT" {
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE manuscripts SET status = -1, review_status = 2, updated_at = NOW() WHERE id = $1`, id)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	http.Error(w, "not found", 404)
}

func (h *ManuscriptHTTPHandler) handleCommentCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	count, _ := strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = $2 WHERE id = $1`, id, count)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *ManuscriptHTTPHandler) handleIncrementComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = comment_count + 1 WHERE id = $1`, id)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *ManuscriptHTTPHandler) handleDecrementComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = $1`, id)
	w.Write([]byte(`{"status":"ok"}`))
}

// ---- 公开稿件 ----

// manuscriptToMap 将 pb 稿件编码为 snake_case JSON map，
// 并给 uploader 补 nickname/username（Flutter 读取）。
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func convertKeysToCamel(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, vv := range val {
			ck := snakeToCamel(k)
			out[ck] = convertKeysToCamel(vv)
		}
		return out
	case []interface{}:
		for i, e := range val {
			val[i] = convertKeysToCamel(e)
		}
		return val
	default:
		return v
	}
}

func manuscriptToMap(info *pb.ManuscriptInfo) map[string]interface{} {
	b, _ := json.Marshal(info)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m == nil {
		m = map[string]interface{}{}
	}
	m = convertKeysToCamel(m).(map[string]interface{})
	if up, ok := m["uploader"].(map[string]interface{}); ok {
		if name, ok := up["name"].(string); ok {
			up["nickname"] = name
			up["username"] = name
		}
	}
	return m
}

func manuscriptListToJSON(infos []*pb.ManuscriptInfo) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(infos))
	for _, info := range infos {
		out = append(out, manuscriptToMap(info))
	}
	return out
}

func (h *ManuscriptHTTPHandler) handleRecommended(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	resp, err := h.manuscriptSvc.ListRecommended(r.Context(), &pb.ListRecommendedRequest{UserId: uid})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

func (h *ManuscriptHTTPHandler) handleHot(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	resp, err := h.manuscriptSvc.ListHot(r.Context(), &pb.ListHotRequest{UserId: uid})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

func (h *ManuscriptHTTPHandler) handleManuscriptDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid manuscript id", "data": nil})
		return
	}
	uid := httputil.GetUserIDFromHeader(r)
	resp, err := h.manuscriptSvc.GetManuscriptWithVideos(r.Context(), &pb.GetManuscriptWithVideosRequest{Id: id, CurrentUserId: uid})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptToMap(resp.Manuscript))
}

func (h *ManuscriptHTTPHandler) handleCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	page, size := httputil.ParsePageParams(r)
	resp, err := h.manuscriptSvc.ListByCategory(r.Context(), &pb.ListByCategoryRequest{CategoryId: id, Page: page, PageSize: size})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

func (h *ManuscriptHTTPHandler) handleUserManuscripts(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	status, _ := strconv.ParseInt(r.URL.Query().Get("status"), 10, 32)
	page, size := httputil.ParsePageParams(r)
	resp, err := h.manuscriptSvc.ListUserManuscripts(r.Context(), &pb.ListUserManuscriptsRequest{
		UserId: id, Status: int32(status), Page: page, PageSize: size,
	})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

func (h *ManuscriptHTTPHandler) handleUserSearch(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	keyword := r.URL.Query().Get("keyword")
	sort := r.URL.Query().Get("sort")
	resp, err := h.manuscriptSvc.SearchUserManuscripts(r.Context(), &pb.SearchUserManuscriptsRequest{UserId: id, Keyword: keyword, Sort: sort})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

func (h *ManuscriptHTTPHandler) handleManuscriptList(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	status, _ := strconv.ParseInt(r.URL.Query().Get("status"), 10, 32)
	page, size := httputil.ParsePageParams(r)
	resp, err := h.manuscriptSvc.ListUserManuscripts(r.Context(), &pb.ListUserManuscriptsRequest{
		UserId: uid, Status: int32(status), Page: page, PageSize: size,
	})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, manuscriptListToJSON(resp.Manuscripts))
}

// ---- 更新 / 删除 / 统计 / 收藏点赞列表 ----

// handleUpdateManuscript PUT /api/v1/manuscript/{id} — 更新稿件信息（multipart 或 JSON）。
func (h *ManuscriptHTTPHandler) handleUpdateManuscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid manuscript id", "data": nil})
		return
	}
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		CategoryID  int64    `json:"category_id"`
		Tags        []string `json:"tags"`
	}
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid multipart form", "data": nil})
			return
		}
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		catID, _ := strconv.ParseInt(r.FormValue("categoryId"), 10, 64)
		req.CategoryID = catID
		if tags := r.Form["tags"]; len(tags) > 0 {
			req.Tags = tags
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid body", "data": nil})
			return
		}
	}

	var owner int64
	if err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM manuscripts WHERE id = $1`, id).Scan(&owner); err != nil {
		errors.WriteHTTPError(w, errors.ErrNotFound("manuscript not found"))
		return
	}
	if owner != uid {
		errors.WriteHTTPError(w, errors.ErrPermissionDenied("forbidden"))
		return
	}
	if req.Title != "" {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET title = $1, updated_at = NOW() WHERE id = $2`, req.Title, id)
	}
	if req.Description != "" {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET description = $1, updated_at = NOW() WHERE id = $2`, req.Description, id)
	}
	if req.CategoryID > 0 {
		_, _ = h.db.ExecContext(r.Context(), `UPDATE manuscripts SET category_id = $1, updated_at = NOW() WHERE id = $2`, req.CategoryID, id)
	}
	if len(req.Tags) > 0 {
		var firstVideoID int64
		_ = h.db.QueryRowContext(r.Context(), `SELECT id FROM videos WHERE manuscript_id = $1 ORDER BY video_order LIMIT 1`, id).Scan(&firstVideoID)
		if firstVideoID > 0 {
			_, _ = h.db.ExecContext(r.Context(), `DELETE FROM video_tags WHERE video_id = $1`, firstVideoID)
			for _, t := range req.Tags {
				t = strings.TrimSpace(t)
				if t == "" {
					continue
				}
				var tid int64
				_ = h.db.QueryRowContext(r.Context(),
					`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`, t).Scan(&tid)
				_, _ = h.db.ExecContext(r.Context(),
					`INSERT INTO video_tags (video_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, firstVideoID, tid)
			}
		}
	}
	httputil.WriteOK(w, map[string]interface{}{"id": id, "status": "ok"})
}

// handleDeleteManuscript DELETE /api/v1/manuscript/{id} — 删除稿件。
func (h *ManuscriptHTTPHandler) handleDeleteManuscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid manuscript id", "data": nil})
		return
	}
	_, err := h.manuscriptSvc.DeleteManuscript(r.Context(), &pb.DeleteManuscriptRequest{Id: id, UserId: uid})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}

// handleUserManuscriptStats GET /api/v1/manuscript/user/{id}/stats — 用户稿件统计。
func (h *ManuscriptHTTPHandler) handleUserManuscriptStats(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid user id", "data": nil})
		return
	}
	var total, published, views, likes int64
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE user_id = $1`, id).Scan(&total)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM manuscripts WHERE user_id = $1 AND status >= 0`, id).Scan(&published)
	_ = h.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(SUM(view_count),0), COALESCE(SUM(like_count),0) FROM manuscripts WHERE user_id = $1`, id).Scan(&views, &likes)
	httputil.WriteOK(w, map[string]interface{}{
		"total": total, "published": published, "views": views, "likes": likes,
	})
}

func (h *ManuscriptHTTPHandler) handleMyManuscriptStats(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	r.SetPathValue("id", strconv.FormatInt(uid, 10))
	h.handleUserManuscriptStats(w, r)
}

// manuscriptsByIDs 按 id 集合返回完整稿件列表（保持 id 顺序）。
func (h *ManuscriptHTTPHandler) manuscriptsByIDs(ctx context.Context, ids []int64) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		resp, err := h.manuscriptSvc.GetManuscript(ctx, &pb.GetManuscriptRequest{Id: id})
		if err != nil {
			continue
		}
		out = append(out, manuscriptToMap(resp.Manuscript))
	}
	return out
}

// handleMyCollections GET /api/v1/manuscript/user/collections — 我收藏的稿件。
func (h *ManuscriptHTTPHandler) handleMyCollections(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	page, size := httputil.ParsePageParams(r)
	ids, _ := h.interactionSvc.Repo().GetInteractionIDs(r.Context(), uid, "MANUSCRIPT", "COLLECT")
	// 简单分页：排序后取当前页
	sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })
	start := int((page - 1) * size)
	end := start + int(size)
	if start > len(ids) {
		start = len(ids)
	}
	if end > len(ids) {
		end = len(ids)
	}
	httputil.WriteOK(w, map[string]interface{}{
		"list": h.manuscriptsByIDs(r.Context(), ids[start:end]), "total": len(ids), "page": page, "size": size,
	})
}

// handleMyLikes GET /api/v1/manuscript/user/likes — 我点赞的稿件。
func (h *ManuscriptHTTPHandler) handleMyLikes(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	page, size := httputil.ParsePageParams(r)
	ids, _ := h.interactionSvc.Repo().GetInteractionIDs(r.Context(), uid, "MANUSCRIPT", "LIKE")
	sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })
	start := int((page - 1) * size)
	end := start + int(size)
	if start > len(ids) {
		start = len(ids)
	}
	if end > len(ids) {
		end = len(ids)
	}
	httputil.WriteOK(w, map[string]interface{}{
		"list": h.manuscriptsByIDs(r.Context(), ids[start:end]), "total": len(ids), "page": page, "size": size,
	})
}

// handleFavoriteFolders GET|POST /api/v1/manuscript/favorite/folders — 收藏夹列表/创建。
func (h *ManuscriptHTTPHandler) handleFavoriteFolders(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := h.db.QueryContext(r.Context(),
			`SELECT f.id, f.name, COALESCE(fc.cnt,0)
			 FROM favorite_folders f
			 LEFT JOIN (SELECT folder_id, COUNT(*) AS cnt FROM favorite_folder_videos GROUP BY folder_id) fc ON fc.folder_id = f.id
			 WHERE f.user_id = $1 ORDER BY f.created_at DESC`, uid)
		if err != nil {
			errors.WriteHTTPError(w, errors.ErrInternal("database error"))
			return
		}
		defer rows.Close()
		type folder struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			VideoCount int64  `json:"video_count"`
		}
		list := []folder{}
		for rows.Next() {
			var f folder
			_ = rows.Scan(&f.ID, &f.Name, &f.VideoCount)
			list = append(list, f)
		}
		httputil.WriteOK(w, list)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if strings.TrimSpace(req.Name) == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "name required", "data": nil})
			return
		}
		var id int64
		err := h.db.QueryRowContext(r.Context(),
			`INSERT INTO favorite_folders (user_id, name) VALUES ($1,$2) RETURNING id`, uid, req.Name).Scan(&id)
		if err != nil {
			errors.WriteHTTPError(w, errors.ErrInternal("create folder failed"))
			return
		}
		httputil.WriteOK(w, map[string]interface{}{"id": id, "name": req.Name})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *ManuscriptHTTPHandler) handleFavoriteFolderByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req struct {
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if strings.TrimSpace(req.Name) == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "name required", "data": nil})
			return
		}
		_, err := h.db.ExecContext(r.Context(),
			`UPDATE favorite_folders SET name = $1 WHERE id = $2 AND user_id = $3`, req.Name, id, uid)
		if err != nil {
			errors.WriteHTTPError(w, errors.ErrInternal("update failed"))
			return
		}
		httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
	case http.MethodDelete:
		_, _ = h.db.ExecContext(r.Context(), `DELETE FROM favorite_folder_videos WHERE folder_id = $1`, id)
		_, err := h.db.ExecContext(r.Context(),
			`DELETE FROM favorite_folders WHERE id = $1 AND user_id = $2`, id, uid)
		if err != nil {
			errors.WriteHTTPError(w, errors.ErrInternal("delete failed"))
			return
		}
		httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
	}
}

// handleFavoriteFolderVideos GET /api/v1/manuscript/favorite/folders/{id}/videos — 收藏夹内稿件列表。
func (h *ManuscriptHTTPHandler) handleFavoriteFolderVideos(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	// 校验收藏夹归属
	var owner int64
	err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM favorite_folders WHERE id = $1`, id).Scan(&owner)
	if err == sql.ErrNoRows {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]interface{}{"code": 404, "message": "folder not found", "data": nil})
		return
	}
	if err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("database error"))
		return
	}
	if owner != uid {
		httputil.WriteJSON(w, http.StatusForbidden, map[string]interface{}{"code": 403, "message": "forbidden", "data": nil})
		return
	}
	sortOrder := r.URL.Query().Get("sortOrder")
	order := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		order = "ASC"
	}
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT manuscript_id, created_at FROM favorite_folder_videos WHERE folder_id = $1 ORDER BY created_at `+order, id)
	if err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("database error"))
		return
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var mid int64
		var createdAt any
		_ = rows.Scan(&mid, &createdAt)
		ids = append(ids, mid)
	}
	list := h.manuscriptsByIDs(r.Context(), ids)
	httputil.WriteOK(w, list)
}

// ---- 上传分片（web-ts 直传兼容） ----

func uploadWorkDir() string {
	dir := os.Getenv("MYBILIBILI_UPLOAD_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "mybilibili-uploads")
	}
	return dir
}

// handleUploadChunk POST /api/v1/manuscript/upload-chunk — 保存单个分片到磁盘。
func (h *ManuscriptHTTPHandler) handleUploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid multipart form", "data": nil})
		return
	}
	uploadID := r.FormValue("uploadId")
	if uploadID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "uploadId required", "data": nil})
		return
	}
	var owner int64
	if err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM upload_sessions WHERE id = $1`, uploadID).Scan(&owner); err != nil {
		errors.WriteHTTPError(w, errors.ErrNotFound("upload session not found"))
		return
	}
	if owner != uid {
		errors.WriteHTTPError(w, errors.ErrPermissionDenied("forbidden"))
		return
	}
	partIndex := r.FormValue("partIndex")
	chunkIndex := r.FormValue("chunkIndex")
	if partIndex == "" {
		partIndex = "0"
	}
	if chunkIndex == "" {
		chunkIndex = "0"
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "file required", "data": nil})
		return
	}
	defer file.Close()

	dir := filepath.Join(uploadWorkDir(), uploadID, partIndex)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("create upload dir failed"))
		return
	}
	dst, err := os.Create(filepath.Join(dir, chunkIndex))
	if err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("create chunk file failed"))
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("save chunk failed"))
		return
	}
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE upload_sessions SET uploaded_chunks = uploaded_chunks + 1, updated_at = NOW() WHERE id = $1`, uploadID)
	httputil.WriteOK(w, map[string]interface{}{"status": "ok", "chunk_index": chunkIndex})
}

// handleUploadCompleteWeb POST /api/v1/manuscript/upload-complete — 合并分片并创建稿件。
func (h *ManuscriptHTTPHandler) handleUploadCompleteWeb(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid multipart form", "data": nil})
		return
	}
	uploadID := r.FormValue("uploadId")
	if uploadID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "uploadId required", "data": nil})
		return
	}
	var owner int64
	if err := h.db.QueryRowContext(r.Context(), `SELECT user_id FROM upload_sessions WHERE id = $1`, uploadID).Scan(&owner); err != nil {
		errors.WriteHTTPError(w, errors.ErrNotFound("upload session not found"))
		return
	}
	if owner != uid {
		errors.WriteHTTPError(w, errors.ErrPermissionDenied("forbidden"))
		return
	}

	var title, desc, tags, videos string
	var catID sql.NullInt64
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = $1`,
		uploadID).Scan(&title, &desc, &catID, &tags, &videos); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("load upload session failed"))
		return
	}

	// 封面
	var coverPath string
	cover, coverHeader, err := r.FormFile("cover")
	if err == nil {
		dir := filepath.Join(uploadWorkDir(), uploadID)
		_ = os.MkdirAll(dir, 0o755)
		ext := filepath.Ext(coverHeader.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		coverPath = filepath.Join(dir, "cover"+ext)
		f, err := os.Create(coverPath)
		if err == nil {
			_, _ = io.Copy(f, cover)
			_ = f.Close()
		}
		cover.Close()
		// 压缩为 WebP
		if webpPath, err := imageutil.CompressAndReplace(coverPath); err == nil {
			coverPath = filepath.Join(dir, webpPath)
		}
	}
	coverURL := ""
	if coverPath != "" {
		coverURL = "/uploads/" + uploadID + "/" + filepath.Base(coverPath)
	}

	// 解析会话中的视频元数据（fileName/size/durationSeconds 等）
	var vv []map[string]interface{}
	_ = json.Unmarshal([]byte(videos), &vv)
	partDirs, _ := os.ReadDir(filepath.Join(uploadWorkDir(), uploadID))
	parts := []int{}
	for _, d := range partDirs {
		if !d.IsDir() {
			continue
		}
		n, err := strconv.Atoi(d.Name())
		if err != nil {
			continue
		}
		parts = append(parts, n)
	}
	sort.Ints(parts)
	if len(parts) == 0 {
		errors.WriteHTTPError(w, errors.ErrInvalidArgument("no uploaded video parts"))
		return
	}

	categoryID := int64(0)
	if catID.Valid {
		categoryID = catID.Int64
	}
	var msID int64
	if err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO manuscripts (title, description, cover_url, user_id, category_id, status, review_status, upload_time, updated_at)
		 VALUES ($1,$2,$3,$4,$5,0,0,NOW(),NOW()) RETURNING id`,
		title, desc, coverURL, uid, categoryID).Scan(&msID); err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("create manuscript failed"))
		return
	}

	var allTags []string
	_ = json.Unmarshal([]byte(tags), &allTags)

	for i, part := range parts {
		chunkFiles, _ := os.ReadDir(filepath.Join(uploadWorkDir(), uploadID, strconv.Itoa(part)))
		chunks := []int{}
		for _, c := range chunkFiles {
			if c.IsDir() {
				continue
			}
			n, err := strconv.Atoi(c.Name())
			if err != nil {
				continue
			}
			chunks = append(chunks, n)
		}
		sort.Ints(chunks)
		partDir := filepath.Join(uploadWorkDir(), uploadID, strconv.Itoa(part))
		mergedPath := filepath.Join(partDir, "merged")
		out, err := os.Create(mergedPath)
		if err != nil {
			continue
		}
		for _, c := range chunks {
			src, err := os.Open(filepath.Join(partDir, strconv.Itoa(c)))
			if err != nil {
				continue
			}
			_, _ = io.Copy(out, src)
			_ = src.Close()
		}
		_ = out.Close()

		playURL := "/uploads/" + uploadID + "/" + strconv.Itoa(part) + "/merged"
		vt := "P" + strconv.Itoa(i+1)
		if i < len(vv) {
			if t, ok := vv[i]["title"].(string); ok && t != "" {
				vt = t
			}
		}
		var vid int64
		if err := h.db.QueryRowContext(r.Context(),
			`INSERT INTO videos (manuscript_id, video_order, title, play_url_hd, source_video_url, process_status, upload_time, updated_at)
			 VALUES ($1,$2,$3,$4,$4,0,NOW(),NOW()) RETURNING id`,
			msID, i, vt, playURL).Scan(&vid); err != nil {
			continue
		}
		for _, t := range allTags {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			var tid int64
			_ = h.db.QueryRowContext(r.Context(),
				`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`, t).Scan(&tid)
			_, _ = h.db.ExecContext(r.Context(),
				`INSERT INTO video_tags (video_id, tag_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, vid, tid)
		}
	}
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE upload_sessions SET status = 'uploaded', updated_at = NOW() WHERE id = $1`, uploadID)
	httputil.WriteOK(w, map[string]interface{}{"manuscript_id": msID, "status": "uploaded"})
}

// ---- 互动 ----

func (h *ManuscriptHTTPHandler) handleInteractionStatus(w http.ResponseWriter, r *http.Request) {	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid manuscript id", "data": nil})
		return
	}
	uid := httputil.GetUserIDFromHeader(r)
	statusResp, err := h.interactionSvc.GetInteractionStatus(r.Context(), &pb.GetInteractionStatusRequest{ManuscriptId: id, UserId: uid})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	infoResp, err := h.manuscriptSvc.GetManuscript(r.Context(), &pb.GetManuscriptRequest{Id: id})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	coined := false
	if uid > 0 {
		coined, _ = h.interactionSvc.Repo().HasInteraction(r.Context(), uid, "MANUSCRIPT", "COIN", id)
	}
	m := infoResp.Manuscript
	httputil.WriteOK(w, map[string]interface{}{
		"liked":        statusResp.Liked,
		"coined":       coined,
		"collected":    statusResp.Collected,
		"likeCount":    m.LikeCount,
		"coinCount":    m.CoinCount,
		"collectCount": m.CollectCount,
	})
}

func (h *ManuscriptHTTPHandler) handleLike(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		_, err := h.interactionSvc.LikeManuscript(r.Context(), &pb.LikeManuscriptRequest{ManuscriptId: id, UserId: uid})
		if err != nil {
			errors.WriteHTTPError(w, err)
			return
		}
	case http.MethodDelete:
		_, err := h.interactionSvc.UnlikeManuscript(r.Context(), &pb.UnlikeManuscriptRequest{ManuscriptId: id, UserId: uid})
		if err != nil {
			errors.WriteHTTPError(w, err)
			return
		}
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}

func (h *ManuscriptHTTPHandler) handleCoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	count := int32(2)
	if q := r.URL.Query().Get("coinCount"); q != "" {
		n, _ := strconv.ParseInt(q, 10, 32)
		if n > 0 {
			count = int32(n)
		}
	} else {
		var body struct {
			Count int32 `json:"count"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Count > 0 {
			count = body.Count
		}
	}
	_, err := h.interactionSvc.CoinManuscript(r.Context(), &pb.CoinManuscriptRequest{ManuscriptId: id, UserId: uid, CoinCount: count})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}

func (h *ManuscriptHTTPHandler) handleCollect(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		var body struct {
			FolderId int64 `json:"folderId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, err := h.interactionSvc.CollectManuscript(r.Context(), &pb.CollectManuscriptRequest{ManuscriptId: id, UserId: uid, FolderId: body.FolderId})
		if err != nil {
			errors.WriteHTTPError(w, err)
			return
		}
	case http.MethodDelete:
		_, err := h.interactionSvc.UncollectManuscript(r.Context(), &pb.UncollectManuscriptRequest{ManuscriptId: id, UserId: uid})
		if err != nil {
			errors.WriteHTTPError(w, err)
			return
		}
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}

func (h *ManuscriptHTTPHandler) handleShare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	id, _ := strconv.ParseInt(httputil.PathValue(r, "id"), 10, 64)
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Channel string `json:"channel"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	_, err := h.interactionSvc.ShareManuscript(r.Context(), &pb.ShareManuscriptRequest{ManuscriptId: id, UserId: uid, Channel: body.Channel})
	if err != nil {
		errors.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}