package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.user.entity.OperationTask;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface OperationTaskMapper extends BaseMapper<OperationTask> {

    @Select("<script>" +
            "SELECT * FROM operation_tasks " +
            "<where>" +
            "<if test='taskType != null and taskType != \"\"'>" +
            "AND task_type = #{taskType}" +
            "</if>" +
            "<if test='status != null and status != \"\"'>" +
            "AND status = #{status}" +
            "</if>" +
            "<if test='targetKeyword != null and targetKeyword != \"\"'>" +
            "AND (target_type LIKE CONCAT('%', #{targetKeyword}, '%') OR target_id LIKE CONCAT('%', #{targetKeyword}, '%'))" +
            "</if>" +
            "<if test='keyword != null and keyword != \"\"'>" +
            "AND (task_key LIKE CONCAT('%', #{keyword}, '%') OR task_name LIKE CONCAT('%', #{keyword}, '%') OR message LIKE CONCAT('%', #{keyword}, '%'))" +
            "</if>" +
            "<if test='startTime != null and startTime != \"\"'>" +
            "AND created_at &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null and endTime != \"\"'>" +
            "AND created_at &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "ORDER BY updated_at DESC, id DESC " +
            "LIMIT #{offset}, #{size}" +
            "</script>")
    List<OperationTask> selectByCondition(@Param("taskType") String taskType,
                                          @Param("status") String status,
                                          @Param("targetKeyword") String targetKeyword,
                                          @Param("keyword") String keyword,
                                          @Param("startTime") String startTime,
                                          @Param("endTime") String endTime,
                                          @Param("offset") Integer offset,
                                          @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM operation_tasks " +
            "<where>" +
            "<if test='taskType != null and taskType != \"\"'>" +
            "AND task_type = #{taskType}" +
            "</if>" +
            "<if test='status != null and status != \"\"'>" +
            "AND status = #{status}" +
            "</if>" +
            "<if test='targetKeyword != null and targetKeyword != \"\"'>" +
            "AND (target_type LIKE CONCAT('%', #{targetKeyword}, '%') OR target_id LIKE CONCAT('%', #{targetKeyword}, '%'))" +
            "</if>" +
            "<if test='keyword != null and keyword != \"\"'>" +
            "AND (task_key LIKE CONCAT('%', #{keyword}, '%') OR task_name LIKE CONCAT('%', #{keyword}, '%') OR message LIKE CONCAT('%', #{keyword}, '%'))" +
            "</if>" +
            "<if test='startTime != null and startTime != \"\"'>" +
            "AND created_at &gt;= #{startTime}" +
            "</if>" +
            "<if test='endTime != null and endTime != \"\"'>" +
            "AND created_at &lt;= #{endTime}" +
            "</if>" +
            "</where>" +
            "</script>")
    Integer countByCondition(@Param("taskType") String taskType,
                             @Param("status") String status,
                             @Param("targetKeyword") String targetKeyword,
                             @Param("keyword") String keyword,
                             @Param("startTime") String startTime,
                             @Param("endTime") String endTime);
}
