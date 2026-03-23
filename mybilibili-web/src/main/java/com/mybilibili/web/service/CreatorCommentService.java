package com.mybilibili.web.service;

import com.mybilibili.common.vo.CreatorCommentVO;
import com.mybilibili.common.vo.ReplyVO;

import java.util.List;

/**
 * 创作者评论管理服务接口
 */
public interface CreatorCommentService {

    /**
     * 获取创作者所有稿件的评论列表（分页）
     *
     * @param userId       创作者用户ID
     * @param manuscriptId 稿件ID（可选，用于筛选）
     * @param page         页码
     * @param size         每页数量
     * @return 评论列表
     */
    List<CreatorCommentVO> getCreatorComments(Integer userId, Integer manuscriptId, Integer page, Integer size);

    /**
     * 统计创作者评论总数
     *
     * @param userId       创作者用户ID
     * @param manuscriptId 稿件ID（可选，用于筛选）
     * @return 评论总数
     */
    int countCreatorComments(Integer userId, Integer manuscriptId);

    /**
     * 删除评论（创作者权限，只能删除自己稿件下的评论）
     *
     * @param commentId 评论ID
     * @param userId    创作者用户ID
     */
    void deleteCommentByCreator(Integer commentId, Integer userId);

    /**
     * 回复评论
     *
     * @param commentId    评论ID
     * @param userId       回复者用户ID
     * @param content      回复内容
     * @param replyToUserId 回复目标用户ID（可选）
     * @return 回复视图对象
     */
    ReplyVO replyComment(Integer commentId, Integer userId, String content, Integer replyToUserId);
}
