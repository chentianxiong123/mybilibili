package com.mybilibili.accountsocial.local;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.message.service.UserLookupPort;
import com.mybilibili.user.service.UserService;
import org.springframework.stereotype.Component;

@Component
public class MessageUserClientLocalAdapter implements UserLookupPort {

    private final UserService userService;

    public MessageUserClientLocalAdapter(UserService userService) {
        this.userService = userService;
    }

    @Override
    public Result<UserVO> getUserById(Integer id) {
        return userService.getUserById(id);
    }
}
