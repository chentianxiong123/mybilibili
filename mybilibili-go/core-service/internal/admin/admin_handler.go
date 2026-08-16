package admin

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mybilibili/core-service/internal/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/admin/login", h.handleLogin)
	mux.HandleFunc("/api/v1/admin/register", h.handleRegister)
	mux.HandleFunc("/api/v1/admin/list", h.handleListAdmins)
	mux.HandleFunc("/api/v1/admin/roles", h.handleRoles)
	mux.HandleFunc("/api/v1/admin/roles/", h.handleRolesByID)
	mux.HandleFunc("/api/v1/admin/permissions", h.handlePermissions)
	mux.HandleFunc("/api/v1/admin/audit-logs", h.handleAuditLogs)
	mux.HandleFunc("/api/v1/admin/audit-logs/", h.handleAuditLogByID)
	mux.HandleFunc("/api/v1/admin/login-logs", h.handleLoginLogs)
	mux.HandleFunc("/api/v1/admin/login-logs/list", h.handleLoginLogs)
	mux.HandleFunc("/api/v1/admin/login-logs/user/", h.handleUserLoginLogs)
	mux.HandleFunc("/api/v1/admin/security-settings", h.handleSecuritySettings)
	mux.HandleFunc("/api/v1/admin/storage/migrate", h.handleStorageMigrate)
	mux.HandleFunc("/api/v1/admin/operation-tasks", h.handleOperationTasks)
	mux.HandleFunc("/api/v1/admin/operation-tasks/", h.handleOperationTaskByID)
	mux.HandleFunc("/api/v1/admin/", h.handleAdminByID)
	mux.HandleFunc("/api/v1/user/admin/", h.handleUserAdminRoute)
}

