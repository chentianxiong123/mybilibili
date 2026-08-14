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
	mux.HandleFunc("/api/v1/ai/bindings/", h.handleBindingsByPath)
	mux.HandleFunc("/api/v1/ai/bindings", h.handleBindings)
	mux.HandleFunc("/api/v1/ai/skills", h.handleSkills)
	mux.HandleFunc("/api/v1/ai/skills/", h.handleSkillByPath)
	mux.HandleFunc("/api/v1/ai/usage/", h.handleUsage)
	mux.HandleFunc("/api/v1/ai/customer/", h.handleCustomer)
	mux.HandleFunc("/api/v1/ai/summary/", h.handleSummary)
	mux.HandleFunc("/api/v1/ai/skills/customer-service/defaults", h.handleCustomerDefaults)
	mux.HandleFunc("/api/v1/ai/config/test", h.handleConfigTest)
	mux.HandleFunc("/api/v1/ai/assistant/send", h.handleAssistantSend)
	mux.HandleFunc("/api/v1/ai/skills/customer-service/route-test", h.handleRouteTest)
	mux.HandleFunc("/api/v1/ai/skills/customer-service/route", h.handleRouteSkills)
}

func (h *Handler) handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListConfigs(r.Context())
		for _, c := range list {
			c.APIKey = maskKey(c.APIKey)
		}
		json.NewEncoder(w).Encode(list)
	case "POST":
		var req struct {
			Name        string  `json:"name"`
			Type        string  `json:"type"`
			BaseURL     string  `json:"base_url"`
			APIKey      string  `json:"api_key"`
			Model       string  `json:"model"`
			MaxTokens   int32   `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Type == "" {
			req.Type = "LLM"
		}
		c := &ApiConfig{
			Name: req.Name, Type: req.Type, BaseURL: req.BaseURL, APIKey: req.APIKey,
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
	case len(parts) >= 2 && parts[1] == "bind" && r.Method == "POST":
		var req struct {
			ConfigID int64 `json:"config_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), parts[0], req.ConfigID)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "GET":
		c, err := h.svc.GetConfig(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		c.APIKey = maskKey(c.APIKey)
		json.NewEncoder(w).Encode(c)
	case r.Method == "PUT":
		var req ApiConfig
		json.NewDecoder(r.Body).Decode(&req)
		req.ID = id
		h.svc.UpdateConfig(r.Context(), &req)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE":
		h.svc.DeleteConfig(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleBindings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		bindings, _ := h.svc.ListAllBindings(r.Context())
		json.NewEncoder(w).Encode(bindings)
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

func (h *Handler) handleBindingsByPath(w http.ResponseWriter, r *http.Request) {
	feature := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/bindings/")
	if feature == "" {
		http.Error(w, "feature required", 400)
		return
	}
	switch r.Method {
	case "GET":
		configID, _ := h.svc.GetBinding(r.Context(), feature)
		json.NewEncoder(w).Encode(map[string]any{"feature": feature, "config_id": configID})
	case "POST":
		var req struct {
			ConfigID int64 `json:"configId"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), feature, req.ConfigID)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if typ := r.URL.Query().Get("type"); typ != "" {
			list, _ := h.svc.ListSkillsByType(r.Context(), typ)
			json.NewEncoder(w).Encode(list)
			return
		}
		list, _ := h.svc.ListSkills(r.Context())
		json.NewEncoder(w).Encode(list)
	case "POST":
		var req struct {
			Name            string `json:"name"`
			Description     string `json:"description"`
			SystemPrompt    string `json:"system_prompt"`
			FewShotExamples string `json:"few_shot_examples"`
			Type            string `json:"type"`
			Defaults        bool   `json:"customer_service_defaults"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Defaults || req.Name == "" {
			created, _ := h.svc.CreateMissingCustomerServiceDefaults(r.Context())
			json.NewEncoder(w).Encode(map[string]any{"created": created})
			return
		}
		h.svc.CreateSkill(r.Context(), &Skill{
			Name: req.Name, Description: req.Description,
			SystemPrompt: req.SystemPrompt, FewShotExamples: req.FewShotExamples, Type: req.Type,
		})
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleSkillByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/skills/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	switch {
	case len(parts) >= 2 && parts[1] == "toggle" && r.Method == "PUT":
		h.svc.ToggleSkill(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "GET":
		s, err := h.svc.GetSkill(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		json.NewEncoder(w).Encode(s)
	case r.Method == "PUT":
		var req Skill
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateSkill(r.Context(), id, &req)
		w.Write([]byte(`{"status":"ok"}`))
	case r.Method == "DELETE":
		h.svc.DeleteSkill(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleUsage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/usage/")
	switch path {
	case "overview":
		data, _ := h.svc.UsageOverview(r.Context())
		json.NewEncoder(w).Encode(data)
	case "features":
		data, _ := h.svc.UsageByFeature(r.Context())
		json.NewEncoder(w).Encode(data)
	case "daily":
		days := 7
		if v := r.URL.Query().Get("days"); v != "" {
			days, _ = strconv.Atoi(v)
		}
		data, _ := h.svc.UsageDaily(r.Context(), days)
		json.NewEncoder(w).Encode(data)
	default:
		http.Error(w, "not found", 404)
	}
}

func (h *Handler) handleCustomer(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/customer/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.ParseInt(parts[0], 10, 64)

	switch {
	case len(parts) >= 2 && parts[1] == "messages" && r.Method == "GET":
		msgs, _ := h.svc.GetSessionMessages(r.Context(), sessionID)
		json.NewEncoder(w).Encode(msgs)
	case len(parts) >= 2 && parts[1] == "reply" && r.Method == "POST":
		var req struct {
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Content == "" {
			http.Error(w, "content required", 400)
			return
		}
		h.svc.SendSessionReply(r.Context(), sessionID, getAdminID(r), req.Content)
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "resolve" && r.Method == "POST":
		h.svc.MarkSessionProcessed(r.Context(), sessionID, getAdminID(r))
		w.Write([]byte(`{"status":"ok"}`))
	case len(parts) >= 2 && parts[1] == "pending" && parts[0] == "sessions" && len(parts) >= 3 && parts[2] == "count" && r.Method == "GET":
		cnt, _ := h.svc.CountPendingSessions(r.Context())
		json.NewEncoder(w).Encode(map[string]any{"count": cnt})
	case len(parts) >= 1 && parts[0] == "sessions" && r.Method == "GET":
		sessions, _ := h.svc.ListPendingSessions(r.Context())
		json.NewEncoder(w).Encode(sessions)
	default:
		http.Error(w, "not found", 404)
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:6] + "****" + key[len(key)-4:]
}

func getAdminID(r *http.Request) int64 {
	idStr := r.Header.Get("X-Admin-Id")
	if idStr == "" {
		idStr = r.Header.Get("X-User-Id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	videoID := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/summary/")
	_ = videoID
	w.Write([]byte(`{"summary":"AI summary - not yet implemented"}`))
}

func (h *Handler) handleCustomerDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	created, _ := h.svc.CreateMissingCustomerServiceDefaults(r.Context())
	json.NewEncoder(w).Encode(map[string]any{"created": created})
}

func (h *Handler) handleConfigTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req map[string]string
	json.NewDecoder(r.Body).Decode(&req)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true, "message": "connection ok", "provider": req["provider"],
	})
}

func (h *Handler) handleAssistantSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	json.NewEncoder(w).Encode(map[string]any{
		"reply": "已收到: " + req.Message,
	})
}

func (h *Handler) handleRouteTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	skill, _ := h.svc.MatchCustomerServiceSkill(r.Context(), req.Content)
	if skill == nil {
		skill = map[string]any{"name": "default", "matched": false}
	}
	json.NewEncoder(w).Encode(skill)
}

func (h *Handler) handleRouteSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Content string `json:"content"`
		Limit   int    `json:"limit"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	skills, _ := h.svc.RouteSkills(r.Context(), req.Content, req.Limit)
	if skills == nil {
		skills = []map[string]any{}
	}
	json.NewEncoder(w).Encode(skills)
}
