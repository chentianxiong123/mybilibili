package com.mybilibili.admin.mapper;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Update;

import java.util.List;
import java.util.Map;

@Mapper
public interface CommentReviewMapper {

    List<Map<String, Object>> selectPendingComments(@Param("offset") Integer offset, @Param("size") Integer size);

    int countPendingComments();

    List<Map<String, Object>> selectPendingReplies(@Param("offset") Integer offset, @Param("size") Integer size);

    int countPendingReplies();

    List<Map<String, Object>> selectComments(@Param("status") Integer status, @Param("offset") Integer offset, @Param("size") Integer size);

    int countComments(@Param("status") Integer status);

    List<Map<String, Object>> selectReplies(@Param("status") String status, @Param("offset") Integer offset, @Param("size") Integer size);

    int countReplies(@Param("status") String status);

    @Update("UPDATE comments SET status = 0 WHERE id = #{id}")
    int restoreComment(@Param("id") Integer id);

    @Update("UPDATE replies SET status = 'NORMAL' WHERE id = #{id}")
    int restoreReply(@Param("id") Integer id);

    @Update("DELETE FROM comments WHERE id = #{id}")
    int deleteComment(@Param("id") Integer id);

    @Update("DELETE FROM replies WHERE id = #{id}")
    int deleteReply(@Param("id") Integer id);
}
