package core

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	pb "mybilibili/internal/core/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// PublicAPIHandler 提供评论的公开 HTTP JSON 端点（Flutter App 与 web-ts 直接消费），
// 内部复用现有 gRPC service 的逻辑，不经过 gRPC 二进制协议。
// 稿件与互动端点统一由 ManuscriptHTTPHandler 的 /api/v1/manuscript/ 子树分发。
type PublicAPIHandler struct {
	commentSvc *CommentService
}

func NewPublicAPIHandler(commentSvc *CommentService) *PublicAPIHandler {
	return &PublicAPIHandler{commentSvc: commentSvc}
}

func (h *PublicAPIHandler) Register(mux *http.ServeMux) {
	// 评论
	mux.HandleFunc("/api/v1/comment/list", h.handleCommentList)
	mux.HandleFunc("/api/v1/comment/add", h.handleCommentAdd)
	mux.HandleFunc("/api/v1/comment/reply", h.handleCommentReply)
	mux.HandleFunc("/api/v1/comment/{id}/like", h.handleCommentLike)
	mux.HandleFunc("/api/v1/comment/{id}/replies", h.handleCommentReplies)
	mux.HandleFunc("/api/v1/comment/reply/{id}/like", h.handleReplyLike)
	mux.HandleFunc("/api/v1/comment/batch-like-counts", h.handleBatchLikeCounts)
}

// ---- 响应辅助 ----

func writeOK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"code": 200, "data": data, "message": "ok"})
}

func writeJSON(w http.ResponseWriter, httpStatus int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err error) {
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
	writeJSON(w, httpStatus, map[string]interface{}{"code": apiCode, "message": err.Error(), "data": nil})
}

func requireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	uid := getUserIDFromHeader(r)
	if uid == 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized", "data": nil})
		return 0, false
	}
	return uid, true
}

// pathValue 兼容 Go 1.22 ServeMux 的 {id} 占位符。
func pathValue(r *http.Request, key string) string {
	v := r.PathValue(key)
	if v != "" {
		return v
	}
	return ""
}

// ---- 评论序列化 ----

// commentToMap 将 pb 评论编码为 camelCase map（protojson），
// 并补 username（Flutter 读取 username，web 读 userName）。
func commentToMap(c *pb.CommentInfo) map[string]interface{} {
	b, _ := protojson.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m == nil {
		m = map[string]interface{}{}
	}
	if name, ok := m["userName"].(string); ok {
		m["username"] = name
	}
	if ct, ok := m["createdAt"].(string); ok {
		m["createTime"] = ct
	}
	if reps, ok := m["replies"].([]interface{}); ok {
		for i, r := range reps {
			rm, _ := r.(map[string]interface{})
			if rm != nil {
				if name, ok := rm["userName"].(string); ok {
					rm["username"] = name
				}
				if ct, ok := rm["createdAt"].(string); ok {
					rm["createTime"] = ct
				}
			}
			m["replies"].([]interface{})[i] = rm
		}
	}
	return m
}

func commentListToJSON(infos []*pb.CommentInfo) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(infos))
	for _, info := range infos {
		out = append(out, commentToMap(info))
	}
	return out
}

func replyListToJSON(infos []*pb.ReplyInfo) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(infos))
	for _, info := range infos {
		b, _ := protojson.Marshal(info)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		if m == nil {
			m = map[string]interface{}{}
		}
		if name, ok := m["userName"].(string); ok {
			m["username"] = name
		}
		if ct, ok := m["createdAt"].(string); ok {
			m["createTime"] = ct
		}
		out = append(out, m)
	}
	return out
}

func replyToMapJSON(info *pb.ReplyInfo) map[string]interface{} {
	if info == nil {
		return map[string]interface{}{}
	}
	list := replyListToJSON([]*pb.ReplyInfo{info})
	return list[0]
}

// decodeCommentBody 兼容 Flutter(application/json) 与 web-ts(x-www-form-urlencoded)。
func decodeCommentBody(r *http.Request, fields map[string]*string) error {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}
		for k, v := range fields {
			*v = r.Form.Get(k)
		}
		return nil
	}
	var m map[string]interface{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	for k, v := range fields {
		if val, ok := m[k]; ok {
			switch t := val.(type) {
			case string:
				*v = t
			case float64:
				*v = strconv.FormatInt(int64(t), 10)
			case json.Number:
				*v = t.String()
			}
		}
	}
	return nil
}

// ---- 评论 ----

