package social

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type SocialHandler struct {
	followSvc  *FollowService
	dynamicSvc *DynamicService
	collectSvc *CollectionService
	shareRepo  *ShareRepository
}

func NewSocialHandler(followSvc *FollowService, dynamicSvc *DynamicService, collectSvc *CollectionService, shareRepo *ShareRepository) *SocialHandler {
	return &SocialHandler{followSvc: followSvc, dynamicSvc: dynamicSvc, collectSvc: collectSvc, shareRepo: shareRepo}
}

func (h *SocialHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/dynamic/", h.handleDynamic)
	mux.HandleFunc("/api/v1/dynamic/comment/list", h.handleDynamicCommentList)
	mux.HandleFunc("/api/v1/dynamic/comment/add", h.handleDynamicCommentAdd)
	mux.HandleFunc("/api/v1/dynamic/comment/replies", h.handleDynamicCommentReplies)
	mux.HandleFunc("/api/v1/dynamic/comment/delete/", h.handleDynamicCommentDelete)
	mux.HandleFunc("/api/v1/dynamic/comment/like/", h.handleDynamicCommentLike)
	mux.HandleFunc("/api/v1/collection/", h.handleCollection)
	mux.HandleFunc("/api/v1/share/", h.handleShare)
	mux.HandleFunc("/api/v1/watch-history/", h.handleWatchHistory)
}

func (h *SocialHandler) handleDynamicCommentList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	dynamicID, _ := strconv.ParseInt(r.URL.Query().Get("dynamicId"), 10, 64)
	page, limit := parsePage(r)
	list, _ := h.dynamicSvc.ListComments(r.Context(), dynamicID, page, limit)
	json.NewEncoder(w).Encode(list)
}

