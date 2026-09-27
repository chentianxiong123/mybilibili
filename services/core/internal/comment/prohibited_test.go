package comment

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStoreDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// ==================== 违禁词匹配 ====================

func TestProhibitedMatch_Contains(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "广告", MatchType: matchTypeContains}}
	assert.Equal(t, "广告", s.match("这是广告啦"))
	assert.Equal(t, "", s.match("正常内容"))
}

func TestProhibitedMatch_ContainsCaseInsensitive(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "ABC", MatchType: matchTypeContains}}
	assert.Equal(t, "ABC", s.match("这里有abc东西"))
	assert.Equal(t, "", s.match("这里没有"))
}

// 原实现无视 match_type，EXACT 的词被当包含匹配，导致"这傻子真逗"被误伤。
func TestProhibitedMatch_ExactOnlyMatchesWholeContent(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "傻子", MatchType: matchTypeExact}}

	assert.Equal(t, "傻子", s.match("傻子"))
	// EXACT 语义下包含不算命中
	assert.Equal(t, "", s.match("这傻子真逗"))
	assert.Equal(t, "", s.match("傻子吧"))
}

func TestProhibitedMatch_ExactIgnoresSurroundingSpaceAndCase(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "傻子", MatchType: matchTypeExact}}
	assert.Equal(t, "傻子", s.match("  傻子  "))
	assert.Equal(t, "", s.match("傻子啊"))
}

func TestProhibitedMatch_UnknownMatchTypeFallsBackToContains(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	// 历史数据 match_type 可能为空或脏值
	s.words = []ProhibitedWord{{Word: "广告", MatchType: ""}, {Word: "赌博", MatchType: "WHATEVER"}}
	assert.Equal(t, "广告", s.match("发广告了"))
	assert.Equal(t, "赌博", s.match("搞赌博"))
}

// ==================== 缓存加载 ====================

// 关键回归：查询必须带 is_enabled 过滤，否则后台"禁用"某个词完全不起作用。
func TestProhibitedRefresh_FiltersDisabled(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Minute)

	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).
			AddRow("广告", "CONTAINS").
			AddRow("傻子", "EXACT"))

	require.NoError(t, s.refresh(context.Background()))
	assert.Equal(t, 2, s.size())
	assert.Equal(t, "广告", s.match("这是广告"))
	assert.Equal(t, "", s.match("这傻子真逗"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProhibitedRefresh_EmptyMatchTypeNormalized(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Minute)

	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", nil))

	require.NoError(t, s.refresh(context.Background()))
	assert.Equal(t, matchTypeContains, s.snapshot()[0].MatchType)
}

func TestProhibitedRefresh_NilMatchType(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Minute)
	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("x", nil))
	require.NoError(t, s.refresh(context.Background()))
	assert.Equal(t, "x", s.match("axb"))
}

// 加载失败必须保留旧缓存，否则线上一次抖动就会让所有违禁词失效。
func TestProhibitedRefresh_FailureKeepsOldCache(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Minute)

	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))
	require.NoError(t, s.refresh(context.Background()))

	mock.ExpectQuery(`WHERE is_enabled = 1`).WillReturnError(errors.New("db down"))
	assert.Error(t, s.refresh(context.Background()))

	// 旧词仍然生效
	assert.Equal(t, "广告", s.match("这是广告"))
	assert.Equal(t, 1, s.size())
}

func TestProhibitedRefresh_ScanErrorKeepsOldCache(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Minute)
	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))
	require.NoError(t, s.refresh(context.Background()))

	// 列数不对 → Scan 报错
	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word"}).AddRow("新词"))
	assert.Error(t, s.refresh(context.Background()))
	assert.Equal(t, "广告", s.match("这是广告"))
}

func TestProhibitedRefresh_NilDBIsNoop(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	require.NoError(t, s.refresh(context.Background()))
	assert.Equal(t, 0, s.size())
}

// ==================== 匹配边界 ====================

