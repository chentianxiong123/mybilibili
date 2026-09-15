package transcoder

import (
	"encoding/json"
	"net/http"

	"mybilibili/pkg/httputil"
)

// RegisterRoutes 注册 transcoder HTTP 路由。
// 采用 MinIO 对象引用：请求体传 bucket+source_key，转码产物写回 MinIO 并返回播放地址。
func RegisterRoutes(mux *http.ServeMux, svc *Service) {
	mux.HandleFunc("/api/v1/transcode", handleTranscode(svc))
	mux.HandleFunc("/api/v1/transcode/transcode", handleTranscode(svc))
}

func handleTranscode(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", 405)
			return
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json: "+err.Error(), 400)
			return
		}
		res, err := svc.Process(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		httputil.WriteOK(w, res)
	}
}