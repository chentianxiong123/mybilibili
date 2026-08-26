package video

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/imageutil"
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
	mux.HandleFunc("/api/v1/video/", h.handleVideo)
	mux.HandleFunc("/api/v1/category", h.handleCategory)
	mux.HandleFunc("/api/v1/category/", h.handleCategoryByID)
	mux.HandleFunc("/api/v1/banner/", h.handleBanner)
	mux.HandleFunc("/api/v1/banner-images/", h.handleBannerImages)
	mux.HandleFunc("/api/v1/statistics", h.handleStatistics)
	mux.HandleFunc("/api/v1/statistics/", h.handleStatisticsByPath)
}

func (h *Handler) handleVideo(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/video/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, map[string]any{"code": 404, "message": "not found"})
		return
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	if len(parts) == 1 && r.Method == "GET" {
		v, err := h.svc.GetVideo(r.Context(), id)
		if err != nil {
			writeJSON(w, map[string]any{"code": 404, "message": "not found"})
			return
		}
		writeJSON(w, v)
	} else if len(parts) >= 2 && parts[1] == "manuscript" {
		msID, _ := strconv.ParseInt(parts[0], 10, 64)
		list, _ := h.svc.ListByManuscript(r.Context(), msID)
		writeJSON(w, list)
	} else if len(parts) >= 3 && parts[0] == "user" && parts[2] == "ids" && r.Method == "GET" {
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		ids, _ := h.svc.ListUserManuscriptIDs(r.Context(), uid)
		writeJSON(w, ids)
	} else if len(parts) >= 3 && parts[0] == "user" && parts[2] == "video-ids" && r.Method == "GET" {
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		ids, _ := h.svc.ListUserVideoIDs(r.Context(), uid)
		writeJSON(w, ids)
	} else {
		writeJSON(w, map[string]any{"code": 404, "message": "not found"})
	}
}

