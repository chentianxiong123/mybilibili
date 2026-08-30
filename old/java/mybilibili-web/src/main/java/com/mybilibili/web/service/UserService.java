package com.mybilibili.web.service;

import com.mybilibili.common.dto.UserDTO;
import com.mybilibili.common.dto.UserUpdateDTO;
import com.mybilibili.common.vo.UserVO;
import org.springframework.web.multipart.MultipartFile;

public interface UserService {
    // 注册
    UserVO register(UserDTO userDTO);

    // 登录
    String login(String username, String password);

    // 根据ID获取用户信息
    UserVO getUserById(Integer id);

    // 更新用户信息
    UserVO updateUser(Integer id, UserUpdateDTO userUpdateDTO);

    // 上传用户头像
    UserVO uploadAvatar(Integer userId, MultipartFile avatarFile);
}