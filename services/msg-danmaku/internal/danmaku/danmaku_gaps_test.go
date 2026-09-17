package danmaku

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRepo = errors.New("repo error")

// Send 携带 manuscript_id > 0 时应累加当日指标（弹幕发送核心 -> 持久化路径）。
func TestHandleDanmakuSend_WithManuscript_UpsertsMetric(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).
		WithArgs(int64(88), int64(1001), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rr := doDanmaku(t, h, http.MethodPost, "/api/v1/danmaku/send",
		`{"video_id":99,"manuscript_id":88,"content":"指标弹幕","time":1,"color":"#fff","mode":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Send 底层 Create 失败 -> 400。
func TestHandleDanmakuSend_SvcError(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectQuery(`INSERT INTO danmaku`).
		WillReturnError(errRepo)

	rr := doDanmaku(t, h, http.MethodPost, "/api/v1/danmaku/send",
		`{"video_id":99,"content":"x","time":1}`)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handler 层 Delete 失败 -> 400。
func TestHandleDanmakuByPath_Delete_Error(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(7), int64(1001)).
		WillReturnError(errRepo)

	rr := doDanmaku(t, h, http.MethodDelete, "/api/v1/danmaku/7", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handler 层 CreatorDelete 失败 -> 400。
func TestHandleDanmakuCreatorByPath_Delete_Error(t *testing.T) {
	h, mock := newTestDanmaku(t)

	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1 AND video_id IN`).
		WithArgs(int64(5), int64(1001)).
		WillReturnError(errRepo)

	rr := doDanmaku(t, h, http.MethodDelete, "/api/v1/creator/danmaku/5", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// trend 的 manuscriptIDs 全部无效（ids 兜底为空）→ 直接返回空 map。
func TestHandleDanmakuTrend_NoValidIDs(t *testing.T) {
	h, _ := newTestDanmaku(t)

	rr := doDanmaku(t, h, http.MethodGet, "/api/v1/danmaku/trend?ids=abc,xyz", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

type danmakuNoFlusher struct {
	w http.ResponseWriter
}

func (n danmakuNoFlusher) Header() http.Header        { return n.w.Header() }
func (n danmakuNoFlusher) Write(b []byte) (int, error) { return n.w.Write(b) }
func (n danmakuNoFlusher) WriteHeader(s int)          { n.w.WriteHeader(s) }

func TestHandleDanmakuSSE_NoFlusher(t *testing.T) {
	h, _ := newTestDanmaku(t)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/sse/danmaku?video_id=99", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(danmakuNoFlusher{w: rec}, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}