package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

@FeignClient(name = "mybilibili-user", path = "/user")
public interface UserClient {

    @GetMapping("/{id}")
    Result<UserVO> getUserById(@PathVariable("id") Integer id);
}