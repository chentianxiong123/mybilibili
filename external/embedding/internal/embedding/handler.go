package embedding

import (
	"encoding/json"
	"net/http"

	"mybilibili/pkg/httputil"
)

// RegisterRoutes 注册 embedding HTTP 路由。
func RegisterRoutes(mux *http.ServeMux, svc *Service) {
	mux.HandleFunc("/v1/embeddings", handleEmbedding(svc))
	mux.HandleFunc("/api/v1/embeddings", handleEmbedding(svc))
}

func handleEmbedding(svc *Service) http.HandlerFunc {
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