func (h *SocialHandler) handleDynamicCommentAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	userID := getUserID(r)
	dynamicID, _ := strconv.ParseInt(r.URL.Query().Get("dynamicId"), 10, 64)
	content := r.URL.Query().Get("content")
	parentID, _ := strconv.ParseInt(r.URL.Query().Get("parentId"), 10, 64)
	replyUserID, _ := strconv.ParseInt(r.URL.Query().Get("replyUserId"), 10, 64)
	if content == "" {
		var req struct {
			DynamicID   int64  `json:"dynamicId"`
			Content     string `json:"content"`
			ParentID    int64  `json:"parentId"`
			ReplyUserID int64  `json:"replyUserId"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		dynamicID, content, parentID, replyUserID = req.DynamicID, req.Content, req.ParentID, req.ReplyUserID
	}
	dc, err := h.dynamicSvc.AddComment(r.Context(), dynamicID, userID, content, parentID, replyUserID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(dc)
}

func (h *SocialHandler) handleDynamicCommentReplies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	commentID, _ := strconv.ParseInt(r.URL.Query().Get("commentId"), 10, 64)
	page, limit := parsePage(r)
	list, _ := h.dynamicSvc.ListReplies(r.Context(), commentID, page, limit)
	json.NewEncoder(w).Encode(list)
}

func (h *SocialHandler) handleDynamicCommentDelete(w http.ResponseWriter, r *http.Request) {
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/comment/delete/")
	commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
	if r.Method == "DELETE" && commentID > 0 {
		h.dynamicSvc.repo.DeleteComment(r.Context(), commentID)
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *SocialHandler) handleDynamicCommentLike(w http.ResponseWriter, r *http.Request) {
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/comment/like/")
	commentID, _ := strconv.ParseInt(commentIDStr, 10, 64)
	switch r.Method {
	case "POST":
		if commentID > 0 {
			h.dynamicSvc.repo.db.ExecContext(r.Context(),
				`UPDATE dynamic_comments SET like_count = like_count + 1 WHERE id = $1`, commentID)
		}
	case "DELETE":
		if commentID > 0 {
			h.dynamicSvc.repo.db.ExecContext(r.Context(),
				`UPDATE dynamic_comments SET like_count = GREATEST(like_count - 1, 0) WHERE id = $1`, commentID)
		}
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *SocialHandler) handleDynamic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/dynamic/")
	userID := getUserID(r)
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 1 && parts[0] == "publish" && r.Method == "POST":
		content := r.URL.Query().Get("content")
		dynType, _ := strconv.ParseInt(r.URL.Query().Get("type"), 10, 32)
		refID, _ := strconv.ParseInt(r.URL.Query().Get("ref_manuscript_id"), 10, 64)
		d, err := h.dynamicSvc.Publish(r.Context(), userID, content, int32(dynType), "", refID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(d)

	case len(parts) == 1 && parts[0] == "list" && r.Method == "GET":
		page, limit := parsePage(r)
		list, _ := h.dynamicSvc.ListFollowing(r.Context(), userID, page, limit)
		json.NewEncoder(w).Encode(list)

	case len(parts) == 1 && parts[0] == "following" && r.Method == "GET":
		page, limit := parsePage(r)
		list, _ := h.dynamicSvc.ListFollowing(r.Context(), userID, page, limit)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 2 && parts[0] == "user" && r.Method == "GET":
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		page, limit := parsePage(r)
		list, _ := h.dynamicSvc.ListByUser(r.Context(), uid, page, limit)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 2 && parts[0] == "like":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		if r.Method == "POST" {
			h.dynamicSvc.Like(r.Context(), id, userID)
		} else if r.Method == "DELETE" {
			h.dynamicSvc.Unlike(r.Context(), id, userID)
		}
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] == "comment" && r.Method == "POST":
		dynamicID, _ := strconv.ParseInt(parts[1], 10, 64)
		content := r.URL.Query().Get("content")
		parentID, _ := strconv.ParseInt(r.URL.Query().Get("parent_id"), 10, 64)
		replyID, _ := strconv.ParseInt(r.URL.Query().Get("reply_user_id"), 10, 64)
		dc, err := h.dynamicSvc.AddComment(r.Context(), dynamicID, userID, content, parentID, replyID)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(dc)

	case len(parts) >= 2 && parts[0] == "comment" && r.Method == "GET":
		dynamicID, _ := strconv.ParseInt(parts[1], 10, 64)
		page, limit := parsePage(r)
		list, _ := h.dynamicSvc.ListComments(r.Context(), dynamicID, page, limit)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 1 && r.Method == "DELETE":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.dynamicSvc.Delete(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "not found", 404)
	}
}

func (h *SocialHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/collection/")
	userID := getUserID(r)
	parts := strings.Split(path, "/")

	switch {
	case len(parts) == 1 && parts[0] == "" && r.Method == "POST":
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			CoverURL    string `json:"cover_url"`
			Status      int32  `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Title == "" {
			http.Error(w, "title required", 400)
			return
		}
		c, err := h.collectSvc.Create(r.Context(), userID, req.Title, req.Description, req.CoverURL, req.Status)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(c)

	case len(parts) >= 2 && parts[0] == "user" && r.Method == "GET":
		uid, _ := strconv.ParseInt(parts[1], 10, 64)
		list, _ := h.collectSvc.ListByUser(r.Context(), uid)
		json.NewEncoder(w).Encode(list)

	case len(parts) >= 1 && parts[0] != "" && r.Method == "GET" && !strings.Contains(parts[0], "manuscript"):
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		c, _ := h.collectSvc.GetByID(r.Context(), id)
		json.NewEncoder(w).Encode(c)

	case len(parts) >= 1 && parts[0] != "" && r.Method == "PUT":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      int32  `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		h.collectSvc.Update(r.Context(), id, userID, req.Title, req.Description, req.Status)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 1 && parts[0] != "" && r.Method == "DELETE" && !strings.Contains(parts[0], "manuscript"):
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		h.collectSvc.Delete(r.Context(), id, userID)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 2 && parts[0] != "" && parts[1] == "manuscripts" && r.Method == "GET":
		id, _ := strconv.ParseInt(parts[0], 10, 64)
		page, limit := parsePage(r)
		ids, _ := h.collectSvc.ListManuscripts(r.Context(), id, page, limit)
		json.NewEncoder(w).Encode(ids)

	case len(parts) >= 3 && parts[1] == "manuscript" && r.Method == "POST":
		collectionID, _ := strconv.ParseInt(parts[0], 10, 64)
		manuscriptID, _ := strconv.ParseInt(parts[2], 10, 64)
		h.collectSvc.AddManuscript(r.Context(), collectionID, manuscriptID, userID)
		w.Write([]byte(`{"status":"ok"}`))

	case len(parts) >= 3 && parts[1] == "manuscript" && r.Method == "DELETE":
		collectionID, _ := strconv.ParseInt(parts[0], 10, 64)
		manuscriptID, _ := strconv.ParseInt(parts[2], 10, 64)
		h.collectSvc.RemoveManuscript(r.Context(), collectionID, manuscriptID, userID)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "not found", 404)
	}
}

func (h *SocialHandler) handleShare(w http.ResponseWriter, r *http.Request) {
	manuscriptID, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/api/v1/share/"), 10, 64)
	if r.Method == "POST" && manuscriptID > 0 {
		userID := getUserID(r)
		channel := r.URL.Query().Get("channel")
		ip := r.Header.Get("X-Forwarded-For")
		h.shareRepo.Record(r.Context(), userID, manuscriptID, channel, ip)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (h *SocialHandler) handleWatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	switch r.Method {
	case "GET":
		page, limit := parsePage(r)
		offset := (page - 1) * limit
		rows, err := h.shareRepo.db.QueryContext(r.Context(),
			`SELECT manuscript_id, progress_seconds, watched_at FROM watch_history WHERE user_id = $1 ORDER BY watched_at DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		type WH struct {
			ManuscriptID int64  `json:"manuscript_id"`
			Progress     int    `json:"progress_seconds"`
			WatchedAt    string `json:"watched_at"`
		}
		var list []WH
		for rows.Next() {
			var wh WH
			rows.Scan(&wh.ManuscriptID, &wh.Progress, &wh.WatchedAt)
			list = append(list, wh)
		}
		json.NewEncoder(w).Encode(list)

	case "POST":
		manuscriptID, _ := strconv.ParseInt(r.URL.Query().Get("manuscript_id"), 10, 64)
		progress, _ := strconv.ParseInt(r.URL.Query().Get("progress_seconds"), 10, 32)
		h.shareRepo.db.ExecContext(r.Context(),
			`INSERT INTO watch_history (user_id, manuscript_id, progress_seconds) VALUES ($1,$2,$3)
			 ON CONFLICT (user_id, manuscript_id) DO UPDATE SET progress_seconds=$3, watched_at=NOW()`,
			userID, manuscriptID, progress)
		w.Write([]byte(`{"status":"ok"}`))

	case "DELETE":
		h.shareRepo.db.ExecContext(r.Context(), `DELETE FROM watch_history WHERE user_id = $1`, userID)
		w.Write([]byte(`{"status":"ok"}`))
	}
}
