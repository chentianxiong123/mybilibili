package com.mybilibili.user.controller;

import com.mybilibili.common.dto.LoginDTO;
import com.mybilibili.common.dto.UserDTO;
import com.mybilibili.common.dto.UserUpdateDTO;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.user.service.EmailCodeService;
import com.mybilibili.user.service.LoginLogService;
import com.mybilibili.user.service.UserService;
import com.mybilibili.user.entity.LoginLog;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/user")
@Tag(name = "用户管理", description = "用户注册、登录、信息管理接口")
public class UserController {

    @Autowired
    private UserService userService;

    @Autowired
    private EmailCodeService emailCodeService;

    @Autowired
    private LoginLogService loginLogService;

    @PostMapping("/register")
    @Operation(summary = "用户注册", description = "新用户注册，创建账号")
    public Result<Map<String, Object>> register(@RequestBody UserDTO userDTO) {
        return userService.register(userDTO);
    }

    @PostMapping("/login")
    @Operation(summary = "用户登录", description = "用户名密码登录，返回JWT令牌")
    public Result<Map<String, Object>> login(@RequestBody LoginDTO loginDTO, HttpServletRequest request) {
        return userService.login(loginDTO, request);
    }

    @PostMapping("/token/refresh")
    @Operation(summary = "刷新令牌", description = "使用refreshToken获取新的accessToken")
    public Result<Map<String, Object>> refreshToken(@RequestBody Map<String, String> body) {
        String refreshToken = body.get("refreshToken");
        if (refreshToken == null || refreshToken.isBlank()) {
            return Result.error("refreshToken不能为空");
        }
        return userService.refreshToken(refreshToken);
    }

    @PostMapping("/email/code")
    @Operation(summary = "发送邮箱验证码（用于注册等）")
    public Result<Void> sendEmailCode(@RequestBody Map<String, String> request) {
        String email = request.get("email");
        if (email == null || email.isBlank()) return Result.error("邮箱不能为空");
        emailCodeService.sendCode(email);
        return Result.success("验证码已发送", null);
    }

