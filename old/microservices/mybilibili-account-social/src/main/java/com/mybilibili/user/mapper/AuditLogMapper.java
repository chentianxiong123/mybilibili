package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.user.entity.AuditLog;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface AuditLogMapper extends BaseMapper<AuditLog> {

    @Select("<script>" +
            "SELECT * FROM audit_logs " +
            "<where>" +
            "<if test='operatorKeyword != null and operatorKeyword != \"\"'>" +
            "AND (CAST(operator_id AS CHAR) = #{operatorKeyword} OR operator_name LIKE CONCAT('%', #{operatorKeyword}, '%'))" +
            "</if>" +
            "<if test='module != null and module != \"\"'>" +
            "AND module = #{module}" +
            "</if>" +
            "<if test='action != null and action != \"\"'>" +
            "AND action = #{action}" +
            "</if>" +
            "<if test='result != null'>" +
            "AND result = #{result}" +
            "</if>" +
            "<if test='targetKeyword != null and targetKeyword != \"\"'>" +
            "AND (target_type LIKE CONCAT('%', #{targetKeyword}, '%') OR target_id LIKE CONCAT('%', #{targetKeyword}, '%'))" +
            "</if>" +
            "<if test='startTime != null and startTime != \"\"'>" +
            "AND created_at &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null and endTime != \"\"'>" +
            "AND created_at &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "ORDER BY created_at DESC " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<AuditLog> selectByCondition(@Param("operatorKeyword") String operatorKeyword,
                                     @Param("module") String module,
                                     @Param("action") String action,
                                     @Param("result") Integer result,
                                     @Param("targetKeyword") String targetKeyword,
                                     @Param("startTime") String startTime,
                                     @Param("endTime") String endTime,
                                     @Param("offset") Integer offset,
                                     @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM audit_logs " +
            "<where>" +
            "<if test='operatorKeyword != null and operatorKeyword != \"\"'>" +
            "AND (CAST(operator_id AS CHAR) = #{operatorKeyword} OR operator_name LIKE CONCAT('%', #{operatorKeyword}, '%'))" +
            "</if>" +
            "<if test='module != null and module != \"\"'>" +
            "AND module = #{module}" +
            "</if>" +
            "<if test='action != null and action != \"\"'>" +
            "AND action = #{action}" +
            "</if>" +
            "<if test='result != null'>" +
            "AND result = #{result}" +
            "</if>" +
            "<if test='targetKeyword != null and targetKeyword != \"\"'>" +
            "AND (target_type LIKE CONCAT('%', #{targetKeyword}, '%') OR target_id LIKE CONCAT('%', #{targetKeyword}, '%'))" +
            "</if>" +
            "<if test='startTime != null and startTime != \"\"'>" +
            "AND created_at &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null and endTime != \"\"'>" +
            "AND created_at &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "</script>")
    Integer countByCondition(@Param("operatorKeyword") String operatorKeyword,
                             @Param("module") String module,
                             @Param("action") String action,
                             @Param("result") Integer result,
                             @Param("targetKeyword") String targetKeyword,
                             @Param("startTime") String startTime,
                             @Param("endTime") String endTime);
}
