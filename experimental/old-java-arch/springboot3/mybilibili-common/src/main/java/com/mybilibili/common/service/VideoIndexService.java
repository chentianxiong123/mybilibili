package com.mybilibili.common.service;

import com.mybilibili.common.entity.Video;

import java.util.List;

public interface VideoIndexService {

    /**
     * 索引单个视频
     *
     * @param video 视频实体
     */
    void indexVideo(Video video);

    /**
     * 根据视频ID删除索引
     *
     * @param videoId 视频ID
     */
    void deleteVideo(Integer videoId);

    /**
     * 更新视频索引
     *
     * @param video 视频实体
     */
    void updateVideo(Video video);

    /**
     * 批量索引视频
     *
     * @param videos 视频实体列表
     */
    void bulkIndex(List<Video> videos);
}
