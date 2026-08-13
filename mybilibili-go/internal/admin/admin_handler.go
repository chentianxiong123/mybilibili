package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	mux.HandleFunc("/api/v1/admin/login-logs/user/", h.handleUserLoginLogs)
	mux.HandleFunc("/api/v1/admin/storage/migrate", h.handleStorageMigrate)
	mux.HandleFunc("/api/v1/admin/operation-tasks", h.handleOperationTasks)
	mux.HandleFunc("/api/v1/admin/operation-tasks/", h.handleOperationTaskByID)
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
	page, size := parsePage(r)
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
	if err := h.svc.CreateAdmin(r.Context(), req.Username, req.Password, req.Nickname, req.Level); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
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
		w.Write([]byte(`{"status":"ok"}`))
	case "DELETE":
		h.svc.DeleteRole(r.Context(), id)
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
			Description:     "适合维护视频、字幕、分类、轮播图、直播和会议资源",
			PermissionCodes: []string{"video:manage", "category:manage", "banner:manage", "live:manage", "meeting:manage"},
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
				"live:manage", "meeting:manage", "storage:manage", "banner:manage", "search:manage",
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
	page, size := parsePage(r)
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
	page, size := parsePage(r)
	list, _ := h.svc.ListLoginLogs(r.Context(), userID, page, size)
	json.NewEncoder(w).Encode(map[string]interface{}{"list": list, "page": page, "size": size})
}

func (h *Handler) handleLoginLogs(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	page, size := parsePage(r)
	list, _ := h.svc.ListLoginLogs(r.Context(), userID, page, size)
	json.NewEncoder(w).Encode(list)
}

func parsePage(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	size, _ := strconv.ParseInt(r.URL.Query().Get("page_size"), 10, 32)
	if size < 1 {
		size, _ = strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 32)
	}
	if size < 1 {
		size, _ = strconv.ParseInt(r.URL.Query().Get("size"), 10, 32)
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return int32(page), int32(size)
}
