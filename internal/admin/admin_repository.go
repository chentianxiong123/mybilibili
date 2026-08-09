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
	ID         int64
	Username   string
	Password   string
	Nickname   string
	AdminLevel int32
}

type Role struct {
	ID          int64
	Name        string
	Description string
}

type Permission struct {
	ID          int64
	Name        string
	Code        string
	URL         string
	Method      string
	ParentID    int64
	Description string
}

type AuditLog struct {
	ID            int64
	OperatorID    int64
	OperatorName  string
	OperatorRole  string
	Module        string
	Action        string
	TargetType    string
	TargetID      string
	RequestMethod string
	RequestURI    string
	ClientIP      string
	UserAgent     string
	Result        int32
	Message       string
	Detail        string
	CreatedAt     time.Time
}

type OperationTask struct {
	ID           int64
	TaskKey      string
	TaskType     string
	TaskName     string
	TargetType   string
	TargetID     string
	Status       string
	Progress     int32
	Stage        string
	Message      string
	ErrorMessage string
	OperatorID   int64
	OperatorName string
	StartedAt    *time.Time
	FinishedAt   *time.Time
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
		`SELECT id, username, password, nickname, admin_level FROM admin_users WHERE username = $1 AND password = $2`,
		username, fmt.Sprintf("%x", hash)).Scan(&u.ID, &u.Username, &u.Password, &u.Nickname, &u.AdminLevel)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) CreateAdmin(ctx context.Context, username, password, nickname string, level int32) (int64, error) {
	hash := sha256.Sum256([]byte(password))
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO admin_users (username, password, nickname, admin_level) VALUES ($1,$2,$3,$4) RETURNING id`,
		username, fmt.Sprintf("%x", hash), nickname, level).Scan(&id)
	return id, err
}

func (r *Repository) ListAdmins(ctx context.Context) ([]*AdminUser, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, username, password, nickname, admin_level FROM admin_users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*AdminUser
	for rows.Next() {
		u := &AdminUser{}
		rows.Scan(&u.ID, &u.Username, &u.Password, &u.Nickname, &u.AdminLevel)
		list = append(list, u)
	}
	return list, nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]*Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, COALESCE(description,'') FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Role
	for rows.Next() {
		role := &Role{}
		rows.Scan(&role.ID, &role.Name, &role.Description)
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

func (r *Repository) RecordLoginLog(ctx context.Context, userID int64, ip, userAgent string, status int32) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO login_logs (user_id, ip, user_agent, status) VALUES ($1,$2,$3,$4)`,
		userID, ip, userAgent, status)
	return err
}

func (r *Repository) ListLoginLogs(ctx context.Context, userID int64, page, size int32) ([]map[string]interface{}, error) {
	offset := (page - 1) * size
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs WHERE user_id = $1 ORDER BY login_time DESC LIMIT $2 OFFSET $3`,
		userID, size, offset)
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

func (s *Service) Login(ctx context.Context, username, password string) (*AdminUser, error) {
	return s.repo.AdminLogin(ctx, username, password)
}

func (s *Service) CreateAdmin(ctx context.Context, username, password, nickname string, level int32) error {
	_, err := s.repo.CreateAdmin(ctx, username, password, nickname, level)
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

func (s *Service) SetAdminRoles(ctx context.Context, adminID int64, roleIDs []int64) error {
	return s.repo.SetAdminRoles(ctx, adminID, roleIDs)
}

func (s *Service) GetAdminRoles(ctx context.Context, adminID int64) ([]int64, error) {
	return s.repo.GetAdminRoles(ctx, adminID)
}

func (s *Service) ListAuditLogs(ctx context.Context, page, size int32) ([]*AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, page, size)
}

func (s *Service) RecordLoginLog(ctx context.Context, userID int64, ip, ua string, status int32) error {
	return s.repo.RecordLoginLog(ctx, userID, ip, ua, status)
}

func (s *Service) ListLoginLogs(ctx context.Context, userID int64, page, size int32) ([]map[string]interface{}, error) {
	return s.repo.ListLoginLogs(ctx, userID, page, size)
}
