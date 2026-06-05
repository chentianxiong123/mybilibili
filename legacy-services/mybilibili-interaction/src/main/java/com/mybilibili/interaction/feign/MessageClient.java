package com.mybilibili.interaction.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

@FeignClient(name = "mybilibili-account-social", contextId = "interactionMessageClient")
public interface MessageClient {

    @PostMapping("/message/notify/like")
    Result<?> sendLikeNotification(
            @RequestParam("senderId") Integer senderId,
            @RequestParam("receiverId") Integer receiverId,
            @RequestParam("videoId") Integer videoId,
            @RequestParam("videoTitle") String videoTitle
    );

    @PostMapping("/message/notify/system")
    Result<?> sendSystemNotification(
            @RequestParam("userId") Integer userId,
            @RequestParam("title") String title,
            @RequestParam("content") String content
    );
}
