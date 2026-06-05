package com.mybilibili.message.service;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;

public interface UserLookupPort {

    Result<UserVO> getUserById(Integer id);
}
