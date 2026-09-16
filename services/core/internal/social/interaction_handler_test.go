package social

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockInteractionHandler(t *testing.T) (*GenericInteractionHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewGenericInteractionHandler(NewInteractionRepository(db)), mock
}

func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	return out
}

func muxForInteraction(h *GenericInteractionHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestInteractionHandler_Register(t *testing.T) {
	h, _ := newMockInteractionHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	w := doReq(mux, "GET", "/api/v1/interaction/count?targetType=video&targetId=9", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInteractionHandler_handleLike_Unauthorized(t *testing.T) {
	h, _ := newMockInteractionHandler(t)
	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/like", `{"targetType":"video","targetId":9}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInteractionHandler_handleLike_InvalidBody(t *testing.T) {
	h, _ := newMockInteractionHandler(t)
	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/like", `not-json`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInteractionHandler_handleLike_MissingFields(t *testing.T) {
	h, _ := newMockInteractionHandler(t)
	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/like", `{"targetType":""}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInteractionHandler_handleLike_Post(t *testing.T) {
	h, mock := newMockInteractionHandler(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WithArgs(int64(1), "video", int64(9), "like").WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/like", `{"targetType":"video","targetId":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleLike_Post_Error(t *testing.T) {
	h, mock := newMockInteractionHandler(t)
	mock.ExpectExec(`INSERT INTO user_interactions`).WillReturnError(errors.New("boom"))
	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/like", `{"targetType":"video","targetId":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInteractionHandler_handleLike_Delete(t *testing.T) {
	h, mock := newMockInteractionHandler(t)
	mock.ExpectExec(`DELETE FROM user_interactions`).WithArgs(int64(1), "video", int64(9), "like").WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForInteraction(h), "DELETE", "/api/v1/interaction/like", `{"targetType":"video","targetId":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleLike_MethodNotAllowed(t *testing.T) {
	h, _ := newMockInteractionHandler(t)
	w := doReq(muxForInteraction(h), "PUT", "/api/v1/interaction/like", `{"targetType":"video","targetId":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestInteractionHandler_handleStatus(t *testing.T) {
	h, mock := newMockInteractionHandler(t)

	w := doReq(muxForInteraction(h), "GET", "/api/v1/interaction/status?targetType=video&targetId=9", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), "video", int64(9), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doReq(muxForInteraction(h), "GET", "/api/v1/interaction/status?targetType=video&targetId=9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	out := parseBody(t, w)
	assert.Equal(t, true, out["data"].(map[string]interface{})["liked"])

	w = doReq(muxForInteraction(h), "GET", "/api/v1/interaction/status", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleBatchStatus(t *testing.T) {
	h, mock := newMockInteractionHandler(t)

	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/batch/status", `{"targetType":"video","targetIds":[9,10]}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doReq(muxForInteraction(h), "POST", "/api/v1/interaction/batch/status", `bad`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), "video", int64(9), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).WithArgs(int64(1), "video", int64(10), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w = doReq(muxForInteraction(h), "POST", "/api/v1/interaction/batch/status", `{"targetType":"video","targetIds":[9,10]}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleCount(t *testing.T) {
	h, mock := newMockInteractionHandler(t)

	w := doReq(muxForInteraction(h), "GET", "/api/v1/interaction/count", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WithArgs("video", int64(9), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	w = doReq(muxForInteraction(h), "GET", "/api/v1/interaction/count?targetType=video&targetId=9", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	out := parseBody(t, w)
	assert.Equal(t, float64(3), out["data"].(map[string]interface{})["count"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleBatchCount(t *testing.T) {
	h, mock := newMockInteractionHandler(t)

	w := doReq(muxForInteraction(h), "POST", "/api/v1/interaction/batch/count", `bad`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WithArgs("video", int64(9), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions WHERE target_type`).WithArgs("video", int64(10), "like").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doReq(muxForInteraction(h), "POST", "/api/v1/interaction/batch/count", `{"targetType":"video","targetIds":[9,10]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInteractionHandler_handleRouter_NotFound(t *testing.T) {
	h, _ := newMockInteractionHandler(t)

	w := doReq(muxForInteraction(h), "GET", "/api/v1/interaction/bad", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = doReq(muxForInteraction(h), "GET", "/api/v1/interaction/batch/bad", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = doReq(muxForInteraction(h), "GET", "/api/v1/interaction/batch", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
