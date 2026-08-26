package proxy

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"mybilibili/bili-proxy-service/internal/bilibili"
)

const cdnCacheTTL = 5 * time.Minute

type cdnCacheEntry struct {
	url  string
	kind string
	ts   time.Time
}

// Handler 提供 B 站视频流的 Range 代理。
type Handler struct {
	db     *sql.DB
	client *bilibili.Client
	mu     sync.Mutex
	cache  map[string]cdnCacheEntry
}

func NewHandler(db *sql.DB, client *bilibili.Client) *Handler {
	return &Handler{db: db, client: client, cache: make(map[string]cdnCacheEntry)}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/stream/", h.handleStream)
	mux.HandleFunc("/resolve/", h.handleResolve)
}

// videoRow 记录 B 站外部视频的解析所需字段。
type videoRow struct {
	ID  int64
	CID int64
}

func (h *Handler) loadVideo(r *http.Request, videoID int64) (*videoRow, error) {
	row := &videoRow{}
	err := h.db.QueryRowContext(r.Context(),
		`SELECT v.id, COALESCE(v.cid,0) FROM videos v
		   JOIN manuscripts m ON v.manuscript_id = m.id
		  WHERE v.id = $1 AND m.source_type = 'bilibili'`, videoID,
	).Scan(&row.ID, &row.CID)
	if err != nil {
		return nil, err
	}
	if row.CID == 0 {
		return nil, fmt.Errorf("video %d has no cid", videoID)
	}
	return row, nil
}

func (h *Handler) handleResolve(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Path[len("/resolve/"):], 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid video id")
		return
	}
	row, err := h.loadVideo(r, id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "video not found")
		return
	}
	qn, _ := strconv.Atoi(r.URL.Query().Get("qn"))
	if qn == 0 {
		qn = 64
	}
	// 需要从中取 bvid
	var bvid string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT bvid FROM manuscripts WHERE id = (SELECT manuscript_id FROM videos WHERE id=$1)`, id,
	).Scan(&bvid); err != nil {
		writeErr(w, http.StatusNotFound, "manuscript not found")
		return
	}
	u, kind, err := h.getCDN(r, bvid, row.CID, qn)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"url": u, "kind": kind})
}

func (h *Handler) handleStream(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Path[len("/stream/"):], 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid video id")
		return
	}
	row, err := h.loadVideo(r, id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "video not found")
		return
	}
	var bvid string
	if err := h.db.QueryRowContext(r.Context(),
		`SELECT m.bvid FROM manuscripts m JOIN videos v ON v.manuscript_id = m.id WHERE v.id = $1`, id,
	).Scan(&bvid); err != nil {
		writeErr(w, http.StatusNotFound, "manuscript not found")
		return
	}
	qn, _ := strconv.Atoi(r.URL.Query().Get("qn"))
	if qn == 0 {
		qn = 64
	}
	u, kind, err := h.getCDN(r, bvid, row.CID, qn)
	if err != nil {
		log.Printf("[bili] resolve stream video=%d err=%v", id, err)
		writeErr(w, http.StatusBadGateway, "resolve stream failed")
		return
	}
	h.proxyRange(w, r, u, kind)
}

// getCDN 从缓存或实时解析获取 CDN 直链。
func (h *Handler) getCDN(r *http.Request, bvid string, cid int64, qn int) (string, string, error) {
	key := fmt.Sprintf("%s/%d/%d", bvid, cid, qn)
	cacheKey := key
	h.mu.Lock()
	if e, ok := h.cache[cacheKey]; ok && time.Since(e.ts) < cdnCacheTTL {
		h.mu.Unlock()
		return e.url, e.kind, nil
	}
	h.mu.Unlock()

	u, kind, err := h.client.ResolveStream(bvid, cid, qn)
	if err != nil {
		return "", "", err
	}
	cacheKey = key
	h.mu.Lock()
	h.cache[cacheKey] = cdnCacheEntry{url: u, kind: kind, ts: time.Now()}
	h.mu.Unlock()
	return u, kind, nil
}

// proxyRange 将浏览器 Range 请求透传给 CDN，带 Referer 防盗链，返回对应状态码。
func (h *Handler) proxyRange(w http.ResponseWriter, r *http.Request, target, kind string) {
	req, err := http.NewRequestWithContext(r.Context(), "GET", target, nil)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "build request failed")
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if rng := r.Header.Get("Range"); rng != "" {
		req.Header.Set("Range", rng)
	}

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "upstream request failed")
		return
	}
	defer resp.Body.Close()

	// 透传响应头
	for _, hk := range []string{"Content-Range", "Accept-Ranges", "Content-Length", "Content-Type", "ETag", "Last-Modified"} {
		if v := resp.Header.Get(hk); v != "" {
			w.Header().Set(hk, v)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", kind)
	}
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"code": code, "message": msg})
}