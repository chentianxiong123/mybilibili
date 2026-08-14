package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	pb "mybilibili/internal/core/pb"
)

func TestMessageRoutes(t *testing.T) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		t.Skip("PG_DSN not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	_, _ = db.ExecContext(ctx, `TRUNCATE messages, conversations, users RESTART IDENTITY CASCADE`)
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (950001, 'msg_u1', 'x'), (950002, 'msg_u2', 'x') ON CONFLICT DO NOTHING`)

	messageRepo := NewMessageRepository(db)
	notif := NewNotificationBroadcaster()
	h := NewMessageHTTPHandler(messageRepo, notif)

	// send a message from 950001 to 950002
	rec := httptest.NewRecorder()
	body := `{"receiver_id":950002,"content":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/message/send", strings.NewReader(body))
	req.Header.Set("X-User-Id", "950001")
	h.handleSend(rec, req)
	if rec.Code != 200 {
		t.Fatalf("send status %d body %s", rec.Code, rec.Body.String())
	}
	var msg struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &msg)
	if msg.ID == 0 {
		t.Fatal("expected message id")
	}

	// getMessageById
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/message/"+strconv.FormatInt(msg.ID, 10), nil)
	req.Header.Set("X-User-Id", "950002")
	h.handleMessageByID(rec, req)
	if rec.Code != 200 {
		t.Fatalf("get by id status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "hello") {
		t.Fatalf("get by id missing content: %s", rec.Body.String())
	}

	// getMessagesByUserId
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/message/user/950002", nil)
	req.Header.Set("X-User-Id", "950002")
	h.handleMessagesByUser(rec, req)
	if rec.Code != 200 {
		t.Fatalf("get by user status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "hello") {
		t.Fatalf("get by user missing content: %s", rec.Body.String())
	}

	// updateConversation (set unread_count to 0)
	var convID int64
	_ = db.QueryRowContext(ctx, `SELECT id FROM conversations WHERE user_id=950002`).Scan(&convID)
	rec = httptest.NewRecorder()
	body = `{"unread_count":0}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/message/conversations/"+strconv.FormatInt(convID, 10), strings.NewReader(body))
	req.Header.Set("X-User-Id", "950002")
	h.handleConversationByID(rec, req)
	if rec.Code != 200 {
		t.Fatalf("update conversation status %d", rec.Code)
	}

	// clearMessagesByUserId
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/message/user/950002", nil)
	req.Header.Set("X-User-Id", "950002")
	h.handleMessagesByUser(rec, req)
	if rec.Code != 200 {
		t.Fatalf("clear status %d", rec.Code)
	}
	var cnt int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages WHERE receiver_id=950002`).Scan(&cnt)
	if cnt != 0 {
		t.Fatalf("expected 0 messages after clear, got %d", cnt)
	}
}

func TestNotificationHooks(t *testing.T) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		t.Skip("PG_DSN not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	_, _ = db.ExecContext(ctx, `TRUNCATE messages, conversations, manuscripts, categories, users RESTART IDENTITY CASCADE`)
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (951001, 'owner', 'x'), (951002, 'liker', 'x') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO categories (id, name) VALUES (951001, 'cat') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO manuscripts (id, title, user_id, category_id) VALUES (951001, 'ms', 951001, 951001) ON CONFLICT DO NOTHING`)

	messageRepo := NewMessageRepository(db)
	interactionRepo := NewInteractionRepository(db)
	interactionSvc := NewInteractionService(interactionRepo)
	interactionSvc.SetMessageRepo(messageRepo)

	// LikeManuscript should trigger sendLikeNotification
	likeReq := &pb.LikeManuscriptRequest{UserId: 951002, ManuscriptId: 951001}
	_, err = interactionSvc.LikeManuscript(ctx, likeReq)
	if err != nil {
		t.Fatalf("LikeManuscript: %v", err)
	}
	var cnt int
	_ = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE receiver_id=951001 AND sender_id=951002 AND message_type=4`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 like notification, got %d", cnt)
	}

	// Also test comment service notification hooks
	commentRepo := NewCommentRepository(db)
	commentSvc := NewCommentService(commentRepo)
	commentSvc.SetMessageRepo(messageRepo)

	// Add a comment from 951001
	commentReq := &pb.AddCommentRequest{ManuscriptId: 951001, UserId: 951001, Content: "my comment"}
	commentResp, err := commentSvc.AddComment(ctx, commentReq)
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	// LikeComment from 951002 should trigger sendCommentLikeNotification
	_, err = commentSvc.LikeComment(ctx, &pb.LikeCommentRequest{CommentId: commentResp.Comment.Id, UserId: 951002})
	if err != nil {
		t.Fatalf("LikeComment: %v", err)
	}
	_ = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE receiver_id=951001 AND sender_id=951002 AND message_type=6`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 comment like notification, got %d", cnt)
	}

	// AddReply from 951002 should trigger sendReplyNotification
	_, err = commentSvc.AddReply(ctx, &pb.AddReplyRequest{
		CommentId: commentResp.Comment.Id, UserId: 951002, Content: "my reply",
	})
	if err != nil {
		t.Fatalf("AddReply: %v", err)
	}
	_ = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE receiver_id=951001 AND sender_id=951002 AND message_type=2`).Scan(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 reply notification, got %d", cnt)
	}
}