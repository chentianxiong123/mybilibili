package search

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
	mux.HandleFunc("/api/v1/search/videos", h.handleSearch)
	mux.HandleFunc("/api/v1/search/hot", h.handleHot)
	mux.HandleFunc("/api/v1/recommend/related/", h.handleRelated)
	mux.HandleFunc("/api/v1/recommend/for-you", h.handleForYou)
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	page, size := parsePage(r)
	list, _ := h.svc.Search(r.Context(), keyword, categoryID, page, size)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleHot(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.Hot(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleRelated(w http.ResponseWriter, r *http.Request) {
	videoID, _ := strconv.ParseInt(r.URL.Path[len("/api/v1/recommend/related/"):], 10, 64)
	size, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 32)
	if size == 0 {
		size = 10
	}
	list, _ := h.svc.Related(r.Context(), videoID, int32(size))
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleForYou(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func parsePage(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}
	return int32(page), int32(size)
}
