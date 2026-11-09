package com.mybilibili.search.service;

import com.mybilibili.common.vo.VideoRecommendVO;

import java.util.List;

public interface VideoRecommendService {

    List<VideoRecommendVO> getRelatedVideos(Integer videoId, int size);

    List<VideoRecommendVO> getHotVideos(Integer categoryId, int size);

    List<VideoRecommendVO> getRecommendedVideosForUser(Integer userId, int size);
}