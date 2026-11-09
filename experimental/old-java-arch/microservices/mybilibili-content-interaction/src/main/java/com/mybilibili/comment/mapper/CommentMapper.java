package com.mybilibili.comment.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Comment;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface CommentMapper extends BaseMapper<Comment> {

    List<Comment> selectByManuscriptId(
            @Param("manuscriptId") Integer manuscriptId,
            @Param("offset") int offset,
            @Param("size") int size,
            @Param("sort") String sort);

    int updateLikeCountDirect(@Param("id") Integer id, @Param("count") int count);

    int updateReplyCount(@Param("id") Integer id, @Param("count") int count);

    List<Comment> selectByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId,
            @Param("offset") int offset,
            @Param("size") int size,
            @Param("sort") String sort);

    int countByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId);

    int countByManuscriptId(@Param("manuscriptId") Integer manuscriptId);

    @Select("<script>" +
            "SELECT * FROM comments WHERE 1=1 " +
            "<if test='keyword != null'> AND content LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='manuscriptId != null'> AND manuscript_id = #{manuscriptId} </if>" +
            "ORDER BY created_at DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Comment> selectAdminList(@Param("keyword") String keyword, @Param("manuscriptId") Integer manuscriptId, @Param("offset") int offset, @Param("size") int size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM comments WHERE 1=1 " +
            "<if test='keyword != null'> AND content LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='manuscriptId != null'> AND manuscript_id = #{manuscriptId} </if>" +
            "</script>")
    int countAdminList(@Param("keyword") String keyword, @Param("manuscriptId") Integer manuscriptId);

    @Select("SELECT * FROM comments WHERE status = 1 ORDER BY created_at DESC LIMIT #{offset}, #{size}")
    List<Comment> selectPendingComments(@Param("offset") int offset, @Param("size") int size);

    @Select("SELECT COUNT(*) FROM comments WHERE status = 1")
    int countPendingComments();

    @Select("<script>" +
            "SELECT * FROM comments WHERE 1=1 " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "ORDER BY created_at DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Comment> selectAllComments(@Param("status") Integer status, @Param("offset") int offset, @Param("size") int size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM comments WHERE 1=1 " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "</script>")
    int countAllComments(@Param("status") Integer status);
}
