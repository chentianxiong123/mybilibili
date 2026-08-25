package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"mybilibili/pkg/auth"
	"mybilibili/pkg/httputil"
)

type Handler struct {
	svc       *Service
	jwt       *auth.JWT
	scheduler *Scheduler
}

func NewHandler(svc *Service, jwt *auth.JWT) *Handler {
	return &Handler{svc: svc, jwt: jwt}
}

func (h *Handler) requirePermission(r *http.Request, permission string) (int64, bool) {
	adminID := httputil.GetAdminIDFromHeader(r)
	if adminID == 0 {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr != "" {
			claims, err := h.jwt.Parse(tokenStr)
			if err == nil {
				adminID = claims.UserId
			}
		}
	}
	if adminID == 0 {
		return 0, false
	}
	codes, err := h.svc.repo.GetAdminPermissionCodes(r.Context(), adminID)
	if err != nil {
		return adminID, false
	}
	for _, c := range codes {
		if c == permission {
			return adminID, true
		}
	}
	return adminID, false
}

func (h *Handler) requirePerm(perm string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := h.requirePermission(r, perm); !ok {
			http.Error(w, "forbidden", 403)
			return
		}
		next(w, r)
	}
}

func (h *Handler) SetScheduler(s *Scheduler) {
	h.scheduler = s
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/admin/login", h.handleLogin)
	mux.HandleFunc("/api/v1/admin/register", h.requirePerm("admin:manage", h.handleRegister))
	mux.HandleFunc("/api/v1/admin/list", h.requirePerm("admin:manage", h.handleListAdmins))
	mux.HandleFunc("/api/v1/admin/roles", h.requirePerm("role:manage", h.handleRoles))
	mux.HandleFunc("/api/v1/admin/roles/", h.requirePerm("role:manage", h.handleRolesByID))
	mux.HandleFunc("/api/v1/admin/permissions", h.requirePerm("role:manage", h.handlePermissions))
	mux.HandleFunc("/api/v1/admin/audit-logs", h.requirePerm("audit:manage", h.handleAuditLogs))
	mux.HandleFunc("/api/v1/admin/audit-logs/", h.requirePerm("audit:manage", h.handleAuditLogByID))
	mux.HandleFunc("/api/v1/admin/login-logs", h.requirePerm("security:manage", h.handleLoginLogs))
	mux.HandleFunc("/api/v1/admin/login-logs/list", h.requirePerm("security:manage", h.handleLoginLogs))
	mux.HandleFunc("/api/v1/admin/login-logs/user/", h.requirePerm("security:manage", h.handleUserLoginLogs))
	mux.HandleFunc("/api/v1/admin/security-settings", h.requirePerm("security:manage", h.handleSecuritySettings))
	mux.HandleFunc("/api/v1/admin/transcode-config", h.requirePerm("transcode:manage", h.handleTranscodeConfig))
	mux.HandleFunc("/api/v1/admin/storage/migrate", h.requirePerm("storage:manage", h.handleStorageMigrate))
	mux.HandleFunc("/api/v1/admin/operation-tasks", h.requirePerm("operation:manage", h.handleOperationTasks))
	mux.HandleFunc("/api/v1/admin/operation-tasks/list", h.requirePerm("operation:manage", h.handleOperationTasks))
	mux.HandleFunc("/api/v1/admin/operation-tasks/", h.requirePerm("operation:manage", h.handleOperationTaskByID))
	mux.HandleFunc("/api/v1/admin/audit-logs/list", h.requirePerm("audit:manage", h.handleAuditLogs))
	mux.HandleFunc("/api/v1/admin/scheduled-tasks", h.requirePerm("scheduled:manage", h.handleScheduledTasks))
	mux.HandleFunc("/api/v1/admin/scheduled-tasks/toggle", h.requirePerm("scheduled:manage", h.handleScheduledTaskToggle))
	mux.HandleFunc("/api/v1/admin/scheduled-tasks/trigger", h.requirePerm("scheduled:manage", h.handleScheduledTaskTrigger))
	mux.HandleFunc("/api/v1/admin/", h.handleAdminByID)
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
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []map[string]interface{}{}, "total": 0, "page": page, "size": size})
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
	var total int64
	h.svc.repo.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM operation_tasks`).Scan(&total)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size,
	})
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
	token, err := h.jwt.GenerateAdmin(admin.ID)
	if err != nil {
		http.Error(w, "token generation failed", 500)
		return
	}
	refreshToken, _ := h.jwt.GenerateRefresh(admin.ID)
	role := "管理员"
	if admin.AdminLevel >= 2 {
		role = "超级管理员"
	}
	permissions := []string{}
	if role == "超级管理员" {
		perms, _ := h.svc.repo.ListPermissions(r.Context())
		for _, p := range perms {
			permissions = append(permissions, p.Code)
		}
	} else {
		perms, _ := h.svc.repo.GetAdminPermissionCodes(r.Context(), admin.ID)
		permissions = perms
	}
	httputil.WriteOK(w, map[string]interface{}{
		"token":         token,
		"refresh_token": refreshToken,
		"admin_id":      admin.ID,
		"username":      admin.Username,
		"admin_level":   admin.AdminLevel,
		"role":          role,
		"permissions":   permissions,
	})
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
	h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), req.Username, "admin", "CREATE_ADMIN", "admin_users", req.Username, 0, "创建管理员", "")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(r, "admin:manage"); !ok {
		http.Error(w, "forbidden", 403)
		return
	}
	list, _ := h.svc.ListAdmins(r.Context())
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) handleRoles(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(r, "role:manage"); !ok {
		http.Error(w, "forbidden", 403)
		return
	}
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
			h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "role", "SET_ROLE_PERMISSIONS", "role_permissions", strconv.FormatInt(id, 10), 0, "设置角色权限", "")
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
		h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "role", "APPLY_ROLE_TEMPLATE", "role_permissions", strconv.FormatInt(id, 10), 0, "应用岗位模板 "+code, "")
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
		h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "role", "UPDATE_ROLE", "roles", strconv.FormatInt(id, 10), 0, "更新角色", "")
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteRole(r.Context(), id)
		h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "role", "DELETE_ROLE", "roles", strconv.FormatInt(id, 10), 0, "删除角色", "")
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
			Description:     "适合维护视频、分类、轮播图、直播、字幕、搜索与定时任务",
			PermissionCodes: []string{"video:manage", "category:manage", "banner:manage", "live:manage", "subtitle:manage", "search:manage", "operation:manage", "scheduled:manage"},
		},
		"content-review": {
			Code: "content-review", Name: "内容审核",
			Description:     "适合处理稿件审核、内容审核、举报、评论和违禁词",
			PermissionCodes: []string{"review:manage", "comment:manage"},
		},
		"customer-support": {
			Code: "customer-support", Name: "客服专员",
			Description:     "适合处理工单、客服会话、系统通知和用户查询",
			PermissionCodes: []string{"message:manage", "user:manage", "operation:manage"},
		},
		"data-analyst": {
			Code: "data-analyst", Name: "数据分析员",
			Description:     "只读查看统计数据与审计日志",
			PermissionCodes: []string{"statistics:manage", "audit:manage"},
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
				"scheduled:manage", "transcode:manage", "subtitle:manage",
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
	if list == nil {
		list = []*AuditLog{}
	}
	var total int64
	h.svc.repo.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM audit_logs`).Scan(&total)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size,
	})
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
	if list == nil {
		list = []map[string]interface{}{}
	}
	var total int64
	if userID > 0 {
		h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT COUNT(*) FROM login_logs WHERE user_id = $1`, userID).Scan(&total)
	} else {
		h.svc.repo.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM login_logs`).Scan(&total)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size,
	})
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

