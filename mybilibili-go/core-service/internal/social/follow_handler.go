package social

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type FollowHandler struct {
	svc *FollowService
	db  *sql.DB
}

func NewFollowHandler(svc *FollowService, db *sql.DB) *FollowHandler {
	return &FollowHandler{svc: svc, db: db}
}

func (h *FollowHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/follow/", h.handleFollow)
	mux.HandleFunc("/api/v1/follow/check/", h.handleCheck)
	mux.HandleFunc("/api/v1/follow/me/followers", h.handleMyFollowers)
	mux.HandleFunc("/api/v1/follow/me/following", h.handleMyFollowing)
	mux.HandleFunc("/api/v1/follow/user/", h.handleUserFollows)
}

func (h *FollowHandler) handleFollow(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	if parts[0] == "me" || parts[0] == "check" || parts[0] == "user" {
		return
	}

	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}

	targetID, _ := strconv.ParseInt(parts[0], 10, 64)
	if targetID == 0 {
		http.Error(w, "invalid user id", 400)
		return
	}

	switch r.Method {
	case "POST":
		if err := h.svc.Follow(r.Context(), userID, targetID); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.Unfollow(r.Context(), userID, targetID)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *FollowHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/check/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "missing user id", 400)
		return
	}
	targetID, _ := strconv.ParseInt(parts[0], 10, 64)
	userID := httputil.GetUserIDFromHeader(r)
	ok, _ := h.svc.IsFollowing(r.Context(), userID, targetID)
	json.NewEncoder(w).Encode(map[string]bool{"following": ok})
}

func (h *FollowHandler) handleMyFollowers(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	page, pageSize := httputil.ParsePageParams(r)
	ids, _ := h.svc.ListFollowers(r.Context(), userID, page, pageSize)
	users := h.loadUserBriefs(r.Context(), ids)
	json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": users})
}

func (h *FollowHandler) handleMyFollowing(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	page, pageSize := httputil.ParsePageParams(r)
	ids, _ := h.svc.ListFollowing(r.Context(), userID, page, pageSize)
	users := h.loadUserBriefs(r.Context(), ids)
	json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": users})
}

func (h *FollowHandler) handleUserFollows(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/user/"), "/")
	if len(parts) < 2 {
		http.Error(w, "not found", 404)
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
		http.Error(w, "not found", 404)
		return
	}
	users := h.loadUserBriefs(r.Context(), ids)
	json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": users})
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
