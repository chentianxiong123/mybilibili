package video

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/video/", h.handleVideo)
	mux.HandleFunc("/api/v1/category", h.handleCategory)
	mux.HandleFunc("/api/v1/category/", h.handleCategoryByID)
	mux.HandleFunc("/api/v1/banner/", h.handleBanner)
	mux.HandleFunc("/api/v1/statistics", h.handleStatistics)
}

func (h *Handler) handleVideo(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/video/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	if len(parts) == 1 && r.Method == "GET" {
		v, err := h.svc.GetVideo(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(v)
	} else if len(parts) >= 2 && parts[1] == "manuscript" {
		msID, _ := strconv.ParseInt(parts[0], 10, 64)
		list, _ := h.svc.ListByManuscript(r.Context(), msID)
		json.NewEncoder(w).Encode(list)
	}
}

func (h *Handler) handleCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		list, _ := h.svc.ListCategories(r.Context())
		json.NewEncoder(w).Encode(list)
	} else if r.Method == "POST" {
		var req struct {
			Name      string `json:"name"`
			Icon      string `json:"icon"`
			SortOrder int32  `json:"sort_order"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name == "" {
			http.Error(w, "name required", 400)
			return
		}
		h.svc.CreateCategory(r.Context(), req.Name, req.Icon, req.SortOrder)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/category/"), 10, 64)
	switch r.Method {
	case "GET":
		c, err := h.svc.repo.GetCategoryByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(c)
	case "PUT":
		var req struct {
			Name      string `json:"name"`
			Icon      string `json:"icon"`
			SortOrder int32  `json:"sort_order"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateCategory(r.Context(), id, req.Name, req.Icon, req.SortOrder)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteCategory(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleBanner(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/banner/")
	parts := strings.Split(path, "/")
	switch {
	case len(parts) >= 1 && parts[0] == "home" && r.Method == "GET":
		list, _ := h.svc.ListBanners(r.Context(), 1)
		json.NewEncoder(w).Encode(list)
	case len(parts) >= 2 && parts[0] == "category" && r.Method == "GET":
		_, _ = strconv.ParseInt(parts[1], 10, 64)
		list, _ := h.svc.ListBanners(r.Context(), 2)
		json.NewEncoder(w).Encode(list)
	case len(parts) >= 1 && parts[0] == "upload" && r.Method == "POST":
		var req struct {
			Title    string `json:"title"`
			ImageURL string `json:"image_url"`
			LinkURL  string `json:"link_url"`
			Type     int32  `json:"type"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		b := &BannerImage{Title: req.Title, ImageURL: req.ImageURL, LinkURL: req.LinkURL, Type: req.Type, SortOrder: 0, Status: 1}
		h.svc.CreateBanner(r.Context(), b)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.svc.DeleteBanner(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.svc.Statistics(r.Context())
	json.NewEncoder(w).Encode(stats)
}
