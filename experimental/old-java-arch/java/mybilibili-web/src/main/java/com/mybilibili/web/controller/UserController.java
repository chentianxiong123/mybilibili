package com.mybilibili.web.controller;

import com.mybilibili.common.dto.LoginDTO;
import com.mybilibili.common.dto.UserDTO;
import com.mybilibili.common.dto.UserUpdateDTO;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.service.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/user")
@Tag(name = "用户相关接口", description = "用户注册、登录、获取信息等操作")
public class UserController {

    @Autowired
    private UserService userService;

    @PostMapping("/register")
    @Operation(summary = "用户注册", description = "注册新用户")
    public Result<UserVO> register(@RequestBody UserDTO userDTO) {
        try {
            UserVO userVO = userService.register(userDTO);
            return Result.success("注册成功", userVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/login")
    @Operation(summary = "用户登录", description = "用户登录并获取JWT令牌")
    public Result<Map<String, Object>> login(@RequestBody LoginDTO loginDTO) {
        try {
            String username = loginDTO.getUsername();
            String password = loginDTO.getPassword();
            String token = userService.login(username, password);
            
            // 获取用户信息
            Integer userId = JwtUtils.getUserIdFromToken(token);
            UserVO userVO = userService.getUserById(userId);
            
            Map<String, Object> data = new HashMap<>();
            data.put("user", userVO);
            data.put("token", token);
            data.put("refreshToken", token); // 简化处理，实际应该生成不同的refresh token
            
            return Result.success("登录成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取用户信息", description = "根据用户ID获取用户信息")
    @SecurityRequirement(name = "JWT")
    public Result<UserVO> getUserById(@PathVariable Integer id) {
        try {
            UserVO userVO = userService.getUserById(id);
            return Result.success("获取成功", userVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新用户信息", description = "更新用户个人信息")
    @SecurityRequirement(name = "JWT")
    public Result<UserVO> updateUser(@PathVariable Integer id, @RequestBody UserUpdateDTO userUpdateDTO) {
        try {
            UserVO userVO = userService.updateUser(id, userUpdateDTO);
            return Result.success("更新成功", userVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/avatar")
    @Operation(summary = "上传用户头像", description = "上传用户头像图片，支持JPG、PNG格式，大小不超过2M")
    @SecurityRequirement(name = "JWT")
    public Result<UserVO> uploadAvatar(@PathVariable Integer id, @RequestParam("avatar") MultipartFile avatarFile) {
        try {
            UserVO userVO = userService.uploadAvatar(id, avatarFile);
            return Result.success("头像上传成功", userVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}