package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.UserDynamic;
import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.UserDynamicMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.FollowService;
import com.mybilibili.web.service.LikeService;
import com.mybilibili.web.service.UserDynamicService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class UserDynamicServiceImpl implements UserDynamicService {

    @Autowired
    private UserDynamicMapper userDynamicMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private LikeService likeService;

    @Autowired
    private FollowService followService;

    private static final String TARGET_TYPE_DYNAMIC = "DYNAMIC";

    @Override
    @Transactional
    public Result<?> publishDynamic(Integer userId, String content, String imageUrl, Integer dynamicType, Integer refVideoId, Integer refManuscriptId) {
        try {
            UserDynamic dynamic = new UserDynamic();
            dynamic.setUserId(userId);
            dynamic.setContent(content);
            dynamic.setImageUrl(imageUrl);
            dynamic.setDynamicType(dynamicType != null ? dynamicType : 0);
            dynamic.setRefVideoId(refVideoId);
            dynamic.setRefManuscriptId(refManuscriptId);
            dynamic.setLikeCount(0);
            dynamic.setCommentCount(0);
            dynamic.setShareCount(0);
            dynamic.setCreatedAt(new Date());
            dynamic.setStatus(0);

            int result = userDynamicMapper.insert(dynamic);
            if (result > 0) {
                return Result.success("发布成功", dynamic);
            } else {
                return Result.error("发布失败");
            }
        } catch (Exception e) {
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
                // 获取关注列表
                List<UserVO> followingList = followService.getFollowingList(userId);
                List<Integer> followedUserIds = new ArrayList<>();
                for (UserVO user : followingList) {
                    followedUserIds.add(user.getId());
                }

                if (followedUserIds.isEmpty()) {
                    return Result.success("暂无动态", new ArrayList<>());
                }
                dynamicList = userDynamicMapper.getByUserIds(followedUserIds, offset, limit);
            }

            List<DynamicVO> voList = convertToVOList(dynamicList, userId);
            return Result.success("获取成功", voList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicVO>> getAllDynamicList(Integer page, Integer limit) {
        try {
            int offset = (page - 1) * limit;
            List<UserDynamic> dynamicList = userDynamicMapper.getAllDynamics(offset, limit);
            List<DynamicVO> voList = convertToVOList(dynamicList, null);
            return Result.success("获取成功", voList);
        } catch (Exception e) {
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

            // 使用新的统一点赞服务
            boolean result = likeService.like(userId, TARGET_TYPE_DYNAMIC, dynamicId);
            if (!result) {
                return Result.error("已经点赞过了");
            }

            // 更新动态点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_DYNAMIC, dynamicId);
            userDynamicMapper.updateLikeCount(dynamicId, newLikeCount);

            // 返回最新的点赞数和状态
            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newLikeCount);
            data.put("isLiked", true);
            return Result.success("点赞成功", data);
        } catch (Exception e) {
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

            // 使用新的统一点赞服务
            boolean result = likeService.unlike(userId, TARGET_TYPE_DYNAMIC, dynamicId);
            if (!result) {
                return Result.error("尚未点赞");
            }

            // 更新动态点赞数
            int newLikeCount = likeService.getLikeCount(TARGET_TYPE_DYNAMIC, dynamicId);
            userDynamicMapper.updateLikeCount(dynamicId, newLikeCount);

            // 返回最新的点赞数和状态
            Map<String, Object> data = new HashMap<>();
            data.put("likeCount", newLikeCount);
            data.put("isLiked", false);
            return Result.success("取消点赞成功", data);
        } catch (Exception e) {
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
            return Result.error(e.getMessage());
        }
    }

    private List<DynamicVO> convertToVOList(List<UserDynamic> dynamics, Integer currentUserId) {
        List<DynamicVO> voList = new ArrayList<>();

        if (dynamics.isEmpty()) {
            return voList;
        }

        // 批量获取动态ID列表
        List<Integer> dynamicIds = new ArrayList<>();
        for (UserDynamic dynamic : dynamics) {
            dynamicIds.add(dynamic.getId());
        }

        // 批量查询点赞状态
        java.util.Map<Integer, Boolean> likeStatusMap = new java.util.HashMap<>();
        if (currentUserId != null) {
            likeStatusMap = likeService.batchIsLiked(currentUserId, TARGET_TYPE_DYNAMIC, dynamicIds);
        } else {
            for (Integer id : dynamicIds) {
                likeStatusMap.put(id, false);
            }
        }

        // 批量查询点赞数
        java.util.Map<Integer, Integer> likeCountMap = likeService.batchGetLikeCount(TARGET_TYPE_DYNAMIC, dynamicIds);

        for (UserDynamic dynamic : dynamics) {
            DynamicVO vo = convertToVO(dynamic, currentUserId, likeStatusMap, likeCountMap);
            voList.add(vo);
        }
        return voList;
    }

    private DynamicVO convertToVO(UserDynamic dynamic, Integer currentUserId) {
        java.util.Map<Integer, Boolean> likeStatusMap = new java.util.HashMap<>();
        java.util.Map<Integer, Integer> likeCountMap = new java.util.HashMap<>();

        if (currentUserId != null) {
            likeStatusMap.put(dynamic.getId(), likeService.isLiked(currentUserId, TARGET_TYPE_DYNAMIC, dynamic.getId()));
        } else {
            likeStatusMap.put(dynamic.getId(), false);
        }
        likeCountMap.put(dynamic.getId(), likeService.getLikeCount(TARGET_TYPE_DYNAMIC, dynamic.getId()));

        return convertToVO(dynamic, currentUserId, likeStatusMap, likeCountMap);
    }

    private DynamicVO convertToVO(UserDynamic dynamic, Integer currentUserId,
                                   java.util.Map<Integer, Boolean> likeStatusMap,
                                   java.util.Map<Integer, Integer> likeCountMap) {
        DynamicVO vo = new DynamicVO();
        vo.setId(dynamic.getId());
        vo.setUserId(dynamic.getUserId());
        vo.setContent(dynamic.getContent());
        vo.setDynamicType(dynamic.getDynamicType());
        vo.setRefVideoId(dynamic.getRefVideoId());
        vo.setRefManuscriptId(dynamic.getRefManuscriptId());

        // 使用批量查询的点赞数
        vo.setLikeCount(likeCountMap.getOrDefault(dynamic.getId(), 0));
        vo.setCommentCount(dynamic.getCommentCount());
        vo.setShareCount(dynamic.getShareCount());
        vo.setCreatedAt(dynamic.getCreatedAt());
        vo.setStatus(dynamic.getStatus());

        if (dynamic.getImageUrl() != null && !dynamic.getImageUrl().isEmpty()) {
            vo.setImageUrls(Arrays.asList(dynamic.getImageUrl().split(",")));
        } else {
            vo.setImageUrls(new ArrayList<>());
        }

        // 使用批量查询的点赞状态
        vo.setIsLiked(likeStatusMap.getOrDefault(dynamic.getId(), false));

        User user = userMapper.findById(dynamic.getUserId());
        if (user != null) {
            UserVO userVO = new UserVO();
            userVO.setId(user.getId());
            userVO.setUsername(user.getUsername());
            userVO.setAvatar(user.getAvatar());
            vo.setUser(userVO);
        }

        // 查询引用稿件信息
        if (dynamic.getRefManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectById(dynamic.getRefManuscriptId());
            if (manuscript != null) {
                DynamicVO.VideoRefVO videoRefVO = new DynamicVO.VideoRefVO();
                videoRefVO.setId(manuscript.getId());
                videoRefVO.setTitle(manuscript.getTitle());
                videoRefVO.setCover(manuscript.getCoverUrl());
                videoRefVO.setDuration(manuscript.getDurationSeconds() != null ? String.valueOf(manuscript.getDurationSeconds()) : null);
                videoRefVO.setViewCount(manuscript.getViewCount());
                vo.setRefVideo(videoRefVO);
            }
        }

        return vo;
    }
}
