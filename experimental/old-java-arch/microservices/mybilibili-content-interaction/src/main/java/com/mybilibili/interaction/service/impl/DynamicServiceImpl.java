package com.mybilibili.interaction.service.impl;

import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.UserDynamic;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.interaction.mapper.UserDynamicMapper;
import com.mybilibili.interaction.mapper.UserMapper;
import com.mybilibili.interaction.mapper.ManuscriptMapper;
import com.mybilibili.interaction.service.DynamicService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class DynamicServiceImpl implements DynamicService {

    @Autowired
    private UserDynamicMapper userDynamicMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String TARGET_TYPE_DYNAMIC = "DYNAMIC";
    private static final String KEY_LIKE_USER_TARGET = "like:user:%s:%s";
    private static final String KEY_LIKE_COUNT = "like:count:%s:%s";
    private static final String KEY_FOLLOWING = "user:following:%s";

    @Override
    public Result<List<DynamicVO>> getAllDynamicList(Integer page, Integer limit, Integer userId) {
        try {
            int offset = (page - 1) * limit;
            List<UserDynamic> dynamicList = userDynamicMapper.getAllDynamics(offset, limit);
            List<DynamicVO> voList = convertToVOList(dynamicList, userId);
            return Result.success("获取成功", voList);
        } catch (Exception e) {
            log.error("获取动态列表失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicVO>> getFollowingDynamicList(Integer userId, Integer page, Integer limit, Integer filterUserId) {
        try {
            int offset = (page - 1) * limit;
            List<UserDynamic> dynamicList;

            if (filterUserId != null) {
                dynamicList = userDynamicMapper.getByUserId(filterUserId, offset, limit);
            } else {
                String key = String.format(KEY_FOLLOWING, userId);
                Set<String> followingIds = redisTemplate.opsForSet().members(key);
                
                if (followingIds == null || followingIds.isEmpty()) {
                    return Result.success("暂无动态", new ArrayList<>());
                }

                List<Integer> followedUserIds = followingIds.stream()
                        .map(Integer::parseInt)
                        .collect(Collectors.toList());

                String userIds = followedUserIds.stream()
                        .map(String::valueOf)
                        .collect(Collectors.joining(","));

                dynamicList = userDynamicMapper.getByUserIds(userIds, offset, limit);
            }

            List<DynamicVO> voList = convertToVOList(dynamicList, userId);
            return Result.success("获取成功", voList);
        } catch (Exception e) {
            log.error("获取关注动态列表失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicVO>> getUserDynamicList(Integer userId, Integer page, Integer limit, Integer currentUserId) {
        try {
            int offset = (page - 1) * limit;
            List<UserDynamic> dynamicList = userDynamicMapper.getByUserId(userId, offset, limit);
            List<DynamicVO> voList = convertToVOList(dynamicList, currentUserId);
            return Result.success("获取成功", voList);
        } catch (Exception e) {
            log.error("获取用户动态列表失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<DynamicVO> getDynamicById(Integer dynamicId, Integer currentUserId) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            DynamicVO vo = convertToVO(dynamic, currentUserId);
            return Result.success("获取成功", vo);
        } catch (Exception e) {
            log.error("获取动态详情失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<DynamicVO> publishDynamic(Integer userId, String content, Integer refVideoId, List<String> imageUrls) {
        try {
            UserDynamic dynamic = new UserDynamic();
            dynamic.setUserId(userId);
            dynamic.setContent(content);
            dynamic.setDynamicType(refVideoId != null ? 2 : 1);
            dynamic.setRefVideoId(refVideoId);
            dynamic.setLikeCount(0);
            dynamic.setCommentCount(0);
            dynamic.setShareCount(0);
            dynamic.setStatus(0);
            dynamic.setCreatedAt(new Date());

            if (imageUrls != null && !imageUrls.isEmpty()) {
                dynamic.setImageUrl(String.join(",", imageUrls));
            }

            userDynamicMapper.insert(dynamic);

            DynamicVO vo = convertToVO(dynamic, userId);
            return Result.success("发布成功", vo);
        } catch (Exception e) {
            log.error("发布动态失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> deleteDynamic(Integer userId, Integer dynamicId) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            if (!dynamic.getUserId().equals(userId)) {
                return Result.error("无权删除此动态");
            }

            userDynamicMapper.deleteById(dynamicId);
            return Result.success("删除成功");
        } catch (Exception e) {
            log.error("删除动态失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> likeDynamic(Integer userId, Integer dynamicId) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            String likeKey = String.format(KEY_LIKE_USER_TARGET, userId, TARGET_TYPE_DYNAMIC);
            Boolean isMember = redisTemplate.opsForSet().isMember(likeKey, String.valueOf(dynamicId));
            
            if (Boolean.TRUE.equals(isMember)) {
                return Result.error("已经点赞过了");
            }

            redisTemplate.opsForSet().add(likeKey, String.valueOf(dynamicId));
            
            String countKey = String.format(KEY_LIKE_COUNT, TARGET_TYPE_DYNAMIC, dynamicId);
            Long newCount = redisTemplate.opsForValue().increment(countKey);
            
            if (newCount != null) {
                userDynamicMapper.updateLikeCount(dynamicId, newCount.intValue());
            }

            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newCount);
            data.put("isLiked", true);
            return Result.success("点赞成功", data);
        } catch (Exception e) {
            log.error("点赞动态失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> unlikeDynamic(Integer userId, Integer dynamicId) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            String likeKey = String.format(KEY_LIKE_USER_TARGET, userId, TARGET_TYPE_DYNAMIC);
            Long removed = redisTemplate.opsForSet().remove(likeKey, String.valueOf(dynamicId));
            
            if (removed == null || removed == 0) {
                return Result.error("尚未点赞");
            }

            String countKey = String.format(KEY_LIKE_COUNT, TARGET_TYPE_DYNAMIC, dynamicId);
            Long newCount = redisTemplate.opsForValue().decrement(countKey);
            
            if (newCount != null && newCount < 0) {
                redisTemplate.opsForValue().set(countKey, "0");
                newCount = 0L;
            }
            
            if (newCount != null) {
                userDynamicMapper.updateLikeCount(dynamicId, newCount.intValue());
            }

            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newCount);
            data.put("isLiked", false);
            return Result.success("取消点赞成功", data);
        } catch (Exception e) {
            log.error("取消点赞动态失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> shareDynamic(Integer userId, Integer dynamicId) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            int newShareCount = dynamic.getShareCount() + 1;
            userDynamicMapper.updateShareCount(dynamicId, newShareCount);
            return Result.success("分享成功");
        } catch (Exception e) {
            log.error("分享动态失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> incrementCommentCount(Integer dynamicId, Integer delta) {
        try {
            UserDynamic dynamic = userDynamicMapper.getById(dynamicId);
            if (dynamic == null) {
                return Result.error("动态不存在");
            }

            int newCount = dynamic.getCommentCount() + delta;
            if (newCount < 0) {
                newCount = 0;
            }
            userDynamicMapper.updateCommentCount(dynamicId, newCount);
            return Result.success("更新成功", newCount);
        } catch (Exception e) {
            log.error("更新评论数量失败", e);
            return Result.error(e.getMessage());
        }
    }

    @Override
    public boolean isLiked(Integer userId, String targetType, Integer targetId) {
        String likeKey = String.format(KEY_LIKE_USER_TARGET, userId, targetType);
        Boolean isMember = redisTemplate.opsForSet().isMember(likeKey, String.valueOf(targetId));
        return Boolean.TRUE.equals(isMember);
    }

    private List<DynamicVO> convertToVOList(List<UserDynamic> dynamics, Integer currentUserId) {
        List<DynamicVO> voList = new ArrayList<>();

        if (dynamics.isEmpty()) {
            return voList;
        }

        for (UserDynamic dynamic : dynamics) {
            DynamicVO vo = convertToVO(dynamic, currentUserId);
            voList.add(vo);
        }
        return voList;
    }

    private DynamicVO convertToVO(UserDynamic dynamic, Integer currentUserId) {
        DynamicVO vo = new DynamicVO();
        vo.setId(dynamic.getId());
        vo.setUserId(dynamic.getUserId());
        vo.setContent(dynamic.getContent());
        vo.setDynamicType(dynamic.getDynamicType());
        vo.setRefManuscriptId(dynamic.getRefManuscriptId());

        if (dynamic.getRefManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectBasicFieldsById(dynamic.getRefManuscriptId());
            if (manuscript != null) {
                DynamicVO.RefVideo refVideo = new DynamicVO.RefVideo();
                refVideo.setId(manuscript.getId());
                refVideo.setTitle(manuscript.getTitle());
                refVideo.setCover(manuscript.getCoverUrl());
                refVideo.setDuration(manuscript.getDuration());
                refVideo.setViewCount(manuscript.getViewCount());
                vo.setRefVideo(refVideo);
            }
        }

        String countKey = String.format(KEY_LIKE_COUNT, TARGET_TYPE_DYNAMIC, dynamic.getId());
        String countStr = redisTemplate.opsForValue().get(countKey);
        vo.setLikeCount(countStr != null ? Integer.parseInt(countStr) : dynamic.getLikeCount());
        
        vo.setCommentCount(dynamic.getCommentCount());
        vo.setShareCount(dynamic.getShareCount());
        vo.setCreatedAt(dynamic.getCreatedAt());
        vo.setStatus(dynamic.getStatus());

        if (dynamic.getImageUrl() != null && !dynamic.getImageUrl().isEmpty()) {
            vo.setImageUrls(Arrays.asList(dynamic.getImageUrl().split(",")));
        } else {
            vo.setImageUrls(new ArrayList<>());
        }

        if (currentUserId != null) {
            vo.setIsLiked(isLiked(currentUserId, TARGET_TYPE_DYNAMIC, dynamic.getId()));
        } else {
            vo.setIsLiked(false);
        }

        User user = userMapper.findById(dynamic.getUserId());
        if (user != null) {
            UserVO userVO = new UserVO();
            userVO.setId(user.getId());
            userVO.setUsername(user.getUsername());
            userVO.setNickname(user.getNickname());
            userVO.setAvatar(user.getAvatar());
            vo.setUser(userVO);
        }

        return vo;
    }
}
