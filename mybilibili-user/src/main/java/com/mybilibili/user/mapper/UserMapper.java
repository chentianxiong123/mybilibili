package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface UserMapper extends BaseMapper<User> {

    @Select("SELECT * FROM users WHERE username = #{username}")
    User selectByUsername(String username);

    @Select("SELECT * FROM users WHERE email = #{email}")
    User selectByEmail(@Param("email") String email);

    @Select("SELECT COUNT(*) FROM user_interactions WHERE user_id = #{userId} AND target_type = 'user' AND interaction_type = 'follow'")
    Integer countFollowing(@Param("userId") Integer userId);

    @Select("SELECT COUNT(*) FROM user_interactions WHERE target_id = #{userId} AND target_type = 'user' AND interaction_type = 'follow'")
    Integer countFollowers(@Param("userId") Integer userId);

    @Select("SELECT COUNT(*) FROM user_dynamics WHERE user_id = #{userId}")
    Integer countDynamics(@Param("userId") Integer userId);

    @Select("SELECT COALESCE(SUM(view_count), 0) FROM manuscripts WHERE user_id = #{userId}")
    Integer sumViewCountByUserId(@Param("userId") Integer userId);

    @Select("SELECT COALESCE(SUM(like_count), 0) FROM manuscripts WHERE user_id = #{userId}")
    Integer sumLikeCountByUserId(@Param("userId") Integer userId);

    @Select("<script>" +
            "SELECT * FROM users " +
            "<where>" +
            "<if test='keyword != null and keyword != \"\"'>" +
            "AND (username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%'))" +
            "</if>" +
            "</where>" +
            "ORDER BY created_at DESC " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<User> selectUserList(@Param("keyword") String keyword, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM users " +
            "<where>" +
            "<if test='keyword != null and keyword != \"\"'>" +
            "AND (username LIKE CONCAT('%', #{keyword}, '%') OR nickname LIKE CONCAT('%', #{keyword}, '%'))" +
            "</if>" +
            "</where>" +
            "</script>")
    Integer countUserList(@Param("keyword") String keyword);

    @Update("UPDATE users SET password = #{password} WHERE id = #{id}")
    void updatePassword(@Param("id") Integer id, @Param("password") String password);

    @Update("UPDATE users SET status = #{status} WHERE id = #{id}")
    void updateStatus(@Param("id") Integer id, @Param("status") Integer status);

    @Update("UPDATE users SET following_count = following_count + 1 WHERE id = #{id}")
    void incrementFollowingCount(@Param("id") Integer id);

    @Update("UPDATE users SET following_count = GREATEST(following_count - 1, 0) WHERE id = #{id}")
    void decrementFollowingCount(@Param("id") Integer id);

    @Update("UPDATE users SET follower_count = follower_count + 1 WHERE id = #{id}")
    void incrementFollowerCount(@Param("id") Integer id);

    @Update("UPDATE users SET follower_count = GREATEST(follower_count - 1, 0) WHERE id = #{id}")
    void decrementFollowerCount(@Param("id") Integer id);

    @Update("UPDATE users SET coin_count = coin_count + #{coinAmount}")
    int addDailyCoins(@Param("coinAmount") int coinAmount);

    @Update("UPDATE users SET experience = experience + #{experienceAmount} WHERE id = #{userId}")
    int addExperience(@Param("userId") Integer userId, @Param("experienceAmount") int experienceAmount);

    @Select("SELECT COUNT(*) FROM users")
    Integer countAllUsers();

    @Select("SELECT COUNT(*) FROM manuscripts WHERE user_id = #{userId}")
    Integer countManuscriptsByUserId(@Param("userId") Integer userId);
}