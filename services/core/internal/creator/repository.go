package creator

import (
	"context"
	"database/sql"
	"time"
)

// Coupon 流量券：平台发给创作者的一张曝光资源。
//
//	status 0 未使用 / 1 已使用 / -1 已过期作废
type Coupon struct {
	ID        int64
	UserID    int64
	Amount    int
	Reason    string
	TaskID    int64
	Status    int
	GrantedAt time.Time
	UsedAt    sql.NullTime
	ExpiresAt sql.NullTime
}

// CouponUse 一次券的使用记录（把券的曝光点数投到某个稿件）。
type CouponUse struct {
	ID           int64
	CouponID     int64
	UserID       int64
	ManuscriptID int64
	Amount       int
	CreatedAt    time.Time
}

// Task 任务定义。
//
//	period once / daily / weekly / total
type Task struct {
	ID           int64
	Code         string
	Name         string
	Description  string
	RewardCoupon int
	Target       int
	Period       string
	Enabled      bool
	CreatedAt    time.Time
}

// TaskProgress 用户在某任务上的进度。
type TaskProgress struct {
	ID            int64
	UserID        int64
	TaskID        int64
	Progress      int
	Done          bool
	RewardClaimed bool
	PeriodStart   time.Time
	UpdatedAt     time.Time
}

// Repo 数据访问层。
type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) DB() *sql.DB { return r.db }

// ---- 流量券 ----

// GrantCoupon 给创作者发一张券。
func (r *Repo) GrantCoupon(ctx context.Context, c *Coupon) (int64, error) {
	return grantCouponTx(ctx, r.db, c)
}

// grantCouponTx 在给定事务/连接里插入一张券（FINDING-003 修复：与标记
// reward_claimed 同事务提交，消除"已标记已领但券未发"的丢失/重复发奖窗口）。
func grantCouponTx(ctx context.Context, exec interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, c *Coupon) (int64, error) {
	var id int64
	err := exec.QueryRowContext(ctx, `
		INSERT INTO traffic_coupons (user_id, amount, reason, task_id, status, expires_at)
		VALUES ($1, $2, $3, $4, 0, $5)
		RETURNING id`, c.UserID, c.Amount, c.Reason, c.TaskID, c.ExpiresAt).Scan(&id)
	return id, err
}

