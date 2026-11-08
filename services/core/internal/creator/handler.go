package creator

import (
	"encoding/json"
	"net/http"

	myerr "mybilibili/pkg/errors"
	"mybilibili/pkg/httputil"
)

// PermChecker 后台权限校验（由 admin.Handler 实现）。
type PermChecker interface {
	CheckPermission(r *http.Request, permission string) (int64, bool)
}

// Handler 创作者中心「流量/激励」后端。
type Handler struct {
	svc  *Service
	perm PermChecker
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SetPermChecker(p PermChecker) { h.perm = p }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/creator/coupons", h.handleCoupons)
	mux.HandleFunc("/api/v1/creator/coupons/balance", h.handleBalance)
	mux.HandleFunc("/api/v1/creator/coupons/redeem", h.handleRedeem)
	mux.HandleFunc("/api/v1/creator/coupons/grant", h.handleGrant)
	mux.HandleFunc("/api/v1/creator/tasks", h.handleTasks)
	mux.HandleFunc("/api/v1/creator/tasks/progress", h.handleTaskProgress)
}

func (h *Handler) handleCoupons(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	list, err := h.svc.Coupons(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	httputil.WriteOK(w, list)
}

func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	bal, err := h.svc.Balance(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	httputil.WriteOK(w, map[string]any{"balance": bal})
}

func (h *Handler) handleRedeem(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	if uid == 0 {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		CouponID     int64 `json:"coupon_id"`
		ManuscriptID int64 `json:"manuscript_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	amount, err := h.svc.Redeem(r.Context(), uid, req.CouponID, req.ManuscriptID)
	if err != nil {
		myerr.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]any{"amount": amount})
}

func (h *Handler) handleGrant(w http.ResponseWriter, r *http.Request) {
	if h.perm == nil {
		httputil.WriteError(w, http.StatusForbidden, "权限校验未配置")
		return
	}
	if _, ok := h.perm.CheckPermission(r, "statistics:manage"); !ok {
		httputil.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		UserID int64  `json:"user_id"`
		Amount int    `json:"amount"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	id, err := h.svc.Grant(r.Context(), req.UserID, req.Amount, req.Reason)
	if err != nil {
		myerr.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]any{"id": id})
}

func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	uid := httputil.GetUserIDFromHeader(r)
	list, err := h.svc.Tasks(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database error")
		return
	}
	httputil.WriteOK(w, list)
}

func (h *Handler) handleTaskProgress(w http.ResponseWriter, r *http.Request) {
	uid, ok := httputil.RequireUser(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		TaskID int64 `json:"task_id"`
		Delta  int   `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	reward, err := h.svc.RecordTaskProgress(r.Context(), uid, req.TaskID, req.Delta)
	if err != nil {
		myerr.WriteHTTPError(w, err)
		return
	}
	httputil.WriteOK(w, map[string]any{"reward_coupon": reward})
}
