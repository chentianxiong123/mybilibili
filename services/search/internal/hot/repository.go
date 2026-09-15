package hot

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	rankKey       = "hot_search:rank"
	detailPrefix  = "hot_search:detail:%s"
	baseScore     = 10.0
	timeDecay     = 0.1
	expireDays    = 30
	maxKeywords   = 100
)

type Repository struct {
	rdb *redis.Client
}

func NewRepository(rdb *redis.Client) *Repository {
	return &Repository{rdb: rdb}
}

func scoreIncrement() float64 {
	hours := time.Now().Unix() / 3600
	return baseScore / (1.0 + timeDecay*math.Log1p(float64(hours%10000)))
}

// Increment 搜索时累加关键词热度
func (r *Repository) Increment(ctx context.Context, keyword string) error {
	now := time.Now().UnixMilli()
	score := scoreIncrement()
	pipe := r.rdb.TxPipeline()
	pipe.ZIncrBy(ctx, rankKey, score, keyword)
	detailKey := r.detailKey(keyword)
	pipe.HIncrBy(ctx, detailKey, "count", 1)
	pipe.HSet(ctx, detailKey, "lastSearchTime", strconv.FormatInt(now, 10))
	pipe.HSetNX(ctx, detailKey, "firstSearchTime", strconv.FormatInt(now, 10))
	pipe.Expire(ctx, rankKey, expireDays*24*time.Hour)
	pipe.Expire(ctx, detailKey, expireDays*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// Top 返回热度最高的前 n 个关键词
func (r *Repository) Top(ctx context.Context, n int64) ([]map[string]interface{}, error) {
	res, err := r.rdb.ZRevRangeWithScores(ctx, rankKey, 0, n-1).Result()
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(res))
	for i, z := range res {
		keyword := z.Member.(string)
		list = append(list, map[string]interface{}{
			"rank":    i + 1,
			"keyword": keyword,
			"score":   int64(z.Score),
		})
	}
	return list, nil
}

// UpdateScore 管理员手动设置某关键词热度分
func (r *Repository) UpdateScore(ctx context.Context, keyword string, newScore float64) error {
	pipe := r.rdb.TxPipeline()
	pipe.ZAdd(ctx, rankKey, redis.Z{Score: newScore, Member: keyword})
	pipe.Expire(ctx, rankKey, expireDays*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// SetRank 管理员设置排名（通过调整分数实现）
func (r *Repository) SetRank(ctx context.Context, keyword string, rank int64) error {
	return r.UpdateScore(ctx, keyword, float64(1000000-rank*100))
}

// Get 查询单个关键词热度
func (r *Repository) Get(ctx context.Context, keyword string) (map[string]interface{}, error) {
	score, err := r.rdb.ZScore(ctx, rankKey, keyword).Result()
	if err != nil {
		return nil, err
	}
	detailKey := r.detailKey(keyword)
	detail, err := r.rdb.HGetAll(ctx, detailKey).Result()
	if err != nil {
		detail = nil
	}
	result := map[string]interface{}{
		"keyword": keyword,
		"score":   int64(score),
	}
	if detail != nil {
		if v, ok := detail["count"]; ok {
			result["search_count"], _ = strconv.ParseInt(v, 10, 64)
		}
		if v, ok := detail["firstSearchTime"]; ok {
			result["first_search_time"] = v
		}
		if v, ok := detail["lastSearchTime"]; ok {
			result["last_search_time"] = v
		}
	}
	return result, nil
}

// Delete 删除关键词
func (r *Repository) Delete(ctx context.Context, keyword string) error {
	pipe := r.rdb.TxPipeline()
	pipe.ZRem(ctx, rankKey, keyword)
	pipe.Del(ctx, r.detailKey(keyword))
	_, err := pipe.Exec(ctx)
	return err
}

// CleanExpired 裁剪超过 maxKeywords 条之后的数据，按分数保留靠前的
func (r *Repository) CleanExpired(ctx context.Context, keep int64) error {
	total, err := r.rdb.ZCard(ctx, rankKey).Result()
	if err != nil {
		return err
	}
	if total <= keep {
		return nil
	}
	res, err := r.rdb.ZRevRange(ctx, rankKey, keep, -1).Result()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return nil
	}
	pipe := r.rdb.TxPipeline()
	members := make([]interface{}, 0, len(res))
	for _, m := range res {
		members = append(members, m)
		pipe.Del(ctx, r.detailKey(m))
	}
	pipe.ZRem(ctx, rankKey, members...)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *Repository) detailKey(keyword string) string {
	return "hot_search:detail:" + keyword
}

func ScoreString(s float64) string {
	return strconv.FormatFloat(s, 'f', 0, 64)
}