// ListCoupons 列出用户所有券（含状态），按发放时间倒序。
func (r *Repo) ListCoupons(ctx context.Context, userID int64) ([]Coupon, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, amount, reason, task_id, status, granted_at, used_at, expires_at
		FROM traffic_coupons WHERE user_id = $1
		ORDER BY granted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Coupon
	for rows.Next() {
		var c Coupon
		if err := rows.Scan(&c.ID, &c.UserID, &c.Amount, &c.Reason, &c.TaskID, &c.Status,
			&c.GrantedAt, &c.UsedAt, &c.ExpiresAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// UsableCouponAmount 用户当前可用的券点数（未使用且未过期）。
func (r *Repo) UsableCouponAmount(ctx context.Context, userID int64) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM traffic_coupons
		WHERE user_id = $1 AND status = 0
		  AND (expires_at IS NULL OR expires_at > NOW())`, userID).Scan(&total)
	return total, err
}

// RedeemCoupon 把一张未使用的券投到稿件上，返回该券面值。
// 事务：标记券已用 + 写使用记录 + 累加稿件曝光权重。
func (r *Repo) RedeemCoupon(ctx context.Context, couponID, userID, manuscriptID int64) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var amount int
	var owner int64
	var status int
	var expiresAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT amount, user_id, status, expires_at FROM traffic_coupons WHERE id = $1`,
		couponID).Scan(&amount, &owner, &status, &expiresAt)
	if err != nil {
		return 0, err
	}
	if owner != userID {
		return 0, sql.ErrNoRows
	}
	if status != 0 {
		return 0, errCouponUsed
	}
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return 0, errCouponExpired
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE traffic_coupons SET status = 1, used_at = NOW() WHERE id = $1`, couponID); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO traffic_coupon_uses (coupon_id, user_id, manuscript_id, amount)
		VALUES ($1, $2, $3, $4)`, couponID, userID, manuscriptID, amount); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE manuscripts SET exposure_weight = exposure_weight + $1 WHERE id = $2`,
		amount, manuscriptID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return amount, nil
}

// ListCouponUses 用户在指定稿件上的券使用记录。
func (r *Repo) ListCouponUses(ctx context.Context, userID int64) ([]CouponUse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, coupon_id, user_id, manuscript_id, amount, created_at
		FROM traffic_coupon_uses WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CouponUse
	for rows.Next() {
		var u CouponUse
		if err := rows.Scan(&u.ID, &u.CouponID, &u.UserID, &u.ManuscriptID, &u.Amount, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// ---- 任务 ----

// ListTasks 列出启用的任务。
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, description, reward_coupon, target, period, enabled, created_at
		FROM creator_tasks WHERE enabled = true ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.RewardCoupon,
			&t.Target, &t.Period, &t.Enabled, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// GetTaskProgress 取用户某任务进度；不存在返回 nil。
func (r *Repo) GetTaskProgress(ctx context.Context, userID, taskID int64) (*TaskProgress, error) {
	var p TaskProgress
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, task_id, progress, done, reward_claimed, period_start, updated_at
		FROM creator_task_progress WHERE user_id = $1 AND task_id = $2`,
		userID, taskID).Scan(&p.ID, &p.UserID, &p.TaskID, &p.Progress, &p.Done,
		&p.RewardClaimed, &p.PeriodStart, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpsertTaskProgress 记一次进度。若本次达标且未领奖，返回 (rewardToGrant, true)。
// reward 非 nil 时，达标发奖与标记 reward_claimed 在同一事务内完成（FINDING-003 修复）。
func (r *Repo) UpsertTaskProgress(ctx context.Context, userID, taskID int64, delta, target int, reward *Coupon) (int, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	var current int
	var done bool
	err = tx.QueryRowContext(ctx, `
		SELECT progress, done FROM creator_task_progress WHERE user_id = $1 AND task_id = $2`,
		userID, taskID).Scan(&current, &done)
	if err == sql.ErrNoRows {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO creator_task_progress (user_id, task_id, progress, done, reward_claimed)
			VALUES ($1, $2, $3, $4, false)`, userID, taskID, delta, delta >= target); err != nil {
			return 0, false, err
		}
		current = delta
		done = delta >= target
	} else if err != nil {
		return 0, false, err
	} else {
		// 已有记录：若任务已达过标（done=true 且已领奖）则不再累加，避免重复刷。
		var claimed bool
		_ = tx.QueryRowContext(ctx, `
			SELECT reward_claimed FROM creator_task_progress WHERE user_id = $1 AND task_id = $2`,
			userID, taskID).Scan(&claimed)
		if done && claimed {
			return 0, false, tx.Commit()
		}
		current += delta
		if current > target {
			current = target
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE creator_task_progress
			SET progress = $3, done = ($3 >= $4), updated_at = NOW()
			WHERE user_id = $1 AND task_id = $2`, userID, taskID, current, target); err != nil {
			return 0, false, err
		}
		done = current >= target
	}

	// 达标且还没发过奖 → 标记已领 + 同事务发券，返回应发券数
	if done {
		var claimed bool
		_ = tx.QueryRowContext(ctx, `
			SELECT reward_claimed FROM creator_task_progress WHERE user_id = $1 AND task_id = $2`,
			userID, taskID).Scan(&claimed)
		if !claimed {
			if _, err := tx.ExecContext(ctx, `
				UPDATE creator_task_progress SET reward_claimed = true, updated_at = NOW()
				WHERE user_id = $1 AND task_id = $2`, userID, taskID); err != nil {
				return 0, false, err
			}
			// FINDING-003 修复：发券与标记 reward_claimed 同事务，消除中间态
			// （此前标记已领后 GrantCoupon 在独立事务，失败则券永久丢失）。
			if reward != nil && reward.Amount > 0 {
				if _, err := grantCouponTx(ctx, tx, reward); err != nil {
					return 0, false, err
				}
			}
			if err := tx.Commit(); err != nil {
				return 0, false, err
			}
			if reward != nil && reward.Amount > 0 {
				return reward.Amount, true, nil
			}
			return target, true, nil
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, false, err
	}
	return 0, false, nil
}

// GetTaskByID 取任务（含未启用，用于回填 reason）。
func (r *Repo) GetTaskByID(ctx context.Context, taskID int64) (*Task, error) {
	var t Task
	err := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, description, reward_coupon, target, period, enabled, created_at
		FROM creator_tasks WHERE id = $1`, taskID).Scan(&t.ID, &t.Code, &t.Name,
		&t.Description, &t.RewardCoupon, &t.Target, &t.Period, &t.Enabled, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
