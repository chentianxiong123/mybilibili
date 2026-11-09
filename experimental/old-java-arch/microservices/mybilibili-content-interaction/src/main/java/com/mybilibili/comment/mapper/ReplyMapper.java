package com.mybilibili.comment.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Reply;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface ReplyMapper extends BaseMapper<Reply> {

    List<Reply> selectByCommentId(Integer commentId, int offset, int size);

    List<Reply> selectPreviewByCommentIds(@Param("commentIds") List<Integer> commentIds,
                                          @Param("limitPerComment") int limitPerComment);

    int updateLikeCountDirect(@Param("id") Integer id, @Param("count") int count);

    int countByCommentId(Integer commentId);

    @Select("SELECT * FROM replies WHERE status = 'REMOVED' ORDER BY created_at DESC LIMIT #{offset}, #{size}")
    List<Reply> selectPendingReplies(@Param("offset") int offset, @Param("size") int size);

    @Select("SELECT COUNT(*) FROM replies WHERE status = 'REMOVED'")
    int countPendingReplies();

    @Select("<script>" +
            "SELECT * FROM replies WHERE 1=1 " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "ORDER BY created_at DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Reply> selectAllReplies(@Param("status") String status, @Param("offset") int offset, @Param("size") int size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM replies WHERE 1=1 " +
            "<if test='status != null'> AND status = #{status} </if>" +
            "</script>")
    int countAllReplies(@Param("status") String status);

    List<Reply> selectByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId,
            @Param("offset") int offset,
            @Param("size") int size,
            @Param("sort") String sort);

    int countByCreatorId(
            @Param("userId") Integer userId,
            @Param("manuscriptId") Integer manuscriptId);
}
