package manuscript

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestFavoriteFolderDelete_DestroysOtherUsersLinks pins a transaction-atomicity
// / cross-user corruption bug in manuscript_http_handler.go:handleFavoriteFolderByID
// DELETE branch (lines ~1323-1326).
//
// ROOT CAUSE: the two DELETEs are not wrapped in a transaction and the FIRST
// one — `DELETE FROM favorite_folder_videos WHERE folder_id = $1` — is NOT
// scoped to the requesting user (unlike the second, owner-scoped
// `DELETE FROM favorite_folders WHERE id = $1 AND user_id = $2`). If the folder
// belongs to another user, the attacker still destroys that user's
// favorite_folder_videos links, then the owner-scoped folder delete affects 0
// rows (no error) and the handler returns 200.
//
// IMPACT: a user can destroy another user's favorite-folder video associations
// (data corruption) while the response claims success. Also non-atomic: if the
// folder delete fails, the links are already gone.
func TestFavoriteFolderDelete_DestroysOtherUsersLinks(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// First delete: removes the video links for folder 5 — NOT scoped to the
	// caller. Succeeds (1 row), even though folder 5 belongs to another user.
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id = \$1`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 3))
	// Second delete: owner-scoped, affects 0 rows (folder not owned by caller) —
	// but the code does not check RowsAffected, so the handler returns OK.
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(5), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := newReq("DELETE", "/api/v1/manuscript/favorite/folders/5", "")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // BUG: reports success

	// BUG confirmed: another user's favorite_folder_videos were deleted
	// (first Exec affected 3 rows) while the folder itself was NOT deleted
	// (second Exec affected 0 rows, error ignored).
}
