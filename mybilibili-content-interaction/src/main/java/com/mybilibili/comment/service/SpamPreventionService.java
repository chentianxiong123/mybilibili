package com.mybilibili.comment.service;

public interface SpamPreventionService {
    /** 检查用户是否超出发布频率，true=超出限制 */
    boolean isRateLimited(Integer userId, String action);
    /** 记录用户发布行为（发布成功后调用） */
    void recordAction(Integer userId, String action);
    /** 获取剩余可发布次数 */
    long getRemainingCount(Integer userId, String action);
}