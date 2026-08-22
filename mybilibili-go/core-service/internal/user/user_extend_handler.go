package user

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
	"mybilibili/pkg/errors"
	"mybilibili/pkg/imageutil"
	pb "mybilibili/pkg/pb"
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
	mux.HandleFunc("/api/v1/user/me", h.handleMe)
	mux.HandleFunc("/api/v1/user/me/avatar", h.handleMeAvatar)
	mux.HandleFunc("/api/v1/user/token/refresh", h.handleRefresh)
	mux.HandleFunc("/api/v1/user/", h.handleUserByID)
	mux.HandleFunc("/api/v1/user/email/code", h.handleEmailCode)
	mux.HandleFunc("/api/v1/user/email/verify", h.handleEmailVerify)
	mux.HandleFunc("/api/v1/user/batch", h.handleBatch)
	mux.HandleFunc("/api/v1/user/default-avatar", h.handleDefaultAvatar)
	mux.HandleFunc("/api/v1/user/password/forgot", h.handleForgotPassword)
	mux.HandleFunc("/api/v1/user/add-experience", h.handleAddExperience)
	mux.HandleFunc("/api/v1/user/pinned-video", h.handlePinnedVideo)
	mux.HandleFunc("/api/v1/user/login-logs", h.handleLoginLogs)
	mux.HandleFunc("/api/v1/user/login-logs/count", h.handleLoginLogCount)
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
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	h.svc.repo.db.ExecContext(r.Context(),
		`INSERT INTO login_logs (user_id, ip, user_agent, status) VALUES ($1, $2, $3, 0)`,
		resp.UserId, ip, r.UserAgent())
	refreshToken, _ := h.svc.jwt.GenerateRefresh(resp.UserId)
	user, _ := h.svc.repo.FindByID(r.Context(), resp.UserId)
	avatar := ""
	if user != nil {
		avatar = user.Avatar
	}
	httputil.WriteOK(w, map[string]interface{}{
		"token":         resp.Token,
		"refresh_token": refreshToken,
		"user_id":       resp.UserId,
		"nickname":      resp.Nickname,
		"avatar":        avatar,
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
	refreshToken, _ := h.svc.jwt.GenerateRefresh(resp.UserId)
	httputil.WriteOK(w, map[string]interface{}{
		"token":         resp.Token,
		"refresh_token": refreshToken,
		"user_id":       resp.UserId,
		"nickname":      req.Nickname,
		"avatar":        "",
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
	userID, err := h.svc.jwt.ParseUserID(req.RefreshToken)
	if err != nil {
		errors.WriteHTTPError(w, errors.ErrUnauthenticated("invalid or expired refresh token"))
		return
	}
	user, err := h.svc.repo.FindByID(r.Context(), userID)
	if err != nil || user.Status != 1 {
		errors.WriteHTTPError(w, errors.ErrUnauthenticated("account is disabled"))
		return
	}
	newToken, _ := h.svc.jwt.Generate(userID)
	newRefresh, _ := h.svc.jwt.GenerateRefresh(userID)
	httputil.WriteOK(w, map[string]interface{}{
		"token":         newToken,
		"refresh_token": newRefresh,
		"user_id":       userID,
	})
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
	if req.Email == "" {
		http.Error(w, "email required", 400)
		return
	}
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	_, _ = h.svc.repo.db.ExecContext(r.Context(),
		`INSERT INTO verification_codes (identifier, code_type, code_value, expires_at)
		 VALUES ($1, 'email_code', $2, NOW() + INTERVAL '10 minutes')`,
		req.Email, code)
	sendEmail(req.Email, code)
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
	if req.Email == "" || req.Code == "" {
		http.Error(w, "email and code required", 400)
		return
	}
	var id int64
	err := h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT id FROM verification_codes
		 WHERE identifier = $1 AND code_type = 'email_code' AND code_value = $2 AND used = 0
		  AND expires_at > NOW() ORDER BY expires_at DESC LIMIT 1`,
		req.Email, req.Code).Scan(&id)
	if err != nil {
		w.Write([]byte(`{"status":"failed","message":"invalid or expired code"}`))
		return
	}
	_, _ = h.svc.repo.db.ExecContext(r.Context(), `UPDATE verification_codes SET used = 1 WHERE id = $1`, id)
	w.Write([]byte(`{"status":"verified"}`))
}

func (h *UserExtendHandler) handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var ids []int64
	json.NewDecoder(r.Body).Decode(&ids)
	if len(ids) == 0 {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := h.svc.repo.db.QueryContext(r.Context(),
		`SELECT id, username, nickname, avatar, level FROM users WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var id, level int64
		var username, nickname, avatar string
		rows.Scan(&id, &username, &nickname, &avatar, &level)
		list = append(list, map[string]interface{}{
			"id": id, "username": username, "nickname": nickname, "avatar": avatar, "level": level,
		})
	}
	json.NewEncoder(w).Encode(list)
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
	userID := httputil.GetUserIDFromHeader(r)
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
	userID := httputil.GetUserIDFromHeader(r)
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

// GET /api/v1/user/login-logs/count — 按条件计数（countUserLogs / countByCondition）
func (h *UserExtendHandler) handleLoginLogCount(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
	conds := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 0

	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		argIdx++
		uid, _ := strconv.ParseInt(uidStr, 10, 64)
		args = append(args, uid)
		conds += fmt.Sprintf(" AND user_id = $%d", argIdx)
	} else {
		argIdx++
		args = append(args, userID)
		conds += fmt.Sprintf(" AND user_id = $%d", argIdx)
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		argIdx++
		args = append(args, statusStr)
		conds += fmt.Sprintf(" AND status = $%d", argIdx)
	}
	if start := r.URL.Query().Get("start_time"); start != "" {
		argIdx++
		args = append(args, start)
		conds += fmt.Sprintf(" AND login_time >= $%d", argIdx)
	}
	if end := r.URL.Query().Get("end_time"); end != "" {
		argIdx++
		args = append(args, end)
		conds += fmt.Sprintf(" AND login_time <= $%d", argIdx)
	}

	var total int64
	_ = h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM login_logs `+conds, args...).Scan(&total)
	json.NewEncoder(w).Encode(map[string]int64{"total": total})
}

func (h *UserExtendHandler) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserIDFromHeader(r)
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
	userID := httputil.GetUserIDFromHeader(r)
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
	userID := httputil.GetUserIDFromHeader(r)
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
	userID := httputil.GetUserIDFromHeader(r)
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

func (h *UserExtendHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		user, err := h.svc.repo.FindByID(r.Context(), uid)
		if err != nil {
			errors.WriteHTTPError(w, errors.ErrNotFound("user not found"))
			return
		}
		var followerCount, followingCount, likeCount, manuscriptCount int64
		_ = h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM follows WHERE following_id = $1`, uid).Scan(&followerCount)
		_ = h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM follows WHERE follower_id = $1`, uid).Scan(&followingCount)
		_ = h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COALESCE(SUM(like_count),0) FROM manuscripts WHERE user_id = $1`, uid).Scan(&likeCount)
		_ = h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM manuscripts WHERE user_id = $1`, uid).Scan(&manuscriptCount)
		httputil.WriteOK(w, map[string]interface{}{
			"id":               user.ID,
			"username":         user.Username,
			"nickname":         user.Nickname,
			"avatar":           user.Avatar,
			"avatar_url":       user.Avatar,
			"introduction":     "",
			"sign":             "",
			"level":            user.Level,
			"follower_count":   followerCount,
			"following_count":  followingCount,
			"like_count":       likeCount,
			"manuscript_count": manuscriptCount,
			"created_at":       user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	case http.MethodPut:
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid body", "data": nil})
			return
		}
		for k, v := range body {
			switch k {
			case "nickname":
				if s, ok := v.(string); ok && s != "" {
					var exists int
					_ = h.svc.repo.db.QueryRowContext(r.Context(),
						`SELECT COUNT(*) FROM users WHERE nickname = $1 AND id != $2`, s, uid).Scan(&exists)
					if exists > 0 {
						httputil.WriteJSON(w, http.StatusConflict, map[string]interface{}{"code": 409, "message": "昵称已存在", "data": nil})
						return
					}
					_, _ = h.svc.repo.db.ExecContext(r.Context(),
						`UPDATE users SET nickname = $1, updated_at = NOW() WHERE id = $2`, s, uid)
				}
			case "avatar", "avatar_url":
				if s, ok := v.(string); ok && s != "" {
					_, _ = h.svc.repo.db.ExecContext(r.Context(),
						`UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2`, s, uid)
				}
			}
		}
		httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
	default:
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
	}
}

func (h *UserExtendHandler) handleMeAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Avatar string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Avatar == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "avatar required", "data": nil})
		return
	}
	_, err := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2`, body.Avatar, uid)
	if err != nil {
		errors.WriteHTTPError(w, errors.ErrInternal("update avatar failed"))
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"status": "ok"})
}

