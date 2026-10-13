package social

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"mybilibili/pkg/httputil"
)

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// handleSpaceVisit POST /api/v1/space/visit — 记录个人空间被访问。
// 登录用户记 visitor_user_id，未登录记 IP；同一访客同一天只记一条（唯一索引去重）。
// 上报失败不影响主流程，允许游客调用。
func (h *FollowHandler) handleSpaceVisit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		HostUserID int64 `json:"host_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.HostUserID <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid host user id", "data": nil})
		return
	}
	if !h.userExists(r.Context(), req.HostUserID) {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "用户不存在", "data": nil})
		return
	}
	visitorID := h.getUserID(r)
	var visitorParam interface{}
	ip := ""
	if visitorID > 0 {
		visitorParam = visitorID
	} else {
		visitorParam = nil
		ip = clientIP(r)
	}
	_, _ = h.db.ExecContext(r.Context(),
		`INSERT INTO space_visits (host_user_id, visitor_user_id, visitor_ip) VALUES ($1,$2,$3)
		 ON CONFLICT DO NOTHING`, req.HostUserID, visitorParam, ip)
	httputil.WriteOK(w, map[string]any{"status": "ok"})
}
