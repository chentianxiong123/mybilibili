package com.mybilibili.ai.service;

import com.mybilibili.ai.pipeline.PipelineTask;

import java.util.List;
import java.util.Map;

/**
 * 视频全流程处理服务接口
 */
public interface VideoPipelineService {

    /**
     * 提交全流程处理任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @param uploaderId 上传者ID
     * @return 是否提交成功
     */
    boolean submitPipelineTask(Integer manuscriptId, Integer videoId, Integer uploaderId);

    /**
     * 取消任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @return 是否取消成功
     */
    boolean cancelTask(Integer manuscriptId, Integer videoId);

    /**
     * 获取任务状态
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @return 任务信息
     */
    PipelineTask getTaskStatus(Integer manuscriptId, Integer videoId);

    /**
     * 获取队列统计信息
     *
     * @return 统计信息
     */
    Map<String, Object> getStatistics();

    /**
     * 获取队列中的任务列表
     *
     * @return 任务列表
     */
    List<PipelineTask> getQueuedTasks();

    /**
     * 获取正在处理的任务
     *
     * @return 任务列表
     */
    List<PipelineTask> getProcessingTasks();

    /**
     * 获取已完成的任务
     *
     * @return 任务列表
     */
    List<PipelineTask> getCompletedTasks();

    /**
     * 获取当前正在处理的任务
     *
     * @return 当前任务
     */
    PipelineTask getCurrentTask();

    /**
     * 清空队列
     */
    void clearQueue();

    /**
     * 重试失败的任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @return 是否重试成功
     */
    boolean retryTask(Integer manuscriptId, Integer videoId);
}
