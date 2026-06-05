package com.mybilibili.message.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

/**
 * Legacy standalone message-service client.
 * <p>
 * In mybilibili-account-social runtime, message code must use UserLookupPort.
 * This Feign contract must not be used as a new internal call path.
 */
@Deprecated(since = "2026-06-05", forRemoval = false)
@FeignClient(name = "mybilibili-account-social", contextId = "messageUserClient")
public interface UserClient {

    @GetMapping("/user/{id}")
    Result<UserVO> getUserById(@PathVariable("id") Integer id);
}
