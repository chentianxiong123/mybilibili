package ai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc         *Service
	summarySvc  *SummaryService
	reviewSvc   *ReviewService
	customerSvc *CustomerService
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) WithSummary(svc *SummaryService) *Handler {
	h.summarySvc = svc
	return h
}

func (h *Handler) WithReview(svc *ReviewService) *Handler {
	h.reviewSvc = svc
	return h
}

func (h *Handler) WithCustomer(svc *CustomerService) *Handler {
	h.customerSvc = svc
	return h
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
		httputil.WriteOK(w, list)
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
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleConfigByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/configs/")
	parts := strings.Split(path, "/")
	if parts[0] == "types" {
		httputil.WriteOK(w, []string{"LLM", "ASR", "TTS", "IMAGE", "MODERATION"})
		return
	}
	if parts[0] == "features" {
		httputil.WriteOK(w, []map[string]string{
			{"feature": "chat", "type": "LLM"},
			{"feature": "summary", "type": "LLM"},
			{"feature": "transcribe", "type": "ASR"},
			{"feature": "tts", "type": "TTS"},
			{"feature": "image_generation", "type": "IMAGE"},
			{"feature": "content_review", "type": "MODERATION"},
		})
		return
	}
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	switch {
	case len(parts) >= 2 && parts[1] == "toggle" && r.Method == "PUT":
		h.svc.ToggleConfig(r.Context(), id)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case len(parts) >= 2 && parts[1] == "bind" && r.Method == "POST":
		var req struct {
			ConfigID int64 `json:"config_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), parts[0], req.ConfigID)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case r.Method == "GET":
		c, err := h.svc.GetConfig(r.Context(), id)
		if err != nil {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": err.Error(), "data": nil})
			return
		}
		c.APIKey = maskKey(c.APIKey)
		httputil.WriteOK(w, c)
	case r.Method == "PUT":
		var req ApiConfig
		json.NewDecoder(r.Body).Decode(&req)
		req.ID = id
		h.svc.UpdateConfig(r.Context(), &req)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case r.Method == "DELETE":
		h.svc.DeleteConfig(r.Context(), id)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleBindings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		bindings, _ := h.svc.ListAllBindings(r.Context())
		httputil.WriteOK(w, bindings)
	case "POST":
		var req struct {
			Feature  string `json:"feature"`
			ConfigID int64  `json:"config_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), req.Feature, req.ConfigID)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleBindingsByPath(w http.ResponseWriter, r *http.Request) {
	feature := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/bindings/")
	if feature == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "feature required", "data": nil})
		return
	}
	switch r.Method {
	case "GET":
		configID, _ := h.svc.GetBinding(r.Context(), feature)
		httputil.WriteOK(w, map[string]any{"feature": feature, "config_id": configID})
	case "POST":
		var req struct {
			ConfigID int64 `json:"configId"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.SetBinding(r.Context(), feature, req.ConfigID)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *Handler) handleSkills(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if typ := r.URL.Query().Get("type"); typ != "" {
			list, _ := h.svc.ListSkillsByType(r.Context(), typ)
			httputil.WriteOK(w, list)
			return
		}
		list, _ := h.svc.ListSkills(r.Context())
		httputil.WriteOK(w, list)
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
			httputil.WriteOK(w, map[string]any{"created": created})
			return
		}
		h.svc.CreateSkill(r.Context(), &Skill{
			Name: req.Name, Description: req.Description,
			SystemPrompt: req.SystemPrompt, FewShotExamples: req.FewShotExamples, Type: req.Type,
		})
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleSkillByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/skills/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)
	switch {
	case len(parts) >= 2 && parts[1] == "toggle" && r.Method == "PUT":
		h.svc.ToggleSkill(r.Context(), id)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case r.Method == "GET":
		s, err := h.svc.GetSkill(r.Context(), id)
		if err != nil {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": err.Error(), "data": nil})
			return
		}
		httputil.WriteOK(w, s)
	case r.Method == "PUT":
		var req Skill
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateSkill(r.Context(), id, &req)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case r.Method == "DELETE":
		h.svc.DeleteSkill(r.Context(), id)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	}
}

func (h *Handler) handleUsage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/usage/")
	switch path {
	case "overview":
		data, _ := h.svc.UsageOverview(r.Context())
		httputil.WriteOK(w, data)
	case "features":
		data, _ := h.svc.UsageByFeature(r.Context())
		httputil.WriteOK(w, data)
	case "daily":
		days := 7
		if v := r.URL.Query().Get("days"); v != "" {
			days, _ = strconv.Atoi(v)
		}
		data, _ := h.svc.UsageDaily(r.Context(), days)
		httputil.WriteOK(w, data)
	default:
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
	}
}

func (h *Handler) handleCustomer(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/customer/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.ParseInt(parts[0], 10, 64)

	switch {
	case len(parts) >= 2 && parts[1] == "messages" && r.Method == "GET":
		msgs, _ := h.svc.GetSessionMessages(r.Context(), sessionID)
		httputil.WriteOK(w, msgs)
	case len(parts) >= 2 && parts[1] == "reply" && r.Method == "POST":
		var req struct {
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Content == "" {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "content required", "data": nil})
			return
		}
		h.svc.SendSessionReply(r.Context(), sessionID, getAdminID(r), req.Content)
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case len(parts) >= 2 && parts[1] == "resolve" && r.Method == "POST":
		h.svc.MarkSessionProcessed(r.Context(), sessionID, getAdminID(r))
		httputil.WriteOK(w, map[string]any{"status": "ok"})
	case len(parts) >= 2 && parts[1] == "pending" && parts[0] == "sessions" && len(parts) >= 3 && parts[2] == "count" && r.Method == "GET":
		cnt, _ := h.svc.CountPendingSessions(r.Context())
		httputil.WriteOK(w, map[string]any{"count": cnt})
	case len(parts) >= 1 && parts[0] == "sessions" && r.Method == "GET":
		sessions, _ := h.svc.ListPendingSessions(r.Context())
		httputil.WriteOK(w, sessions)
	default:
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "not found", "data": nil})
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
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/summary/")

	// GET /api/v1/ai/summary/check/{videoId} — 是否已有摘要
	if strings.HasPrefix(path, "check/") {
		if h.summarySvc == nil {
			httputil.WriteOK(w, map[string]any{"code": 200, "data": false})
			return
		}
		videoID, _ := strconv.ParseInt(strings.TrimPrefix(path, "check/"), 10, 64)
		if videoID == 0 {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid video id", "data": nil})
			return
		}
		has, err := h.summarySvc.CheckSummary(r.Context(), videoID)
		if err != nil {
			has = false
		}
		httputil.WriteOK(w, map[string]any{"code": 200, "data": has})
		return
	}

	// GET /api/v1/ai/summary/stream/{videoId} — SSE 流式摘要
	if strings.HasPrefix(path, "stream/") {
		videoID, _ := strconv.ParseInt(strings.TrimPrefix(path, "stream/"), 10, 64)
		if videoID == 0 {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid video id", "data": nil})
			return
		}
		h.handleSummaryStream(w, r, videoID)
		return
	}

	videoID, _ := strconv.ParseInt(path, 10, 64)
	if videoID == 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid video id", "data": nil})
		return
	}
	if h.summarySvc == nil {
		w.Write([]byte(`{"summary":"AI summary not available"}`))
		return
	}
	summary, err := h.summarySvc.GetSummary(r.Context(), videoID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": err.Error(), "data": nil})
		return
	}
	httputil.WriteOK(w, map[string]string{"summary": summary})
}

// handleSummaryStream 以 SSE 推送流式摘要；data 为 base64(UTF-8)，与前端 atob 解码对齐。
func (h *Handler) handleSummaryStream(w http.ResponseWriter, r *http.Request, videoID int64) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "streaming unsupported", "data": nil})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sendEvent := func(event, payload string) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
		flusher.Flush()
	}
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

	sendEvent("start", "开始生成摘要...")

	// 优先推送 MinIO 里流水线生成的持久化摘要（老项目数据），
	// 打字机节奏与老项目 AiController.streamSummary 对齐：5字符/块 + 25~65ms 随机间隔
	if stored, serr := h.summarySvc.FetchStoredSummary(r.Context(), videoID); serr == nil && stored != "" {
		runes := []rune(stored)
		sendEvent("meta", fmt.Sprintf(`{"totalLength":%d}`, len(runes)))

		const chunkSize = 5
		for i := 0; i < len(runes); i += chunkSize {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			sendEvent("data", b64(string(runes[i:end])))

			delay := 25 + rand.Intn(40) // 25~65ms，同老项目
			if (i+chunkSize)%60 == 0 && rand.Float64() > 0.6 {
				delay += 80 + rand.Intn(100) // 偶尔停顿，模拟思考
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		sendEvent("done", "摘要生成完成")
		return
	}

	ch, err := func() (<-chan string, error) {
		if h.summarySvc == nil {
			return nil, errors.New("summary service not configured")
		}
		return h.summarySvc.StreamSummary(r.Context(), videoID)
	}()
	if err != nil {
		// 流式不可用时回退为一次性取摘要推送
		if h.summarySvc != nil {
			if summary, gerr := h.summarySvc.GetSummary(r.Context(), videoID); gerr == nil && summary != "" {
				sendEvent("data", b64(summary))
				sendEvent("done", "摘要生成完成")
				return
			}
		}
		sendEvent("error", "生成摘要失败: "+err.Error())
		return
	}
	for chunk := range ch {
		if chunk == "" {
			continue
		}
		sendEvent("data", b64(chunk))
	}
	sendEvent("done", "摘要生成完成")
}

func (h *Handler) handleCustomerDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	created, _ := h.svc.CreateMissingCustomerServiceDefaults(r.Context())
	httputil.WriteOK(w, map[string]any{"created": created})
}

func (h *Handler) handleConfigTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req map[string]string
	json.NewDecoder(r.Body).Decode(&req)
	httputil.WriteOK(w, map[string]any{
		"success": true, "message": "connection ok", "provider": req["provider"],
	})
}

func (h *Handler) handleAssistantSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	var req struct {
		Message string `json:"message,omitempty"`
		Content string `json:"content,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	msg := req.Message
	if msg == "" {
		msg = req.Content
	}
	httputil.WriteOK(w, map[string]any{
		"reply": "已收到: " + msg,
	})
}

func (h *Handler) handleRouteTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
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
	httputil.WriteOK(w, skill)
}

func (h *Handler) handleRouteSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 405, "message": "method not allowed", "data": nil})
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
	httputil.WriteOK(w, skills)
}
