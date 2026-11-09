package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;

import java.util.List;

@FeignClient(name = "mybilibili-account-social", contextId = "commentUserClient", path = "/user")
public interface UserClient {

    @GetMapping("/{id}")
    Result<UserVO> getUserById(@PathVariable("id") Integer id);

    @PostMapping("/batch")
    Result<List<UserVO>> getUsersByIds(@RequestBody List<Integer> ids);
}
