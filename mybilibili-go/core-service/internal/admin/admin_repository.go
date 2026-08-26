package admin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type AdminUser struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"-"`
	AdminLevel int32  `json:"-"`
	CreatedAt  string `json:"created_at"`
	Roles      []Role `json:"roles,omitempty"`
}

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreateTime  string    `json:"create_time"`
}

type Permission struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	URL         string `json:"url"`
	Method      string `json:"method"`
	ParentID    int64  `json:"parent_id"`
	Description string `json:"description"`
}

type AuditLog struct {
	ID            int64     `json:"id"`
	OperatorID    int64     `json:"operator_id"`
	OperatorName  string    `json:"operator_name"`
	OperatorRole  string    `json:"operator_role"`
	Module        string    `json:"module"`
	Action        string    `json:"action"`
	TargetType    string    `json:"target_type"`
	TargetID      string    `json:"target_id"`
	RequestMethod string    `json:"request_method"`
	RequestURI    string    `json:"request_uri"`
	ClientIP      string    `json:"client_ip"`
	UserAgent     string    `json:"user_agent"`
	Result        int32     `json:"result"`
	Message       string    `json:"message"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"created_at"`
}

type OperationTask struct {
	ID           int64      `json:"id"`
	TaskKey      string     `json:"task_key"`
	TaskType     string     `json:"task_type"`
	TaskName     string     `json:"task_name"`
	TargetType   string     `json:"target_type"`
	TargetID     string     `json:"target_id"`
	Status       string     `json:"status"`
	Progress     int32      `json:"progress"`
	Stage        string     `json:"stage"`
	Message      string     `json:"message"`
	ErrorMessage string     `json:"error_message"`
	OperatorID   int64      `json:"operator_id"`
	OperatorName string     `json:"operator_name"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

type ScheduledTask struct {
	ID            int64      `json:"id"`
	TaskKey       string     `json:"task_key"`
	TaskName      string     `json:"task_name"`
	Description   string     `json:"description"`
	CronExpr      string     `json:"cron_expr"`
	TaskType      string     `json:"task_type"`
	TaskConfig    string     `json:"task_config"`
	Enabled       int32      `json:"enabled"`
	LastRunAt     *time.Time `json:"last_run_at"`
	LastRunResult string     `json:"last_run_result"`
	LastRunMsg    string     `json:"last_run_message"`
	NextRunAt     *time.Time `json:"next_run_at"`
	RunCount      int32      `json:"run_count"`
	MaxRetries    int32      `json:"max_retries"`
	RetryCount    int32      `json:"retry_count"`
	TimeoutSecs   int32      `json:"timeout_seconds"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AdminLogin(ctx context.Context, username, password string) (*AdminUser, error) {
	hash := sha256.Sum256([]byte(password))
	u := &AdminUser{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password, COALESCE(admin_level,1) FROM admin_users WHERE username = $1 AND password = $2`,
		username, fmt.Sprintf("%x", hash)).Scan(&u.ID, &u.Username, &u.Password, &u.AdminLevel)
	if err != nil {
		return nil, err
	}
	u.Roles = r.getRolesForAdmin(ctx, u.ID)
	return u, nil
}

func (r *Repository) CreateAdmin(ctx context.Context, username, password string, level int32) (int64, error) {
	hash := sha256.Sum256([]byte(password))
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO admin_users (username, password, admin_level) VALUES ($1,$2,$3) RETURNING id`,
		username, fmt.Sprintf("%x", hash), level).Scan(&id)
	return id, err
}

func (r *Repository) ListAdmins(ctx context.Context) ([]*AdminUser, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, username, password, COALESCE(admin_level,1), COALESCE(created_at::text,'') FROM admin_users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*AdminUser
	for rows.Next() {
		u := &AdminUser{}
		rows.Scan(&u.ID, &u.Username, &u.Password, &u.AdminLevel, &u.CreatedAt)
		u.Roles = r.getRolesForAdmin(ctx, u.ID)
		list = append(list, u)
	}
	return list, nil
}

func (r *Repository) GetAdminByID(ctx context.Context, id int64) (*AdminUser, error) {
	u := &AdminUser{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, COALESCE(admin_level,1) FROM admin_users WHERE id=$1`, id).
		Scan(&u.ID, &u.Username, &u.AdminLevel)
	if err != nil {
		return nil, err
	}
	u.Roles = r.getRolesForAdmin(ctx, u.ID)
	return u, nil
}