func TestProhibitedMatch_EmptyContentNeverMatches(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "", MatchType: matchTypeContains}, {Word: "广告", MatchType: matchTypeContains}}
	assert.Equal(t, "", s.match(""))
	assert.Equal(t, "", s.match("   "))
}

func TestProhibitedMatch_BlankWordSkipped(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	s.words = []ProhibitedWord{{Word: "   ", MatchType: matchTypeContains}}
	assert.Equal(t, "", s.match("任何内容"))
}

func TestProhibitedMatch_EmptyCacheAllowsEverything(t *testing.T) {
	s := newProhibitedWordStore(nil, time.Minute)
	assert.Equal(t, "", s.match("随便发什么"))
}

// ==================== TTL ====================

func TestProhibitedStore_SetTTL(t *testing.T) {
	s := newProhibitedWordStore(nil, 0)
	assert.Equal(t, 5*time.Minute, s.currentTTL())

	s.setTTL(90 * time.Second)
	assert.Equal(t, 90*time.Second, s.currentTTL())

	// 非法值退回默认，不接受 0/负数
	s.setTTL(0)
	assert.Equal(t, 5*time.Minute, s.currentTTL())
	s.setTTL(-1)
	assert.Equal(t, 5*time.Minute, s.currentTTL())
}

func TestProhibitedStore_StartRefresher_StopsOnContextCancel(t *testing.T) {
	db, mock := newStoreDB(t)
	ctx, cancel := context.WithCancel(context.Background())

	mock.ExpectQuery(`WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))

	s := newProhibitedWordStore(db, 10*time.Millisecond)
	s.StartRefresher(ctx, func() time.Duration { return 10 * time.Millisecond })

	// 首次加载完成后取消，协程应退出
	assert.Eventually(t, func() bool { return s.size() == 1 }, time.Second, 5*time.Millisecond)
	cancel()
}

// ==================== 安全设置读取 ====================

func TestSecurityConfig_DefaultsWhenNoRow(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`SELECT config_value FROM system_configs WHERE config_key=\$1`).
		WillReturnError(sql.ErrNoRows)

	p := newSecurityConfigProvider(db)
	got := p.get(context.Background())
	assert.Equal(t, defaultCommentMaxCount, got.CommentMaxCount)
	assert.Equal(t, defaultCommentWindowSecs, got.CommentWindowSeconds)
	assert.Equal(t, defaultReplyMaxCount, got.ReplyMaxCount)
	assert.Equal(t, defaultReplyWindowSecs, got.ReplyWindowSeconds)
	assert.Equal(t, defaultCacheRefreshSecs, got.CacheRefreshIntervalSeconds)
}

func TestSecurityConfig_ReadsStoredValues(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).
			AddRow(`{"commentMaxCount":3,"commentWindowSeconds":60,"replyMaxCount":5,"replyWindowSeconds":90,"cacheRefreshIntervalSeconds":120}`))

	p := newSecurityConfigProvider(db)
	got := p.get(context.Background())
	assert.Equal(t, 3, got.CommentMaxCount)
	assert.Equal(t, 60, got.CommentWindowSeconds)
	assert.Equal(t, 5, got.ReplyMaxCount)
	assert.Equal(t, 90, got.ReplyWindowSeconds)
	assert.Equal(t, 120, got.CacheRefreshIntervalSeconds)
}

// 只存了部分字段时，其余用默认值补齐。
func TestSecurityConfig_PartialStoredFillsDefaults(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":7}`))

	p := newSecurityConfigProvider(db)
	got := p.get(context.Background())
	assert.Equal(t, 7, got.CommentMaxCount)
	assert.Equal(t, defaultCommentWindowSecs, got.CommentWindowSeconds)
	assert.Equal(t, defaultReplyMaxCount, got.ReplyMaxCount)
}

