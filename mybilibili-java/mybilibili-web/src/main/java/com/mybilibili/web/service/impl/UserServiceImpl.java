package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.UserDTO;
import com.mybilibili.common.dto.UserUpdateDTO;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.UserService;
import com.mybilibili.web.utils.UploadFilePathUtils;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.FileOutputStream;

@Service
public class UserServiceImpl implements UserService {

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    private BCryptPasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    @Override
    public UserVO register(UserDTO userDTO) {
        // 检查用户名是否已存在
        User existingUser = userMapper.findByUsername(userDTO.getUsername());
        if (existingUser != null) {
            throw new RuntimeException("用户名已存在");
        }

        // 创建新用户
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

        // 保存用户
        userMapper.insert(user);

        // 转换为VO返回
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);
        return userVO;
    }

    @Override
    public String login(String username, String password) {
        // 根据用户名查询用户
        User user = userMapper.findByUsername(username);
        if (user == null) {
            throw new RuntimeException("用户名或密码错误");
        }

        // 验证密码
        if (!passwordEncoder.matches(password, user.getPassword())) {
            throw new RuntimeException("用户名或密码错误");
        }

        // 生成JWT令牌
        return JwtUtils.generateToken(user.getId(), user.getUsername());
    }

    @Override
    public UserVO getUserById(Integer id) {
        User user = userMapper.findById(id);
        if (user == null) {
            throw new RuntimeException("用户不存在");
        }

        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);

        // 将 Date 类型的 birthdate 转换为 String
        if (user.getBirthdate() != null) {
            userVO.setBirthdate(new java.text.SimpleDateFormat("yyyy-MM-dd").format(user.getBirthdate()));
        }

        // 实时查询统计数据
        int followingCount = userMapper.countFollowing(id);
        int followerCount = userMapper.countFollowers(id);
        int dynamicCount = userMapper.countDynamics(id);

        // 统计用户所有稿件的总播放数和总获赞数
        Integer totalViewCount = manuscriptMapper.sumViewCountByUserId(id);
        Integer totalLikeCount = manuscriptMapper.sumLikeCountByUserId(id);

        userVO.setFollowingCount(followingCount);
        userVO.setFollowerCount(followerCount);
        userVO.setDynamicCount(dynamicCount);
        userVO.setTotalViewCount(totalViewCount != null ? totalViewCount : 0);
        userVO.setTotalLikeCount(totalLikeCount != null ? totalLikeCount : 0);

        return userVO;
    }

    @Override
    public UserVO updateUser(Integer id, UserUpdateDTO userUpdateDTO) {
        // 检查用户是否存在
        User user = userMapper.findById(id);
        if (user == null) {
            throw new RuntimeException("用户不存在");
        }

        // 更新用户信息
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
                user.setBirthdate(new java.text.SimpleDateFormat("yyyy-MM-dd").parse(userUpdateDTO.getBirthdate()));
            } catch (Exception e) {
                throw new RuntimeException("出生日期格式错误，应为yyyy-MM-dd");
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

        // 保存更新
        userMapper.update(user);

        // 转换为VO返回
        UserVO userVO = new UserVO();
        BeanUtils.copyProperties(user, userVO);

        // 将 Date 类型的 birthdate 转换为 String
        if (user.getBirthdate() != null) {
            userVO.setBirthdate(new java.text.SimpleDateFormat("yyyy-MM-dd").format(user.getBirthdate()));
        }

        return userVO;
    }

    @Override
    public UserVO uploadAvatar(Integer userId, MultipartFile avatarFile) {
        // 检查用户是否存在
        User user = userMapper.findById(userId);
        if (user == null) {
            throw new RuntimeException("用户不存在");
        }

        // 验证文件
        if (avatarFile == null || avatarFile.isEmpty()) {
            throw new RuntimeException("请选择要上传的图片");
        }

        // 验证文件大小（2M）
        if (avatarFile.getSize() > 2 * 1024 * 1024) {
            throw new RuntimeException("图片大小不能超过2M");
        }

        // 验证文件格式
        String contentType = avatarFile.getContentType();
        if (contentType == null || !(contentType.equals("image/jpeg") || contentType.equals("image/jpg") || contentType.equals("image/png"))) {
            throw new RuntimeException("只支持JPG、PNG格式的图片");
        }

        try {
            // 创建用户头像目录
            uploadFilePathUtils.createUserAvatarDirectory(userId);

            // 获取头像保存路径
            String avatarPath = uploadFilePathUtils.getAvatarPath(userId);

            // 保存文件
            File destFile = new File(avatarPath);
            byte[] bytes = avatarFile.getBytes();
            FileOutputStream fos = new FileOutputStream(destFile);
            fos.write(bytes);
            fos.close();

            // 获取头像访问URL
            String avatarUrl = uploadFilePathUtils.getAvatarUrl(userId);

            // 更新用户头像URL
            user.setAvatar(avatarUrl);
            userMapper.update(user);

            // 转换为VO返回
            UserVO userVO = new UserVO();
            BeanUtils.copyProperties(user, userVO);
            return userVO;

        } catch (Exception e) {
            throw new RuntimeException("头像上传失败: " + e.getMessage());
        }
    }
}