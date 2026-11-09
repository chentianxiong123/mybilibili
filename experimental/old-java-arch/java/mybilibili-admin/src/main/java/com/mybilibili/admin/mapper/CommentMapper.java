package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Comment;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;
import java.util.Map;

@Mapper
public interface CommentMapper {

    // 用于视频级联删除的方法
    @Delete("DELETE FROM comments WHERE video_id = #{videoId}")
    int deleteByVideoId(Integer videoId);

    @Select("SELECT COUNT(*) FROM comments WHERE video_id = #{videoId}")
    int countByVideoId(Integer videoId);

    // 评论管理相关方法
    @Select("<script>" +
            "SELECT * FROM comments " +
            "<where>" +
            "  <if test='keyword != null and keyword != \"\"'>" +
            "    AND (content LIKE CONCAT('%', #{keyword}, '%'))" +
            "  </if>" +
            "  <if test='videoId != null'>" +
            "    AND video_id = #{videoId}" +
            "  </if>" +
            "</where>" +
            "ORDER BY id DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Comment> selectCommentsByKeyword(Integer offset, Integer size, String keyword, Integer videoId);

    @Select("<script>" +
            "SELECT COUNT(*) FROM comments " +
            "<where>" +
            "  <if test='keyword != null and keyword != \"\"'>" +
            "    AND (content LIKE CONCAT('%', #{keyword}, '%'))" +
            "  </if>" +
            "  <if test='videoId != null'>" +
            "    AND video_id = #{videoId}" +
            "  </if>" +
            "</where>" +
            "</script>")
    int countCommentsByKeyword(String keyword, Integer videoId);

    @Select("SELECT * FROM comments WHERE id = #{id}")
    Comment selectById(Integer id);

    @Delete("DELETE FROM comments WHERE id = #{id}")
    int delete(Integer id);

    @Update("UPDATE comments SET status = #{status} WHERE id = #{id}")
    int updateStatus(Integer id, Integer status);

    // 统计相关方法
    @Select("SELECT DATE(create_time) as date, COUNT(*) as count FROM comments WHERE create_time BETWEEN #{startDate} AND #{endDate} GROUP BY DATE(create_time) ORDER BY date")
    List<Map<String, Object>> getDailyCommentCount(String startDate, String endDate);

    @Select("SELECT COUNT(*) FROM comments WHERE create_time BETWEEN #{startDate} AND #{endDate}")
    Integer getTotalCommentCount(String startDate, String endDate);
}
