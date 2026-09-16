package errors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
)

func grpcCode(err error) codes.Code {
	s, _ := status.FromError(err)
	if s == nil {
		return codes.Unknown
	}
	return s.Code()
}

func TestErrInvalidArgument(t *testing.T) {
	err := ErrInvalidArgument("bad value")
	assert.Equal(t, codes.InvalidArgument, grpcCode(err))
	assert.Contains(t, err.Error(), "bad value")
}

func TestErrAlreadyExists(t *testing.T) {
	err := ErrAlreadyExists("dup")
	assert.Equal(t, codes.AlreadyExists, grpcCode(err))
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("missing")
	assert.Equal(t, codes.NotFound, grpcCode(err))
}

func TestErrInternal(t *testing.T) {
	err := ErrInternal("boom")
	assert.Equal(t, codes.Internal, grpcCode(err))
}

func TestErrUnauthenticated(t *testing.T) {
	err := ErrUnauthenticated("no token")
	assert.Equal(t, codes.Unauthenticated, grpcCode(err))
}

func TestErrPermissionDenied(t *testing.T) {
	err := ErrPermissionDenied("forbidden")
	assert.Equal(t, codes.PermissionDenied, grpcCode(err))
}

func TestErrResourceExhausted(t *testing.T) {
	err := ErrResourceExhausted("slow down")
	assert.Equal(t, codes.ResourceExhausted, grpcCode(err))
}

func TestWriteHTTPError_CodeMapping(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    int
		wantMsg string
	}{
		{"not found", ErrNotFound("missing"), http.StatusNotFound, "missing"},
		{"invalid argument", ErrInvalidArgument("bad"), http.StatusBadRequest, "bad"},
		{"unauthenticated", ErrUnauthenticated("anon"), http.StatusUnauthorized, "anon"},
		{"permission denied", ErrPermissionDenied("deny"), http.StatusForbidden, "deny"},
		{"internal fallback", ErrInternal("oops"), http.StatusInternalServerError, "oops"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteHTTPError(rec, tt.err)
			assert.Equal(t, tt.want, rec.Code)
			body := rec.Body.String()
			assert.Contains(t, body, `"message":"`+tt.wantMsg+`"`)
			// 不暴露 grpc 内部格式
			assert.NotContains(t, body, "rpc error")
		})
	}
}

func TestWriteHTTPError_NonGRPCError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteHTTPError(rec, nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), `"code":500`))
}
