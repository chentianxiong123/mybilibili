package com.mybilibili.video.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

@FeignClient(name = "mybilibili-message")
public interface MessageClient {

    @PostMapping("/message/notify/system")
    Result<?> sendSystemNotification(
            @RequestParam("userId") Integer userId,
            @RequestParam("title") String title,
            @RequestParam("content") String content
    );
}
