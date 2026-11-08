package comment

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "mybilibili/pkg/pb"
)

// ==================== 纯函数：parseMentionNames / truncateRunes ====================

func TestParseMentionNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "@alice", []string{"alice"}},
		{"two", "@alice @bob", []string{"alice", "bob"}},
		{"dedup", "@alice @alice", []string{"alice"}},
		{"trailingPunct", "@alice，", []string{"alice"}},
		{"chinesePunctSep", "@a。@b", []string{"a", "b"}},
		{"doubleAt", "@@", nil},
		{"inline", "hi @name world", []string{"name"}},
		{"unicode", "@张三 你好", []string{"张三"}},
		{"newlineSep", "@a\n@b", []string{"a", "b"}},
		{"tabSep", "@a\t@b", []string{"a", "b"}},
		{"commaJoin", "@a,@b", []string{"a", "b"}},
		{"atEnd", "hi @", nil},
		{"atSpace", "@ ", nil},
		{"parenPunct", "@a）", []string{"a"}},
		{"onlyAt", "@", nil},
		{"nameWithTrailingLatinPunct", "@a,", []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMentionNames(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	assert.Equal(t, "abc", truncateRunes("abc", 5))
	assert.Equal(t, "abc", truncateRunes("abc", 3))
	assert.Equal(t, "ab", truncateRunes("abc", 2))
	assert.Equal(t, "你好", truncateRunes("你好世界", 2))
	assert.Equal(t, "你好", truncateRunes("你好", 5))
	assert.Equal(t, "", truncateRunes("", 3))
	assert.Equal(t, "", truncateRunes("", 0))
}

// ==================== service 错误路径 ====================

func TestListComments_DBError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`FROM comments WHERE manuscript_id = \$1`).WillReturnError(sql.ErrConnDone)
	_, err := svc.ListComments(context.Background(), &pb.ListCommentsRequest{ManuscriptId: 9, Page: 1, PageSize: 20})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReplies_DBError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1`).WillReturnError(sql.ErrConnDone)
	_, err := svc.GetReplies(context.Background(), &pb.GetRepliesRequest{CommentId: 9, Page: 1, PageSize: 20})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_InternalError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WillReturnError(sql.ErrConnDone)
	_, err := svc.LikeComment(context.Background(), &pb.LikeCommentRequest{CommentId: 9, UserId: 1})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnlikeComment_Error(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(sql.ErrConnDone)
	_, err := svc.UnlikeComment(context.Background(), &pb.UnlikeCommentRequest{CommentId: 9, UserId: 1})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeReply_InternalError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WillReturnError(sql.ErrConnDone)
	_, err := svc.LikeReply(context.Background(), &pb.LikeReplyRequest{ReplyId: 9, UserId: 1})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== 通知内部函数分支 ====================

func TestSendCommentLikeNotification_ReplyFindError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`FROM replies WHERE id = \$1`).WithArgs(int64(9)).WillReturnError(sql.ErrNoRows)
	svc.sendCommentLikeNotification(context.Background(), "reply", 9, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCommentLikeNotification_ReplyOwnerIsSender(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`FROM replies WHERE id = \$1`).WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at"}).
			AddRow(9, 3, 1, nil, "r", 0, "NORMAL", time.Now(), time.Now()))
	svc.sendCommentLikeNotification(context.Background(), "reply", 9, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCommentLikeNotification_CommentOwnerZero(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(9, 7, 0, "c", 0, 0, 0, time.Now(), time.Now()))
	svc.sendCommentLikeNotification(context.Background(), "comment", 9, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSendReplyNotification_FindError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id, manuscript_id, user_id, content`).WithArgs(int64(9)).WillReturnError(sql.ErrNoRows)
	svc.sendReplyNotification(context.Background(), 9, 1, "hi", 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== notifyMentions ====================

func TestNotifyMentions_NoDB(t *testing.T) {
	svc := NewCommentService(NewCommentRepository(nil))
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2) // db nil → no-op
}

func TestNotifyMentions_NoNotifier(t *testing.T) {
	svc, _ := newCommentSvc(t)
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2) // notifier nil → no-op
}

func TestNotifyMentions_NoMentions(t *testing.T) {
	svc, _ := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	svc.notifyMentions(context.Background(), "hello world", 1, 5, 2)
}

