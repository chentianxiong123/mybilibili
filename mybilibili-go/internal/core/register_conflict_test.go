package core

import (
	"net/http"
	"strings"
	"testing"
)

func TestPublicAPIHandlerMuxNoConflict(t *testing.T) {
	mux := http.NewServeMux()
	mh := NewManuscriptHTTPHandler(nil, nil, nil, nil)
	mh.Register(mux)
	pub := NewPublicAPIHandler(nil)
	pub.Register(mux)
	t.Log("core registration OK")
}

func TestManuscriptRouterDispatches(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/api/v1/manuscript/recommended", "recommended"},
		{"/api/v1/manuscript/hot", "hot"},
		{"/api/v1/manuscript/list", "list"},
		{"/api/v1/manuscript/me/list", "meList"},
		{"/api/v1/manuscript/me/other", ""},
		{"/api/v1/manuscript/category/5", "category"},
		{"/api/v1/manuscript/user/7", "userManuscripts"},
		{"/api/v1/manuscript/user/7/search", "userSearch"},
		{"/api/v1/manuscript/user/7/stats", "userStats"},
		{"/api/v1/manuscript/user/collections", "userCollections"},
		{"/api/v1/manuscript/user/likes", "userLikes"},
		{"/api/v1/manuscript/favorite/folders", "favoriteFolders"},
		{"/api/v1/manuscript/upload-chunk", "uploadChunk"},
		{"/api/v1/manuscript/upload-complete", "uploadCompleteWeb"},
		{"/api/v1/manuscript/123", "detail"},
		{"/api/v1/manuscript/123/status", "status"},
		{"/api/v1/manuscript/123/like", "like"},
		{"/api/v1/manuscript/123/coin", "coin"},
		{"/api/v1/manuscript/123/collect", "collect"},
		{"/api/v1/manuscript/123/share", "share"},
		{"/api/v1/manuscript/upload-session", "uploadSession"},
		{"/api/v1/manuscript/upload-session/abc", "uploadSessionByID"},
		{"/api/v1/manuscript/upload-session/abc/complete", "uploadSessionComplete"},
		{"/api/v1/manuscript/upload-session/a/b", ""},
		{"/api/v1/manuscript/fix-durations", "fixDurations"},
		{"/api/v1/manuscript/internal/5/take-down", "internal"},
		{"/api/v1/manuscript/123/comment-count", "commentCount"},
		{"/api/v1/manuscript/123/increment-comment", "incrementComment"},
		{"/api/v1/manuscript/123/decrement-comment", "decrementComment"},
		{"/api/v1/manuscript/123/nope", ""},
	}
	for _, c := range cases {
		path := strings.TrimPrefix(c.path, "/api/v1/manuscript/")
		got := manuscriptRouteName(strings.Split(path, "/"))
		if got != c.want {
			t.Errorf("path %s: got %q, want %q", c.path, got, c.want)
		}
	}
}
