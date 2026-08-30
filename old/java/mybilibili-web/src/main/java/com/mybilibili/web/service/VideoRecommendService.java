package com.mybilibili.web.service;

import com.mybilibili.common.vo.VideoRecommendVO;

import java.util.List;

/**
 * 视频推荐服务接口
 */
public interface VideoRecommendService {

    /**
     * 获取相关视频推荐
     * 使用 More Like This 查询，基于标题、描述、标签相似度
     *
     * @param videoId 当前视频ID
     * @param size    推荐数量
     * @return 相关视频列表
     */
    List<VideoRecommendVO> getRelatedVideos(Integer videoId, int size);

    /**
     * 获取热门视频推荐
     * 按播放量排序，支持分类过滤
     *
     * @param categoryId 分类ID（可选，为null时返回全部分类）
     * @param size       推荐数量
     * @return 热门视频列表
     */
    List<VideoRecommendVO> getHotVideos(Integer categoryId, int size);

    /**
     * 获取个性化推荐视频（简化版）
     * 基于用户浏览历史推荐相似标签的视频
     *
     * @param userId 用户ID
     * @param size   推荐数量
     * @return 个性化推荐视频列表
     */
    List<VideoRecommendVO> getRecommendedVideosForUser(Integer userId, int size);
}