func TestNotifyMentions_SendsNotification(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname = \$1 OR username = \$1`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyMentions_SelfMentionSkipped(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname = \$1 OR username = \$1`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // 自己@自己
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyMentions_QueryErrorSkipped(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname = \$1 OR username = \$1`).WithArgs("alice").
		WillReturnError(sql.ErrConnDone)
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyMentions_ZeroUIDSkipped(t *testing.T) {
	svc, mock := newCommentSvc(t)
	svc.SetNotifier(&fakeNotifier{})
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname = \$1 OR username = \$1`).WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
	svc.notifyMentions(context.Background(), "@alice hi", 1, 5, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== 安全配置下发 ====================

func TestSetSecurityConfig_NilDB(t *testing.T) {
	svc := NewCommentService(NewCommentRepository(nil))
	svc.SetSecurityConfig(context.Background(), nil) // db nil → no-op, no panic
	assert.Nil(t, svc.securityCfg)
}

func TestSetSecurityConfig_WithDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	svc := NewCommentService(NewCommentRepository(db))
	svc.SetDB(db)

	// securityCfg.get() → 回源读 system_configs
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":4,"commentWindowSeconds":30,"replyMaxCount":6,"replyWindowSeconds":40,"cacheRefreshIntervalSeconds":60}`))
	// StartRefresher 协程的首次 refresh → 读违禁词表
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.SetSecurityConfig(ctx, db)

	require.NotNil(t, svc.securityCfg)
	require.NotNil(t, svc.prohibited)
	// 限流桶已按新配置下发
	assert.Equal(t, 4, svc.commentLimiter.maxCount)
	assert.Equal(t, 6, svc.replyLimiter.maxCount)
	assert.Equal(t, 30*time.Second, svc.commentLimiter.window)
	assert.Equal(t, 60*time.Second, svc.prohibited.currentTTL())

	// 等待协程完成首次 refresh 并加载违禁词
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && svc.prohibited.size() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	assert.Equal(t, 1, svc.prohibited.size())
	cancel()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInvalidateSecurityConfig_NilCfg(t *testing.T) {
	svc, _ := newCommentSvc(t)
	svc.InvalidateSecurityConfig(context.Background()) // securityCfg nil → refresh no-op
}

func TestInvalidateSecurityConfig_WithCfg(t *testing.T) {
	svc, mock := newCommentSvc(t)
	// 先加载并缓存一份配置
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":9}`))
	svc.securityCfg = newSecurityConfigProvider(svc.db)
	_ = svc.securityCfg.get(context.Background())

	// invalidate 强制回源 → 重新读 system_configs 并下发
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":12}`))
	svc.InvalidateSecurityConfig(context.Background())
	assert.Equal(t, 12, svc.commentLimiter.maxCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshSecurityConfig_Nil(t *testing.T) {
	svc, _ := newCommentSvc(t)
	svc.refreshSecurityConfig(context.Background()) // securityCfg nil → no-op
}

func TestRefreshSecurityConfig_Applies(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`system_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(`{"commentMaxCount":5,"commentWindowSeconds":60,"replyMaxCount":7,"replyWindowSeconds":90,"cacheRefreshIntervalSeconds":120}`))
	svc.securityCfg = newSecurityConfigProvider(svc.db)
	svc.refreshSecurityConfig(context.Background())
	assert.Equal(t, 5, svc.commentLimiter.maxCount)
	assert.Equal(t, 7, svc.replyLimiter.maxCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== public API：get-up-like / get-like-and-dislike ====================

func TestHandleGetUpLike_NoUID(t *testing.T) {
	h, _ := newPublicHandler(t)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-up-like", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"code":200,"data":[],"message":"ok"}`, w.Body.String())
}

func TestHandleGetUpLike_Success(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions WHERE user_id = \$1 AND target_type IN \('COMMENT','REPLY'\)`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(int64(1)).AddRow(int64(5)))
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-up-like?uid=42", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]interface{})
	assert.Len(t, data, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetUpLike_QueryError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions WHERE user_id = \$1`).
		WithArgs(int64(42)).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-up-like?uid=42", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetUpLike_ScanErrorSkipsRow(t *testing.T) {
	h, mock := newPublicHandler(t)
	// 第一行是合法 id，第二行无法 scan 成 int64 → 被跳过
	mock.ExpectQuery(`SELECT target_id FROM user_interactions WHERE user_id = \$1`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(int64(1)).AddRow("not-a-number"))
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-up-like?uid=42", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]interface{})
	assert.Len(t, data, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetLikeAndDislike_NoUID(t *testing.T) {
	h, _ := newPublicHandler(t)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-like-and-dislike", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]interface{})
	assert.Equal(t, []interface{}{}, data["userLike"])
	assert.Equal(t, []interface{}{}, data["userDislike"])
}

func TestHandleGetLikeAndDislike_Success(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions WHERE user_id = \$1 AND target_type = 'COMMENT'`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(int64(3)).AddRow(int64(9)))
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-like-and-dislike?uid=7", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]interface{})
	liked := data["userLike"].([]interface{})
	assert.Len(t, liked, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleGetLikeAndDislike_QueryError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT target_id FROM user_interactions WHERE user_id = \$1`).
		WithArgs(int64(7)).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/get-like-and-dislike?uid=7", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBatchLikeCounts_ReplyType(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT id, like_count FROM replies WHERE id IN`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "like_count"}).AddRow(1, 5))
	mock.ExpectQuery(`SELECT target_id FROM user_interactions`).WithArgs(int64(0), "REPLY", int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(1))
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/batch-like-counts?ids=1&type=reply", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== public API：list / add 服务错误 ====================

func TestHandleCommentList_ServiceError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`FROM comments WHERE manuscript_id = \$1`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/list?manuscriptId=9", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCommentAdd_ServiceError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`INSERT INTO comments`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/add", `{"manuscriptId":9,"content":"hello"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDecodeCommentBody_InvalidForm(t *testing.T) {
	r := httptest.NewRequest("POST", "/x", strings.NewReader("%zz"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := decodeCommentBody(r, map[string]*string{"content": new(string)})
	assert.Error(t, err)
}

// ==================== repository 错误路径 ====================

func TestListRepliesByComment_QueryError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1`).WillReturnError(sql.ErrConnDone)
	_, err := svc.Repo().ListRepliesByComment(context.Background(), 9, 1, 20)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteReply_Error(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED'`).WillReturnError(sql.ErrConnDone)
	err := svc.Repo().DeleteReply(context.Background(), 9, 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeTarget_ReplyExistsError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WillReturnError(sql.ErrConnDone)
	err := svc.Repo().LikeTarget(context.Background(), "reply", 9, 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== rate limiter remaining 下限 ====================

func TestCommentRateLimiter_RemainingClampsAtZero(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 2)
	now := time.Now()
	// 直接塞入超过上限的窗口内记录（模拟窗口收紧等边缘情况）
	l.mu.Lock()
	l.counts[1] = []time.Time{now, now, now}
	l.mu.Unlock()
	assert.Equal(t, 0, l.remaining(1, now))
}

// ==================== creator handler 分支 ====================

// newCreatorHandlerConv 与 newCreatorHandler 相同，但给两个 mock db 都配置了
// argConverter（[]int64 → Postgres 数组字面量），因为 findManuscriptsByIDs /
// FindUsersByIDs 会把 []int64 直接作为 ANY($1) 参数传入。
func newCreatorHandlerConv(t *testing.T) (*CreatorCommentHTTPHandler, sqlmock.Sqlmock, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.ValueConverterOption(argConverter{}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	repoDB, repoMock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.ValueConverterOption(argConverter{}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { repoDB.Close() })

	repo := NewCommentRepository(repoDB)
	svc := NewCommentService(repo)
	svc.SetDB(db)
	h := NewCreatorCommentHTTPHandler(repo, svc, db)
	return h, mock, repoMock
}

func TestHandleList_CommentTypeComment(t *testing.T) {
	h, mock, repoMock := newCreatorHandlerConv(t)
	now := time.Now()
	repoMock.ExpectQuery(`SELECT comments\.id[\s\S]*FROM comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow(1, 9, 200, "hello", 3, 1, 0, now, now))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	repoMock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}).AddRow(9, "Title", "cover.jpg"))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest("GET", "/api/v1/creator/comments?manuscriptId=9&commentType=comment", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, repoMock.ExpectationsWereMet())
}

func TestHandleList_CommentTypeReply(t *testing.T) {
	h, mock, repoMock := newCreatorHandlerConv(t)
	now := time.Now()
	repoMock.ExpectQuery(`SELECT r\.[\s\S]*FROM replies r`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}).
			AddRow(1, 5, 200, nil, "r", 1, "NORMAL", now, now, 9))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	repoMock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).AddRow(200, "u", "nick", "a", 1))
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}).AddRow(9, "Title", "cover.jpg"))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest("GET", "/api/v1/creator/comments?commentType=reply", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, repoMock.ExpectationsWereMet())
}

func TestHandleList_QueryError(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectQuery(`SELECT comments\.id[\s\S]*FROM comments`).WillReturnError(sql.ErrConnDone)
	req := httptest.NewRequest("GET", "/api/v1/creator/comments?commentType=comment", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleByPath_DELETE_Error(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM comments WHERE id = \$1 AND manuscript_id IN`).WillReturnError(sql.ErrConnDone)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/123", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleByPath_POST_Success(t *testing.T) {
	h, mock, repoMock := newCreatorHandler(t)
	repoMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(123), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	repoMock.ExpectQuery(`INSERT INTO replies`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))
	repoMock.ExpectExec(`UPDATE comments SET reply_count = reply_count \+ 1`).WithArgs(int64(123)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users`).WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(1, 0))
	mock.ExpectExec(`UPDATE users SET experience`).WithArgs(int64(2), int64(1), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/creator/comments/123?content=hello&replyToUserId=5", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, repoMock.ExpectationsWereMet())
}

func TestHandleDeleteReply_Error(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectExec(`DELETE FROM replies WHERE id = \$1 AND comment_id IN`).WillReturnError(sql.ErrConnDone)
	req := httptest.NewRequest("DELETE", "/api/v1/creator/comments/reply/456", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleDeleteReply(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== findManuscriptsByIDs ====================

func TestFindManuscriptsByIDs_Rows(t *testing.T) {
	h, mock, _ := newCreatorHandlerConv(t)
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}).AddRow(9, "Title", "cover.jpg"))
	out := h.findManuscriptsByIDs(context.Background(), []int64{9})
	assert.Equal(t, "Title", out[9].Title)
	assert.Equal(t, "cover.jpg", out[9].Cover)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindManuscriptsByIDs_NullFields(t *testing.T) {
	h, mock, _ := newCreatorHandlerConv(t)
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}).AddRow(9, nil, nil))
	out := h.findManuscriptsByIDs(context.Background(), []int64{9})
	assert.Equal(t, "", out[9].Title)
	assert.Equal(t, "", out[9].Cover)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFindManuscriptsByIDs_QueryError(t *testing.T) {
	h, mock, _ := newCreatorHandlerConv(t)
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).WillReturnError(sql.ErrConnDone)
	out := h.findManuscriptsByIDs(context.Background(), []int64{9})
	assert.Empty(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== public API：reply / like / replies 错误路径 ====================

func TestHandleCommentReply_InvalidBody(t *testing.T) {
	h, _ := newPublicHandler(t)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/reply", `bad`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCommentReply_ServiceError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`INSERT INTO replies`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/reply", `{"commentId":9,"content":"hi"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCommentLike_PostError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCommentLike_DeleteError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "DELETE", "/api/v1/comment/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCommentReplies_ServiceError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`FROM replies WHERE comment_id = \$1`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "GET", "/api/v1/comment/9/replies", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleReplyLike_Unauthorized(t *testing.T) {
	h, _ := newPublicHandler(t)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/reply/9/like", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleReplyLike_PostError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM replies`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "POST", "/api/v1/comment/reply/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleReplyLike_DeleteError(t *testing.T) {
	h, mock := newPublicHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WillReturnError(sql.ErrConnDone)
	w := doReq(muxForPublic(h), "DELETE", "/api/v1/comment/reply/9/like", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== 序列化 marshal 错误分支 ====================

func TestCommentToMap_MarshalError(t *testing.T) {
	c := &pb.CommentInfo{Id: 1, Content: string([]byte{0xff, 0xfe})}
	m := commentToMap(c)
	assert.NotNil(t, m)
}

func TestReplyListToJSON_MarshalError(t *testing.T) {
	info := &pb.ReplyInfo{Id: 1, Content: string([]byte{0xff, 0xfe})}
	out := replyListToJSON([]*pb.ReplyInfo{info})
	require.Len(t, out, 1)
	assert.NotNil(t, out[0])
}

// ==================== repository 扫描/执行错误 ====================

func TestListByManuscript_ScanError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`FROM comments WHERE manuscript_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow("not-int", 9, 200, "c", 0, 0, 0, time.Now(), time.Now()))
	_, err := svc.Repo().ListByManuscript(context.Background(), 9, 1, 20, "hot")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_ExecError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectExec(`UPDATE comments SET status = 1`).WillReturnError(sql.ErrConnDone)
	err := svc.Repo().Delete(context.Background(), 9, 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListCommentsByCreator_ScanError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT comments\.id[\s\S]*FROM comments`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}).
			AddRow("bad", 9, 200, "c", 0, 0, 0, time.Now(), time.Now()))
	_, err := svc.Repo().ListCommentsByCreator(context.Background(), 10, 9, 1, 20, "latest", "comment")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListRepliesByCreator_ScanError(t *testing.T) {
	svc, mock := newCommentSvc(t)
	mock.ExpectQuery(`SELECT r\.[\s\S]*FROM replies r`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}).
			AddRow("bad", 5, 200, nil, "r", 1, "NORMAL", time.Now(), time.Now(), 9))
	_, err := svc.Repo().ListRepliesByCreator(context.Background(), 10, 9, 1, 20, "oldest")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== creator handler：默认参数 / 回复目标 / 创建失败 ====================

func TestHandleList_DefaultCommentTypeAndSort(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	// 不传 sort/commentType → 默认 latest/all；结果为空
	repoMock.ExpectQuery(`SELECT comments\.id[\s\S]*FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "user_id", "content", "like_count", "reply_count", "status", "created_at", "updated_at"}))
	repoMock.ExpectQuery(`SELECT r\.[\s\S]*FROM replies r`).WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/creator/comments", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, repoMock.ExpectationsWereMet())
}

func TestHandleList_ReplyToUser(t *testing.T) {
	h, mock, repoMock := newCreatorHandlerConv(t)
	now := time.Now()
	repoMock.ExpectQuery(`SELECT r\.[\s\S]*FROM replies r`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "comment_id", "user_id", "reply_to_user_id", "content", "like_count", "status", "created_at", "updated_at", "manuscript_id"}).
			AddRow(1, 5, 200, int64(300), "r", 1, "NORMAL", now, now, 9))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	repoMock.ExpectQuery(`SELECT id, username, nickname, avatar, level FROM users WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nickname", "avatar", "level"}).
			AddRow(200, "u", "nick", "a", 1).AddRow(300, "u3", "nick3", "b", 2))
	mock.ExpectQuery(`SELECT id, title, cover_url FROM manuscripts WHERE id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url"}).AddRow(9, "Title", "cover.jpg"))
	repoMock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest("GET", "/api/v1/creator/comments?commentType=reply", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleList(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, repoMock.ExpectationsWereMet())
}

func TestHandleByPath_POST_CreateReplyError(t *testing.T) {
	h, _, repoMock := newCreatorHandler(t)
	repoMock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM comments`).WithArgs(int64(123), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	repoMock.ExpectQuery(`INSERT INTO replies`).WillReturnError(sql.ErrConnDone)
	req := httptest.NewRequest("POST", "/api/v1/creator/comments/123?content=hello", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleByPath(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== 违禁词 / 安全配置补充 ====================

func TestSecurityConfig_LoadQueryError(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`system_configs`).WillReturnError(sql.ErrConnDone)
	p := newSecurityConfigProvider(db)
	got := p.get(context.Background())
	assert.Equal(t, defaultCommentMaxCount, got.CommentMaxCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProhibitedStore_StartRefresher_NilProvider(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Hour)
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))
	ctx, cancel := context.WithCancel(context.Background())
	s.StartRefresher(ctx, nil) // nil provider → 默认 currentTTL
	time.Sleep(100 * time.Millisecond)
	cancel()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProhibitedStore_StartRefresher_ZeroTTLFallback(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, 5*time.Minute)
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))
	ctx, cancel := context.WithCancel(context.Background())
	s.StartRefresher(ctx, func() time.Duration { return 0 }) // wait<=0 → 5min 兜底
	time.Sleep(100 * time.Millisecond)
	cancel()
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== 违禁词缓存刷新错误分支 ====================

func TestProhibitedRefresh_RowsErr(t *testing.T) {
	db, mock := newStoreDB(t)
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).
			AddRow("广告", "CONTAINS").RowError(0, sql.ErrConnDone))
	s := newProhibitedWordStore(db, time.Minute)
	err := s.refresh(context.Background())
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProhibitedStore_StartRefresher_InitialRefreshError(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Hour)
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnError(sql.ErrConnDone)
	ctx, cancel := context.WithCancel(context.Background())
	s.StartRefresher(ctx, nil) // 初始 refresh 失败 → log，但继续循环
	time.Sleep(100 * time.Millisecond)
	cancel()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProhibitedStore_StartRefresher_TimerRefreshError(t *testing.T) {
	db, mock := newStoreDB(t)
	s := newProhibitedWordStore(db, time.Hour)
	// 首次 refresh 成功
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnRows(sqlmock.NewRows([]string{"word", "match_type"}).AddRow("广告", "CONTAINS"))
	// 定时器触发后的 refresh 失败 → log 但保留旧缓存
	mock.ExpectQuery(`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`).
		WillReturnError(sql.ErrConnDone)
	ctx, cancel := context.WithCancel(context.Background())
	s.StartRefresher(ctx, func() time.Duration { return 5 * time.Millisecond })
	time.Sleep(80 * time.Millisecond)
	cancel()
	require.NoError(t, mock.ExpectationsWereMet())
}
