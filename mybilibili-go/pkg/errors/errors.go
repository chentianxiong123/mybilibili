package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
