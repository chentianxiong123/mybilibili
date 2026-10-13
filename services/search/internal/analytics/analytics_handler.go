package analytics

import (
	"net/http"
	"strconv"

	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/creator/stats/overview", h.handleOverview)
	mux.HandleFunc("/api/v1/creator/stats/trend", h.handleTrend)
	mux.HandleFunc("/api/v1/creator/stats/ranking", h.handleRanking)
	mux.HandleFunc("/api/v1/creator/stats/latest-comments", h.handleLatestComments)
	mux.HandleFunc("/api/v1/creator/stats/fans-trend", h.handleFansTrend)
	mux.HandleFunc("/api/v1/creator/stats/fans-ranking", h.handleFansRanking)
	mux.HandleFunc("/api/v1/creator/stats/manuscript-trend", h.handleManuscriptTrend)
	mux.HandleFunc("/api/v1/creator/stats/fans-by-manuscript", h.handleFansByManuscript)
	mux.HandleFunc("/api/v1/creator/stats/manuscript-insights", h.handleManuscriptInsights)
	mux.HandleFunc("/api/v1/creator/stats/income", h.handleIncome)
}

func (h *Handler) handleOverview(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.Overview(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleTrend(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.Trend(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleRanking(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	sortBy := r.URL.Query().Get("sortBy")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, _ := h.svc.Ranking(r.Context(), userID, sortBy, limit)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleLatestComments(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, _ := h.svc.LatestComments(r.Context(), userID, limit)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleFansTrend(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.FansTrend(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleFansRanking(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	typ := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, _ := h.svc.FansRanking(r.Context(), userID, typ, limit)
	httputil.WriteOK(w, data)
}

func (h *Handler) handleManuscriptTrend(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.ManuscriptTrend(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}

// handleFansByManuscript 各稿件带来的涨粉数（关注来源归因，历史累积）。
func (h *Handler) handleFansByManuscript(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	data, _ := h.svc.FansByManuscript(r.Context(), userID)
	httputil.WriteOK(w, data)
}

// handleManuscriptInsights 各稿件洞察：粉丝观看率/游客播放占比/平均播放进度 + 粉丝播放趋势。
// 口径说明：watch_history 按（用户，稿件）去重，fan=观看者关注了作者，游客=总播放-登录观众。
func (h *Handler) handleManuscriptInsights(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.ManuscriptInsights(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}

// handleIncome 收益数据：总收益/自定义充电（分）。目前平台无充值能力，真实为 0。
func (h *Handler) handleIncome(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days != 7 && days != 30 {
		days = 30
	}
	data, _ := h.svc.Income(r.Context(), userID, days)
	httputil.WriteOK(w, data)
}