func (r *Repository) getRolesForAdmin(ctx context.Context, adminID int64) []Role {
	roleRows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.name, COALESCE(r.description,''), COALESCE(r.create_time::text,'')
		 FROM admin_user_roles aur JOIN roles r ON r.id=aur.role_id
		 WHERE aur.admin_user_id=$1 ORDER BY r.id`, adminID)
	if err != nil {
		return nil
	}
	defer roleRows.Close()
	var roles []Role
	for roleRows.Next() {
		rl := Role{}
		roleRows.Scan(&rl.ID, &rl.Name, &rl.Description, &rl.CreateTime)
		roles = append(roles, rl)
	}
	return roles
}

func (r *Repository) UpdateAdmin(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admin_users SET updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *Repository) ListRoles(ctx context.Context) ([]*Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, COALESCE(description,''), COALESCE(create_time::text,'') FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Role
	for rows.Next() {
		role := &Role{}
		rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreateTime)
		list = append(list, role)
	}
	return list, nil
}

func (r *Repository) CreateRole(ctx context.Context, name, desc string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO roles (name, description) VALUES ($1,$2) RETURNING id`, name, desc).Scan(&id)
	return id, err
}

func (r *Repository) UpdateRole(ctx context.Context, id int64, name, desc string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE roles SET name=$1, description=$2, update_time=NOW() WHERE id=$3`, name, desc, id)
	return err
}

func (r *Repository) DeleteRole(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]*Permission, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, code, COALESCE(url,''), COALESCE(method,''), COALESCE(parent_id,0), COALESCE(description,'')
		 FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Permission
	for rows.Next() {
		p := &Permission{}
		rows.Scan(&p.ID, &p.Name, &p.Code, &p.URL, &p.Method, &p.ParentID, &p.Description)
		list = append(list, p)
	}
	return list, nil
}

func (r *Repository) SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	for _, pid := range permIDs {
		tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, roleID, pid)
	}
	return tx.Commit()
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT permission_id FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) GetPermissionIDsByCodes(ctx context.Context, codes []string) (map[string]int64, error) {
	result := make(map[string]int64, len(codes))
	if len(codes) == 0 {
		return result, nil
	}
	query := `SELECT id, code FROM permissions WHERE code = ANY($1)`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(codes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		result[code] = id
	}
	return result, nil
}

func (r *Repository) SetAdminRoles(ctx context.Context, adminID int64, roleIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.ExecContext(ctx, `DELETE FROM admin_user_roles WHERE admin_user_id = $1`, adminID)
	for _, rid := range roleIDs {
		tx.ExecContext(ctx, `INSERT INTO admin_user_roles (admin_user_id, role_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, adminID, rid)
	}
	return tx.Commit()
}