func (h *Handler) handleStorageMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.To == "" {
		req.To = "s3"
	}
	_, err := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE videos SET source_video_url = replace(source_video_url, $1, $2)
		 WHERE source_video_url LIKE '%' || $1 || '%'`, req.From, req.To)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok", "from": req.From, "to": req.To,
	})
}

func (h *Handler) handleOperationTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	page, size := httputil.ParsePageParams(r)
	rows, err := h.svc.repo.db.QueryContext(r.Context(),
		`SELECT id, task_key, task_type, task_name, target_type, COALESCE(target_id,0),
		        status, COALESCE(progress,0), COALESCE(error_message,''), created_at, updated_at
		 FROM operation_tasks ORDER BY id DESC LIMIT $1 OFFSET $2`, size, (page-1)*size)
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, targetID int64
		var key, typ, name, targetType, status string
		var progress int32
		var errMsg string
		var created, updated string
		rows.Scan(&id, &key, &typ, &name, &targetType, &targetID, &status, &progress, &errMsg, &created, &updated)
		list = append(list, map[string]interface{}{
			"id": id, "task_key": key, "task_type": typ, "task_name": name,
			"target_type": targetType, "target_id": targetID, "status": status,
			"progress": progress, "error_message": errMsg, "created_at": created, "updated_at": updated,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleOperationTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/operation-tasks/"), 10, 64)
	row := h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT id, task_key, task_type, task_name, target_type, COALESCE(target_id,0),
		        status, COALESCE(progress,0), COALESCE(error_message,''), created_at, updated_at
		 FROM operation_tasks WHERE id=$1`, id)
	var tid, targetID int64
	var key, typ, name, targetType, status string
	var progress int32
	var errMsg string
	var created, updated string
	if err := row.Scan(&tid, &key, &typ, &name, &targetType, &targetID, &status, &progress, &errMsg, &created, &updated); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": tid, "task_key": key, "task_type": typ, "task_name": name,
		"target_type": targetType, "target_id": targetID, "status": status,
		"progress": progress, "error_message": errMsg, "created_at": created, "updated_at": updated,
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	admin, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	h.svc.repo.db.ExecContext(r.Context(),
		`INSERT INTO login_logs (user_id, ip, user_agent, status) VALUES ($1, $2, $3, 0)`,
		admin.ID, ip, r.UserAgent())
	json.NewEncoder(w).Encode(admin)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Level    int32  `json:"level"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.svc.CreateAdmin(r.Context(), req.Username, req.Password, req.Level); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	h.svc.RecordAudit(r.Context(), getAdminID(r), req.Username, "admin", "CREATE_ADMIN", "admin_users", req.Username, 0, "创建管理员", "")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListAdmins(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleRoles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListRoles(r.Context())
		json.NewEncoder(w).Encode(list)
	case "POST":
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.CreateRole(r.Context(), req.Name, req.Description)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *Handler) handleRolesByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/roles/")
	parts := strings.Split(path, "/")
	id, _ := strconv.ParseInt(parts[0], 10, 64)

	if parts[0] == "templates" && r.Method == "GET" {
		json.NewEncoder(w).Encode(roleTemplates())
		return
	}

	if len(parts) >= 2 && parts[1] == "permissions" {
		if r.Method == "GET" {
			ids, _ := h.svc.GetRolePermissions(r.Context(), id)
			json.NewEncoder(w).Encode(ids)
		} else if r.Method == "PUT" {
			var req struct {
				PermissionIDs []int64 `json:"permission_ids"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			h.svc.SetRolePermissions(r.Context(), id, req.PermissionIDs)
			h.svc.RecordAudit(r.Context(), getAdminID(r), "", "role", "SET_ROLE_PERMISSIONS", "role_permissions", strconv.FormatInt(id, 10), 0, "设置角色权限", "")
			w.Write([]byte(`{"status":"ok"}`))
		}
		return
	}

	if len(parts) >= 3 && parts[1] == "template" && r.Method == "PUT" {
		code := parts[2]
		tpl, ok := roleTemplates()[code]
		if !ok {
			http.Error(w, "岗位模板不存在", 400)
			return
		}
		permMap, err := h.svc.GetPermissionIDsByCodes(r.Context(), tpl.PermissionCodes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		missing := []string{}
		for _, c := range tpl.PermissionCodes {
			if _, ok := permMap[c]; !ok {
				missing = append(missing, c)
			}
		}
		if len(missing) > 0 {
			http.Error(w, "权限码不存在: "+strings.Join(missing, ", "), 400)
			return
		}
		ids := []int64{}
		for _, c := range tpl.PermissionCodes {
			ids = append(ids, permMap[c])
		}
		if err := h.svc.SetRolePermissions(r.Context(), id, ids); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		h.svc.RecordAudit(r.Context(), getAdminID(r), "", "role", "APPLY_ROLE_TEMPLATE", "role_permissions", strconv.FormatInt(id, 10), 0, "应用岗位模板 "+code, "")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	switch r.Method {
	case "PUT":
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.svc.UpdateRole(r.Context(), id, req.Name, req.Description)
		h.svc.RecordAudit(r.Context(), getAdminID(r), "", "role", "UPDATE_ROLE", "roles", strconv.FormatInt(id, 10), 0, "更新角色", "")
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteRole(r.Context(), id)
		h.svc.RecordAudit(r.Context(), getAdminID(r), "", "role", "DELETE_ROLE", "roles", strconv.FormatInt(id, 10), 0, "删除角色", "")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

type roleTemplate struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
}

func roleTemplates() map[string]roleTemplate {
	return map[string]roleTemplate{
		"platform-operation": {
			Code: "platform-operation", Name: "平台运营",
			Description:     "适合处理工单、运营任务、推荐策略、索引和运营审计",
			PermissionCodes: []string{"operation:manage", "search:manage", "audit:manage", "statistics:manage"},
		},
		"content-review": {
			Code: "content-review", Name: "内容审核",
			Description:     "适合处理稿件审核、内容审核、举报、评论和违禁词",
			PermissionCodes: []string{"review:manage", "comment:manage"},
		},
		"ai-manager": {
			Code: "ai-manager", Name: "AI 管理",
			Description:     "适合维护 AI 渠道、技能、用量和客服会话",
			PermissionCodes: []string{"ai:manage"},
		},
		"media-manager": {
			Code: "media-manager", Name: "媒体管理",
			Description:     "适合维护视频、字幕、分类、轮播图、直播资源",
			PermissionCodes: []string{"video:manage", "category:manage", "banner:manage", "live:manage"},
		},
		"system-manager": {
			Code: "system-manager", Name: "系统管理",
			Description:     "适合维护管理员、角色权限和安全日志",
			PermissionCodes: []string{"admin:manage", "role:manage", "security:manage"},
		},
		"super-admin": {
			Code: "super-admin", Name: "超级管理员",
			Description: "完整后台权限模板，仅用于初始化或修复超级管理员角色",
			PermissionCodes: []string{
				"user:manage", "video:manage", "comment:manage", "category:manage", "tag:manage",
				"review:manage", "statistics:manage", "role:manage", "admin:manage", "security:manage",
				"live:manage", "storage:manage", "banner:manage", "search:manage",
				"ai:manage", "message:manage", "audit:manage", "operation:manage",
			},
		},
	}
}

func (h *Handler) handlePermissions(w http.ResponseWriter, r *http.Request) {
	list, _ := h.svc.ListPermissions(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, size := httputil.ParsePageParams(r)
	list, _ := h.svc.ListAuditLogs(r.Context(), page, size)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleAuditLogByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/audit-logs/"), 10, 64)
	l, err := h.svc.GetAuditLogByID(r.Context(), id)
	if err != nil {
		http.Error(w, "audit log not found", 404)
		return
	}
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) handleUserLoginLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/login-logs/user/"), 10, 64)
	page, size := httputil.ParsePageParams(r)
	list, _ := h.svc.ListLoginLogs(r.Context(), userID, page, size)
	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "page": page, "size": size})
}

