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
	} else if len(parts) >= 3 && parts[0] == "user" && parts[2] == "ids" && r.Method == "GET" {
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		ids, _ := h.svc.ListUserManuscriptIDs(r.Context(), uid)
		json.NewEncoder(w).Encode(ids)
	} else if len(parts) >= 3 && parts[0] == "user" && parts[2] == "video-ids" && r.Method == "GET" {
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		ids, _ := h.svc.ListUserVideoIDs(r.Context(), uid)
		json.NewEncoder(w).Encode(ids)
	} else {
		http.Error(w, "not found", 404)
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
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}

	switch parts[0] {
	case "home":
		h.handleBannerHome(w, r, parts)
	case "category":
		h.handleBannerCategory(w, r, parts)
	case "background":
		h.handleBannerSingle(w, r, 3)
	case "user-profile":
		h.handleBannerSingle(w, r, 4)
	case "upload":
		if r.Method == "POST" {
			var req struct {
				Title    string `json:"title"`
				ImageURL string `json:"image_url"`
				LinkURL  string `json:"link_url"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(map[string]string{"url": req.ImageURL})
			return
		}
		http.Error(w, "method not allowed", 405)
	default:
		http.Error(w, "not found", 404)
	}
}

func (h *Handler) handleBannerHome(w http.ResponseWriter, r *http.Request, parts []string) {
	switch {
	case r.Method == "GET":
		list, _ := h.svc.ListBanners(r.Context(), 1)
		json.NewEncoder(w).Encode(list)
	case r.Method == "POST":
		b := decodeBanner(r)
		b.Type = 1
		h.svc.CreateBanner(r.Context(), b)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "PUT" && len(parts) >= 2:
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		b := decodeBanner(r)
		b.Type = 1
		h.svc.UpdateBanner(r.Context(), id, b)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE" && len(parts) >= 2:
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.svc.DeleteBanner(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleBannerCategory(w http.ResponseWriter, r *http.Request, parts []string) {
	categoryID, _ := strconv.ParseInt(parts[1], 10, 64)
	switch {
	case r.Method == "GET":
		list, _ := h.svc.ListBannersByCategory(r.Context(), 2, categoryID)
		json.NewEncoder(w).Encode(list)
	case r.Method == "POST":
		b := decodeBanner(r)
		b.Type = 2
		b.CategoryID = categoryID
		h.svc.CreateBanner(r.Context(), b)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "PUT" && len(parts) >= 3:
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		b := decodeBanner(r)
		b.Type = 2
		b.CategoryID = categoryID
		h.svc.UpdateBanner(r.Context(), id, b)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE" && len(parts) >= 3:
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		h.svc.DeleteBanner(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleBannerSingle(w http.ResponseWriter, r *http.Request, bannerType int32) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListBanners(r.Context(), bannerType)
		var b *BannerImage
		if len(list) > 0 {
			b = list[0]
		}
		json.NewEncoder(w).Encode(b)
	case "POST":
		b := decodeBanner(r)
		b.Type = bannerType
		h.svc.CreateBanner(r.Context(), b)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		list, _ := h.svc.ListBanners(r.Context(), bannerType)
		for _, b := range list {
			h.svc.DeleteBanner(r.Context(), b.ID)
		}
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func decodeBanner(r *http.Request) *BannerImage {
	var req struct {
		Title     string `json:"title"`
		ImageURL  string `json:"image_url"`
		LinkURL   string `json:"link_url"`
		SortOrder int32  `json:"sort_order"`
		Status    int32  `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == 0 {
		req.Status = 1
	}
	return &BannerImage{Title: req.Title, ImageURL: req.ImageURL, LinkURL: req.LinkURL, SortOrder: req.SortOrder, Status: req.Status}
}

func (h *Handler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.svc.Statistics(r.Context())
	json.NewEncoder(w).Encode(stats)
}
