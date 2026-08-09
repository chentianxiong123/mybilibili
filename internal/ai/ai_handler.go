package ai

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
	mux.HandleFunc("/api/v1/ai/configs", h.handleConfigs)
	mux.HandleFunc("/api/v1/ai/configs/", h.handleConfigByID)
	mux.HandleFunc("/api/v1/ai/bindings", h.handleBindings)
	mux.HandleFunc("/api/v1/ai/summary/", h.handleSummary)
}

func (h *Handler) handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListConfigs(r.Context())
		json.NewEncoder(w).Encode(list)
	case "POST":
		var req struct {
			Name        string  `json:"name"`
			BaseURL     string  `json:"base_url"`
			APIKey      string  `json:"api_key"`
			Model       string  `json:"model"`
			MaxTokens   int32   `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		c := &ApiConfig{
			Name: req.Name, BaseURL: req.BaseURL, APIKey: req.APIKey,
			Model: req.Model, MaxTokens: req.MaxTokens, Temperature: req.Temperature,
		}
		h.svc.CreateConfig(r.Context(), c)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleConfigByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/configs/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	switch {
	case len(parts) >= 2 && parts[1] == "toggle" && r.Method == "PUT":
		h.svc.ToggleConfig(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE":
		h.svc.DeleteConfig(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleBindings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		feature := r.URL.Query().Get("feature")
		configID, _ := h.svc.GetBinding(r.Context(), feature)
		json.NewEncoder(w).Encode(map[string]interface{}{"feature": feature, "config_id": configID})
	case "POST":
		var req struct {
			Feature  string `json:"feature"`
			ConfigID int64  `json:"config_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), req.Feature, req.ConfigID)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	videoID := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/summary/")
	_ = videoID
	w.Write([]byte(`{"summary":"AI summary - not yet implemented"}`))
}
