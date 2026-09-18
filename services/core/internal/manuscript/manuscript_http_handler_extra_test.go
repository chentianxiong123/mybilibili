package manuscript

import (
	"bytes"
	"database/sql"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	pb "mybilibili/pkg/pb"
)

// --- convertManuscriptStatusParam ---

func TestConvertManuscriptStatusParam(t *testing.T) {
	tests := []struct {
		input    string
		expected int32
	}{
		{"", -100},
		{"0", 0},
		{"3", 3},
		{"-1", -1},
		{"pending", 0},
		{"reviewing", 0},
		{"draft", 0},
		{"processing", 1},
		{"published", 3},
		{"rejected", 4},
		{"failed", 5},
		{"unpublished", -1},
		{"PUBLISHED", 3},
		{"unknown", -100},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertManuscriptStatusParam(tt.input))
		})
	}
}

// --- snakeToCamel ---

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"id", "id"},
		{"cover_url", "coverUrl"},
		{"view_count", "viewCount"},
		{"first_video_id", "firstVideoId"},
		{"review_status", "reviewStatus"},
		{"single", "single"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, snakeToCamel(tt.input))
		})
	}
}

// --- convertKeysToCamel ---

func TestConvertKeysToCamel(t *testing.T) {
	input := map[string]interface{}{
		"cover_url":   "test.jpg",
		"view_count":  int64(100),
		"nested_key":  map[string]interface{}{"inner_key": "val"},
		"simple":      "value",
		"array_field": []interface{}{map[string]interface{}{"item_id": 1}},
	}
	result := convertKeysToCamel(input).(map[string]interface{})
	assert.Equal(t, "test.jpg", result["coverUrl"])
	assert.Equal(t, int64(100), result["viewCount"])
	assert.Equal(t, "value", result["simple"])

	nested := result["nestedKey"].(map[string]interface{})
	assert.Equal(t, "val", nested["innerKey"])

	arr := result["arrayField"].([]interface{})
	item := arr[0].(map[string]interface{})
	assert.Equal(t, 1, item["itemId"])
}

// --- manuscriptToMap ---

func TestManuscriptToMap_NilUploader(t *testing.T) {
	info := &pb.ManuscriptInfo{
		Id:          1,
		Title:       "Test",
		Description: "Desc",
		Uploader:    nil,
	}
	m := manuscriptToMap(info)
	assert.Equal(t, int64(1), m["id"])
	assert.Nil(t, m["uploader"])
}

func TestManuscriptToMap_WithUploaderAndTags(t *testing.T) {
	info := &pb.ManuscriptInfo{
		Id:    1,
		Title: "Test",
		Uploader: &pb.UserInfo{
			Id:   10,
			Name: "user1",
		},
		Tags: []string{"go", "test"},
		Videos: []*pb.VideoItem{
			{Id: 100, Title: "video1"},
		},
	}
	m := manuscriptToMap(info)
	up := m["uploader"].(map[string]interface{})
	assert.Equal(t, "user1", up["nickname"])
	assert.Equal(t, "user1", up["username"])
	assert.Equal(t, []string{"go", "test"}, m["tags"])
	videos := m["videos"].([]map[string]interface{})
	assert.Len(t, videos, 1)
}

// --- manuscriptListPageJSON ---

func TestManuscriptListPageJSON_ZeroPageSize(t *testing.T) {
	resp := &pb.ListUserManuscriptsResponse{
		Manuscripts: []*pb.ManuscriptInfo{{Id: 1}},
		Total:       5,
		Page:        1,
		PageSize:    0,
	}
	result := manuscriptListPageJSON(resp)
	// size defaults to 10 when PageSize <= 0
	assert.Equal(t, int32(10), result["size"])
	assert.Equal(t, int64(1), result["totalPages"])
}

func TestManuscriptListPageJSON_Empty(t *testing.T) {
	resp := &pb.ListUserManuscriptsResponse{
		Manuscripts: []*pb.ManuscriptInfo{},
		Total:       0,
		Page:        1,
		PageSize:    10,
	}
	result := manuscriptListPageJSON(resp)
	assert.Equal(t, int64(0), result["totalPages"])
}

