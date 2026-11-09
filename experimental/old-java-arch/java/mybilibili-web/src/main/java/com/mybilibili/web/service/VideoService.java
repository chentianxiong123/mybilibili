package com.mybilibili.web.service;

import com.mybilibili.common.dto.VideoUploadDTO;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.VideoVO;

import java.util.List;

public interface VideoService {
    VideoVO uploadVideo(VideoUploadDTO videoUploadDTO, Integer userId) throws Exception;
    VideoVO getVideoById(Integer id);
    VideoVO getVideoById(Integer id, Integer currentUserId);

    /**
     * 获取视频详情（包含稿件信息）
     *
     * @param id            视频ID
     * @param currentUserId 当前用户ID
     * @param includeManuscript 是否包含稿件信息
     * @return 视频VO
     */
    VideoVO getVideoById(Integer id, Integer currentUserId, boolean includeManuscript);

    List<VideoVO> getVideosByUserId(Integer userId);
    List<VideoVO> getVideosByUserId(Integer userId, String sort);
    List<VideoVO> getVideosByCategoryId(Integer categoryId);
    List<VideoVO> getRecommendedVideos();
    List<VideoVO> getHotVideos();
    void updateViewCount(Integer id);
    List<VideoVO> getVideoList(Integer page, Integer size);
    VideoVO updateVideo(Integer id, VideoUploadDTO videoUploadDTO, Integer userId) throws Exception;
    void deleteVideo(Integer id, Integer userId) throws Exception;
    void transcodeAllVideos();
    void transcodeVideo(Integer id);

    /**
     * 根据稿件ID获取视频（用于稿件详情页）
     *
     * @param manuscriptId 稿件ID
     * @param p            第几个视频（从1开始）
     * @param currentUserId 当前用户ID
     * @return 视频VO（包含稿件下的所有视频列表）
     */
    VideoVO getVideoByManuscriptId(Integer manuscriptId, Integer p, Integer currentUserId);
}