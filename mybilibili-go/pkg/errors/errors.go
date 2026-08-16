package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mybilibili/pkg/httputil"
)

func ErrInvalidArgument(msg string) error {
	return status.Errorf(codes.InvalidArgument, msg)
}

func ErrAlreadyExists(msg string) error {
	return status.Errorf(codes.AlreadyExists, msg)
}

func ErrNotFound(msg string) error {
	return status.Errorf(codes.NotFound, msg)
}

func ErrInternal(msg string) error {
	return status.Errorf(codes.Internal, msg)
}

func ErrUnauthenticated(msg string) error {
	return status.Errorf(codes.Unauthenticated, msg)
}

func ErrPermissionDenied(msg string) error {
	return status.Errorf(codes.PermissionDenied, msg)
}

func ErrResourceExhausted(msg string) error {
	return status.Errorf(codes.ResourceExhausted, msg)
}

func WriteHTTPError(w http.ResponseWriter, err error) {
	code := codes.Internal
	if s, ok := status.FromError(err); ok {
		code = s.Code()
	}
	httpStatus := http.StatusInternalServerError
	switch code {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	}
	httputil.WriteError(w, httpStatus, err.Error())
}