// --- handleInteractionStatus ---

func TestHTTPHandler_InteractionStatus_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/abc/status")
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.handleInteractionStatus(w, req)
	assert.Equal(t, 400, w.Code)
}

// --- handleUploadChunk ---

func TestHTTPHandler_UploadChunk_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/upload-chunk")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_UploadChunk_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", nil)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHTTPHandler_UploadChunk_MissingUploadId(t *testing.T) {
	h, _ := newHTTPHandler(t)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

// --- handleUploadCompleteWeb ---

func TestHTTPHandler_UploadCompleteWeb_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := newReq("GET", "/api/v1/manuscript/upload-complete")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHTTPHandler_UploadCompleteWeb_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", nil)
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHTTPHandler_UploadCompleteWeb_MissingUploadId(t *testing.T) {
	h, _ := newHTTPHandler(t)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHTTPHandler_UploadCompleteWeb_SessionNotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id`).
		WillReturnError(sql.ErrNoRows)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("uploadId", "nonexist")
	writer.Close()
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHTTPHandler_UploadCompleteWeb_WrongOwner(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("uploadId", "abc")
	writer.Close()
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)
	assert.Equal(t, 403, w.Code)
}

// --- handleDeleteManuscript error paths ---

func TestHTTPHandler_DeleteManuscript_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM manuscripts WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := newReq("DELETE", "/api/v1/manuscript/1")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- handleUpdateManuscript multipart ---

func TestHTTPHandler_UpdateManuscript_Multipart(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET title`).WillReturnResult(sqlmock.NewResult(0, 1))

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("title", "new title")
	writer.Close()
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

// --- manuscriptRouteName edge cases ---

func TestManuscriptRouteName_Empty(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{}))
}

func TestManuscriptRouteName_UploadSession_Complete(t *testing.T) {
	assert.Equal(t, "uploadSessionComplete", manuscriptRouteName([]string{"upload-session", "abc", "complete"}))
}

func TestManuscriptRouteName_UploadSession_UnknownPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"upload-session", "abc", "unknown"}))
}

func TestManuscriptRouteName_FixDurations(t *testing.T) {
	assert.Equal(t, "fixDurations", manuscriptRouteName([]string{"fix-durations"}))
}

func TestManuscriptRouteName_FixDurations_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"fix-durations", "extra"}))
}

func TestManuscriptRouteName_UploadChunk_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"upload-chunk", "extra"}))
}

func TestManuscriptRouteName_UploadComplete_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"upload-complete", "extra"}))
}

func TestManuscriptRouteName_Recommended_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"recommended", "extra"}))
}

func TestManuscriptRouteName_Hot_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"hot", "extra"}))
}

func TestManuscriptRouteName_List_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"list", "extra"}))
}

func TestManuscriptRouteName_Me_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"me", "unknown"}))
}

func TestManuscriptRouteName_Category_ExtraPath(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"category", "1", "extra"}))
}

func TestManuscriptRouteName_User_Collections(t *testing.T) {
	assert.Equal(t, "userCollections", manuscriptRouteName([]string{"user", "collections"}))
}

func TestManuscriptRouteName_User_Likes(t *testing.T) {
	assert.Equal(t, "userLikes", manuscriptRouteName([]string{"user", "likes"}))
}

func TestManuscriptRouteName_Favorite_Folders(t *testing.T) {
	assert.Equal(t, "favoriteFolders", manuscriptRouteName([]string{"favorite", "folders"}))
}

func TestManuscriptRouteName_Favorite_FoldersByID(t *testing.T) {
	assert.Equal(t, "favoriteFolderByID", manuscriptRouteName([]string{"favorite", "folders", "1"}))
}

func TestManuscriptRouteName_Favorite_FoldersVideos(t *testing.T) {
	assert.Equal(t, "favoriteFolderVideos", manuscriptRouteName([]string{"favorite", "folders", "1", "videos"}))
}

func TestManuscriptRouteName_Favorite_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"favorite", "unknown"}))
}

func TestManuscriptRouteName_ID_Status(t *testing.T) {
	assert.Equal(t, "status", manuscriptRouteName([]string{"1", "status"}))
}

