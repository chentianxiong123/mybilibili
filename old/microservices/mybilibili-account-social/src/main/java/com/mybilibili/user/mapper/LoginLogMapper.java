package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.user.entity.LoginLog;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface LoginLogMapper extends BaseMapper<LoginLog> {

    @Select("SELECT * FROM login_logs WHERE user_id = #{userId} ORDER BY login_time DESC LIMIT #{offset}, #{size}")
    List<LoginLog> selectByUserId(@Param("userId") Integer userId, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("SELECT COUNT(*) FROM login_logs WHERE user_id = #{userId}")
    Integer countByUserId(@Param("userId") Integer userId);

    @Select("<script>" +
            "SELECT * FROM login_logs " +
            "<where>" +
            "<if test='ip != null and ip != \"\"'>" +
            "AND ip LIKE CONCAT('%', #{ip}, '%')" +
            "</if>" +
            "<if test='userId != null'>" +
            "AND user_id = #{userId}" +
            "</if>" +
            "<if test='status != null'>" +
            "AND status = #{status}" +
            "</if>" +
            "<if test='startTime != null'>" +
            "AND login_time &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null'>" +
            "AND login_time &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "ORDER BY login_time DESC " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<LoginLog> selectByCondition(@Param("ip") String ip, @Param("userId") Integer userId,
                                    @Param("status") Integer status, @Param("startTime") String startTime,
                                    @Param("endTime") String endTime, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM login_logs " +
            "<where>" +
            "<if test='ip != null and ip != \"\"'>" +
            "AND ip LIKE CONCAT('%', #{ip}, '%')" +
            "</if>" +
            "<if test='userId != null'>" +
            "AND user_id = #{userId}" +
            "</if>" +
            "<if test='status != null'>" +
            "AND status = #{status}" +
            "</if>" +
            "<if test='startTime != null'>" +
            "AND login_time &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null'>" +
            "AND login_time &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "</script>")
    Integer countByCondition(@Param("ip") String ip, @Param("userId") Integer userId,
                            @Param("status") Integer status, @Param("startTime") String startTime,
                            @Param("endTime") String endTime);
}