package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;
import java.util.Map;

@Mapper
public interface UserMapper {
    @Select("SELECT id, username, password, nickname, avatar, email, phone, gender, birthdate, signature, level, following_count as followingCount, follower_count as followerCount, video_count as videoCount, liked_count as likedCount, coin_count as coinCount, point_count as pointCount, experience, bio, announcement, status, created_at as createdAt, updated_at as updatedAt FROM users WHERE username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%') LIMIT #{offset}, #{size}")
    List<User> selectUsersByKeyword(Integer offset, Integer size, String keyword);
    
    @Select("SELECT COUNT(*) FROM users WHERE username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%')")
    int countUsersByKeyword(String keyword);
    
    @Select("SELECT id, username, password, nickname, avatar, email, phone, gender, birthdate, signature, level, following_count as followingCount, follower_count as followerCount, video_count as videoCount, liked_count as likedCount, coin_count as coinCount, point_count as pointCount, experience, bio, announcement, status, created_at as createdAt, updated_at as updatedAt FROM users WHERE id = #{id}")
    User selectById(Integer id);
    
    @Update("UPDATE users SET password = #{password} WHERE id = #{id}")
    int updatePassword(Integer id, String password);
    
    @Update("UPDATE users SET status = #{status} WHERE id = #{id}")
    int updateStatus(Integer id, Integer status);
    
    // 统计相关方法
    @Select("SELECT DATE(create_time) as date, COUNT(*) as userCount FROM users WHERE create_time BETWEEN #{startDate} AND #{endDate} GROUP BY DATE(create_time) ORDER BY date")
    List<Map<String, Object>> getDailyUserCount(String startDate, String endDate);
    
    @Select("SELECT COUNT(*) FROM users")
    Integer getTotalUserCount();
}