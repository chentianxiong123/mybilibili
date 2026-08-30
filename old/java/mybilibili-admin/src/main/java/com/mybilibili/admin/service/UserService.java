package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;

import java.util.Map;

public interface UserService {
    Result<?> getUserList(Integer page, Integer size, String keyword);
    Result<?> getUserById(Integer id);
    Result<?> updateUserStatus(Integer id, Integer status);
    Result<?> resetPassword(Integer id, String newPassword);
}