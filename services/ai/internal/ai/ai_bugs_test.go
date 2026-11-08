package ai

// Pinning tests for confirmed AI-service defects. Each test asserts CURRENT
// (buggy) behavior; it fails once the bug is fixed.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "mybilibili/pkg/pb"
)

// BUG: maskKey reveals the whole key for short keys. For a key of length <= 10
// the first 6 + last 4 characters cover the entire key, so maskKey("abcdefghij")
// yields "abcdef****ghij" which contains all 10 characters — the secret is
// leaked in the GET /configs and GET /configs/{id} responses.
// File: ai_handler.go:299-304.
func TestMaskKey_ShortKeyFullyRevealed_Pinned(t *testing.T) {
	key := "abcdefghij" // length 10
	masked := maskKey(key)
	// FINDING-070 修复：len<=10 完全打码，剥离 * 后不含任何真实字符。
	assert.Equal(t, "****", masked)
	assert.NotEqual(t, key, strings.ReplaceAll(masked, "*", ""),
		"FINDING-070 修复: 短 key 不应泄露")
}

// BUG: PUT /api/v1/ai/configs/{id} decodes the body into ApiConfig; when the
// body omits api_key (a partial update), APIKey is "" and UpdateConfig writes
// api_key='' — silently wiping stored credentials. File: ai_handler.go:135-140;
// ai_repository.go:98-103.
func TestUpdateConfig_PartialUpdateWipesAPIKey_Pinned(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	// PUT body without api_key -> api_key arg is "" (wiped).
	mock.ExpectExec(`UPDATE ai_api_configs SET name=\$1`).
		WithArgs("new-name", "", "", "", "gpt-4o", int32(0), float64(0), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	svc := NewService(repo)
	h := NewHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/ai/configs/1",
		strings.NewReader(`{"name":"new-name","model":"gpt-4o"}`))
	req.Header.Set("X-Admin-Id", "1") // FINDING-024 修复：需 admin 身份
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.NoError(t, mock.ExpectationsWereMet())
	// Pinned: api_key was overwritten with "" on a partial update.
	assert.Equal(t, http.StatusOK, rr.Code)
}

// BUG: RouteSkills with empty content returns every enabled CUSTOMER_SERVICE
// skill as "matched" (score 50) because strings.Contains(name, "") is always
// true. Empty input should yield no matches. File: ai_repository.go:496-553.
func TestRouteSkills_EmptyContentMatchesAll_Pinned(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).
		WithArgs("CUSTOMER_SERVICE").
		WillReturnRows(skillsRows(
			skillRow(1, "账号与登录问题", 1),
			skillRow(2, "视频上传与发布", 1),
		))
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "", 3)
	require.NoError(t, err)
	// Pinned: empty content matches all enabled skills.
	require.Len(t, out, 2, "pinned: empty content should match nothing but returns all skills")
}

// BUG: GrpcServer.ReviewContent dereferences s.reviewSvc without a nil check
// (unlike GetSummary which guards nil summarySvc), so a server built with
// NewGrpcServer(nil, summarySvc) panics on a nil pointer dereference.
// File: grpc_server.go:24.
func TestGrpcServer_ReviewContentNilReviewSvc_Panics_Pinned(t *testing.T) {
	s := NewGrpcServer(nil, nil)
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		_, _ = s.ReviewContent(context.Background(), &pb.ReviewContentRequest{Content: "x"})
	}()
	assert.False(t, panicked, "FINDING-072 修复: ReviewContent 不应 panic 于 nil reviewSvc")
}
