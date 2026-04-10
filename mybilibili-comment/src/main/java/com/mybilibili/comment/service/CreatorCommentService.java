package com.mybilibili.comment.service;

import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;

import java.util.List;

public interface CreatorCommentService {

    List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size);

    int countCreatorComments(Integer userId, Integer manuscriptId);

    void deleteCommentByCreator(Integer commentId, Integer userId);

    ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId);
}