package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCaptchaAndEmailCode(t *testing.T) {
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

	_, _ = db.ExecContext(ctx, `TRUNCATE verification_codes RESTART IDENTITY CASCADE`)

	userRepo := NewRepository(db)
	userSvc := NewService(userRepo, "test-secret")
	h := &UserExtendHandler{svc: userSvc}

	// captcha new
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/captcha/new", nil)
	h.handleCaptcha(rec, req)
	if rec.Code != 200 {
		t.Fatalf("captcha new status %d", rec.Code)
	}
	var captchaResp struct {
		CaptchaID string `json:"captchaId"`
		Question  string `json:"question"`
	}
	json.Unmarshal(rec.Body.Bytes(), &captchaResp)
	if captchaResp.CaptchaID == "" || captchaResp.Question == "" {
		t.Fatalf("captcha new missing fields: %s", rec.Body.String())
	}

	// verify with wrong answer
	rec = httptest.NewRecorder()
	body := `{"captchaId":"` + captchaResp.CaptchaID + `","answer":"0"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/captcha/verify", strings.NewReader(body))
	h.handleCaptcha(rec, req)
	if rec.Code != 200 {
		t.Fatalf("captcha verify status %d", rec.Code)
	}
	var verifyResp struct {
		Verified bool `json:"verified"`
	}
	json.Unmarshal(rec.Body.Bytes(), &verifyResp)
	if verifyResp.Verified {
		t.Fatalf("expected false for wrong answer")
	}

	// email code
	rec = httptest.NewRecorder()
	body = `{"email":"test@example.com"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/email/code", strings.NewReader(body))
	h.handleEmailCode(rec, req)
	if rec.Code != 200 {
		t.Fatalf("email code status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "code_sent") {
		t.Fatalf("expected code_sent: %s", rec.Body.String())
	}

	// read the actual code from DB
	var code string
	_ = db.QueryRowContext(ctx, `SELECT code_value FROM verification_codes WHERE identifier='test@example.com' AND code_type='email_code'`).Scan(&code)
	if code == "" {
		t.Fatal("expected code in DB")
	}

	// verify with correct code
	rec = httptest.NewRecorder()
	body = `{"email":"test@example.com","code":"` + code + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/email/verify", strings.NewReader(body))
	h.handleEmailVerify(rec, req)
	if rec.Code != 200 {
		t.Fatalf("email verify status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "verified") {
		t.Fatalf("expected verified: %s", rec.Body.String())
	}

	// verify with wrong code (should fail)
	rec = httptest.NewRecorder()
	body = `{"email":"test@example.com","code":"000000"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/email/verify", strings.NewReader(body))
	h.handleEmailVerify(rec, req)
	if !strings.Contains(rec.Body.String(), "failed") {
		t.Fatalf("expected failed for wrong code: %s", rec.Body.String())
	}
}

func TestLoginLogCount(t *testing.T) {
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

	_, _ = db.ExecContext(ctx, `TRUNCATE login_logs, users RESTART IDENTITY CASCADE`)
	_, _ = db.ExecContext(ctx, `INSERT INTO users (id, username, password) VALUES (960001, 'log_u', 'x') ON CONFLICT DO NOTHING`)
	_, _ = db.ExecContext(ctx, `INSERT INTO login_logs (user_id, ip, status) VALUES (960001, '1.1.1.1', 0), (960001, '2.2.2.2', 1), (960001, '3.3.3.3', 0)`)

	userRepo := NewRepository(db)
	userSvc := NewService(userRepo, "test-secret")
	h := &UserExtendHandler{svc: userSvc}

	// countUserLogs (by user_id)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/login-logs/count?user_id=960001", nil)
	req.Header.Set("X-User-Id", "960001")
	h.handleLoginLogCount(rec, req)
	if rec.Code != 200 {
		t.Fatalf("count status %d", rec.Code)
	}
	var countResp struct {
		Total int64 `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &countResp)
	if countResp.Total != 3 {
		t.Fatalf("expected 3 total, got %d", countResp.Total)
	}

	// countByCondition (status=0)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/user/login-logs/count?user_id=960001&status=0", nil)
	req.Header.Set("X-User-Id", "960001")
	h.handleLoginLogCount(rec, req)
	json.Unmarshal(rec.Body.Bytes(), &countResp)
	if countResp.Total != 2 {
		t.Fatalf("expected 2 with status=0, got %d", countResp.Total)
	}
}