package com.mybilibili.comment.service;

import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import java.util.List;
import java.util.Map;

public interface CommentService {

    CommentVO addComment(Integer manuscriptId, Integer userId, String content);

    List<CommentVO> getCommentsByManuscriptIdOnly(Integer manuscriptId, Integer page, Integer size, Integer userId);

    CommentVO addCommentByManuscriptId(Integer manuscriptId, Integer userId, String content);

    List<CommentVO> getCommentsByManuscriptId(Integer manuscriptId, Integer page, Integer size, Integer userId, String sort);

    void deleteComment(Integer commentId, Integer userId);

    ReplyVO addReply(Integer commentId, Integer userId, String content, Integer replyToUserId);

    List<ReplyVO> getRepliesByCommentId(Integer commentId, Integer page, Integer size, Integer userId);

    void deleteReply(Integer replyId, Integer userId);

    void likeComment(Integer commentId, Integer userId);

    void unlikeComment(Integer commentId, Integer userId);

    void likeReply(Integer replyId, Integer userId);

    void unlikeReply(Integer replyId, Integer userId);

    Map<Integer, Boolean> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds);

    Map<Integer, Integer> batchGetLikeCount(String targetType, List<Integer> targetIds);

    int getLikeCount(String targetType, Integer targetId);

    boolean isLiked(Integer userId, String targetType, Integer targetId);

    Result<?> getAdminCommentList(Integer page, Integer size, String keyword, Integer manuscriptId);

    Result<?> getAdminCommentById(Integer id);

    void deleteCommentByAdmin(Integer id);

    void updateCommentStatusByAdmin(Integer id, Integer status);

    Result<?> getPendingList(String contentType, Integer page, Integer size);

    Result<?> getAllContent(String contentType, String status, Integer page, Integer size);

    Result<?> restoreContent(String type, Integer id);

    Result<?> deleteContent(String type, Integer id);

    Result<?> batchProcess(String action, List<Map<String, Object>> items);
}
