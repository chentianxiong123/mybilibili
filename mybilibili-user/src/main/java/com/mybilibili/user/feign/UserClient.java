package com.mybilibili.user.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

/**
 * Legacy standalone user-service self client.
 * <p>
 * In mybilibili-account-social runtime this call is handled by
 * UserSelfClientLocalAdapter and must not be used as a new internal Feign path.
 */
@Deprecated(since = "2026-06-05", forRemoval = false)
@FeignClient(name = "mybilibili-account-social", contextId = "userSelfClient", path = "/user")
public interface UserClient {

    @GetMapping("/{id}")
    Result<UserVO> getUserById(@PathVariable Integer id);
}