func (r *Repository) GetAdminRoles(ctx context.Context, adminID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT role_id FROM admin_user_roles WHERE admin_user_id = $1`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (operator_id, operator_name, operator_role, module, action, target_type, target_id,
		 request_method, request_uri, client_ip, user_agent, result, message, detail)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		log.OperatorID, log.OperatorName, log.OperatorRole, log.Module, log.Action,
		log.TargetType, log.TargetID, log.RequestMethod, log.RequestURI, log.ClientIP,
		log.UserAgent, log.Result, log.Message, log.Detail)
	return err
}

// RecordAudit 便捷审计方法（对齐旧版 AuditLogService.record）。
func (s *Service) RecordAudit(ctx context.Context, operatorID int64, operatorName, module, action, targetType, targetID string, result int32, message, detail string) error {
	return s.repo.CreateAuditLog(ctx, &AuditLog{
		OperatorID: operatorID, OperatorName: operatorName, Module: module, Action: action,
		TargetType: targetType, TargetID: targetID, Result: result, Message: message, Detail: detail,
	})
}

func (r *Repository) ListAuditLogs(ctx context.Context, page, size int32) ([]*AuditLog, error) {
	offset := (page - 1) * size
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, operator_id, operator_name, operator_role, module, action, target_type, target_id,
		        request_method, request_uri, client_ip, user_agent, result, message, detail, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*AuditLog
	for rows.Next() {
		l := &AuditLog{}
		rows.Scan(&l.ID, &l.OperatorID, &l.OperatorName, &l.OperatorRole, &l.Module, &l.Action,
			&l.TargetType, &l.TargetID, &l.RequestMethod, &l.RequestURI, &l.ClientIP,
			&l.UserAgent, &l.Result, &l.Message, &l.Detail, &l.CreatedAt)
		list = append(list, l)
	}
	return list, nil
}

func (r *Repository) GetAuditLogByID(ctx context.Context, id int64) (*AuditLog, error) {
	l := &AuditLog{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, operator_id, operator_name, operator_role, module, action, target_type, target_id,
		        request_method, request_uri, client_ip, user_agent, result, message, detail, created_at
		 FROM audit_logs WHERE id = $1`, id).
		Scan(&l.ID, &l.OperatorID, &l.OperatorName, &l.OperatorRole, &l.Module, &l.Action,
			&l.TargetType, &l.TargetID, &l.RequestMethod, &l.RequestURI, &l.ClientIP,
			&l.UserAgent, &l.Result, &l.Message, &l.Detail, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *Repository) RecordLoginLog(ctx context.Context, userID int64, ip, userAgent string, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO login_logs (user_id, ip, user_agent, status) VALUES ($1,$2,$3,$4)`,
		userID, ip, userAgent, status)
	return err
}

func (r *Repository) ListLoginLogs(ctx context.Context, userID int64, page, size int32) ([]map[string]interface{}, error) {
	offset := (page - 1) * size
	var query string
	var args []interface{}
	if userID > 0 {
		query = `SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs WHERE user_id = $1 ORDER BY login_time DESC LIMIT $2 OFFSET $3`
		args = []interface{}{userID, size, offset}
	} else {
		query = `SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs ORDER BY login_time DESC LIMIT $1 OFFSET $2`
		args = []interface{}{size, offset}
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
	return list, nil
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// InitPermissions 初始化权限表，确保所有权限码存在
func (r *Repository) InitPermissions(ctx context.Context) error {
	perms := []struct {
		Name string
		Code string
		Desc string
	}{
		{"用户管理", "user:manage", "查询/禁用/重置密码用户"},
		{"稿件管理", "video:manage", "查询/审核/发布/下架稿件"},
		{"评论管理", "comment:manage", "删除/审核评论"},
		{"分类管理", "category:manage", "增删改视频分类"},
		{"标签管理", "tag:manage", "增删改标签"},
		{"内容审核", "review:manage", "审核稿件/评论/举报"},
		{"统计查看", "statistics:manage", "查看仪表盘统计"},
		{"角色管理", "role:manage", "增删改角色/岗位模板"},
		{"管理员管理", "admin:manage", "增删改管理员/分配角色"},
		{"安全设置", "security:manage", "修改安全策略/登录日志"},
		{"直播管理", "live:manage", "管理直播间/强制下播"},
		{"存储管理", "storage:manage", "触发存储迁移"},
		{"轮播图管理", "banner:manage", "增删改轮播图/背景图"},
		{"搜索管理", "search:manage", "管理搜索索引/推荐配置"},
		{"AI 管理", "ai:manage", "管理 AI 渠道/技能/用量"},
		{"消息管理", "message:manage", "发送系统通知/公告"},
		{"审计日志", "audit:manage", "查看审计日志"},
		{"运营任务", "operation:manage", "查看/管理运营任务"},
		{"定时任务", "scheduled:manage", "管理定时任务"},
		{"转码配置", "transcode:manage", "修改转码编码器配置"},
		{"字幕管理", "subtitle:manage", "管理字幕上传/审核"},
	}
	for _, p := range perms {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO permissions (name, code, description)
			 VALUES ($1, $2, $3) ON CONFLICT (code) DO UPDATE SET name=$1, description=$3`,
			p.Name, p.Code, p.Desc)
		if err != nil {
			return err
		}
	}
	// 初始化超级管理员角色（id=1）全部权限
	permRows, err := r.db.QueryContext(ctx, `SELECT id FROM permissions`)
	if err != nil {
		return err
	}
	defer permRows.Close()
	var ids []int64
	for permRows.Next() {
		var id int64
		permRows.Scan(&id)
		ids = append(ids, id)
	}
	for _, pid := range ids {
		r.db.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES (1, $1) ON CONFLICT DO NOTHING`, pid)
	}
	return nil
}

// GetAdminPermissionCodes 获取指定管理员的所有权限码
func (r *Repository) GetAdminPermissionCodes(ctx context.Context, adminID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT p.code
		 FROM admin_user_roles aur
		 JOIN role_permissions rp ON rp.role_id = aur.role_id
		 JOIN permissions p ON p.id = rp.permission_id
		 WHERE aur.admin_user_id = $1`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var code string
		rows.Scan(&code)
		codes = append(codes, code)
	}
	return codes, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*AdminUser, error) {
	return s.repo.AdminLogin(ctx, username, password)
}

