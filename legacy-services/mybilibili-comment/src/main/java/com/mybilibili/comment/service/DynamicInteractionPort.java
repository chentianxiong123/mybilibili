package com.mybilibili.comment.service;

import com.mybilibili.common.vo.Result;

public interface DynamicInteractionPort {

    Result<?> incrementCommentCount(Integer dynamicId, int delta);
}
