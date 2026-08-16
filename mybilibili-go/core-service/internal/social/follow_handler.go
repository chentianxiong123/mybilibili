package social

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/pkg/httputil"
)

type FollowHandler struct {
	svc *FollowService
}

func NewFollowHandler(svc *FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
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
	json.NewEncoder(w).Encode(ids)
}

func (h *FollowHandler) handleMyFollowing(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	page, pageSize := httputil.ParsePageParams(r)
	ids, _ := h.svc.ListFollowing(r.Context(), userID, page, pageSize)
	json.NewEncoder(w).Encode(ids)
}

func (h *FollowHandler) handleUserFollows(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/follow/user/"), "/")
	if len(parts) < 2 {
		http.Error(w, "not found", 404)
		return
	}
	userID, _ := strconv.ParseInt(parts[0], 10, 64)
	page, pageSize := httputil.ParsePageParams(r)
	switch parts[1] {
	case "following":
		ids, _ := h.svc.ListFollowing(r.Context(), userID, page, pageSize)
		json.NewEncoder(w).Encode(ids)
	case "followers":
		ids, _ := h.svc.ListFollowers(r.Context(), userID, page, pageSize)
		json.NewEncoder(w).Encode(ids)
	default:
		http.Error(w, "not found", 404)
	}
}