func (s *Service) CreateAdmin(ctx context.Context, username, password string, level int32) error {
	_, err := s.repo.CreateAdmin(ctx, username, password, level)
	return err
}

func (s *Service) ListAdmins(ctx context.Context) ([]*AdminUser, error) {
	return s.repo.ListAdmins(ctx)
}

func (s *Service) ListRoles(ctx context.Context) ([]*Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) CreateRole(ctx context.Context, name, desc string) error {
	_, err := s.repo.CreateRole(ctx, name, desc)
	return err
}

func (s *Service) UpdateRole(ctx context.Context, id int64, name, desc string) error {
	return s.repo.UpdateRole(ctx, id, name, desc)
}

func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	return s.repo.DeleteRole(ctx, id)
}

func (s *Service) ListPermissions(ctx context.Context) ([]*Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *Service) SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	return s.repo.SetRolePermissions(ctx, roleID, permIDs)
}

func (s *Service) GetRolePermissions(ctx context.Context, roleID int64) ([]int64, error) {
	return s.repo.GetRolePermissions(ctx, roleID)
}

func (s *Service) GetPermissionIDsByCodes(ctx context.Context, codes []string) (map[string]int64, error) {
	return s.repo.GetPermissionIDsByCodes(ctx, codes)
}

func (s *Service) SetAdminRoles(ctx context.Context, adminID int64, roleIDs []int64) error {
	return s.repo.SetAdminRoles(ctx, adminID, roleIDs)
}

func (s *Service) GetAdminRoles(ctx context.Context, adminID int64) ([]int64, error) {
	return s.repo.GetAdminRoles(ctx, adminID)
}

func (s *Service) ListAuditLogs(ctx context.Context, page, size int32) ([]*AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, page, size)
}

func (s *Service) GetAuditLogByID(ctx context.Context, id int64) (*AuditLog, error) {
	return s.repo.GetAuditLogByID(ctx, id)
}

func (s *Service) RecordLoginLog(ctx context.Context, userID int64, ip, ua string, status int32) error {
	return s.repo.RecordLoginLog(ctx, userID, ip, ua, status)
}

func (s *Service) ListLoginLogs(ctx context.Context, userID int64, page, size int32) ([]map[string]interface{}, error) {
	return s.repo.ListLoginLogs(ctx, userID, page, size)
}

// ---- ScheduledTask CRUD ----

