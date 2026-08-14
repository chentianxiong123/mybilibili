package core

import (
	"testing"
	"time"
)

func TestCommentRateLimiter(t *testing.T) {
	l := newCommentRateLimiter(10*time.Minute, 2)
	now := time.Now()
	if l.record(1, now) {
		t.Fatalf("first record should be allowed")
	}
	if l.record(1, now.Add(1*time.Second)) {
		t.Fatalf("second record should be allowed")
	}
	if !l.record(1, now.Add(2*time.Second)) {
		t.Fatalf("third record should be rate-limited")
	}
	if r := l.remaining(1, now.Add(3*time.Second)); r != 0 {
		t.Fatalf("remaining should be 0, got %d", r)
	}
	// 窗口滑动后恢复
	if r := l.remaining(1, now.Add(10*time.Minute+1*time.Second)); r != 2 {
		t.Fatalf("after window remaining should be 2, got %d", r)
	}
	// 不同用户互不影响
	if l.record(2, now) {
		t.Fatalf("other user should not be rate-limited")
	}
}
