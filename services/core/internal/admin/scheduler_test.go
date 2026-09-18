package admin

import (
	"context"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	svc := &Service{}
	s := NewScheduler(svc)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.svc != svc {
		t.Error("expected svc to be set")
	}
}

func TestScheduler_Stop(t *testing.T) {
	s := NewScheduler(&Service{})
	go s.Run(context.Background())
	time.Sleep(10 * time.Millisecond)
	s.Stop()
}

func TestScheduler_Run_CtxCancel(t *testing.T) {
	s := NewScheduler(&Service{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done
}
