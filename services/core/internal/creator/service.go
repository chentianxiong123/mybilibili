package creator

import (
	"context"
	"database/sql"
	"errors"
	"time"

	myerr "mybilibili/pkg/errors"
)

var (
	errCouponUsed    = errors.New("coupon already used")
	errCouponExpired = errors.New("coupon expired")
	errCouponManuscriptNotFound = errors.New("manuscript not found or not owned by user")
)

// Service 创作者流量/激励业务逻辑。
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// ---- 流量券 ----

// Coupon 视图（对外 JSON）。
type CouponView struct {
	ID        int64      `json:"id"`
	Amount    int        `json:"amount"`
	Reason    string     `json:"reason"`
	TaskID    int64      `json:"task_id"`
	Status    int        `json:"status"`
	GrantedAt time.Time  `json:"granted_at"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Grant 平台给创作者发一张券（admin 操作，amount>0）。
func (s *Service) Grant(ctx context.Context, userID int64, amount int, reason string) (int64, error) {
	if amount <= 0 {
		return 0, myerr.ErrInvalidArgument("coupon amount must be positive")
	}
	if userID <= 0 {
		return 0, myerr.ErrInvalidArgument("invalid user")
	}
	if reason == "" {
		reason = "平台发放"
	}
	return s.repo.GrantCoupon(ctx, &Coupon{
		UserID: userID,
		Amount: amount,
		Reason: reason,
	})
}

// Balance 返回用户可用的券点数。
func (s *Service) Balance(ctx context.Context, userID int64) (int, error) {
	return s.repo.UsableCouponAmount(ctx, userID)
}

// Coupons 返回用户所有券。
func (s *Service) Coupons(ctx context.Context, userID int64) ([]CouponView, error) {
	list, err := s.repo.ListCoupons(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]CouponView, 0, len(list))
	for _, c := range list {
		v := CouponView{
			ID: c.ID, Amount: c.Amount, Reason: c.Reason, TaskID: c.TaskID,
			Status: c.Status, GrantedAt: c.GrantedAt,
		}
		if c.UsedAt.Valid {
			t := c.UsedAt.Time
			v.UsedAt = &t
		}
		if c.ExpiresAt.Valid {
			t := c.ExpiresAt.Time
			v.ExpiresAt = &t
		}
		out = append(out, v)
	}
	return out, nil
}

// Redeem 把用户的一张券投到自己的稿件上，返回消耗的曝光点数。
func (s *Service) Redeem(ctx context.Context, userID, couponID, manuscriptID int64) (int, error) {
	if userID <= 0 {
		return 0, myerr.ErrUnauthenticated("unauthorized")
	}
	if couponID <= 0 || manuscriptID <= 0 {
		return 0, myerr.ErrInvalidArgument("invalid coupon or manuscript")
	}
	amount, err := s.repo.RedeemCoupon(ctx, couponID, userID, manuscriptID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, myerr.ErrNotFound("coupon not found or not yours")
		case errors.Is(err, errCouponUsed):
			return 0, myerr.ErrInvalidArgument("coupon already used")
		case errors.Is(err, errCouponExpired):
			return 0, myerr.ErrInvalidArgument("coupon expired")
		case errors.Is(err, errCouponManuscriptNotFound):
			// FINDING-059 修复：稿件不存在或非本人所有，不消耗 coupon。
			return 0, myerr.ErrNotFound("manuscript not found or not owned by you")
		default:
			return 0, myerr.ErrInternal("database error")
		}
	}
	return amount, nil
}

// Uses 返回用户投券记录。
func (s *Service) Uses(ctx context.Context, userID int64) ([]CouponUse, error) {
	return s.repo.ListCouponUses(ctx, userID)
}

// ---- 任务 ----

// TaskView 任务视图（含用户进度）。
type TaskView struct {
	ID            int64  `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	RewardCoupon  int    `json:"reward_coupon"`
	Target        int    `json:"target"`
	Period        string `json:"period"`
	Progress      int    `json:"progress"`
	Done          bool   `json:"done"`
	RewardClaimed bool   `json:"reward_claimed"`
}

// Tasks 返回用户视角的任务列表（带进度）。
func (s *Service) Tasks(ctx context.Context, userID int64) ([]TaskView, error) {
	list, err := s.repo.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TaskView, 0, len(list))
	for _, t := range list {
		v := TaskView{
			ID: t.ID, Code: t.Code, Name: t.Name, Description: t.Description,
			RewardCoupon: t.RewardCoupon, Target: t.Target, Period: t.Period,
		}
		if userID > 0 {
			if p, perr := s.repo.GetTaskProgress(ctx, userID, t.ID); perr == nil && p != nil {
				v.Progress = p.Progress
				v.Done = p.Done
				v.RewardClaimed = p.RewardClaimed
			}
		}
		out = append(out, v)
	}
	return out, nil
}

// RecordTaskProgress 上报一次任务进度（如投稿+1）。达标且未领奖时自动发流量券。
func (s *Service) RecordTaskProgress(ctx context.Context, userID, taskID int64, delta int) (reward int, err error) {
	if userID <= 0 {
		return 0, myerr.ErrUnauthenticated("unauthorized")
	}
	if delta <= 0 {
		delta = 1
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, myerr.ErrNotFound("task not found")
		}
		return 0, myerr.ErrInternal("database error")
	}
	if !task.Enabled {
		return 0, myerr.ErrInvalidArgument("task disabled")
	}
	var rewardCoupon *Coupon
	if task.RewardCoupon > 0 {
		rewardCoupon = &Coupon{UserID: userID, Amount: task.RewardCoupon, Reason: "任务奖励", TaskID: task.ID}
	}
	rewardAmount, done, err := s.repo.UpsertTaskProgress(ctx, userID, taskID, delta, task.Target, rewardCoupon)
	if err != nil {
		return 0, myerr.ErrInternal("database error")
	}
	if done && rewardAmount > 0 {
		// FINDING-003 修复：券已在 UpsertTaskProgress 同事务内发放，无需再 GrantCoupon。
		return rewardAmount, nil
	}
	_ = rewardAmount
	return 0, nil
}
