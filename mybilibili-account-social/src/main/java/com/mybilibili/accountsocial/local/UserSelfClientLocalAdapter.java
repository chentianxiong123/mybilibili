package com.mybilibili.accountsocial.local;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.user.feign.UserClient;
import com.mybilibili.user.service.UserService;
import org.springframework.stereotype.Component;

@Component
@SuppressWarnings("deprecation")
public class UserSelfClientLocalAdapter implements UserClient {

    private final UserService userService;

    public UserSelfClientLocalAdapter(UserService userService) {
        this.userService = userService;
    }

    @Override
    public Result<UserVO> getUserById(Integer id) {
        return userService.getUserById(id);
    }
}
