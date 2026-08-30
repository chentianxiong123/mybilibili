package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface ContentReviewMapper {
    // 视频审核相关
    @Select("SELECT * FROM videos WHERE (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%')) AND (#{status} IS NULL OR status = #{status}) LIMIT #{offset}, #{size}")
    List<Video> selectVideosForReview(Integer offset, Integer size, String keyword, Integer status);
    
    @Select("SELECT COUNT(*) FROM videos WHERE (title LIKE CONCAT('%', #{keyword}, '%') OR description LIKE CONCAT('%', #{keyword}, '%')) AND (#{status} IS NULL OR status = #{status})")
    int countVideosForReview(String keyword, Integer status);
    
    @Update("UPDATE videos SET status = #{status} WHERE id = #{id}")
    int updateVideoStatus(Integer id, Integer status);
    
    // 评论审核相关
    @Select("SELECT * FROM comments WHERE content LIKE CONCAT('%', #{keyword}, '%') AND (#{status} IS NULL OR status = #{status}) LIMIT #{offset}, #{size}")
    List<Comment> selectCommentsForReview(Integer offset, Integer size, String keyword, Integer status);
    
    @Select("SELECT COUNT(*) FROM comments WHERE content LIKE CONCAT('%', #{keyword}, '%') AND (#{status} IS NULL OR status = #{status})")
    int countCommentsForReview(String keyword, Integer status);
    
    @Update("UPDATE comments SET status = #{status} WHERE id = #{id}")
    int updateCommentStatus(Integer id, Integer status);
    
    // 用户审核相关
    @Select("SELECT * FROM users WHERE (username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%')) AND (#{status} IS NULL OR status = #{status}) LIMIT #{offset}, #{size}")
    List<User> selectUsersForReview(Integer offset, Integer size, String keyword, Integer status);
    
    @Select("SELECT COUNT(*) FROM users WHERE (username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%')) AND (#{status} IS NULL OR status = #{status})")
    int countUsersForReview(String keyword, Integer status);
    
    @Update("UPDATE users SET status = #{status} WHERE id = #{id}")
    int updateUserStatus(Integer id, Integer status);
}