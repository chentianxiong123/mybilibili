package com.mybilibili.analytics.mapper;

import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.time.LocalDateTime;

@Mapper
public interface VideoProcessAnalyticsMapper {

    @Insert("INSERT INTO video_process_events " +
            "(video_id, manuscript_id, from_status, to_status, stage, progress, error_message, operator_type, operator_id, created_at) " +
            "VALUES (#{videoId}, #{manuscriptId}, #{fromStatus}, #{toStatus}, #{stage}, #{progress}, #{errorMessage}, #{operatorType}, #{operatorId}, #{occurredAt})")
    int insertProcessEvent(@Param("videoId") Integer videoId,
                           @Param("manuscriptId") Integer manuscriptId,
                           @Param("fromStatus") Integer fromStatus,
                           @Param("toStatus") Integer toStatus,
                           @Param("stage") String stage,
                           @Param("progress") Integer progress,
                           @Param("errorMessage") String errorMessage,
                           @Param("operatorType") String operatorType,
                           @Param("operatorId") Integer operatorId,
                           @Param("occurredAt") LocalDateTime occurredAt);
}
