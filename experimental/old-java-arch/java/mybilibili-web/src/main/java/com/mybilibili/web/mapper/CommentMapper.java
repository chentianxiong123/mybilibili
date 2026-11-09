package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.enums.TargetType;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface CommentMapper {
    int insert(Comment comment);
    Comment selectById(Integer id);

    // 向后兼容：根据 manuscriptId 查询
    List<Comment> selectByManuscriptId(Integer manuscriptId, int offset, int size);

    // 新增：根据 targetType 和 targetId 查询
    List<Comment> selectByTargetTypeAndTargetId(
            @Param("targetType") TargetType targetType,
            @Param("targetId") Integer targetId,
            @Param("offset") int offset,
            @Param("size") int size);

    int updateLikeCount(Integer id, int count);
    int updateLikeCountDirect(@Param("id") Integer id, @Param("count") int count);
    int updateReplyCount(Integer id, int count);
    int delete(Integer id);

    // 向后兼容：根据 manuscriptId 统计
    int countByManuscriptId(Integer manuscriptId);

    // 新增：根据 targetType 和 targetId 统计
    int countByTargetTypeAndTargetId(
            @Param("targetType") TargetType targetType,
            @Param("targetId") Integer targetId);

    // 创作者评论管理：查询创作者所有稿件的评论
    List<Comment> selectByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId,
            @Param("offset") int offset,
            @Param("size") int size);

    // 创作者评论管理：统计创作者所有稿件的评论数
    int countByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId);
}
