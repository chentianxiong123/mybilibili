package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.UserMapper;
import com.mybilibili.admin.service.UserService;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class UserServiceImpl implements UserService {

    @Autowired
    private UserMapper userMapper;

    private final BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    @Override
    public Result<?> getUserList(Integer page, Integer size, String keyword) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<User> users = userMapper.selectUsersByKeyword(offset, size, keyword);
            int total = userMapper.countUsersByKeyword(keyword);

            List<UserVO> userVOs = new ArrayList<>();
            for (User user : users) {
                UserVO userVO = new UserVO();
                BeanUtils.copyProperties(user, userVO);
                userVOs.add(userVO);
            }

            Map<String, Object> data = new HashMap<>();
            data.put("list", userVOs);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取用户列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getUserById(Integer id) {
        try {
            User user = userMapper.selectById(id);
            if (user == null) {
                return Result.error("用户不存在");
            }
            UserVO userVO = new UserVO();
            BeanUtils.copyProperties(user, userVO);
            return Result.success("获取用户详情成功", userVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> updateUserStatus(Integer id, Integer status) {
        try {
            // 验证用户是否存在
            User user = userMapper.selectById(id);
            if (user == null) {
                return Result.error("用户不存在");
            }
            
            // 更新用户状态
            int result = userMapper.updateStatus(id, status);
            if (result > 0) {
                return Result.success("更新用户状态成功", null);
            } else {
                return Result.error("更新用户状态失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> resetPassword(Integer id, String newPassword) {
        try {
            User user = userMapper.selectById(id);
            if (user == null) {
                return Result.error("用户不存在");
            }

            String encryptedPassword = passwordEncoder.encode(newPassword);
            int result = userMapper.updatePassword(id, encryptedPassword);
            if (result > 0) {
                return Result.success("重置密码成功", null);
            } else {
                return Result.error("重置密码失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}