func (h *UserExtendHandler) handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/user/")
	// Handle avatar upload: /api/v1/user/{id}/avatar
	if strings.HasSuffix(idStr, "/avatar") {
		h.handleAvatarUpload(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "not found", 404)
		return
	}
	user, err := h.svc.repo.FindByID(r.Context(), id)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", 404)
			return
		}
		http.Error(w, "database error", 500)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"avatar":     user.Avatar,
		"level":      user.Level,
		"created_at": user.CreatedAt,
	})
}

func (h *UserExtendHandler) handleAvatarUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/user/"), "/avatar")
	uid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || uid <= 0 {
		http.Error(w, "invalid user id", 400)
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
	dir := "/tmp/mybilibili-uploads/avatars"
	os.MkdirAll(dir, 0o755)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	name := fmt.Sprintf("u%d_%d%s", uid, time.Now().Unix(), ext)
	dst := filepath.Join(dir, name)
	f, err := os.Create(dst)
	if err != nil {
		http.Error(w, "create file: "+err.Error(), 500)
		return
	}
	io.Copy(f, file)
	f.Close()
	imageutil.CompressAndReplace(dst)
	avatarURL := "/uploads/avatars/" + filepath.Base(dst)
	_, dbErr := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2`, avatarURL, uid)
	if dbErr != nil {
		http.Error(w, "update avatar: "+dbErr.Error(), 500)
		return
	}
	httputil.WriteOK(w, map[string]interface{}{"url": avatarURL})
}

func (h *UserExtendHandler) handleCaptcha(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/captcha/")
	switch path {
	case "new":
		a := rand.Intn(50) + 1
		b := rand.Intn(50) + 1
		answer := a + b
		captchaID := fmt.Sprintf("%d", time.Now().UnixNano())
		_, _ = h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO verification_codes (identifier, code_type, code_value, expires_at)
			 VALUES ($1, 'captcha', $2, NOW() + INTERVAL '5 minutes')`,
			captchaID, fmt.Sprintf("%d", answer))
		json.NewEncoder(w).Encode(map[string]string{
			"captchaId": captchaID,
			"question":  fmt.Sprintf("%d+%d=?", a, b),
		})
	case "verify":
		var req struct {
			CaptchaID string `json:"captchaId"`
			Answer    string `json:"answer"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.CaptchaID == "" || req.Answer == "" {
			json.NewEncoder(w).Encode(map[string]bool{"verified": false})
			return
		}
		var id int64
		err := h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT id FROM verification_codes
			 WHERE identifier = $1 AND code_type = 'captcha' AND code_value = $2 AND used = 0
			  AND expires_at > NOW() ORDER BY expires_at DESC LIMIT 1`,
			req.CaptchaID, req.Answer).Scan(&id)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]bool{"verified": false})
			return
		}
		_, _ = h.svc.repo.db.ExecContext(r.Context(), `UPDATE verification_codes SET used = 1 WHERE id = $1`, id)
		json.NewEncoder(w).Encode(map[string]bool{"verified": true})
	}
}

func sendEmail(to, code string) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if host == "" || port == "" || from == "" {
		log.Printf("SMTP not configured, skipping email to %s", to)
		return
	}
	auth := smtp.PlainAuth("", user, pass, host)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, to, "Your verification code", fmt.Sprintf("Your verification code is: %s\n\nThis code expires in 10 minutes.", code)))
	addr := host + ":" + port
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		log.Printf("send email to %s: %v", to, err)
	}
}

var _ = time.Now