func TestManuscriptRouteName_ID_Like(t *testing.T) {
	assert.Equal(t, "like", manuscriptRouteName([]string{"1", "like"}))
}

func TestManuscriptRouteName_ID_Coin(t *testing.T) {
	assert.Equal(t, "coin", manuscriptRouteName([]string{"1", "coin"}))
}

func TestManuscriptRouteName_ID_Collect(t *testing.T) {
	assert.Equal(t, "collect", manuscriptRouteName([]string{"1", "collect"}))
}

func TestManuscriptRouteName_ID_Share(t *testing.T) {
	assert.Equal(t, "share", manuscriptRouteName([]string{"1", "share"}))
}

func TestManuscriptRouteName_ID_Favorite(t *testing.T) {
	assert.Equal(t, "videoFavorite", manuscriptRouteName([]string{"1", "favorite"}))
}

func TestManuscriptRouteName_ID_CommentCount(t *testing.T) {
	assert.Equal(t, "commentCount", manuscriptRouteName([]string{"1", "comment-count"}))
}

func TestManuscriptRouteName_ID_IncrementComment(t *testing.T) {
	assert.Equal(t, "incrementComment", manuscriptRouteName([]string{"1", "increment-comment"}))
}

func TestManuscriptRouteName_ID_DecrementComment(t *testing.T) {
	assert.Equal(t, "decrementComment", manuscriptRouteName([]string{"1", "decrement-comment"}))
}

func TestManuscriptRouteName_ID_Publish(t *testing.T) {
	assert.Equal(t, "publishManuscript", manuscriptRouteName([]string{"1", "publish"}))
}

func TestManuscriptRouteName_ID_Unpublish(t *testing.T) {
	assert.Equal(t, "unpublishManuscript", manuscriptRouteName([]string{"1", "unpublish"}))
}

func TestManuscriptRouteName_ID_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"1", "unknown"}))
}

func TestManuscriptRouteName_ID_FavoriteFolders(t *testing.T) {
	assert.Equal(t, "videoFavoriteFolders", manuscriptRouteName([]string{"1", "favorite", "folders"}))
}

func TestManuscriptRouteName_ID_FavoriteFoldersByID(t *testing.T) {
	assert.Equal(t, "videoFavoriteFolderByID", manuscriptRouteName([]string{"1", "favorite", "folders", "2"}))
}

func TestManuscriptRouteName_ID_ThreeParts_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"1", "unknown", "2"}))
}

func TestManuscriptRouteName_ID_FourParts_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"1", "unknown", "2", "3"}))
}

func TestManuscriptRouteName_DefaultDetail(t *testing.T) {
	assert.Equal(t, "detail", manuscriptRouteName([]string{"123"}))
}

func TestManuscriptRouteName_User_Stats(t *testing.T) {
	assert.Equal(t, "userStats", manuscriptRouteName([]string{"user", "10", "stats"}))
}

func TestManuscriptRouteName_User_Search(t *testing.T) {
	assert.Equal(t, "userSearch", manuscriptRouteName([]string{"user", "10", "search"}))
}

func TestManuscriptRouteName_User_Manuscripts(t *testing.T) {
	assert.Equal(t, "userManuscripts", manuscriptRouteName([]string{"user", "10"}))
}

func TestManuscriptRouteName_User_Unknown(t *testing.T) {
	assert.Equal(t, "", manuscriptRouteName([]string{"user", "10", "unknown"}))
}

func TestManuscriptRouteName_UploadSession_ByID(t *testing.T) {
	assert.Equal(t, "uploadSessionByID", manuscriptRouteName([]string{"upload-session", "abc"}))
}

func TestManuscriptRouteName_UploadSession_Single(t *testing.T) {
	assert.Equal(t, "uploadSession", manuscriptRouteName([]string{"upload-session"}))
}

func TestManuscriptRouteName_MeList(t *testing.T) {
	assert.Equal(t, "meList", manuscriptRouteName([]string{"me", "list"}))
}

func TestManuscriptRouteName_MeStats(t *testing.T) {
	assert.Equal(t, "meStats", manuscriptRouteName([]string{"me", "stats"}))
}
