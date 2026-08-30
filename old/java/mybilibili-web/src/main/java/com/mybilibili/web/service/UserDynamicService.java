package com.mybilibili.web.service;

import com.mybilibili.common.entity.UserDynamic;
import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface UserDynamicService {
    Result<?> publishDynamic(Integer userId, String content, String imageUrl, Integer dynamicType, Integer refVideoId, Integer refManuscriptId);
    
    Result<List<DynamicVO>> getUserDynamicList(Integer userId, Integer page, Integer limit, Integer currentUserId);
    
    Result<List<DynamicVO>> getFollowingDynamicList(Integer userId, Integer page, Integer limit, Integer filterUserId);
    
    Result<List<DynamicVO>> getAllDynamicList(Integer page, Integer limit);
    
    Result<?> likeDynamic(Integer userId, Integer dynamicId);
    
    Result<?> unlikeDynamic(Integer userId, Integer dynamicId);
    
    Result<?> shareDynamic(Integer userId, Integer dynamicId);
    
    Result<?> deleteDynamic(Integer userId, Integer dynamicId);
    
    Result<DynamicVO> getDynamicById(Integer dynamicId, Integer currentUserId);
}
