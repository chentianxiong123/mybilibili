package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

@FeignClient(name = "mybilibili-user", path = "/user")
public interface UserClient {

    @GetMapping("/{id}")
    Result<UserVO> getUserById(@PathVariable("id") Integer id);

    @PostMapping("/add-experience")
    Result<?> addExperience(@RequestParam("userId") Integer userId, @RequestParam("experienceAmount") int experienceAmount);
}