package com.mybilibili.web.service;

import java.util.List;

/**
 * 视频标签服务接口（Redis版）
 */
public interface VideoTagService {

    /**
     * 为视频添加标签
     *
     * @param videoId 视频ID
     * @param tagName 标签名
     * @return 是否成功
     */
    boolean addTag(Integer videoId, String tagName);

    /**
     * 为视频添加多个标签
     *
     * @param videoId 视频ID
     * @param tagNames 标签名列表
     * @return 是否成功
     */
    boolean addTags(Integer videoId, List<String> tagNames);

    /**
     * 为视频移除标签
     *
     * @param videoId 视频ID
     * @param tagName 标签名
     * @return 是否成功
     */
    boolean removeTag(Integer videoId, String tagName);

    /**
     * 获取视频的所有标签
     *
     * @param videoId 视频ID
     * @return 标签列表
     */
    List<String> getVideoTags(Integer videoId);

    /**
     * 获取多个视频的标签
     *
     * @param videoIds 视频ID列表
     * @return 标签Map（videoId -> 标签列表）
     */
    List<Object> getVideosTags(List<Integer> videoIds);

    /**
     * 获取使用某标签的所有视频ID
     *
     * @param tagName 标签名
     * @return 视频ID列表
     */
    List<Integer> getVideosByTag(String tagName);

    /**
     * 删除视频的所有标签
     *
     * @param videoId 视频ID
     * @return 是否成功
     */
    boolean clearVideoTags(Integer videoId);

    /**
     * 同步视频标签到Redis（从数据库加载）
     *
     * @param videoId 视频ID
     * @param tagNames 标签名列表
     */
    void syncTags(Integer videoId, List<String> tagNames);
}
