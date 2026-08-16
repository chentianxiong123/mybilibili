package meeting

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/meeting/create", h.handleCreate)
	mux.HandleFunc("/api/v1/meeting/room/", h.handleRoomByCode)
	mux.HandleFunc("/api/v1/meeting/my-rooms", h.handleMyRooms)
	mux.HandleFunc("/api/v1/meeting/join/", h.handleJoin)
	mux.HandleFunc("/api/v1/meeting/leave/", h.handleLeave)
	mux.HandleFunc("/api/v1/meeting/participants/", h.handleParticipants)
	mux.HandleFunc("/api/v1/meeting/end/", h.handleEnd)
	mux.HandleFunc("/api/v1/meeting/reserve", h.handleReserve)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID, userName := getUser(r)
	roomName := r.URL.Query().Get("room_name")
	if roomName == "" {
		var req struct {
			RoomName string `json:"room_name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		roomName = req.RoomName
	}
	m, err := h.svc.Create(r.Context(), roomName, userID, userName)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) handleRoomByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/v1/meeting/room/")
	m, err := h.svc.GetByCode(r.Context(), code)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) handleMyRooms(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUser(r)
	list, _ := h.svc.MyRooms(r.Context(), userID)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleJoin(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/v1/meeting/join/")
	userID, userName := getUser(r)
	m, err := h.svc.Join(r.Context(), code, userID, userName)
	if err != nil {
		http.Error(w, "room not found", 404)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) handleLeave(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/meeting/leave/"), 10, 64)
	userID, _ := getUser(r)
	h.svc.Leave(r.Context(), roomID, userID)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleParticipants(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/meeting/participants/"), 10, 64)
	list, _ := h.svc.Participants(r.Context(), roomID)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleEnd(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/meeting/end/"), 10, 64)
	h.svc.End(r.Context(), roomID)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleReserve(w http.ResponseWriter, r *http.Request) {
	userID, userName := getUser(r)
	var req struct {
		RoomName       string `json:"room_name"`
		ScheduledStart string `json:"scheduled_start"`
		ScheduledEnd   string `json:"scheduled_end"`
		Reason         string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	startT, _ := time.Parse(time.RFC3339, req.ScheduledStart)
	endT, _ := time.Parse(time.RFC3339, req.ScheduledEnd)
	m, err := h.svc.Reserve(r.Context(), req.RoomName, userID, userName, startT, endT, req.Reason)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func getUser(r *http.Request) (int64, string) {
	idStr := r.Header.Get("X-User-Id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id, r.Header.Get("X-Username")
}