// handleTranscodeConfig 读取/更新转码编码器配置。
// GET：返回当前保存的 encoder 配置 + 本机硬件探测结果（是否可用 VAAPI 硬编）；
// PUT：保存 encoder 配置（auto | vaapi | x264）。
func (h *Handler) handleTranscodeConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var cfg string
		err := h.svc.repo.db.QueryRowContext(r.Context(),
			`SELECT config_value FROM system_configs WHERE config_key='transcode_encoder'`).Scan(&cfg)
		if err != nil {
			cfg = "auto"
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"encoder":  cfg,
			"vaapi":    h.detectVAAPI(),
			"vaapiDev": os.Getenv("VAAPI_DEVICE"),
			"options":  []string{"auto", "vaapi", "x264"},
		})
	case "PUT":
		var req struct {
			Encoder string `json:"encoder"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Encoder != "auto" && req.Encoder != "vaapi" && req.Encoder != "x264" {
			http.Error(w, "invalid encoder, must be auto|vaapi|x264", 400)
			return
		}
		_, err := h.svc.repo.db.ExecContext(r.Context(),
			`INSERT INTO system_configs (config_key, config_value, updated_at, updated_by)
			 VALUES ('transcode_encoder', $1, NOW(), $2)
			 ON CONFLICT (config_key) DO UPDATE SET config_value=$1, updated_at=NOW(), updated_by=$2`,
			req.Encoder, httputil.GetAdminIDFromHeader(r))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "encoder": req.Encoder})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

// detectVAAPI 探测本机是否有可用的 VAAPI 硬件编码（/dev/dri 设备 + ffmpeg 支持 h264_vaapi）。
func (h *Handler) detectVAAPI() bool {
	if _, err := os.Stat("/dev/dri/renderD128"); err != nil {
		return false
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-encoders")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "h264_vaapi")
}

// handleAdminByID 分派 /api/v1/admin/{id} 及子路径
func (h *Handler) handleAdminByID(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(r, "admin:manage"); !ok {
		http.Error(w, "forbidden", 403)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "not found", 404)
		return
	}
	// 跳过已知的有独立 handler 的子路径
	switch parts[0] {
	case "roles", "permissions", "login", "register", "list", "login-logs", "audit-logs", "storage", "operation-tasks", "security-settings", "transcode-config":
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
			h.svc.RecordAudit(r.Context(), httputil.GetAdminIDFromHeader(r), "", "admin", "SET_ADMIN_ROLES", "admin_user_roles", strconv.FormatInt(id, 10), 0, "设置管理员角色", "")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}
	http.Error(w, "not found", 404)
}

func (h *Handler) handleScheduledTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		list, _ := h.svc.ListScheduledTasks(r.Context())
		if list == nil {
			list = []*ScheduledTask{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"list": list})
	case "POST":
		var t ScheduledTask
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		if err := h.svc.CreateScheduledTask(r.Context(), &t); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(t)
	case "PUT":
		var t ScheduledTask
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		if err := h.svc.UpdateScheduledTask(r.Context(), &t); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	case "DELETE":
		body := struct {
			ID int64 `json:"id"`
		}{}
		json.NewDecoder(r.Body).Decode(&body)
		if body.ID <= 0 {
			http.Error(w, "id required", 400)
			return
		}
		_ = h.svc.DeleteScheduledTask(r.Context(), body.ID)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (h *Handler) handleScheduledTaskToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		ID      int64 `json:"id"`
		Enabled int32 `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	_ = h.svc.ToggleScheduledTask(r.Context(), req.ID, req.Enabled)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) handleScheduledTaskTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		TaskKey string `json:"task_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TaskKey == "" {
		http.Error(w, "task_key required", 400)
		return
	}
	if h.scheduler != nil {
		h.scheduler.TriggerTaskNow(r.Context(), req.TaskKey)
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
