package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"
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

func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}