package com.mybilibili.comment.service;

import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;

import java.util.List;

public interface CreatorCommentService {

    List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size, String sort, String commentType);

    int countCreatorComments(Integer userId, Integer manuscriptId, String commentType);

    void deleteCommentByCreator(Integer commentId, Integer userId);

    void deleteReplyByCreator(Integer replyId, Integer userId);

    ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId);
}