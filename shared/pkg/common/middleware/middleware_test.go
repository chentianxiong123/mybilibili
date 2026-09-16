package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testInfo(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: "/svc.Method"}
}

func TestLogging_PassesThrough(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "resp", nil
	}
	out, err := Logging(context.Background(), "req", testInfo("/svc.Method"), handler)
	require.NoError(t, err)
	assert.Equal(t, "resp", out)
}

func TestLogging_ForwardsError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("boom")
	}
	_, err := Logging(context.Background(), "req", testInfo(""), handler)
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestRecovery_NormalPassthrough(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	out, err := Recovery(context.Background(), "req", testInfo(""), handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

func TestRecovery_PanicsReturnsInternal(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("crash")
	}
	_, err := Recovery(context.Background(), "req", testInfo("/svc.Method"), handler)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "internal server error")
}

func TestRecovery_HandlerErrorPassthrough(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "missing")
	}
	_, err := Recovery(context.Background(), "req", testInfo(""), handler)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestTimeout_DeadlineSet(t *testing.T) {
	var seenDeadline time.Time
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		seenDeadline, _ = ctx.Deadline()
		return "ok", nil
	}
	interceptor := Timeout(500 * time.Millisecond)
	out, err := interceptor(context.Background(), "req", testInfo(""), handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
	assert.False(t, seenDeadline.IsZero())
	assert.WithinDuration(t, time.Now().Add(500*time.Millisecond), seenDeadline, time.Second)
}

func TestTimeout_ZeroMeansImmediateDeadline(t *testing.T) {
	var deadline time.Time
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		deadline, _ = ctx.Deadline()
		return "ok", nil
	}
	interceptor := Timeout(0)
	start := time.Now()
	out, err := interceptor(context.Background(), "req", testInfo(""), handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
	// d=0: deadline 应设在调用时刻附近（同一毫秒级窗口内），而非远未来
	require.True(t, deadline.After(start.Add(-time.Millisecond)))
	assert.WithinDuration(t, start, deadline, 100*time.Millisecond)
}

func TestTimeout_ExpiredContextCancels(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	interceptor := Timeout(20 * time.Millisecond)
	_, err := interceptor(context.Background(), "req", testInfo(""), handler)
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}
