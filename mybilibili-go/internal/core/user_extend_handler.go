package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "mybilibili/internal/core/pb"
)

type UserExtendHandler struct {
	svc *Service
}

func NewUserExtendHandler(svc *Service) *UserExtendHandler {
	return &UserExtendHandler{svc: svc}
}

func (h *UserExtendHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/user/login", h.handleLogin)
	mux.HandleFunc("/api/v1/user/register", h.handleRegister)
	mux.HandleFunc("/api/v1/user/token/refresh", h.handleRefresh)
	mux.HandleFunc("/api/v1/user/email/code", h.handleEmailCode)
	mux.HandleFunc("/api/v1/user/email/verify", h.handleEmailVerify)
	mux.HandleFunc("/api/v1/user/batch", h.handleBatch)
	mux.HandleFunc("/api/v1/user/default-avatar", h.handleDefaultAvatar)
	mux.HandleFunc("/api/v1/user/password/forgot", h.handleForgotPassword)
	mux.HandleFunc("/api/v1/user/add-experience", h.handleAddExperience)
	mux.HandleFunc("/api/v1/user/pinned-video", h.handlePinnedVideo)
	mux.HandleFunc("/api/v1/user/login-logs", h.handleLoginLogs)
	mux.HandleFunc("/api/v1/user/privacy/", h.handlePrivacy)
	mux.HandleFunc("/api/v1/user/tags", h.handleTags)
	mux.HandleFunc("/api/v1/user/settings/message", h.handleMessageSettings)
	mux.HandleFunc("/api/v1/user/settings/creator", h.handleCreatorSettings)
	mux.HandleFunc("/api/v1/captcha/", h.handleCaptcha)
}

func (h *UserExtendHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	resp, err := h.svc.Login(r.Context(), &pb.LoginRequest{Username: req.Username, Password: req.Password})
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"token":    resp.Token,
			"user_id":  resp.UserId,
			"nickname": resp.Nickname,
		},
	})
}

func (h *UserExtendHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	resp, err := h.svc.Register(r.Context(), &pb.RegisterRequest{
		Username: req.Username, Password: req.Password,
		Nickname: req.Nickname, Email: req.Email,
	})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"token":   resp.Token,
			"user_id": resp.UserId,
		},
	})
}

func (h *UserExtendHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = req
	w.Write([]byte(`{"token":"new-token-here"}`))
}

func (h *UserExtendHandler) handleEmailCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = req
	w.Write([]byte(`{"status":"code_sent"}`))
}

func (h *UserExtendHandler) handleEmailVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	_ = req
	w.Write([]byte(`{"status":"verified"}`))
}

func (h *UserExtendHandler) handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var ids []int64
	json.NewDecoder(r.Body).Decode(&ids)
	_ = ids
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

func (h *UserExtendHandler) handleDefaultAvatar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	name := r.URL.Query().Get("name")
	initial := "U"
	if len(name) > 0 {
		initial = string(name[0])
	}
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
		<circle cx="50" cy="50" r="50" fill="#e0e0e0"/>
		<text x="50" y="55" text-anchor="middle" fill="#666" font-size="40">%s</text></svg>`, initial)
}

func (h *UserExtendHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	hash := sha256.Sum256([]byte(req.NewPassword))
	h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET password = $1 WHERE email = $2`, fmt.Sprintf("%x", hash), req.Email)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *UserExtendHandler) handleAddExperience(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	amount, _ := strconv.ParseInt(r.URL.Query().Get("experienceAmount"), 10, 64)
	h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET experience = experience + $1 WHERE id = $2`, amount, userID)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *UserExtendHandler) handlePinnedVideo(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "POST":
		videoID, _ := strconv.ParseInt(r.URL.Query().Get("videoId"), 10, 64)
		h.svc.repo.db.ExecContext(r.Context(),
			`UPDATE users SET pinned_video_id = $1 WHERE id = $2`, videoID, userID)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.repo.db.ExecContext(r.Context(),
			`UPDATE users SET pinned_video_id = NULL WHERE id = $1`, userID)
		w.Write([]byte(`{"status":"ok"}`))
	case "GET":
		var videoID int64
		h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COALESCE(pinned_video_id,0) FROM users WHERE id = $1`, userID).Scan(&videoID)
		json.NewEncoder(w).Encode(map[string]int64{"video_id": videoID})
	}
}