// 非法 JSON / 非法数值都不能让评论发不出去，一律回默认。
func TestSecurityConfig_InvalidInputsFallBackToDefaults(t *testing.T) {
	for _, raw := range []string{
		`not-json`,
		`{}`,
		`{"commentMaxCount":0,"commentWindowSeconds":-5,"replyMaxCount":null}`,
		``,
	} {
		db, mock := newStoreDB(t)
		mock.ExpectQuery(`system_configs`).
			WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(raw))

		p := newSecurityConfigProvider(db)
		got := p.get(context.Background())
		assert.Equal(t, defaultCommentMaxCount, got.CommentMaxCount, raw)
		assert.Equal(t, defaultCommentWindowSecs, got.CommentWindowSeconds, raw)
		assert.Equal(t, defaultReplyMaxCount, got.ReplyMaxCount, raw)
	}
}

func TestSecurityConfig_CachesWithinTTL(t *testing.T) {
	db, mock := newStoreDB(t)
	// 只允许查一次：第二次 get 应命中 TTL 缓存
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":11}`))

	p := newSecurityConfigProvider(db)
	assert.Equal(t, 11, p.get(context.Background()).CommentMaxCount)
	assert.Equal(t, 11, p.get(context.Background()).CommentMaxCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecurityConfig_InvalidateForcesReload(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":11}`))
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":22}`))

	p := newSecurityConfigProvider(db)
	assert.Equal(t, 11, p.get(context.Background()).CommentMaxCount)
	p.invalidate()
	assert.Equal(t, 22, p.get(context.Background()).CommentMaxCount)
}

func TestSecurityConfig_NilDBReturnsDefaults(t *testing.T) {
	p := newSecurityConfigProvider(nil)
	got := p.get(context.Background())
	assert.Equal(t, defaultCommentMaxCount, got.CommentMaxCount)
}

// ==================== 配置下发到限流桶 ====================

// 后台改配置后，两个桶都要按新阈值工作。
func TestApply_ReconfiguresBothLimiters(t *testing.T) {
	svc := NewCommentService(nil)
	now := time.Now()

	svc.apply(SecuritySettings{
		CommentMaxCount: 2, CommentWindowSeconds: 60,
		ReplyMaxCount: 5, ReplyWindowSeconds: 120,
		CacheRefreshIntervalSeconds: 45,
	})

	assert.False(t, svc.commentLimiter.record(1, now))
	assert.False(t, svc.commentLimiter.record(1, now))
	// 第 3 次超限（上限 2）
	assert.True(t, svc.commentLimiter.record(1, now))

	// 回复桶独立，上限 5
	for i := 0; i < 5; i++ {
		assert.False(t, svc.replyLimiter.record(1, now), i)
	}
	assert.True(t, svc.replyLimiter.record(1, now))
}

func TestApply_UpdatesCacheTTL(t *testing.T) {
	svc := NewCommentService(nil)
	svc.prohibited = newProhibitedWordStore(nil, 5*time.Minute)

	svc.apply(SecuritySettings{CacheRefreshIntervalSeconds: 30})
	assert.Equal(t, 30*time.Second, svc.prohibited.currentTTL())
}

// 记录/查询时窗口内无数据要删掉 map key，否则 counts 随用户数无限增长。
func TestCommentRateLimiter_NoLeakOnExpiredEntries(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 5)
	now := time.Now()

	require.False(t, l.record(1, now))
	l.mu.Lock()
	assert.Len(t, l.counts, 1)
	l.mu.Unlock()

	// 窗口过后再访问，应清掉
	assert.Equal(t, 5, l.remaining(1, now.Add(2*time.Minute)))
	l.mu.Lock()
	assert.Empty(t, l.counts)
	l.mu.Unlock()

	// record 路径同样清理：2 分钟后再记一次，旧记录应被丢弃而不是累加
	require.False(t, l.record(2, now))
	require.False(t, l.record(2, now.Add(2*time.Minute)))
	l.mu.Lock()
	assert.Len(t, l.counts[2], 1, "过期记录应被丢弃，只保留新记录")
	l.mu.Unlock()
}
