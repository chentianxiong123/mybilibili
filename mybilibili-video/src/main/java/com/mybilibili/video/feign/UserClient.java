package com.mybilibili.video.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Map;

@FeignClient(name = "mybilibili-user")
public interface UserClient {

    @GetMapping("/user/{id}")
    Result<UserVO> getUserById(@PathVariable("id") Integer id);

    @GetMapping("/follow/check/{targetUserId}")
    Result<Map<String, Object>> checkFollowStatus(
            @PathVariable("targetUserId") Integer targetUserId,
            @RequestHeader("X-User-Id") Integer currentUserId);
}
