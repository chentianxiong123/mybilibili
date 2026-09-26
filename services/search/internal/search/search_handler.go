package search

import (
	"bytes"
	"encoding/json"
	"io"
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
	mux.HandleFunc("/api/v1/search/health", h.handleHealth)
	mux.HandleFunc("/api/v1/search/videos", h.handleSearch)
	mux.HandleFunc("/api/v1/search/users", h.handleSearchUsers)
	mux.HandleFunc("/api/v1/search/count", h.handleSearchCount)
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

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
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

func (h *Handler) handleSearchCount(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	videoCount, userCount, err := h.svc.Count(r.Context(), keyword)
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": err.Error(), "data": nil})
		return
	}
	writeJSON(w, []int64{videoCount, userCount})
}

func (h *Handler) handleHot(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.Hot(r.Context())
	writeJSON(w, list)
}

func (h *Handler) handleHotIncrement(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	_ = h.svc.IncrementHotSearch(r.Context(), readKeyword(r))
	writeJSON(w, map[string]any{"status": "ok"})
}

// readKeyword 从请求体里取出 keyword 字段，同时认 JSON 与表单。
//
// web 端 /search/word/add 发的是 FormData（multipart/form-data），管理端和
// 测试发 JSON。原来只用 json.NewDecoder 解，multipart 走进去解析失败、req 停在
// 零值，IncrementHotSearch 拿到空串直接返回——热搜从来没有被前端真正写入过，
// 而且是静默成功（永远 200）。
//
// 注意：multipart 的值按原样透传（前端 FormData 本来就是原始 UTF-8），只有
// application/x-www-form-urlencoded 会做百分号解码。JSON 路径不做解码——
// 那里的值是调用方自己给的字面量，猜它是不是编码过的只会把正经的 % 搜索改坏。
func readKeyword(r *http.Request) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	var req struct {
		Keyword string `json:"keyword"`
	}
	if json.Unmarshal(body, &req) == nil {
		return req.Keyword
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil {
		if v := r.FormValue("keyword"); v != "" {
			return v
		}
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err := r.ParseForm(); err == nil {
		return r.FormValue("keyword")
	}
	return ""
}

func (h *Handler) handleHotKeyword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Score   int    `json:"score"`
		Rank    int    `json:"rank"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetKeyword(r.Context(), req.Keyword, req.Score, req.Rank)
	writeJSON(w, map[string]any{"status": "ok"})
}

func (h *Handler) handleHotRank(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Rank    int    `json:"rank"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetRank(r.Context(), req.Keyword, req.Rank)
	writeJSON(w, map[string]any{"status": "ok"})
}

func (h *Handler) handleHotScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
		Score   int    `json:"score"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.SetScore(r.Context(), req.Keyword, req.Score)
	writeJSON(w, map[string]any{"status": "ok"})
}

func (h *Handler) handleCleanExpired(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	_ = h.svc.CleanExpiredHotSearch(r.Context())
	writeJSON(w, map[string]any{"status": "ok"})
}

func (h *Handler) handleHotDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = h.svc.DeleteOne(r.Context(), req.Keyword)
	writeJSON(w, map[string]any{"status": "ok"})
}

func (h *Handler) handleHotGet(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "keyword required", "data": nil})
		return
	}
	result, err := h.svc.GetKeyword(r.Context(), keyword)
	if err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	writeJSON(w, result)
}

func (h *Handler) handleHotScoreGet(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "keyword required", "data": nil})
		return
	}
	score, err := h.svc.GetScore(r.Context(), keyword)
	if err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
		return
	}
	writeJSON(w, map[string]interface{}{"keyword": keyword, "score": score})
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
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	writeJSON(w, st)
}

func (h *Handler) handleIndexValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	st, err := h.svc.GetIndexStatus(r.Context())
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	writeJSON(w, st)
}

func (h *Handler) handleIndexRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	count, err := h.svc.RebuildSearchVector(r.Context())
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	writeJSON(w, map[string]interface{}{
		"status": "success", "message": "全文索引重建完成", "updatedCount": count,
	})
}

func (h *Handler) handleIndexNotNeeded(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":  "success",
		"message": "当前引擎为 PostgreSQL 全文搜索，索引由数据库触发器自动维护，无需手动批量索引",
	})
}

func (h *Handler) handleIndexRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
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
		writeJSON(w, data)
	case "PUT":
		var data map[string]interface{}
		json.NewDecoder(r.Body).Decode(&data)
		b, _ := json.Marshal(data)
		updatedBy := r.Header.Get("X-Username")
		if updatedBy == "" {
			updatedBy = "admin"
		}
		if err := h.svc.UpdateRecommendConfig(r.Context(), string(b), updatedBy); err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
			return
		}
		writeJSON(w, data)
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *Handler) handleRecommendConfigReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
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
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	writeJSON(w, defaults)
}
