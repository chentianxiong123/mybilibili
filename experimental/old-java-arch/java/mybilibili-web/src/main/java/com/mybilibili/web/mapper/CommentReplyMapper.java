package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.CommentReply;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface CommentReplyMapper {
    // 插入评论回复
    int insert(CommentReply commentReply);
    
    // 根据评论ID获取回复列表
    List<CommentReply> getByCommentId(@Param("commentId") Integer commentId, @Param("offset") Integer offset, @Param("limit") Integer limit);
    
    // 根据ID获取回复
    CommentReply getById(Integer id);
    
    // 更新回复点赞数
    int updateLikeCount(@Param("id") Integer id, @Param("count") Integer count);
    
    // 更新回复状态
    int updateStatus(@Param("id") Integer id, @Param("status") Integer status);
    
    // 根据评论ID获取回复数量
    int countByCommentId(Integer commentId);

    // 根据稿件ID获取所有回复数量
    int countByManuscriptId(Integer manuscriptId);
}
