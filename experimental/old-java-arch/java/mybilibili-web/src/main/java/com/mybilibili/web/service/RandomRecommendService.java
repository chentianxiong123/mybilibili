package com.mybilibili.web.service;

import com.mybilibili.common.entity.Manuscript;

import java.util.List;

/**
 * 随机推荐服务接口
 * 提供基于加权随机算法的稿件推荐功能
 */
public interface RandomRecommendService {

    /**
     * 获取加权随机推荐的稿件列表
     * 结合用户历史行为（浏览、点赞）和随机性生成推荐
     *
     * @param userId 用户ID（可为null，表示未登录用户）
     * @param size   推荐数量
     * @return 推荐稿件列表
     */
    List<Manuscript> getRandomRecommendedManuscripts(Integer userId, int size);
}