func (r *Repository) ListScheduledTasks(ctx context.Context) ([]*ScheduledTask, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_key, task_name, description, cron_expr, task_type, task_config,
		        enabled, last_run_at, last_run_result, last_run_message,
		        next_run_at, run_count, max_retries, retry_count, timeout_seconds,
		        created_at, updated_at
		 FROM scheduled_tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*ScheduledTask
	for rows.Next() {
		t := &ScheduledTask{}
		if err := rows.Scan(&t.ID, &t.TaskKey, &t.TaskName, &t.Description,
			&t.CronExpr, &t.TaskType, &t.TaskConfig,
			&t.Enabled, &t.LastRunAt, &t.LastRunResult, &t.LastRunMsg,
			&t.NextRunAt, &t.RunCount, &t.MaxRetries, &t.RetryCount, &t.TimeoutSecs,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *Repository) GetScheduledTaskByKey(ctx context.Context, taskKey string) (*ScheduledTask, error) {
	t := &ScheduledTask{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, task_key, task_name, description, cron_expr, task_type, task_config,
		        enabled, last_run_at, last_run_result, last_run_message,
		        next_run_at, run_count, max_retries, retry_count, timeout_seconds,
		        created_at, updated_at
		 FROM scheduled_tasks WHERE task_key = $1`, taskKey).Scan(
		&t.ID, &t.TaskKey, &t.TaskName, &t.Description,
		&t.CronExpr, &t.TaskType, &t.TaskConfig,
		&t.Enabled, &t.LastRunAt, &t.LastRunResult, &t.LastRunMsg,
		&t.NextRunAt, &t.RunCount, &t.MaxRetries, &t.RetryCount, &t.TimeoutSecs,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) CreateScheduledTask(ctx context.Context, t *ScheduledTask) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO scheduled_tasks (task_key, task_name, description, cron_expr, task_type, task_config,
		 enabled, max_retries, timeout_seconds)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		t.TaskKey, t.TaskName, t.Description, t.CronExpr, t.TaskType, t.TaskConfig,
		t.Enabled, t.MaxRetries, t.TimeoutSecs).Scan(&t.ID)
}

func (r *Repository) UpdateScheduledTask(ctx context.Context, t *ScheduledTask) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET task_name=$1, description=$2, cron_expr=$3,
		 task_type=$4, task_config=$5, enabled=$6, max_retries=$7, timeout_seconds=$8,
		 updated_at=NOW() WHERE id=$9`,
		t.TaskName, t.Description, t.CronExpr, t.TaskType, t.TaskConfig,
		t.Enabled, t.MaxRetries, t.TimeoutSecs, t.ID)
	return err
}

func (r *Repository) ToggleScheduledTask(ctx context.Context, id int64, enabled int32) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET enabled=$1, updated_at=NOW() WHERE id=$2`, enabled, id)
	return err
}

func (r *Repository) DeleteScheduledTask(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM scheduled_tasks WHERE id=$1`, id)
	return err
}

func (r *Repository) UpdateScheduledTaskRun(ctx context.Context, taskKey string, result, message string, nextRunAt *time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET last_run_at=NOW(), last_run_result=$1, last_run_message=$2,
		 next_run_at=$3, run_count=run_count+1, retry_count=0, updated_at=NOW()
		 WHERE task_key=$4`, result, message, nextRunAt, taskKey)
	return err
}

// ---- Service wrappers ----

func (s *Service) ListScheduledTasks(ctx context.Context) ([]*ScheduledTask, error) {
	return s.repo.ListScheduledTasks(ctx)
}

func (s *Service) GetScheduledTaskByKey(ctx context.Context, taskKey string) (*ScheduledTask, error) {
	return s.repo.GetScheduledTaskByKey(ctx, taskKey)
}

func (s *Service) CreateScheduledTask(ctx context.Context, t *ScheduledTask) error {
	return s.repo.CreateScheduledTask(ctx, t)
}

func (s *Service) UpdateScheduledTask(ctx context.Context, t *ScheduledTask) error {
	return s.repo.UpdateScheduledTask(ctx, t)
}

func (s *Service) ToggleScheduledTask(ctx context.Context, id int64, enabled int32) error {
	return s.repo.ToggleScheduledTask(ctx, id, enabled)
}

func (s *Service) DeleteScheduledTask(ctx context.Context, id int64) error {
	return s.repo.DeleteScheduledTask(ctx, id)
}

func (s *Service) UpdateScheduledTaskRun(ctx context.Context, taskKey, result, message string, nextRunAt *time.Time) error {
	return s.repo.UpdateScheduledTaskRun(ctx, taskKey, result, message, nextRunAt)
}
