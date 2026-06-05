package com.mybilibili.user.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

import java.util.Map;

@FeignClient(name = "mybilibili-content-interaction", contextId = "userProfileClient", path = "/profile")
public interface UserProfileClient {

    @PostMapping("/init/{userId}")
    Result<?> initProfile(@PathVariable("userId") Integer userId, @RequestBody Map<String, Object> body);
}