func (h *Handler) handleLoginLogs(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	page, size := httputil.ParsePageParams(r)
	list, _ := h.svc.ListLoginLogs(r.Context(), userID, page, size)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleSecuritySettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(map[string]interface{}{
			"password_policy": map[string]interface{}{
				"min_length":      8,
				"require_upper":   true,
				"require_lower":   true,
				"require_digit":   true,
				"require_special": false,
				"max_age_days":    90,
			},
			"login_policy": map[string]interface{}{
				"max_attempts":         5,
				"lockout_minutes":      30,
				"session_timeout_min":  480,
				"two_factor_required":  false,
				"ip_whitelist_enabled": false,
			},
		})
	case "PUT":
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// handleAdminByID 分派 /api/v1/admin/{id} 及子路径
func (h *Handler) handleAdminByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	// 跳过已知的有独立 handler 的子路径
	switch parts[0] {
	case "roles", "permissions", "login", "register", "list", "login-logs", "audit-logs", "storage", "operation-tasks", "security-settings":
		http.Error(w, "not found", 404)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid admin id", 400)
		return
	}
	if len(parts) == 1 {
		switch r.Method {
		case "GET":
			admin, err := h.svc.repo.GetAdminByID(r.Context(), id)
			if err != nil {
				http.Error(w, "admin not found", 404)
				return
			}
			json.NewEncoder(w).Encode(admin)
		case "PUT":
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if _, ok := body["nickname"].(string); ok {
				_ = h.svc.repo.UpdateAdmin(r.Context(), id)
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	if len(parts) == 2 && parts[1] == "roles" {
		switch r.Method {
		case "GET":
			ids, _ := h.svc.repo.GetAdminRoles(r.Context(), id)
			json.NewEncoder(w).Encode(ids)
		case "PUT":
			var req struct {
				RoleIDs []int64 `json:"role_ids"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			_ = h.svc.repo.SetAdminRoles(r.Context(), id, req.RoleIDs)
			h.svc.RecordAudit(r.Context(), getAdminID(r), "", "admin", "SET_ADMIN_ROLES", "admin_user_roles", strconv.FormatInt(id, 10), 0, "设置管理员角色", "")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	http.Error(w, "not found", 404)
}

// handleUserAdminRoute 分派 /api/v1/user/admin/ 下的子路径
func (h *Handler) handleUserAdminRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/user/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && parts[0] == "list" {
		h.handleUserAdminList(w, r)
		return
	}
	if len(parts) >= 1 {
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "invalid user id", 400)
			return
		}
		if len(parts) == 1 {
			switch r.Method {
			case "GET":
				h.handleUserAdminGet(w, r, id)
			case "PUT":
				h.handleUserAdminUpdate(w, r, id)
			default:
				http.Error(w, "method not allowed", 405)
			}
			return
		}
		if len(parts) == 2 {
			switch parts[1] {
			case "status":
				h.handleUserAdminStatus(w, r, id)
			case "password":
				h.handleUserAdminPassword(w, r, id)
			default:
				http.Error(w, "not found", 404)
			}
			return
		}
	}
	http.Error(w, "not found", 404)
}

func (h *Handler) handleUserAdminList(w http.ResponseWriter, r *http.Request) {
	page, size := httputil.ParsePageParams(r)
	offset := (page - 1) * size
	rows, err := h.svc.repo.db.QueryContext(r.Context(),
		`SELECT id, username, nickname, email, avatar, level, status, created_at
		 FROM users ORDER BY id DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	type userItem struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Nickname  string `json:"nickname"`
		Email     string `json:"email"`
		Avatar    string `json:"avatar"`
		Level     int32  `json:"level"`
		Status    int32  `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	list := []userItem{}
	for rows.Next() {
		var u userItem
		var createdAt string
		rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.Email, &u.Avatar, &u.Level, &u.Status, &createdAt)
		u.CreatedAt = createdAt
		list = append(list, u)
	}
	var total int64
	_ = h.svc.repo.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM users`).Scan(&total)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size,
	})
}

func (h *Handler) handleUserAdminGet(w http.ResponseWriter, r *http.Request, id int64) {
	var uid, level, status int32
	var username, nickname, email, avatar, createdAt string
	err := h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT id, username, nickname, email, avatar, level, status, created_at FROM users WHERE id=$1`, id).
		Scan(&uid, &username, &nickname, &email, &avatar, &level, &status, &createdAt)
	if err != nil {
		http.Error(w, "user not found", 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": uid, "username": username, "nickname": nickname, "email": email,
		"avatar": avatar, "level": level, "status": status, "created_at": createdAt,
	})
}

func (h *Handler) handleUserAdminUpdate(w http.ResponseWriter, r *http.Request, id int64) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	for k, v := range body {
		switch k {
		case "nickname":
			if s, ok := v.(string); ok && s != "" {
				_, _ = h.svc.repo.db.ExecContext(r.Context(),
					`UPDATE users SET nickname=$1, updated_at=NOW() WHERE id=$2`, s, id)
			}
		case "email":
			if s, ok := v.(string); ok {
				_, _ = h.svc.repo.db.ExecContext(r.Context(),
					`UPDATE users SET email=$1, updated_at=NOW() WHERE id=$2`, s, id)
			}
		case "level":
			if n, ok := v.(float64); ok && n > 0 {
				_, _ = h.svc.repo.db.ExecContext(r.Context(),
					`UPDATE users SET level=$1, updated_at=NOW() WHERE id=$2`, int32(n), id)
			}
		}
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleUserAdminStatus(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var body struct {
		Status int32 `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	_, err := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET status=$1, updated_at=NOW() WHERE id=$2`, body.Status, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.svc.RecordAudit(r.Context(), getAdminID(r), "", "user", "UPDATE_USER_STATUS", "users", strconv.FormatInt(id, 10), 0, "更新用户状态", "")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleUserAdminPassword(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.NewPassword == "" {
		http.Error(w, "newPassword required", 400)
		return
	}
	hash := sha256hex(body.NewPassword)
	_, err := h.svc.repo.db.ExecContext(r.Context(),
		`UPDATE users SET password=$1, updated_at=NOW() WHERE id=$2`, hash, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.svc.RecordAudit(r.Context(), getAdminID(r), "", "user", "RESET_USER_PASSWORD", "users", strconv.FormatInt(id, 10), 0, "重置用户密码", "")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}



func getAdminID(r *http.Request) int64 {
	idStr := r.Header.Get("X-Admin-Id")
	if idStr == "" {
		idStr = r.Header.Get("X-User-Id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}