func (h *Handler) handleCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		list, _ := h.svc.ListCategories(r.Context())
		writeJSON(w, list)
	} else if r.Method == "POST" {
		var req struct {
			Name      string `json:"name"`
			Icon      string `json:"icon"`
			SortOrder int32  `json:"sort_order"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name == "" {
			writeJSON(w, map[string]any{"code": 400, "message": "name required"})
			return
		}
		h.svc.CreateCategory(r.Context(), req.Name)
		writeJSON(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleCategoryByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/category/"), 10, 64)
	switch r.Method {
	case "GET":
		c, err := h.svc.repo.GetCategoryByID(r.Context(), id)
		if err != nil {
			writeJSON(w, map[string]any{"code": 404, "message": "not found"})
			return
		}
		writeJSON(w, c)
	case "PUT":
		var req struct {
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if _, err := h.svc.repo.GetCategoryByID(r.Context(), id); err != nil {
			writeJSON(w, map[string]any{"code": 404, "message": "category not found"})
			return
		}
		h.svc.UpdateCategory(r.Context(), id, req.Name)
		writeJSON(w, map[string]any{"status": "ok"})
	case "DELETE":
		if _, err := h.svc.repo.GetCategoryByID(r.Context(), id); err != nil {
			writeJSON(w, map[string]any{"code": 404, "message": "category not found"})
			return
		}
		h.svc.DeleteCategory(r.Context(), id)
		writeJSON(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleBanner(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/banner/")
	if path == r.URL.Path {
		path = strings.TrimPrefix(r.URL.Path, "/api/v1/banner-images/")
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, map[string]any{"code": 404, "message": "not found"})
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
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				writeJSON(w, map[string]any{"code": 400, "message": "parse form: "+err.Error()})
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				writeJSON(w, map[string]any{"code": 400, "message": "file required"})
				return
			}
			dir := "/tmp/mybilibili-uploads/images"
			os.MkdirAll(dir, 0o755)
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".jpg"
			}
			name := time.Now().Format("20060102150405") + "_" + strconv.FormatInt(time.Now().UnixNano(), 36) + ext
			dst := filepath.Join(dir, name)
			f, err := os.Create(dst)
			if err != nil {
				file.Close()
				writeJSON(w, map[string]any{"code": 500, "message": "create file: "+err.Error()})
				return
			}
			io.Copy(f, file)
			f.Close()
			file.Close()
			imageutil.CompressAndReplace(dst)
			writeJSON(w, map[string]string{"url": "/uploads/images/" + filepath.Base(dst)})
			return
		}
		writeJSON(w, map[string]any{"code": 405, "message": "method not allowed"})
	default:
		writeJSON(w, map[string]any{"code": 404, "message": "not found"})
	}
}

func (h *Handler) handleBannerImages(w http.ResponseWriter, r *http.Request) {
	h.handleBanner(w, r)
}

func (h *Handler) handleBannerHome(w http.ResponseWriter, r *http.Request, parts []string) {
	switch {
	case r.Method == "GET":
		list, _ := h.svc.ListBanners(r.Context(), 1)
		if list == nil {
			list = []*BannerImage{}
		}
		writeJSON(w, list)
	case r.Method == "POST":
		b := decodeBanner(r)
		b.Type = 1
		h.svc.CreateBanner(r.Context(), b)
		writeJSON(w, map[string]any{"status": "ok"})
	case r.Method == "PUT" && len(parts) >= 2:
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		b := decodeBanner(r)
		b.Type = 1
		h.svc.UpdateBanner(r.Context(), id, b)
		writeJSON(w, map[string]any{"status": "ok"})
	case r.Method == "DELETE" && len(parts) >= 2:
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.svc.DeleteBanner(r.Context(), id)
		writeJSON(w, map[string]any{"status": "ok"})
	default:
		writeJSON(w, map[string]any{"code": 405, "message": "method not allowed"})
	}
}

func (h *Handler) handleBannerCategory(w http.ResponseWriter, r *http.Request, parts []string) {
	categoryID, _ := strconv.ParseInt(parts[1], 10, 64)
	switch {
	case r.Method == "GET":
		list, _ := h.svc.ListBannersByCategory(r.Context(), 2, categoryID)
		if list == nil {
			list = []*BannerImage{}
		}
		writeJSON(w, list)
	case r.Method == "POST":
		b := decodeBanner(r)
		b.Type = 2
		b.CategoryID = categoryID
		h.svc.CreateBanner(r.Context(), b)
		writeJSON(w, map[string]any{"status": "ok"})
	case r.Method == "PUT" && len(parts) >= 3:
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		b := decodeBanner(r)
		b.Type = 2
		b.CategoryID = categoryID
		h.svc.UpdateBanner(r.Context(), id, b)
		writeJSON(w, map[string]any{"status": "ok"})
	case r.Method == "DELETE" && len(parts) >= 3:
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		h.svc.DeleteBanner(r.Context(), id)
		writeJSON(w, map[string]any{"status": "ok"})
	default:
		writeJSON(w, map[string]any{"code": 405, "message": "method not allowed"})
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
		writeJSON(w, b)
	case "POST":
		b := decodeBanner(r)
		b.Type = bannerType
		h.svc.CreateBanner(r.Context(), b)
		writeJSON(w, map[string]any{"status": "ok"})
	case "DELETE":
		list, _ := h.svc.ListBanners(r.Context(), bannerType)
		for _, b := range list {
			h.svc.DeleteBanner(r.Context(), b.ID)
		}
		writeJSON(w, map[string]any{"status": "ok"})
	default:
		writeJSON(w, map[string]any{"code": 405, "message": "method not allowed"})
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
	writeJSON(w, stats)
}

func (h *Handler) handleStatisticsByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/statistics/")
	switch path {
	case "overview":
		stats, _ := h.svc.Statistics(r.Context())
		writeJSON(w, stats)
	case "manuscript/status":
		data := map[string]interface{}{}
		h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FILTER (WHERE review_status=0), COUNT(*) FILTER (WHERE review_status=1),
			        COUNT(*) FILTER (WHERE status=3), COUNT(*) FILTER (WHERE status=4)
			 FROM manuscripts`).Scan(
			func() *int64 { var v int64; data["pending_review"] = &v; return &v }(),
			func() *int64 { var v int64; data["approved"] = &v; return &v }(),
			func() *int64 { var v int64; data["published"] = &v; return &v }(),
			func() *int64 { var v int64; data["rejected"] = &v; return &v }())
		writeJSON(w, data)
	case "manuscript/recent":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 50 {
			limit = 10
		}
		rows, err := h.svc.repo.db.QueryContext(r.Context(),
			`SELECT id, user_id, title, cover_url, view_count, upload_time
			 FROM manuscripts ORDER BY COALESCE(upload_time, to_timestamp(0)) DESC LIMIT $1`, limit)
		if err != nil {
			writeJSON(w, []interface{}{})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var id, uid, views int64
			var title string
			var cover sql.NullString
			var uploaded sql.NullTime
			if err := rows.Scan(&id, &uid, &title, &cover, &views, &uploaded); err != nil {
				writeJSON(w, []interface{}{})
				return
			}
			item := map[string]interface{}{
				"id": id, "user_id": uid, "title": title,
				"cover_url": cover.String, "view_count": views,
			}
			if uploaded.Valid {
				item["created_at"] = uploaded.Time.Format(time.RFC3339)
			}
			list = append(list, item)
		}
		writeJSON(w, list)
	case "video/play", "user/growth", "comment", "video/hot":
		// 前端已调用但后端暂未实现，返回空结构避免 404 报错
		writeJSON(w, map[string]interface{}{})
	default:
		writeJSON(w, map[string]interface{}{})
	}
}