func (h *PublicAPIHandler) handleCommentList(w http.ResponseWriter, r *http.Request) {
	manuscriptID, _ := strconv.ParseInt(r.URL.Query().Get("manuscriptId"), 10, 64)
	page, size := parsePageParams(r)
	sort := r.URL.Query().Get("sort")
	uid := getUserIDFromHeader(r)
	resp, err := h.commentSvc.ListComments(r.Context(), &pb.ListCommentsRequest{
		ManuscriptId: manuscriptID, Page: page, PageSize: size, Sort: sort, UserId: uid,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, commentListToJSON(resp.Comments))
}

func (h *PublicAPIHandler) handleCommentAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	var manuscriptIDStr, content string
	if err := decodeCommentBody(r, map[string]*string{"manuscriptId": &manuscriptIDStr, "content": &content}); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	manuscriptID, _ := strconv.ParseInt(manuscriptIDStr, 10, 64)
	if manuscriptID <= 0 || strings.TrimSpace(content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "manuscriptId and content required", "data": nil})
		return
	}
	resp, err := h.commentSvc.AddComment(r.Context(), &pb.AddCommentRequest{ManuscriptId: manuscriptID, UserId: uid, Content: content})
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, commentToMap(resp.Comment))
}

func (h *PublicAPIHandler) handleCommentReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	var commentIDStr, content, replyToStr string
	if err := decodeCommentBody(r, map[string]*string{"commentId": &commentIDStr, "content": &content, "replyToUserId": &replyToStr}); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "invalid body", "data": nil})
		return
	}
	commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
	replyToUserID, _ := strconv.ParseInt(replyToStr, 10, 64)
	if commentID <= 0 || strings.TrimSpace(content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "message": "commentId and content required", "data": nil})
		return
	}
	resp, err := h.commentSvc.AddReply(r.Context(), &pb.AddReplyRequest{CommentId: commentID, UserId: uid, Content: content, ReplyToUserId: replyToUserID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, replyToMapJSON(resp.Reply))
}

func (h *PublicAPIHandler) handleCommentLike(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(pathValue(r, "id"), 10, 64)
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		_, err := h.commentSvc.LikeComment(r.Context(), &pb.LikeCommentRequest{CommentId: id, UserId: uid})
		if err != nil {
			writeError(w, err)
			return
		}
	case http.MethodDelete:
		_, err := h.commentSvc.UnlikeComment(r.Context(), &pb.UnlikeCommentRequest{CommentId: id, UserId: uid})
		if err != nil {
			writeError(w, err)
			return
		}
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	writeOK(w, map[string]interface{}{"status": "ok"})
}

func (h *PublicAPIHandler) handleCommentReplies(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(pathValue(r, "id"), 10, 64)
	page, size := parsePageParams(r)
	uid := getUserIDFromHeader(r)
	resp, err := h.commentSvc.GetReplies(r.Context(), &pb.GetRepliesRequest{CommentId: id, Page: page, PageSize: size, UserId: uid})
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, replyListToJSON(resp.Replies))
}

func (h *PublicAPIHandler) handleReplyLike(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(pathValue(r, "id"), 10, 64)
	uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		_, err := h.commentSvc.LikeReply(r.Context(), &pb.LikeReplyRequest{ReplyId: id, UserId: uid})
		if err != nil {
			writeError(w, err)
			return
		}
	case http.MethodDelete:
		_, err := h.commentSvc.UnlikeReply(r.Context(), &pb.UnlikeReplyRequest{ReplyId: id, UserId: uid})
		if err != nil {
			writeError(w, err)
			return
		}
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	writeOK(w, map[string]interface{}{"status": "ok"})
}

// handleBatchLikeCounts 批量查询评论/回复点赞数（对齐旧版 batchGetLikeCount）。
func (h *PublicAPIHandler) handleBatchLikeCounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "message": "method not allowed", "data": nil})
		return
	}
	uid := getUserIDFromHeader(r)
	targetType := r.URL.Query().Get("type")
	if targetType == "" {
		targetType = "comment"
	}
	weirdIDs := strings.Split(r.URL.Query().Get("ids"), ",")
	ids := make([]int64, 0, len(weirdIDs))
	for _, s := range weirdIDs {
		if s == "" {
			continue
		}
		id, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	upper := "COMMENT"
	if targetType == "reply" {
		upper = "REPLY"
	}
	counts, _ := h.commentSvc.repo.BatchGetLikeCounts(r.Context(), upper, ids)
	liked, _ := h.commentSvc.repo.BatchIsLiked(r.Context(), uid, upper, ids)
	out := map[string]interface{}{}
	for _, id := range ids {
		out[strconv.FormatInt(id, 10)] = map[string]interface{}{
			"like_count": counts[id],
			"is_liked":   liked[id],
		}
	}
	writeOK(w, out)
}