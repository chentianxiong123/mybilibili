package com.mybilibili.interaction.service;

import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface DynamicService {
    Result<List<DynamicVO>> getAllDynamicList(Integer page, Integer limit, Integer userId);
    Result<List<DynamicVO>> getFollowingDynamicList(Integer userId, Integer page, Integer limit, Integer filterUserId);
    Result<List<DynamicVO>> getUserDynamicList(Integer userId, Integer page, Integer limit, Integer currentUserId);
    Result<DynamicVO> getDynamicById(Integer dynamicId, Integer currentUserId);
    Result<?> deleteDynamic(Integer userId, Integer dynamicId);
    Result<?> likeDynamic(Integer userId, Integer dynamicId);
    Result<?> unlikeDynamic(Integer userId, Integer dynamicId);
    Result<?> shareDynamic(Integer userId, Integer dynamicId);
    Result<?> incrementCommentCount(Integer dynamicId, Integer delta);
    boolean isLiked(Integer userId, String targetType, Integer targetId);
}
