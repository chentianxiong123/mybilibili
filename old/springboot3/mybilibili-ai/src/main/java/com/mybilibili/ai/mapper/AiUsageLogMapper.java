package com.mybilibili.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.ai.entity.AiUsageLog;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

@Mapper
public interface AiUsageLogMapper extends BaseMapper<AiUsageLog> {

    @Select("SELECT feature, COUNT(*) as count, " +
            "COALESCE(SUM(total_tokens), 0) as totalTokens, " +
            "COALESCE(AVG(duration_ms), 0) as avgDuration, " +
            "SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) as successCount " +
            "FROM ai_usage_logs GROUP BY feature")
    List<Map<String, Object>> selectFeatureStats();

    @Select("SELECT DATE(created_at) as date, COUNT(*) as count, " +
            "COALESCE(SUM(total_tokens), 0) as totalTokens, " +
            "SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) as successCount " +
            "FROM ai_usage_logs " +
            "WHERE created_at >= #{startDate} " +
            "GROUP BY DATE(created_at) ORDER BY date")
    List<Map<String, Object>> selectDailyStats(@Param("startDate") String startDate);

    @Select("SELECT COUNT(*) as totalCount, " +
            "COALESCE(SUM(total_tokens), 0) as totalTokens, " +
            "COALESCE(AVG(duration_ms), 0) as avgDuration, " +
            "SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) as successCount " +
            "FROM ai_usage_logs")
    Map<String, Object> selectOverview();
}
