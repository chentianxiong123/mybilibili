package moderation

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mybilibili/pkg/httputil"
)

// 稿件申诉（前端 /platform/upload-manager/appeal，对齐B站申诉管理）。
//   GET  /api/v1/appeal/me/list?status=all|processing|done&page=&size=  我的申诉列表
//   GET  /api/v1/appeal/me/stats                                       申诉计数
//   GET  /api/v1/appeal/{id}                                           申诉详情（仅本人）

func (h *Handler) registerAppealRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/appeal/me/list", h.handleAppealList)
	mux.HandleFunc("/api/v1/appeal/me/stats", h.handleAppealStats)
	mux.HandleFunc("/api/v1/appeal/", h.handleAppealByID)
}

// appealStatusGroup 将前端 status 参数映射为 SQL 状态集合。
// 进行中 = PENDING + PROCESSING；已完成 = APPROVED + REJECTED；空/其他 = 全部。
func appealStatusGroup(s string) []string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "processing", "pending", "doing":
		return []string{"PENDING", "PROCESSING"}
	case "done", "finished", "completed":
		return []string{"APPROVED", "REJECTED"}
	default:
		return nil
	}
}

func appealStatusText(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PENDING", "PROCESSING":
		return "申诉处理中"
	case "APPROVED", "REJECTED":
		return "申诉完成"
	default:
		return "申诉处理中"
	}
}

func manuscriptAppealStatusText(status int64) string {
	switch status {
	case 0:
		return "审核中"
	case 1:
		return "处理中"
	case 3:
		return "已通过"
	case 4:
		return "未通过"
	case -1:
		return "已下架"
	default:
		return "处理中"
	}
}

func (h *Handler) handleAppealList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	page, size := httputil.ParsePageParams(r)
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}
	statuses := appealStatusGroup(r.URL.Query().Get("status"))

	args := []interface{}{uid}
	statusFilter := ""
	if len(statuses) > 0 {
		placeholders := make([]string, 0, len(statuses))
		for i, s := range statuses {
			placeholders = append(placeholders, "$"+strconv.Itoa(i+2))
			args = append(args, s)
		}
		statusFilter = " AND a.status IN (" + strings.Join(placeholders, ",") + ")"
	}

	var total int64
	_ = h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM appeals a WHERE a.user_id = $1`+statusFilter, args...).Scan(&total)

	offset := (page - 1) * size
	rows, err := h.svc.repo.db.QueryContext(r.Context(),
		`SELECT a.id, a.manuscript_id, COALESCE(m.title,''), COALESCE(m.status,0), COALESCE(m.cover_url,''),
		        a.reason, COALESCE(a.description,''), a.status,
		        COALESCE(a.admin_remark,''), a.created_at
		   FROM appeals a LEFT JOIN manuscripts m ON m.id = a.manuscript_id
		  WHERE a.user_id = $1`+statusFilter+` ORDER BY a.id DESC LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2),
		append(args, size, offset)...)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": 500, "message": "database error", "data": nil})
		return
	}
	defer rows.Close()

	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, manuscriptID int64
		var title, coverUrl, reason, description, status, adminRemark string
		var manuscriptStatus int64
		var createdAt time.Time
		if err := rows.Scan(&id, &manuscriptID, &title, &manuscriptStatus, &coverUrl, &reason, &description, &status, &adminRemark, &createdAt); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id":                    id,
			"manuscriptId":          manuscriptID,
			"title":                 title,
			"coverUrl":              coverUrl,
			"manuscriptStatus":      manuscriptStatus,
			"manuscriptStatusText":  manuscriptAppealStatusText(manuscriptStatus),
			"reason":                reason,
			"description":           description,
			"status":                status,
			"statusText":            appealStatusText(status),
			"adminRemark":           adminRemark,
			"appealTime":            createdAt.Format("2006-01-02 15:04:05"),
			"appealNo":              id,
		})
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(size) - 1) / int64(size)
	}
	httputil.WriteOK(w, map[string]interface{}{
		"list": list, "total": total, "page": page, "size": size, "totalPages": totalPages,
	})
}

func (h *Handler) handleAppealStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	var processing, done int64
	_ = h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM appeals WHERE user_id = $1 AND status IN ('PENDING','PROCESSING')`, uid).Scan(&processing)
	_ = h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM appeals WHERE user_id = $1 AND status IN ('APPROVED','REJECTED')`, uid).Scan(&done)
	httputil.WriteOK(w, map[string]interface{}{
		"total": processing + done, "processing": processing, "done": done,
	})
}

func (h *Handler) handleAppealByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/appeal/")
	idStr = strings.Trim(idStr, "/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid appeal id", "data": nil})
		return
	}
	var manuscriptID int64
	var title, coverUrl, reason, description, status, adminRemark string
	var manuscriptStatus int64
	var createdAt, processedAt sql.NullTime
	err := h.svc.repo.db.QueryRowContext(r.Context(),
		`SELECT a.manuscript_id, COALESCE(m.title,''), COALESCE(m.status,0), COALESCE(m.cover_url,''),
		        a.reason, COALESCE(a.description,''), a.status,
		        COALESCE(a.admin_remark,''), a.created_at, a.processed_at
		   FROM appeals a LEFT JOIN manuscripts m ON m.id = a.manuscript_id
		  WHERE a.id = $1 AND a.user_id = $2`, id, uid).
		Scan(&manuscriptID, &title, &manuscriptStatus, &coverUrl, &reason, &description, &status, &adminRemark, &createdAt, &processedAt)
	if err != nil {
		httputil.WriteJSON(w, http.StatusNotFound, map[string]interface{}{"code": 404, "message": "appeal not found", "data": nil})
		return
	}
	detail := map[string]interface{}{
		"id":                   id,
		"manuscriptId":         manuscriptID,
		"title":                title,
		"coverUrl":             coverUrl,
		"manuscriptStatus":     manuscriptStatus,
		"manuscriptStatusText": manuscriptAppealStatusText(manuscriptStatus),
		"reason":               reason,
		"description":          description,
		"status":               status,
		"statusText":           appealStatusText(status),
		"adminRemark":          adminRemark,
		"appealTime":           createdAt.Time.Format("2006-01-02 15:04:05"),
		"appealNo":             id,
	}
	if processedAt.Valid {
		detail["processedTime"] = processedAt.Time.Format("2006-01-02 15:04:05")
	}
	httputil.WriteOK(w, detail)
}
