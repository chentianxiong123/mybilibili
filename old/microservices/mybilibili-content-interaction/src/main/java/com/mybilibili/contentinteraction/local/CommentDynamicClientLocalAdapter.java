package com.mybilibili.contentinteraction.local;

import com.mybilibili.comment.service.DynamicInteractionPort;
import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.service.DynamicService;
import org.springframework.stereotype.Component;

@Component
public class CommentDynamicClientLocalAdapter implements DynamicInteractionPort {

    private final DynamicService dynamicService;

    public CommentDynamicClientLocalAdapter(DynamicService dynamicService) {
        this.dynamicService = dynamicService;
    }

    @Override
    public Result<?> incrementCommentCount(Integer dynamicId, int delta) {
        return dynamicService.incrementCommentCount(dynamicId, delta);
    }
}
