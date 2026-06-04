package com.mybilibili.ai.mapper;

import com.mybilibili.ai.entity.VideoPipelineTaskEntity;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Options;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface VideoPipelineTaskMapper {

    @Insert("INSERT INTO video_pipeline_tasks " +
            "(task_key, manuscript_id, video_id, uploader_id, video_title, current_step_index, status, create_time) " +
            "VALUES (#{taskKey}, #{manuscriptId}, #{videoId}, #{uploaderId}, #{videoTitle}, #{currentStepIndex}, #{status}, #{createTime})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(VideoPipelineTaskEntity task);

    @Select("SELECT * FROM video_pipeline_tasks WHERE task_key = #{taskKey} ORDER BY id DESC LIMIT 1")
    VideoPipelineTaskEntity selectLatestByTaskKey(@Param("taskKey") String taskKey);

    @Select("SELECT COUNT(*) FROM video_pipeline_tasks WHERE task_key = #{taskKey} AND status IN ('PENDING', 'RUNNING')")
    int countActiveByTaskKey(@Param("taskKey") String taskKey);

    @Select("SELECT * FROM video_pipeline_tasks WHERE worker_id = #{workerId} AND status = 'RUNNING' ORDER BY locked_until DESC, id ASC LIMIT 1")
    VideoPipelineTaskEntity selectClaimedTask(@Param("workerId") String workerId);

    @Select("SELECT * FROM video_pipeline_tasks WHERE status = #{status} ORDER BY id DESC LIMIT #{limit}")
    List<VideoPipelineTaskEntity> selectByStatus(@Param("status") String status, @Param("limit") int limit);

    @Select("SELECT COUNT(*) FROM video_pipeline_tasks WHERE status = #{status}")
    int countByStatus(@Param("status") String status);

    @Update("UPDATE video_pipeline_tasks SET status = 'RUNNING', worker_id = #{workerId}, locked_until = DATE_ADD(NOW(), INTERVAL #{leaseSeconds} SECOND), start_time = IFNULL(start_time, NOW()), error_message = NULL, failed_step = NULL " +
            "WHERE id = (SELECT id FROM (SELECT id FROM video_pipeline_tasks WHERE status = 'PENDING' ORDER BY create_time ASC, id ASC LIMIT 1) pending_task) " +
            "AND status = 'PENDING'")
    int claimNextPendingTask(@Param("workerId") String workerId, @Param("leaseSeconds") int leaseSeconds);

    @Update("UPDATE video_pipeline_tasks SET status = 'RUNNING', worker_id = #{workerId}, locked_until = DATE_ADD(NOW(), INTERVAL #{leaseSeconds} SECOND), start_time = IFNULL(start_time, NOW()), current_step_index = #{currentStepIndex}, error_message = NULL, failed_step = NULL WHERE id = #{id}")
    int markRunning(@Param("id") Long id,
                    @Param("currentStepIndex") int currentStepIndex,
                    @Param("workerId") String workerId,
                    @Param("leaseSeconds") int leaseSeconds);

    @Update("UPDATE video_pipeline_tasks SET locked_until = DATE_ADD(NOW(), INTERVAL #{leaseSeconds} SECOND) WHERE id = #{id} AND status = 'RUNNING' AND worker_id = #{workerId}")
    int renewLease(@Param("id") Long id,
                   @Param("workerId") String workerId,
                   @Param("leaseSeconds") int leaseSeconds);

    @Update("UPDATE video_pipeline_tasks SET current_step_index = #{currentStepIndex} WHERE id = #{id}")
    int updateCurrentStep(@Param("id") Long id, @Param("currentStepIndex") int currentStepIndex);

    @Update("UPDATE video_pipeline_tasks SET status = 'COMPLETED', current_step_index = #{currentStepIndex}, end_time = NOW(), locked_until = NULL, error_message = NULL, failed_step = NULL WHERE id = #{id}")
    int markCompleted(@Param("id") Long id, @Param("currentStepIndex") int currentStepIndex);

    @Update("UPDATE video_pipeline_tasks SET status = 'FAILED', failed_step = #{failedStep}, error_message = #{errorMessage}, locked_until = NULL, end_time = NOW() WHERE id = #{id}")
    int markFailed(@Param("id") Long id, @Param("failedStep") String failedStep, @Param("errorMessage") String errorMessage);

    @Update("UPDATE video_pipeline_tasks SET status = 'CANCELLED', end_time = NOW() WHERE id = #{id} AND status = 'PENDING'")
    int markCancelled(@Param("id") Long id);

    @Update("UPDATE video_pipeline_tasks SET status = 'FAILED', failed_step = 'LEASE_EXPIRED', error_message = 'AI视频处理任务租约过期，任务已中断', locked_until = NULL, end_time = NOW() WHERE status = 'RUNNING' AND locked_until < NOW()")
    int markExpiredRunningTasksInterrupted();

    @Update("UPDATE video_pipeline_tasks SET status = 'CANCELLED', end_time = NOW() WHERE status = 'PENDING'")
    int cancelPendingTasks();
}
