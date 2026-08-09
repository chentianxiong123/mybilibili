package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
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
}

func (h *Handler) handleOverview(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	data, _ := h.svc.Overview(r.Context(), userID)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleTrend(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.Trend(r.Context(), userID, days)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleRanking(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sortBy")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, _ := h.svc.Ranking(r.Context(), sortBy, limit)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleLatestComments(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	data, _ := h.svc.LatestComments(r.Context(), userID, limit)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleFansTrend(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	data, _ := h.svc.FansTrend(r.Context(), userID, days)
	json.NewEncoder(w).Encode(data)
}

func getUserID(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}
