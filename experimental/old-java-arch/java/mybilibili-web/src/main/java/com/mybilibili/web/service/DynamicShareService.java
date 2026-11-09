package com.mybilibili.web.service;

import com.mybilibili.common.entity.DynamicShare;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface DynamicShareService {
    
    /**
     * 转发动态
     * @param userId 转发用户ID
     * @param dynamicId 被转发的动态ID
     * @param content 转发评论内容
     * @return 结果
     */
    Result<?> shareDynamic(Integer userId, Integer dynamicId, String content);
    
    /**
     * 获取动态的转发列表
     * @param dynamicId 动态ID
     * @return 转发列表
     */
    Result<List<DynamicShare>> getShareList(Integer dynamicId);
    
    /**
     * 获取用户的转发列表
     * @param userId 用户ID
     * @return 转发列表
     */
    Result<List<DynamicShare>> getUserShareList(Integer userId);
    
    /**
     * 检查用户是否已转发该动态
     * @param userId 用户ID
     * @param dynamicId 动态ID
     * @return 是否已转发
     */
    boolean isShared(Integer userId, Integer dynamicId);
    
    /**
     * 取消转发
     * @param userId 用户ID
     * @param shareId 转发记录ID
     * @return 结果
     */
    Result<?> cancelShare(Integer userId, Integer shareId);
}
