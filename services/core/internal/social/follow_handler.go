package social

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/auth"
	"mybilibili/pkg/httputil"
)

type FollowHandler struct {
	svc *FollowService
	db  *sql.DB
	jwt *auth.JWT
}

func NewFollowHandler(svc *FollowService, db *sql.DB, jwt *auth.JWT) *FollowHandler {
	return &FollowHandler{svc: svc, db: db, jwt: jwt}
}

// getUserID 先从 X-User-Id header 取，无网关时 fallback 解析 Authorization Bearer token。
func (h *FollowHandler) getUserID(r *http.Request) int64 {
	uid := httputil.GetUserIDFromHeader(r)
	if uid != 0 {
		return uid
	}
	if h.jwt != nil {
		if tokenStr := auth.TokenFromRequest(r); tokenStr != "" {
			if id, err := h.jwt.ParseUserID(tokenStr); err == nil {
				return id
			}
		}
	}
	return 0
}

func (h *FollowHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/follow/", h.handleFollow)
	mux.HandleFunc("/api/v1/follow/check/", h.handleCheck)
	mux.HandleFunc("/api/v1/follow/me/followers", h.handleMyFollowers)
	mux.HandleFunc("/api/v1/follow/me/following", h.handleMyFollowing)
	mux.HandleFunc("/api/v1/follow/me/fans", h.handleMyFans)
	mux.HandleFunc("/api/v1/follow/me/fans/stats", h.handleMyFansStats)
	mux.HandleFunc("/api/v1/follow/user/", h.handleUserFollows)
	mux.HandleFunc("/api/v1/space/visit", h.handleSpaceVisit)
}

func (h *FollowHandler) handleFollow(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	if parts[0] == "me" || parts[0] == "check" || parts[0] == "user" {
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}

	targetID, _ := strconv.ParseInt(parts[0], 10, 64)
	if targetID == 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid user id", "data": nil})
		return
	}

	switch r.Method {
	case "POST":
		if userID == targetID {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "不能关注自己", "data": nil})
			return
		}
		// 校验目标用户存在，避免抛 pq 外键错误给前端
		if !h.userExists(r.Context(), targetID) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "用户不存在", "data": nil})
			return
		}
		// 可选关注来源稿件（播放页关注按钮透传，用于涨粉归因；body 为空时忽略）
		var sourceManuscriptID int64
		var body struct {
			SourceManuscriptID int64 `json:"source_manuscript_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.SourceManuscriptID > 0 {
			var exists bool
			_ = h.db.QueryRowContext(r.Context(),
				`SELECT EXISTS(SELECT 1 FROM manuscripts WHERE id = $1 AND user_id = $2)`,
				body.SourceManuscriptID, targetID).Scan(&exists)
			if exists {
				sourceManuscriptID = body.SourceManuscriptID
			}
		}
		if err := h.svc.FollowWithSource(r.Context(), userID, targetID, sourceManuscriptID); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "关注失败", "data": nil})
			return
		}
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case "DELETE":
		h.svc.Unfollow(r.Context(), userID, targetID)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *FollowHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/check/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "missing user id", "data": nil})
		return
	}
	targetID, _ := strconv.ParseInt(parts[0], 10, 64)
	userID := h.getUserID(r)
	ok, _ := h.svc.IsFollowing(r.Context(), userID, targetID)
	httputil.WriteOK(w, map[string]any{"following": ok})
}

func (h *FollowHandler) handleMyFollowers(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	page, pageSize := httputil.ParsePageParams(r)
	ids, _ := h.svc.ListFollowers(r.Context(), userID, page, pageSize)
	users := h.loadUserBriefs(r.Context(), ids)
	httputil.WriteOK(w, users)
}

func (h *FollowHandler) handleMyFollowing(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	page, pageSize := httputil.ParsePageParams(r)
	ids, _ := h.svc.ListFollowing(r.Context(), userID, page, pageSize)
	users := h.loadUserBriefs(r.Context(), ids)
	httputil.WriteOK(w, users)
}

func (h *FollowHandler) handleMyFans(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	page, pageSize := httputil.ParsePageParams(r)
	mutual := r.URL.Query().Get("mutual") == "true"
	entries, total, err := h.svc.ListFans(r.Context(), userID, page, pageSize, mutual)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "internal error", "data": nil})
		return
	}
	ids := make([]int64, len(entries))
	for i, e := range entries {
		ids[i] = e.FollowerID
	}
	briefs := h.loadUserBriefs(r.Context(), ids)
	briefByID := make(map[int64]map[string]interface{}, len(briefs))
	for _, b := range briefs {
		if id, ok := b["id"].(int64); ok {
			briefByID[id] = b
		}
	}
	list := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		b, ok := briefByID[e.FollowerID]
		if !ok {
			continue
		}
		b["isMutual"] = e.IsMutual
		b["followedAt"] = e.FollowedAt.Format("2006-01-02 15:04:05")
		list = append(list, b)
	}
	httputil.WriteOK(w, map[string]any{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *FollowHandler) handleMyFansStats(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	stats, err := h.svc.FansStats(r.Context(), userID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "internal error", "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]any{
		"totalFans":    stats.TotalFans,
		"mutualFans":   stats.MutualFans,
		"newFansToday": stats.NewFansToday,
		"newFansWeek":  stats.NewFansWeek,
		"newFansMonth": stats.NewFansMonth,
	})
}

func (h *FollowHandler) handleUserFollows(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/user/"), "/")
	if len(parts) < 2 {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	userID, _ := strconv.ParseInt(parts[0], 10, 64)
	page, pageSize := httputil.ParsePageParams(r)
	var ids []int64
	switch parts[1] {
	case "following":
		ids, _ = h.svc.ListFollowing(r.Context(), userID, page, pageSize)
	case "followers":
		ids, _ = h.svc.ListFollowers(r.Context(), userID, page, pageSize)
	default:
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	users := h.loadUserBriefs(r.Context(), ids)
	httputil.WriteOK(w, users)
}

func (h *FollowHandler) userExists(ctx context.Context, userID int64) bool {
	if h.db == nil || userID == 0 {
		return false
	}
	var exists bool
	err := h.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	return err == nil && exists
}

// loadUserBriefs 批量查询用户信息，返回与 Java UserVO 一致的字段：
// id, username, nickname, avatar, level, signature
func (h *FollowHandler) loadUserBriefs(ctx context.Context, ids []int64) []map[string]interface{} {
	if len(ids) == 0 || h.db == nil {
		return []map[string]interface{}{}
	}
	rows, err := h.db.QueryContext(ctx,
		`SELECT id, COALESCE(username,''), COALESCE(nickname,''), COALESCE(avatar,''), COALESCE(level,1), COALESCE(signature,'')
		 FROM users WHERE id = ANY($1)`, ids)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()
	out := make([]map[string]interface{}, 0, len(ids))
	for rows.Next() {
		var id int64
		var username, nickname, avatar, signature string
		var level int
		rows.Scan(&id, &username, &nickname, &avatar, &level, &signature)
		out = append(out, map[string]interface{}{
			"id":        id,
			"username":  username,
			"nickname":  nickname,
			"avatar":    avatar,
			"level":     level,
			"signature": signature,
		})
	}
	return out
}
