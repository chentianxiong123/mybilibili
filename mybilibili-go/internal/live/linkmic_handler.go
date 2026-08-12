package live

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type LinkmicHandler struct {
	svc *LinkmicService
}

func NewLinkmicHandler(svc *LinkmicService) *LinkmicHandler {
	return &LinkmicHandler{svc: svc}
}

func (h *LinkmicHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/live/linkmic/", h.handleLinkmic)
}

func (h *LinkmicHandler) handleLinkmic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/live/linkmic/")
	parts := strings.Split(path, "/")
	userID := getUserID(r)

	switch {
	case len(parts) >= 2 && parts[0] == "apply" && r.Method == "POST":
		roomID, _ := strconv.ParseInt(parts[1], 10, 64)
		lm, err := h.svc.Apply(r.Context(), roomID, userID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(lm)

	case len(parts) >= 2 && parts[0] == "accept" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.svc.Accept(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] == "reject" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.svc.Reject(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] == "disconnect" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		h.svc.Disconnect(r.Context(), id)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] == "active" && r.Method == "GET":
		roomID, _ := strconv.ParseInt(parts[1], 10, 64)
		list, _ := h.svc.Active(r.Context(), roomID)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 2 && parts[0] == "pending" && r.Method == "GET":
		roomID, _ := strconv.ParseInt(parts[1], 10, 64)
		list, _ := h.svc.Pending(r.Context(), roomID)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 2 && parts[0] == "queue-position" && r.Method == "GET":
		roomID, _ := strconv.ParseInt(parts[1], 10, 64)
		pos, _ := h.svc.QueuePosition(r.Context(), roomID, userID)
		json.NewEncoder(w).Encode(map[string]int{"position": pos})

	case len(parts) >= 2 && parts[0] == "toggle-audio" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		enabled, err := h.svc.ToggleAudio(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]int32{"audio_enabled": enabled})

	case len(parts) >= 2 && parts[0] == "toggle-video" && r.Method == "POST":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		enabled, err := h.svc.ToggleVideo(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]int32{"video_enabled": enabled})
	}
}
