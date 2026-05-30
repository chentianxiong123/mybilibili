package com.mybilibili.user.service;

import com.mybilibili.common.dto.LoginDTO;
import com.mybilibili.common.dto.UserDTO;
import com.mybilibili.common.dto.UserUpdateDTO;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.utils.UploadFilePathUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.user.feign.ManuscriptClient;
import com.mybilibili.user.mapper.UserMapper;
import com.mybilibili.user.mapper.UserTagMapper;
import com.mybilibili.common.entity.UserTag;
import com.mybilibili.user.service.EmailCodeService;
import com.mybilibili.user.service.LoginLogService;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;
import jakarta.servlet.http.HttpServletRequest;
import java.io.File;
import java.io.FileOutputStream;
import java.text.SimpleDateFormat;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Service
public class UserService {

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private UserTagMapper userTagMapper;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Autowired(required = false)
    private ManuscriptClient manuscriptClient;

    @Autowired
    private EmailCodeService emailCodeService;

    @Autowired
    private LoginLogService loginLogService;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private final BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    private static final String LOGIN_FAIL_PREFIX = "login:fail:";
    private static final int MAX_FAIL_COUNT = 5;
    private static final long LOCK_MINUTES = 15;

    public Result<Map<String, Object>> register(UserDTO userDTO) {
        if (userDTO.getEmail() == null || userDTO.getEmail().isBlank()) {
            throw new BusinessException("邮箱不能为空");
        }
        if (userDTO.getEmailCode() == null || userDTO.getEmailCode().isBlank()) {
            throw new BusinessException("邮箱验证码不能为空");
        }
        if (!emailCodeService.verify(userDTO.getEmail(), userDTO.getEmailCode())) {
            throw new BusinessException("验证码错误或已过期");
        }

        User existUser = userMapper.selectByUsername(userDTO.getUsername());
        if (existUser != null) {
            throw new BusinessException("用户名已存在");
        }

        User user = new User();
        user.setUsername(userDTO.getUsername());
        user.setPassword(passwordEncoder.encode(userDTO.getPassword()));
        user.setNickname(userDTO.getNickname() != null ? userDTO.getNickname() : userDTO.getUsername());
        user.setEmail(userDTO.getEmail());
        user.setAvatar("https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png");
        user.setLevel(1);
        user.setFollowingCount(0);
        user.setFollowerCount(0);
        user.setManuscriptCount(0);
        user.setLikedCount(0);
        user.setCoinCount(0);
        user.setExperience(0);
        user.setBio("");
        user.setSignature("");
        user.setStatus(1);

        userMapper.insert(user);

        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);
        Map<String, Object> data = new HashMap<>();
        data.put("user", userVO);
        return Result.success("注册成功", data);
    }

    public Result<Map<String, Object>> login(LoginDTO loginDTO, HttpServletRequest request) {
        if (loginDTO != null) {
            loginDTO.setLoginIp(getClientIp(request));
        }
        return login(loginDTO);
    }

    public Result<Map<String, Object>> login(LoginDTO loginDTO) {
        String loginType = loginDTO.getLoginType();

        if ("email_code".equals(loginType)) {
            // 邮箱验证码登录
            return loginWithEmailCode(loginDTO);
        } else {
            // 用户名/邮箱 + 密码登录
            return loginWithPassword(loginDTO);
        }
    }

    private Result<Map<String, Object>> loginWithEmailCode(LoginDTO loginDTO) {
        String email = loginDTO.getEmail();
        String emailCode = loginDTO.getEmailCode();

        if (email == null || emailCode == null) {
            throw new BusinessException("参数不完整");
        }

        // 校验邮箱验证码
        if (!emailCodeService.verify(email, emailCode)) {
            throw new BusinessException("验证码错误或已过期");
        }

        // 根据邮箱查找用户
        User user = userMapper.selectByEmail(email);
        if (user == null) {
            throw new BusinessException("该邮箱未注册");
        }
        if (user.getStatus() == 0) {
            throw new BusinessException("账号已被禁用，请联系管理员");
        }

        String token = JwtUtils.generateToken(user.getId(), user.getUsername());
        UserVO userVO = getUserVO(user);

        // 记录登录日志
        loginLogService.saveLog(user.getId(), loginDTO.getLoginIp(), null, null, 1);

        Map<String, Object> data = new HashMap<>();
        data.put("user", userVO);
        data.put("token", token);
        data.put("refreshToken", token);
        return Result.success("登录成功", data);
    }

    private Result<Map<String, Object>> loginWithPassword(LoginDTO loginDTO) {
        String username = loginDTO.getUsername();

        User user = null;
        if (username != null && !username.isBlank()) {
            user = userMapper.selectByUsername(username);
            if (user == null) {
                user = userMapper.selectByEmail(username);
            }
        }
        if (user == null) {
            throw new BusinessException("用户名或密码错误");
        }

        String accountKey = LOGIN_FAIL_PREFIX + user.getId();
        String lockKey = accountKey + ":lock";
        if (Boolean.TRUE.equals(redisTemplate.hasKey(lockKey))) {
            throw new BusinessException("账号已锁定，请15分钟后再试");
        }

        if (user.getStatus() == 0) {
            throw new BusinessException("账号已被禁用，请联系管理员");
        }

        if (!passwordEncoder.matches(loginDTO.getPassword(), user.getPassword())) {
            loginLogService.saveLog(user.getId(), loginDTO.getLoginIp(), null, null, 0);
            Long failCount = redisTemplate.opsForValue().increment(accountKey);
            redisTemplate.expire(accountKey, LOCK_MINUTES, TimeUnit.MINUTES);
            if (failCount != null && failCount >= MAX_FAIL_COUNT) {
                redisTemplate.opsForValue().set(lockKey, "1", LOCK_MINUTES, TimeUnit.MINUTES);
                redisTemplate.delete(accountKey);
                throw new BusinessException("连续登录失败次数过多，账号已锁定15分钟");
            }
            throw new BusinessException("用户名或密码错误");
        }

        redisTemplate.delete(accountKey);

        String token = JwtUtils.generateToken(user.getId(), user.getUsername());
        UserVO userVO = getUserVO(user);

        // 记录登录日志
        loginLogService.saveLog(user.getId(), loginDTO.getLoginIp(), null, null, 1);

        Map<String, Object> data = new HashMap<>();
        data.put("user", userVO);
        data.put("token", token);
        data.put("refreshToken", token);
        return Result.success("登录成功", data);
    }

    public Result<UserVO> getUserById(Integer id) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        UserVO userVO = getUserVO(user);
        return Result.success(userVO);
    }

    private String getClientIp(HttpServletRequest request) {
        if (request == null) {
            return null;
        }
        String forwarded = request.getHeader("X-Forwarded-For");
        if (forwarded != null && !forwarded.isBlank()) {
            return forwarded.split(",")[0].trim();
        }
        String realIp = request.getHeader("X-Real-IP");
        if (realIp != null && !realIp.isBlank()) {
            return realIp.trim();
        }
        return request.getRemoteAddr();
    }

    private UserVO getUserVO(User user) {
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);

        if (user.getBirthdate() != null) {
            userVO.setBirthdate(new SimpleDateFormat("yyyy-MM-dd").format(user.getBirthdate()));
        }

        Integer followingCount = userMapper.countFollowing(user.getId());
        Integer followerCount = userMapper.countFollowers(user.getId());
        Integer dynamicCount = userMapper.countDynamics(user.getId());
        Integer totalViewCount = userMapper.sumViewCountByUserId(user.getId());
        Integer totalLikeCount = userMapper.sumLikeCountByUserId(user.getId());

        userVO.setFollowingCount(followingCount != null ? followingCount : 0);
        userVO.setFollowerCount(followerCount != null ? followerCount : 0);
        userVO.setDynamicCount(dynamicCount != null ? dynamicCount : 0);
        userVO.setTotalViewCount(totalViewCount != null ? totalViewCount : 0);
        userVO.setTotalLikeCount(totalLikeCount != null ? totalLikeCount : 0);
        userVO.setCoinCount(user.getCoinCount() != null ? user.getCoinCount() : 0);

        List<UserTag> userTags = userTagMapper.selectByUserId(user.getId());
        List<String> tags = userTags.stream()
                .map(UserTag::getTagName)
                .collect(java.util.stream.Collectors.toList());
        userVO.setTags(tags);

        return userVO;
    }

    public Result<UserVO> updateUser(Integer id, UserUpdateDTO userUpdateDTO) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        if (userUpdateDTO.getNickname() != null) {
            user.setNickname(userUpdateDTO.getNickname());
        }
        if (userUpdateDTO.getAvatar() != null) {
            user.setAvatar(userUpdateDTO.getAvatar());
        }
        if (userUpdateDTO.getEmail() != null) {
            user.setEmail(userUpdateDTO.getEmail());
        }
        if (userUpdateDTO.getPhone() != null) {
            user.setPhone(userUpdateDTO.getPhone());
        }
        if (userUpdateDTO.getGender() != null) {
            user.setGender(userUpdateDTO.getGender());
        }
        if (userUpdateDTO.getBirthdate() != null) {
            try {
                user.setBirthdate(new SimpleDateFormat("yyyy-MM-dd").parse(userUpdateDTO.getBirthdate()));
            } catch (Exception e) {
                throw new BusinessException("出生日期格式错误，应为yyyy-MM-dd");
            }
        }
        if (userUpdateDTO.getSignature() != null) {
            user.setSignature(userUpdateDTO.getSignature());
        }
        if (userUpdateDTO.getBio() != null) {
            user.setBio(userUpdateDTO.getBio());
        }
        if (userUpdateDTO.getAnnouncement() != null) {
            user.setAnnouncement(userUpdateDTO.getAnnouncement());
        }

        userMapper.updateById(user);

        UserVO userVO = getUserVO(user);
        return Result.success("更新成功", userVO);
    }

    public Result<UserVO> uploadAvatar(Integer userId, MultipartFile avatarFile) {
        User user = userMapper.selectById(userId);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        if (avatarFile == null || avatarFile.isEmpty()) {
            throw new BusinessException("请选择要上传的图片");
        }

        if (avatarFile.getSize() > 2 * 1024 * 1024) {
            throw new BusinessException("图片大小不能超过2M");
        }

        String contentType = avatarFile.getContentType();
        if (contentType == null || !(contentType.equals("image/jpeg") || contentType.equals("image/jpg") || contentType.equals("image/png"))) {
            throw new BusinessException("只支持JPG、PNG格式的图片");
        }

        try {
            uploadFilePathUtils.createUserAvatarDirectory(userId);

            String avatarPath = uploadFilePathUtils.getAvatarPath(userId);

            File destFile = new File(avatarPath);
            byte[] bytes = avatarFile.getBytes();
            FileOutputStream fos = new FileOutputStream(destFile);
            fos.write(bytes);
            fos.close();

            String avatarUrl = uploadFilePathUtils.getAvatarUrl(userId);

            user.setAvatar(avatarUrl);
            userMapper.updateById(user);

            UserVO userVO = getUserVO(user);
            return Result.success("头像上传成功", userVO);

        } catch (Exception e) {
            throw new BusinessException("头像上传失败: " + e.getMessage());
        }
    }

    public Result<Map<String, Object>> getUserList(Integer page, Integer size, String keyword) {
        if (page == null || page < 1) {
            page = 1;
        }
        if (size == null || size < 1) {
            size = 10;
        }

        Integer offset = (page - 1) * size;
        List<User> users = userMapper.selectUserList(keyword, offset, size);
        Integer total = userMapper.countUserList(keyword);

        List<UserVO> userVOList = new java.util.ArrayList<>();
        for (User user : users) {
            UserVO userVO = new UserVO();
            BeanUtils.copyProperties(user, userVO);
            if (user.getBirthdate() != null) {
                userVO.setBirthdate(new SimpleDateFormat("yyyy-MM-dd").format(user.getBirthdate()));
            }
            Integer followingCount = userMapper.countFollowing(user.getId());
            Integer followerCount = userMapper.countFollowers(user.getId());
            Integer manuscriptCount = userMapper.countManuscriptsByUserId(user.getId());
            userVO.setFollowingCount(followingCount != null ? followingCount : 0);
            userVO.setFollowerCount(followerCount != null ? followerCount : 0);
            userVO.setManuscriptCount(manuscriptCount != null ? manuscriptCount : 0);
            userVOList.add(userVO);
        }

        Map<String, Object> data = new HashMap<>();
        data.put("list", userVOList);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);

        return Result.success(data);
    }

    public Result<UserVO> getAdminUserById(Integer id) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);
        if (user.getBirthdate() != null) {
            userVO.setBirthdate(new SimpleDateFormat("yyyy-MM-dd").format(user.getBirthdate()));
        }
        Integer followingCount = userMapper.countFollowing(user.getId());
        Integer followerCount = userMapper.countFollowers(user.getId());
        userVO.setFollowingCount(followingCount != null ? followingCount : 0);
        userVO.setFollowerCount(followerCount != null ? followerCount : 0);

        return Result.success(userVO);
    }

    public Result<Void> updateUserStatus(Integer id, Integer status) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        userMapper.updateStatus(id, status);
        return Result.success("状态更新成功", null);
    }

    public Result<Void> resetPassword(Integer id, String newPassword) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        String encodedPassword = passwordEncoder.encode(newPassword);
        userMapper.updatePassword(id, encodedPassword);
        return Result.success("密码重置成功", null);
    }

    public Result<ManuscriptVO> getPinnedVideo(Integer userId) {
        User user = userMapper.selectById(userId);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        Integer pinnedVideoId = user.getPinnedVideoId();
        if (pinnedVideoId == null) {
            return Result.success("该用户暂无置顶视频", null);
        }

        if (manuscriptClient == null) {
            throw new BusinessException("稿件服务暂不可用");
        }

        Result<ManuscriptVO> result = manuscriptClient.getManuscriptById(pinnedVideoId);
        if (result == null || result.getCode() != 200 || result.getData() == null) {
            return Result.success("该用户暂无置顶视频", null);
        }
        return Result.success(result.getData());
    }

    public Result<Void> setPinnedVideo(Integer videoId, Integer currentUserId) {
        if (currentUserId == null) {
            throw new BusinessException("请先登录");
        }

        User user = userMapper.selectById(currentUserId);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        if (manuscriptClient == null) {
            throw new BusinessException("稿件服务暂不可用");
        }

        try {
            Result<ManuscriptVO> result = manuscriptClient.getManuscriptById(videoId);
            if (result == null || result.getCode() != 200) {
                throw new BusinessException("稿件服务调用失败: " + (result != null ? result.getMessage() : "null"));
            }
            ManuscriptVO manuscript = result.getData();
            if (manuscript == null) {
                throw new BusinessException("稿件不存在, id=" + videoId);
            }

            if (manuscript.getUserId() == null || !manuscript.getUserId().equals(currentUserId)) {
                throw new BusinessException("只能置顶自己的稿件");
            }

            user.setPinnedVideoId(videoId);
            userMapper.updateById(user);
            return Result.success("置顶视频设置成功", null);
        } catch (BusinessException e) {
            throw e;
        } catch (Exception e) {
            throw new BusinessException("稿件服务调用异常: " + e.getMessage());
        }
    }

    public Result<Void> removePinnedVideo(Integer currentUserId) {
        if (currentUserId == null) {
            throw new BusinessException("请先登录");
        }

        User user = userMapper.selectById(currentUserId);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        user.setPinnedVideoId(null);
        userMapper.updateById(user);
        return Result.success("已取消置顶视频", null);
    }

    public Result<Void> addExperience(Integer userId, int experienceAmount) {
        if (userId == null) {
            throw new BusinessException("用户ID不能为空");
        }
        userMapper.addExperience(userId, experienceAmount);
        return Result.success("经验值添加成功", null);
    }

    public Result<Void> resetPassword(String email, String code, String newPassword) {
        if (email == null || code == null || newPassword == null) {
            throw new BusinessException("参数不完整");
        }
        if (newPassword.length() < 6 || newPassword.length() > 20) {
            throw new BusinessException("密码长度需在6-20个字符之间");
        }
        // 校验邮箱验证码
        if (!emailCodeService.verify(email, code)) {
            throw new BusinessException("验证码错误或已过期");
        }
        // 根据邮箱查找用户
        User user = userMapper.selectByEmail(email);
        if (user == null) {
            throw new BusinessException("该邮箱未注册");
        }
        // 重置密码
        String encodedPassword = passwordEncoder.encode(newPassword);
        userMapper.updatePassword(user.getId(), encodedPassword);
        return Result.success("密码重置成功", null);
    }
}