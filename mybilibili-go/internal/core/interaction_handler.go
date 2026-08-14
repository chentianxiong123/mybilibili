package core

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type GenericInteractionHandler struct {
	repo *InteractionRepository
}

func NewGenericInteractionHandler(repo *InteractionRepository) *GenericInteractionHandler {
	return &GenericInteractionHandler{repo: repo}
}

func (h *GenericInteractionHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/interaction/", h.handleRouter)
}

func (h *GenericInteractionHandler) handleRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/interaction/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, "not found", 404)
		return
	}
	switch parts[0] {
	case "like":
		h.handleLike(w, r)
	case "status":
		h.handleStatus(w, r)
	case "batch":
		if len(parts) >= 2 && parts[1] == "status" {
			h.handleBatchStatus(w, r)
		} else if len(parts) >= 2 && parts[1] == "count" {
			h.handleBatchCount(w, r)
		} else {
			http.Error(w, "not found", 404)
		}
	case "count":
		h.handleCount(w, r)
	default:
		http.Error(w, "not found", 404)
	}
}

func (h *GenericInteractionHandler) handleLike(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		TargetType string `json:"targetType"`
		TargetID   int64  `json:"targetId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid request"})
		return
	}
	if req.TargetType == "" || req.TargetID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "targetType and targetId required"})
		return
	}
	switch r.Method {
	case http.MethodPost:
		if err := h.repo.AddInteraction(r.Context(), uid, req.TargetType, "like", req.TargetID); err != nil {
			writeError(w, ErrInternal("failed to like"))
			return
		}
		writeOK(w, map[string]interface{}{"status": "liked"})
	case http.MethodDelete:
		if err := h.repo.RemoveInteraction(r.Context(), uid, req.TargetType, "like", req.TargetID); err != nil {
			writeError(w, ErrInternal("failed to unlike"))
			return
		}
		writeOK(w, map[string]interface{}{"status": "unliked"})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed"})
	}
}

func (h *GenericInteractionHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	targetType := r.URL.Query().Get("targetType")
	targetID, _ := strconv.ParseInt(r.URL.Query().Get("targetId"), 10, 64)
	if targetType == "" || targetID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "targetType and targetId required"})
		return
	}
	liked, _ := h.repo.HasInteraction(r.Context(), uid, targetType, "like", targetID)
	writeOK(w, map[string]interface{}{"liked": liked})
}

func (h *GenericInteractionHandler) handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	var req struct {
		TargetType string  `json:"targetType"`
		TargetIDs  []int64 `json:"targetIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid request"})
		return
	}
	result := make(map[int64]bool)
	for _, id := range req.TargetIDs {
		liked, _ := h.repo.HasInteraction(r.Context(), uid, req.TargetType, "like", id)
		result[id] = liked
	}
	writeOK(w, result)
}

func (h *GenericInteractionHandler) handleCount(w http.ResponseWriter, r *http.Request) {
	targetType := r.URL.Query().Get("targetType")
	targetID, _ := strconv.ParseInt(r.URL.Query().Get("targetId"), 10, 64)
	if targetType == "" || targetID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "targetType and targetId required"})
		return
	}
	count, _ := h.repo.CountInteraction(r.Context(), targetType, "like", targetID)
	writeOK(w, map[string]interface{}{"count": count})
}

func (h *GenericInteractionHandler) handleBatchCount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetType string  `json:"targetType"`
		TargetIDs  []int64 `json:"targetIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid request"})
		return
	}
	result := make(map[int64]int32)
	for _, id := range req.TargetIDs {
		count, _ := h.repo.CountInteraction(r.Context(), req.TargetType, "like", id)
		result[id] = count
	}
	writeOK(w, result)
}