package user

import (
	"context"
	"database/sql"
	"math"
)

// LevelThreshold 返回从 level 升级到 level+1 所需的累计经验值。
// 公式 floor(100 * level^1.8)，与前端 calculateMaxExperience 保持一致。
func LevelThreshold(level int32) int64 {
	if level < 1 {
		level = 1
	}
	return int64(math.Floor(100 * math.Pow(float64(level), 1.8)))
}

// AwardExperience 给用户增加经验值，并在跨过阈值时自动升级、结转余量。
// 返回是否升级及新等级。
func AwardExperience(ctx context.Context, db *sql.DB, userID int64, amount int32) (bool, int32) {
	if db == nil || userID == 0 || amount <= 0 {
		return false, 0
	}
	var level int32
	var exp int64
	err := db.QueryRowContext(ctx, `SELECT level, experience FROM users WHERE id = $1`, userID).Scan(&level, &exp)
	if err != nil {
		return false, 0
	}
	oldLevel := level
	exp += int64(amount)
	for exp >= LevelThreshold(level) {
		exp -= LevelThreshold(level)
		level++
	}
	_, err = db.ExecContext(ctx,
		`UPDATE users SET experience = $1, level = $2, updated_at = NOW() WHERE id = $3`,
		exp, level, userID)
	if err != nil {
		return false, oldLevel
	}
	return level > oldLevel, level
}
