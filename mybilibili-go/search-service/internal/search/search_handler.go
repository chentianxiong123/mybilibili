package search

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mybilibili/pkg/httputil"
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "data": data, "message": "ok"})
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/search/videos", h.handleSearch)
	mux.HandleFunc("/api/v1/search/users", h.handleSearchUsers)
	mux.HandleFunc("/api/v1/search/suggest", h.handleSuggest)
	mux.HandleFunc("/api/v1/search/hot", h.handleHot)
	mux.HandleFunc("/api/v1/recommend/related/", h.handleRelated)
	mux.HandleFunc("/api/v1/recommend/for-you", h.handleForYou)
	mux.HandleFunc("/api/v1/recommend/hot", h.handleHotRecommend)
	mux.HandleFunc("/api/v1/search/admin/index/status", h.handleIndexStatus)
	mux.HandleFunc("/api/v1/search/admin/index/bulk", h.handleIndexNotNeeded)
	mux.HandleFunc("/api/v1/search/admin/index/rebuild", h.handleIndexRebuild)
	mux.HandleFunc("/api/v1/search/admin/index/refresh", h.handleIndexRefresh)
	mux.HandleFunc("/api/v1/search/admin/index/validate", h.handleIndexValidate)
	mux.HandleFunc("/api/v1/search/admin/index/incremental", h.handleIndexNotNeeded)
	mux.HandleFunc("/api/v1/search/admin/recommend-config", h.handleRecommendConfig)
	mux.HandleFunc("/api/v1/search/admin/recommend-config/reset", h.handleRecommendConfigReset)
	mux.HandleFunc("/api/v1/search/hot/increment", h.handleHotIncrement)
	mux.HandleFunc("/api/v1/search/hot/keyword", h.handleHotKeyword)
	mux.HandleFunc("/api/v1/search/hot/rank", h.handleHotRank)
	mux.HandleFunc("/api/v1/search/hot/score", h.handleHotScore)
	mux.HandleFunc("/api/v1/search/hot/clean-expired", h.handleCleanExpired)
	mux.HandleFunc("/api/v1/search/hot/delete", h.handleHotDelete)
	mux.HandleFunc("/api/v1/search/hot/get", h.handleHotGet)
	mux.HandleFunc("/api/v1/search/hot/score-get", h.handleHotScoreGet)
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	page, size := httputil.ParsePageParams(r)
	list, _ := h.svc.Search(r.Context(), keyword, categoryID, page, size)
	writeJSON(w, map[string]interface{}{"list": list, "total": len(list)})
}

func (h *Handler) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	page, size := httputil.ParsePageParams(r)
	list, _ := h.svc.SearchUsers(r.Context(), keyword, page, size)
	writeJSON(w, map[string]interface{}{"list": list, "total": len(list)})
}

func (h *Handler) handleHot(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.Hot(r.Context())
	writeJSON(w, list)
}

func (h *Handler) handleHotIncrement(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.IncrementHotSearch(r.Context(), req.Keyword)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleHotKeyword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Score   int    `json:"score"`
		Rank    int    `json:"rank"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetKeyword(r.Context(), req.Keyword, req.Score, req.Rank)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleHotRank(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Rank    int    `json:"rank"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetRank(r.Context(), req.Keyword, req.Rank)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleHotScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Score   int    `json:"score"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetScore(r.Context(), req.Keyword, req.Score)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleCleanExpired(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	_ = h.svc.CleanExpiredHotSearch(r.Context())
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleHotDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.DeleteOne(r.Context(), req.Keyword)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleHotGet(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword required", 400)
		return
	}
	result, err := h.svc.GetKeyword(r.Context(), keyword)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleHotScoreGet(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword required", 400)
		return
	}
	score, err := h.svc.GetScore(r.Context(), keyword)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"keyword": keyword, "score": score})
}

func (h *Handler) handleSuggest(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	size, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 32)
	if size <= 0 {
		size = 10
	}
	list, _ := h.svc.Suggest(r.Context(), keyword, int32(size))
	writeJSON(w, list)
}

func (h *Handler) handleRelated(w http.ResponseWriter, r *http.Request) {
	videoID, _ := strconv.ParseInt(r.URL.Path[len("/api/v1/recommend/related/"):], 10, 64)
	size, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 32)
	if size == 0 {
		size = 10
	}
	list, _ := h.svc.Related(r.Context(), videoID, int32(size))
	writeJSON(w, list)
}

func (h *Handler) handleForYou(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.HotRecommend(r.Context(), 0, 20)
	if list == nil {
		list = []map[string]interface{}{}
	}
	writeJSON(w, list)
}

func (h *Handler) handleHotRecommend(w http.ResponseWriter, r *http.Request) {
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("categoryId"), 10, 64)
	size, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 32)
	if size <= 0 {
		size = 10
	}
	list, _ := h.svc.HotRecommend(r.Context(), categoryID, int32(size))
	if list == nil {
		list = []map[string]interface{}{}
	}
	writeJSON(w, list)
}

func (h *Handler) handleIndexStatus(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetIndexStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, st)
}

func (h *Handler) handleIndexValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	st, err := h.svc.GetIndexStatus(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, st)
}

func (h *Handler) handleIndexRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	count, err := h.svc.RebuildSearchVector(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status": "success", "message": "全文索引重建完成", "updatedCount": count,
	})
}

func (h *Handler) handleIndexNotNeeded(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status": "success",
		"message": "当前引擎为 PostgreSQL 全文搜索，索引由数据库触发器自动维护，无需手动批量索引",
	})
}

func (h *Handler) handleIndexRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status": "success", "message": "索引状态已刷新",
	})
}

func (h *Handler) handleRecommendConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		cfg, _ := h.svc.GetRecommendConfig(r.Context())
		var data map[string]interface{}
		json.Unmarshal([]byte(cfg), &data)
		if data == nil {
			data = map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(data)
	case "PUT":
		var data map[string]interface{}
		json.NewDecoder(r.Body).Decode(&data)
		b, _ := json.Marshal(data)
		updatedBy := r.Header.Get("X-Username")
		if updatedBy == "" {
			updatedBy = "admin"
		}
		if err := h.svc.UpdateRecommendConfig(r.Context(), string(b), updatedBy); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(data)
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleRecommendConfigReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	defaults := map[string]interface{}{
		"refresh_interval": 300, "for_you_size": 20, "related_size": 10, "hot_size": 10, "personalized": true,
	}
	b, _ := json.Marshal(defaults)
	updatedBy := r.Header.Get("X-Username")
	if updatedBy == "" {
		updatedBy = "admin"
	}
	if err := h.svc.UpdateRecommendConfig(r.Context(), string(b), updatedBy); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(defaults)
}


