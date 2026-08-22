package social

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"mybilibili/pkg/httputil"
	"mybilibili/pkg/abstraction"
)

// SearchHistoryHandler 把用户搜索历史存 Redis（TTL 滚动），不落 PostgreSQL。
// 匿名用户搜索历史由前端本地存储（storage 抽象层）负责，不经过此接口。
type SearchHistoryHandler struct {
	cache abstraction.CacheStore
}

func NewSearchHistoryHandler(cache abstraction.CacheStore) *SearchHistoryHandler {
	return &SearchHistoryHandler{cache: cache}
}

const searchHistoryTTL = 30 * 24 * time.Hour
const searchHistoryMax = 20

func searchHistoryKey(userID int64) string {
	return "search:history:" + strconv.FormatInt(userID, 10)
}

func (h *SearchHistoryHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search/history", h.handle)
}

func (h *SearchHistoryHandler) handle(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if h.cache == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "cache unavailable")
		return
	}
	ctx := r.Context()
	key := searchHistoryKey(userID)

	switch r.Method {
	case "GET":
		h.get(ctx, w, key)
	case "POST":
		h.add(ctx, w, r, key)
	case "DELETE":
		_ = h.cache.Delete(ctx, key)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SearchHistoryHandler) get(ctx context.Context, w http.ResponseWriter, key string) {
	data, err := h.cache.Get(ctx, key)
	if err != nil {
		httputil.WriteOK(w, []string{})
		return
	}
	var list []string
	if json.Unmarshal(data, &list) != nil {
		list = []string{}
	}
	httputil.WriteOK(w, list)
}

func (h *SearchHistoryHandler) add(ctx context.Context, w http.ResponseWriter, r *http.Request, key string) {
	var req struct {
		Keyword string `json:"keyword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Keyword == "" {
		httputil.WriteError(w, http.StatusBadRequest, "keyword required")
		return
	}
	var list []string
	if data, err := h.cache.Get(ctx, key); err == nil {
		_ = json.Unmarshal(data, &list)
	}
	// 去重后置顶，最多保留 searchHistoryMax 条
	dedup := make([]string, 0, len(list)+1)
	dedup = append(dedup, req.Keyword)
	for _, kw := range list {
		if kw != req.Keyword {
			dedup = append(dedup, kw)
		}
	}
	if len(dedup) > searchHistoryMax {
		dedup = dedup[:searchHistoryMax]
	}
	raw, _ := json.Marshal(dedup)
	_ = h.cache.Set(ctx, key, raw, searchHistoryTTL)
	httputil.WriteOK(w, dedup)
}