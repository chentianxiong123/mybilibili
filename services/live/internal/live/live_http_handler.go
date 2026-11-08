package live

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
	"mybilibili/pkg/imageutil"
)

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "data": data, "message": "ok"})
}

type HTTPHandler struct {
	svc *Service
	hub *Hub
}

func NewHTTPHandler(svc *Service, hub *Hub) *HTTPHandler {
	return &HTTPHandler{svc: svc, hub: hub}
}

func (h *HTTPHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/live/room", h.handleRoom)
	mux.HandleFunc("/api/v1/live/room/list", h.handleListRooms)
	mux.HandleFunc("/api/v1/live/room/", h.handleRoomByID)
	mux.HandleFunc("/api/v1/live/rooms", h.handleListRooms)
	mux.HandleFunc("/api/v1/live/srs/hook", h.handleSRSCallback)
	mux.HandleFunc("/api/v1/live/health", h.handleHealth)
	mux.Handle("/api/v1/live/ws", h.hub.HandleWS(h.svc))
}

func (h *HTTPHandler) handleRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	userID := httputil.GetUserIDFromHeader(r)
	if userID == 0 {
		http.Error(w, "unauthorized", 401)
		return
	}

	var req struct {
		RoomName string `json:"room_name"`
		Cover    string `json:"cover"`
		Category string `json:"category"`
		MaxSeats int32  `json:"max_seats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if req.RoomName == "" {
		http.Error(w, "room_name required", 400)
		return
	}

	room, err := h.svc.CreateRoom(r.Context(), req.RoomName, req.Cover, req.Category, userID, req.MaxSeats)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	writeJSON(w, room)
}

func (h *HTTPHandler) handleRoomByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/live/room/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}

	if parts[0] == "my" {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", 405)
			return
		}
		userID := httputil.GetUserIDFromHeader(r)
		if userID == 0 {
			http.Error(w, "unauthorized", 401)
			return
		}
		room, err := h.svc.GetRoomByHost(r.Context(), userID)
		if err != nil {
			http.Error(w, "no room", 404)
			return
		}
		writeJSON(w, room)
		return
	}

	if parts[0] == "cover" {
		h.handleCoverUpload(w, r)
		return
	}

	roomID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid room id", 400)
		return
	}

	userID := httputil.GetUserIDFromHeader(r)

	if len(parts) >= 2 {
		switch parts[1] {
		case "status":
			if r.Method != "PUT" || userID == 0 {
				http.Error(w, "method not allowed", 405)
				return
			}
			var req struct {
				Status string `json:"status"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			if err := h.svc.UpdateRoomStatus(r.Context(), roomID, userID, req.Status); err != nil {
				http.Error(w, err.Error(), 403)
				return
			}
			w.Write([]byte(`{"status":"ok"}`))
			return
		case "schedule":
			if r.Method != "PUT" || userID == 0 {
				http.Error(w, "method not allowed", 405)
				return
			}
			var req struct {
				ScheduledAt *int64 `json:"scheduled_at"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			var scheduledAt *time.Time
			if req.ScheduledAt != nil {
				t := time.UnixMilli(*req.ScheduledAt)
				scheduledAt = &t
			}
			if err := h.svc.ScheduleRoom(r.Context(), roomID, userID, scheduledAt); err != nil {
				http.Error(w, err.Error(), 403)
				return
			}
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
	}

	switch r.Method {
	case "GET":
		room, err := h.svc.GetRoom(r.Context(), roomID)
		if err != nil {
			http.Error(w, "room not found", 404)
			return
		}
		writeJSON(w, room)
	case "PUT":
		if userID == 0 {
			http.Error(w, "unauthorized", 401)
			return
		}
		var req struct {
			RoomName string `json:"room_name"`
			Cover    string `json:"cover"`
			Category string `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		if err := h.svc.UpdateRoom(r.Context(), roomID, userID, req.RoomName, req.Cover, req.Category); err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *HTTPHandler) handleListRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	// FINDING-012 修复：钳制分页（page=0/超大 → 负 OFFSET/溢出 500）。
	page32, size32 := httputil.SafePage(int32(page), int32(pageSize))

	rooms, err := h.svc.ListLiveRooms(r.Context(), page32, size32)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeJSON(w, rooms)
}

func (h *HTTPHandler) handleSRSCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	var req struct {
		Action    string `json:"action"`
		StreamKey string `json:"stream"`
		ClientID  string `json:"client_id"`
		IP        string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":-1,"msg":"invalid"}`, 400)
		return
	}

	if err := h.svc.HandleSRSCallback(r.Context(), req.Action, req.StreamKey); err != nil {
		log.Printf("SRS callback error: %v", err)
		http.Error(w, `{"code":-1,"msg":"failed"}`, 200)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"code":0}`))
}

func (h *HTTPHandler) handleCoverUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	// FINDING-025 修复：封面上传需用户身份（此前未认证者可上传任意图片 + 硬编码
	// userID=0 导致 UpdateRoom 恒 403 被吞、封面永不更新）。
	userID, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), 400)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", 400)
		return
	}
	defer file.Close()
	roomIDStr := r.FormValue("roomId")
	roomID, _ := strconv.ParseInt(roomIDStr, 10, 64)
	dir := "/tmp/mybilibili-uploads/live-covers"
	os.MkdirAll(dir, 0o755)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	name := fmt.Sprintf("room_%s_%d%s", roomIDStr, time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, name)
	f, err := os.Create(dst)
	if err != nil {
		http.Error(w, "create file: "+err.Error(), 500)
		return
	}
	io.Copy(f, file)
	f.Close()
	imageutil.CompressAndReplace(dst)
	coverURL := "/uploads/live-covers/" + filepath.Base(dst)
	if roomID > 0 {
		// FINDING-025 修复：传真实 userID 并检查错误（此前硬编码 0 + 吞错 → 封面永不更新）。
		if err := h.svc.UpdateRoom(r.Context(), roomID, userID, "", coverURL, ""); err != nil {
			http.Error(w, "update room cover: "+err.Error(), 403)
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "uploaded", "url": coverURL})
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"live ok"}`))
}

var _ = log.Print