func (h *UserExtendHandler) handleLoginLogs(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}
	offset := (page - 1) * size
	rows, err := h.svc.repo.db.QueryContext(r.Context(),
		`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs WHERE user_id = $1 ORDER BY login_time DESC LIMIT $2 OFFSET $3`,
		userID, size, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, uid, st int64
		var ip, ua, t string
		rows.Scan(&id, &uid, &ip, &ua, &st, &t)
		list = append(list, map[string]interface{}{
			"id": id, "user_id": uid, "ip": ip, "user_agent": ua, "status": st, "login_time": t,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *UserExtendHandler) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user/privacy/")
	_ = path
	switch r.Method {
	case "GET":
		var settings map[string]interface{}
		h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT public_collection, public_birthday_tags, public_coin_videos, public_like_videos,
			        public_following_list, public_followers_list
			 FROM user_privacy_settings WHERE user_id = $1`, userID).Scan(
			&settings, &settings, &settings, &settings, &settings, &settings)
		if settings == nil {
			settings = map[string]interface{}{
				"public_collection": 1, "public_birthday_tags": 0, "public_coin_videos": 0,
				"public_like_videos": 0, "public_following_list": 0, "public_followers_list": 0,
			}
		}
		json.NewEncoder(w).Encode(settings)
	case "PUT":
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO user_privacy_settings (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, userID)
		for k, v := range req {
			h.svc.repo.db.ExecContext(r.Context(),
				`UPDATE user_privacy_settings SET `+k+` = $1, updated_at = NOW() WHERE user_id = $2`, v, userID)
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *UserExtendHandler) handleTags(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		rows, _ := h.svc.repo.db.QueryContext(r.Context(),
			`SELECT tag_name FROM user_tags WHERE user_id = $1`, userID)
		defer rows.Close()
		var tags []string
		for rows.Next() {
			var t string
			rows.Scan(&t)
			tags = append(tags, t)
		}
		json.NewEncoder(w).Encode(tags)
	case "POST":
		tagName := r.URL.Query().Get("tagName")
		h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO user_tags (user_id, tag_name) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, tagName)
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		tagName := r.URL.Query().Get("tagName")
		h.svc.repo.db.ExecContext(r.Context(),
			`DELETE FROM user_tags WHERE user_id = $1 AND tag_name = $2`, userID, tagName)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *UserExtendHandler) handleMessageSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		row := h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT private_message_notification, reply_notification, at_notification, like_notification, system_notification
			 FROM message_settings WHERE user_id = $1`, userID)
		var settings = map[string]interface{}{
			"private_message_notification": 1, "reply_notification": 1,
			"at_notification": 1, "like_notification": 1, "system_notification": 1,
		}
		var p, rp, at, lk, sys int32
		if err := row.Scan(&p, &rp, &at, &lk, &sys); err == nil {
			settings["private_message_notification"] = p
			settings["reply_notification"] = rp
			settings["at_notification"] = at
			settings["like_notification"] = lk
			settings["system_notification"] = sys
		}
		json.NewEncoder(w).Encode(settings)
	case "PUT":
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO message_settings (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, userID)
		for k, v := range req {
			h.svc.repo.db.ExecContext(r.Context(),
				`UPDATE message_settings SET `+k+` = $1, updated_at = NOW() WHERE user_id = $2`, v, userID)
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *UserExtendHandler) handleCreatorSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromHeader(r)
	switch r.Method {
	case "GET":
		row := h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COALESCE(default_category_id,0), auto_publish, comment_notify, like_notify, follow_notify
			 FROM creator_settings WHERE user_id = $1`, userID)
		var s = map[string]interface{}{
			"default_category_id": 0, "auto_publish": 0, "comment_notify": 1, "like_notify": 1, "follow_notify": 1,
		}
		var catID, auto, cmt, lk, flw int32
		if err := row.Scan(&catID, &auto, &cmt, &lk, &flw); err == nil {
			s["default_category_id"] = catID
			s["auto_publish"] = auto
			s["comment_notify"] = cmt
			s["like_notify"] = lk
			s["follow_notify"] = flw
		}
		json.NewEncoder(w).Encode(s)
	case "PUT":
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO creator_settings (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, userID)
		for k, v := range req {
			h.svc.repo.db.ExecContext(r.Context(),
				`UPDATE creator_settings SET `+k+` = $1, updated_at = NOW() WHERE user_id = $2`, v, userID)
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *UserExtendHandler) handleCaptcha(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/captcha/")
	switch path {
	case "new":
		json.NewEncoder(w).Encode(map[string]string{
			"captchaId": fmt.Sprintf("%d", time.Now().UnixNano()),
			"question":  "1+1=?",
		})
	case "verify":
		json.NewEncoder(w).Encode(map[string]bool{"verified": true})
	}
}

var _ = time.Now
