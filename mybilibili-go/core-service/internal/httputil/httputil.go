package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetUserIDFromHeader(r *http.Request) int64 {
	idStr := r.Header.Get("X-User-Id")
	if idStr == "" {
		return 0
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func ParsePageParams(r *http.Request) (int32, int32) {
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	sizeStr := r.URL.Query().Get("page_size")
	if sizeStr == "" {
		sizeStr = r.URL.Query().Get("pageSize")
	}
	if sizeStr == "" {
		sizeStr = r.URL.Query().Get("size")
	}
	size, _ := strconv.ParseInt(sizeStr, 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 20
	}
	return int32(page), int32(size)
}

func WriteOK(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{"code": 200, "data": data, "message": "ok"})
}

func WriteJSON(w http.ResponseWriter, httpStatus int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(body)
}

func WriteError(w http.ResponseWriter, err error) {
	code := codes.Internal
	if s, ok := status.FromError(err); ok {
		code = s.Code()
	}
	httpStatus := http.StatusInternalServerError
	apiCode := 500
	switch code {
	case codes.NotFound:
		httpStatus, apiCode = http.StatusNotFound, 404
	case codes.InvalidArgument:
		httpStatus, apiCode = http.StatusBadRequest, 400
	case codes.Unauthenticated:
		httpStatus, apiCode = http.StatusUnauthorized, 401
	case codes.PermissionDenied:
		httpStatus, apiCode = http.StatusForbidden, 403
	}
	WriteJSON(w, httpStatus, map[string]interface{}{"code": apiCode, "message": err.Error(), "data": nil})
}

func RequireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	uid := GetUserIDFromHeader(r)
	if uid == 0 {
		WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized", "data": nil})
		return 0, false
	}
	return uid, true
}

func PathValue(r *http.Request, key string) string {
	v := r.PathValue(key)
	if v != "" {
		return v
	}
	return ""
}