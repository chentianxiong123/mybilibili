package httputil

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgutil "mybilibili/pkg/httputil"
)

func GetUserIDFromHeader(r *http.Request) int64 { return pkgutil.GetUserIDFromHeader(r) }
func ParsePageParams(r *http.Request) (int32, int32) { return pkgutil.ParsePageParams(r) }
func WriteOK(w http.ResponseWriter, data interface{}) { pkgutil.WriteOK(w, data) }
func WriteJSON(w http.ResponseWriter, httpStatus int, body interface{}) { pkgutil.WriteJSON(w, httpStatus, body) }
func RequireUser(w http.ResponseWriter, r *http.Request) (int64, bool) { return pkgutil.RequireUser(w, r) }
func PathValue(r *http.Request, key string) string { return pkgutil.PathValue(r, key) }

func WriteError(w http.ResponseWriter, err error) {
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
	pkgutil.WriteError(w, httpStatus, err.Error())
}