    @PostMapping("/email/verify")
    @Operation(summary = "校验邮箱验证码")
    public Result<Boolean> verifyEmailCode(@RequestBody Map<String, String> request) {
        String email = request.get("email");
        String code = request.get("code");
        if (email == null || code == null) return Result.error("参数不完整");
        boolean ok = emailCodeService.verify(email, code);
        return Result.success(ok);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取用户信息", description = "根据用户ID获取用户详细信息，包含关注数、粉丝数、稿件统计等")
    public Result<UserVO> getUserById(
            @Parameter(description = "用户ID") @PathVariable Integer id) {
        return userService.getUserById(id);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新用户信息", description = "修改当前登录用户的个人信息（昵称、头像、签名等）")
    public Result<UserVO> updateUser(
            @Parameter(description = "用户ID") @PathVariable Integer id,
            @RequestBody UserUpdateDTO userUpdateDTO) {
        return userService.updateUser(id, userUpdateDTO);
    }

    @PostMapping("/{id}/avatar")
    @Operation(summary = "上传用户头像", description = "上传并更新用户头像图片，支持JPG/PNG，最大2MB")
    public Result<UserVO> uploadAvatar(
            @Parameter(description = "用户ID") @PathVariable Integer id,
            @RequestParam("avatar") MultipartFile avatarFile) {
        return userService.uploadAvatar(id, avatarFile);
    }

    @GetMapping("/admin/list")
    @Operation(summary = "获取用户列表（管理员）", description = "分页查询用户列表，支持关键词搜索用户名或昵称")
    public Result<Map<String, Object>> getUserList(
            @Parameter(description = "页码，默认1") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量，默认10") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "搜索关键词（用户名或昵称）") @RequestParam(required = false) String keyword) {
        return userService.getUserList(page, size, keyword);
    }

    @GetMapping("/admin/{id}")
    @Operation(summary = "获取用户详情（管理员）", description = "管理员查看指定用户的详细信息")
    public Result<UserVO> getAdminUserById(
            @Parameter(description = "用户ID") @PathVariable Integer id) {
        return userService.getAdminUserById(id);
    }

    @PutMapping("/admin/{id}/status")
    @Operation(summary = "更新用户状态（管理员）", description = "管理员修改用户状态（正常/封禁等）")
    public Result<Void> updateUserStatus(
            @Parameter(description = "用户ID") @PathVariable Integer id,
            @Parameter(description = "状态值") @RequestParam Integer status) {
        return userService.updateUserStatus(id, status);
    }

    @PutMapping("/admin/{id}/password")
    @Operation(summary = "重置用户密码（管理员）", description = "管理员重置指定用户的密码")
    public Result<Void> resetPassword(
            @Parameter(description = "用户ID") @PathVariable Integer id,
            @Parameter(description = "新密码") @RequestParam String newPassword) {
        return userService.resetPassword(id, newPassword);
    }

    @GetMapping("/{userId}/pinned-video")
    @Operation(summary = "获取用户置顶视频", description = "获取指定用户的置顶视频")
    public Result<ManuscriptVO> getPinnedVideo(
            @Parameter(description = "用户ID") @PathVariable Integer userId) {
        return userService.getPinnedVideo(userId);
    }

    @PostMapping("/pinned-video")
    @Operation(summary = "设置置顶视频", description = "设置当前登录用户的置顶视频")
    public Result<Void> setPinnedVideo(
            @RequestBody Map<String, Integer> request,
            @RequestHeader(value = "X-User-Id", required = false) Integer currentUserId) {
        return userService.setPinnedVideo(request.get("videoId"), currentUserId);
    }

    @DeleteMapping("/pinned-video")
    @Operation(summary = "取消置顶视频", description = "取消当前登录用户的置顶视频")
    public Result<Void> removePinnedVideo(
            @RequestHeader(value = "X-User-Id", required = false) Integer currentUserId) {
        return userService.removePinnedVideo(currentUserId);
    }

    @PostMapping("/add-experience")
    @Operation(summary = "添加经验值", description = "为用户添加经验值")
    public Result<Void> addExperience(
            @RequestParam Integer userId,
            @RequestParam int experienceAmount) {
        return userService.addExperience(userId, experienceAmount);
    }

    @PostMapping("/password/forgot")
    @Operation(summary = "忘记密码", description = "通过邮箱验证码重置密码")
    public Result<Void> forgotPassword(@RequestBody Map<String, String> request) {
        String email = request.get("email");
        String code = request.get("code");
        String newPassword = request.get("newPassword");
        return userService.resetPassword(email, code, newPassword);
    }

    @GetMapping("/login-logs")
    @Operation(summary = "获取登录日志", description = "获取当前用户的登录日志记录")
    public Result<Map<String, Object>> getLoginLogs(
            @RequestHeader(value = "X-User-Id", required = false) Integer currentUserId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size) {
        if (currentUserId == null) return Result.error("请先登录");
        List<LoginLog> logs = loginLogService.getUserLogs(currentUserId, page, size);
        Integer total = loginLogService.countUserLogs(currentUserId);
        Map<String, Object> data = new HashMap<>();
        data.put("list", logs);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);
        return Result.success(data);
    }

    @GetMapping("/default-avatar")
    @Operation(summary = "获取默认头像", description = "返回默认头像图片")
    public void getDefaultAvatar(HttpServletRequest request, HttpServletResponse response) throws IOException {
        response.setContentType("image/svg+xml");
        response.setCharacterEncoding("UTF-8");
        response.setStatus(200);
        String name = request.getParameter("name");
        if (name == null || name.isEmpty()) {
            name = "User";
        }
        String svg = "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"128\" height=\"128\" viewBox=\"0 0 128 128\"><rect width=\"128\" height=\"128\" fill=\"#0D8ABC\" rx=\"64\"/><text x=\"64\" y=\"72\" font-family=\"Arial, sans-serif\" font-size=\"48\" fill=\"white\" text-anchor=\"middle\" dominant-baseline=\"middle\">" + name.substring(0, Math.min(name.length(), 2)).toUpperCase() + "</text></svg>";
        response.getWriter().write(svg);
    }
}