package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

@FeignClient(name = "mybilibili-message", path = "/message")
public interface MessageClient {

    @PostMapping("/internal/reply-notify")
    Result<?> sendReplyNotification(@RequestParam Integer senderId,
                          @RequestParam Integer receiverId,
                          @RequestParam String content,
                          @RequestParam Integer messageType,
                          @RequestParam(required = false) Integer targetId,
                          @RequestParam(required = false) Integer commentId);
}