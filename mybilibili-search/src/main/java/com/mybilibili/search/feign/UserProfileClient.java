package com.mybilibili.search.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

import java.util.Map;

@FeignClient(name = "mybilibili-content-interaction", contextId = "searchUserProfileClient", path = "/profile")
public interface UserProfileClient {

    @GetMapping("/{userId}")
    Result<Map<String, Object>> getProfile(@PathVariable("userId") Integer userId);
}
