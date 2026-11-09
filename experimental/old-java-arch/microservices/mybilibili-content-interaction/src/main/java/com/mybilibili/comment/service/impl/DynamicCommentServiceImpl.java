package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.service.DynamicInteractionPort;
import com.mybilibili.comment.service.DynamicCommentService;
import com.mybilibili.comment.service.LikeInteractionPort;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.vo.DynamicCommentVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.mq.ContentModerationMessage;
import com.mybilibili.mq.ContentReviewMQProducer;
import com.mybilibili.mq.UserEventMQProducer;
import com.mybilibili.mq.UserExperienceEvent;
import com.mybilibili.mq.UserNotificationEvent;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class DynamicCommentServiceImpl implements DynamicCommentService {

    @Autowired
    private DynamicCommentMapper dynamicCommentMapper;

    @Autowired
    private LikeInteractionPort likeInteractionPort;

    @Autowired
    private UserClient userClient;

    @Autowired
    private DynamicInteractionPort dynamicInteractionPort;

    @Autowired
    private UserEventMQProducer userEventMQProducer;

    @Autowired
    private ContentReviewMQProducer contentReviewMQProducer;

    private static final String TARGET_TYPE_DYNAMIC_COMMENT = "DYNAMIC_COMMENT";
    private static final int COMMENT_EXPERIENCE = 5;
    private static final int REPLY_EXPERIENCE = 2;

    private UserVO getUserById(Integer userId) {
        if (userId == null) {
            return null;
        }
        try {
            Result<UserVO> result = userClient.getUserById(userId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            log.error("获取用户信息失败, userId={}, error={}", userId, e.getMessage());
        }
        return null;
    }

    private DynamicCommentVO buildDynamicCommentVO(DynamicComment comment, Integer currentUserId) {
        return buildDynamicCommentVO(comment, currentUserId, Collections.emptyMap(), Collections.emptyMap());
    }

    private DynamicCommentVO buildDynamicCommentVO(DynamicComment comment,
                                                   Integer currentUserId,
                                                   Map<Integer, UserVO> usersById,
                                                   Map<Integer, Boolean> likeStatusMap) {
        DynamicCommentVO vo = new DynamicCommentVO();
        vo.setId(comment.getId());
        vo.setDynamicId(comment.getDynamicId());
        vo.setUserId(comment.getUserId());
        vo.setContent(comment.getContent());
        vo.setParentId(comment.getParentId());
        vo.setReplyUserId(comment.getReplyUserId());
        vo.setReplyToUserId(comment.getReplyUserId());
        vo.setLikeCount(comment.getLikeCount() != null ? comment.getLikeCount() : 0);
        vo.setCreatedAt(comment.getCreatedAt());
        vo.setCreateTime(comment.getCreatedAt());
        vo.setStatus(comment.getStatus());

        UserVO user = resolveUser(comment.getUserId(), usersById);
        String nickname = user != null ? user.getNickname() : null;
        vo.setUserName(isValidNickname(nickname) ? nickname : "用户" + comment.getUserId());
        if (user != null) {
            vo.setUserAvatar(user.getAvatar());
            vo.setUserLevel(user.getLevel());
        }

        if (comment.getReplyUserId() != null) {
            UserVO replyUser = resolveUser(comment.getReplyUserId(), usersById);
            String replyNickname = replyUser != null ? replyUser.getNickname() : null;
            vo.setReplyToUserName(isValidNickname(replyNickname) ? replyNickname : "用户" + comment.getReplyUserId());
        }

        vo.setLiked(resolveLiked(comment.getId(), currentUserId, likeStatusMap));

        return vo;
    }

    private boolean isValidNickname(String nickname) {
        return nickname != null && !nickname.trim().isEmpty() && !"string".equals(nickname.trim());
    }

    @Override
    public Result<DynamicCommentVO> addComment(Integer userId, Integer dynamicId, String content, Integer parentId, Integer replyUserId) {
        if (content == null || content.trim().isEmpty()) {
            return Result.error("评论内容不能为空");
        }

        DynamicComment comment = new DynamicComment();
        comment.setDynamicId(dynamicId);
        comment.setUserId(userId);
        comment.setContent(content.trim());
        comment.setParentId(parentId);
        comment.setReplyUserId(replyUserId);
        comment.setLikeCount(0);
        comment.setStatus(0);  // 0-正常 1-已删除

        dynamicCommentMapper.insert(comment);

        contentReviewMQProducer.sendContentModerationMessage(
                ContentModerationMessage.of(TARGET_TYPE_DYNAMIC_COMMENT, comment.getId(), content.trim()));

        userEventMQProducer.sendExperienceEvent(UserExperienceEvent.of(
                userId, COMMENT_EXPERIENCE, "DYNAMIC_COMMENT", comment.getId(), "评论动态"));

        if (parentId == null) {
            try {
                dynamicInteractionPort.incrementCommentCount(dynamicId, 1);
            } catch (Exception e) {
            }
        }

        return Result.success("评论成功", buildDynamicCommentVO(comment, userId));
    }

    @Override
    public Result<?> deleteComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        if (!comment.getUserId().equals(userId)) {
            return Result.error("没有权限删除此评论");
        }

        int delta = (comment.getParentId() == null) ? -1 : 0;
        dynamicCommentMapper.deleteById(commentId);

        if (delta != 0) {
            try {
                dynamicInteractionPort.incrementCommentCount(comment.getDynamicId(), delta);
            } catch (Exception e) {
            }
        }

        return Result.success("删除成功", null);
    }

    @Override
    public Result<List<DynamicCommentVO>> getCommentsByDynamicId(Integer dynamicId, Integer page, Integer size, String sort, Integer currentUserId) {
        int offset = (page - 1) * size;
        List<DynamicComment> comments = dynamicCommentMapper.selectByDynamicIdWithSort(dynamicId, offset, size, sort);
        Map<Integer, List<DynamicComment>> repliesByCommentId = new HashMap<>();
        List<DynamicComment> allComments = new ArrayList<>(comments);
        for (DynamicComment comment : comments) {
            List<DynamicComment> replies = dynamicCommentMapper.selectRepliesByParentId(comment.getId());
            repliesByCommentId.put(comment.getId(), replies);
            allComments.addAll(replies);
        }
        Map<Integer, UserVO> usersById = fetchUsersByIds(extractUserIds(allComments));
        Map<Integer, Boolean> likeStatusMap = batchIsLiked(currentUserId, allComments);

        List<DynamicCommentVO> voList = new ArrayList<>();
        for (DynamicComment comment : comments) {
            DynamicCommentVO vo = buildDynamicCommentVO(comment, currentUserId, usersById, likeStatusMap);
            List<DynamicCommentVO> replyVOs = new ArrayList<>();
            for (DynamicComment reply : repliesByCommentId.getOrDefault(comment.getId(), Collections.emptyList())) {
                replyVOs.add(buildDynamicCommentVO(reply, currentUserId, usersById, likeStatusMap));
            }
            vo.setReplies(replyVOs);
            vo.setReplyCount(replyVOs.size());
            voList.add(vo);
        }
        return Result.success("获取成功", voList);
    }

    @Override
    public Result<List<DynamicCommentVO>> getRepliesByCommentId(Integer commentId, Integer currentUserId) {
        List<DynamicComment> replies = dynamicCommentMapper.selectRepliesByParentId(commentId);
        Map<Integer, UserVO> usersById = fetchUsersByIds(extractUserIds(replies));
        Map<Integer, Boolean> likeStatusMap = batchIsLiked(currentUserId, replies);
        List<DynamicCommentVO> voList = new ArrayList<>();
        for (DynamicComment reply : replies) {
            voList.add(buildDynamicCommentVO(reply, currentUserId, usersById, likeStatusMap));
        }
        return Result.success("获取成功", voList);
    }

    @Override
    public Result<?> likeComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        Result<?> result = likeInteractionPort.like(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            return Result.error(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        dynamicCommentMapper.updateLikeCountDirect(commentId, newLikeCount);
        userEventMQProducer.sendNotificationEvent(
                UserNotificationEvent.commentLike(userId, comment.getUserId(), commentId, comment.getContent()));
        return Result.success("点赞成功", null);
    }

    @Override
    public Result<?> unlikeComment(Integer userId, Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        Result<?> result = likeInteractionPort.unlike(userId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            return Result.error(result != null ? result.getMessage() : "取消点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_DYNAMIC_COMMENT, commentId);
        dynamicCommentMapper.updateLikeCountDirect(commentId, newLikeCount);
        return Result.success("取消点赞成功", null);
    }

    private Set<Integer> extractUserIds(List<DynamicComment> comments) {
        LinkedHashSet<Integer> userIds = new LinkedHashSet<>();
        if (comments == null) {
            return userIds;
        }
        for (DynamicComment comment : comments) {
            if (comment.getUserId() != null) {
                userIds.add(comment.getUserId());
            }
            if (comment.getReplyUserId() != null) {
                userIds.add(comment.getReplyUserId());
            }
        }
        return userIds;
    }

    private Map<Integer, UserVO> fetchUsersByIds(Collection<Integer> userIds) {
        if (userIds == null || userIds.isEmpty()) {
            return Collections.emptyMap();
        }
        List<Integer> ids = userIds.stream()
                .filter(Objects::nonNull)
                .distinct()
                .collect(Collectors.toList());
        if (ids.isEmpty()) {
            return Collections.emptyMap();
        }
        try {
            Result<List<UserVO>> result = userClient.getUsersByIds(ids);
            if (result == null || result.getCode() == null || result.getCode() != 200 || result.getData() == null) {
                return Collections.emptyMap();
            }
            Map<Integer, UserVO> usersById = new HashMap<>();
            for (UserVO user : result.getData()) {
                if (user != null && user.getId() != null) {
                    usersById.put(user.getId(), user);
                }
            }
            return usersById;
        } catch (Exception e) {
            log.warn("批量获取动态评论用户信息失败: {}", e.getMessage());
            return Collections.emptyMap();
        }
    }

    private Map<Integer, Boolean> batchIsLiked(Integer currentUserId, List<DynamicComment> comments) {
        if (currentUserId == null || comments == null || comments.isEmpty()) {
            return Collections.emptyMap();
        }
        List<Integer> ids = comments.stream()
                .map(DynamicComment::getId)
                .filter(Objects::nonNull)
                .distinct()
                .collect(Collectors.toList());
        if (ids.isEmpty()) {
            return Collections.emptyMap();
        }
        try {
            Result<Map<Integer, Boolean>> result = likeInteractionPort.batchIsLiked(currentUserId, TARGET_TYPE_DYNAMIC_COMMENT, ids);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            log.warn("批量获取动态评论点赞状态失败: {}", e.getMessage());
        }
        return Collections.emptyMap();
    }

    private UserVO resolveUser(Integer userId, Map<Integer, UserVO> usersById) {
        if (userId == null) {
            return null;
        }
        UserVO user = usersById == null ? null : usersById.get(userId);
        return user != null ? user : getUserById(userId);
    }

    private boolean resolveLiked(Integer commentId, Integer currentUserId, Map<Integer, Boolean> likeStatusMap) {
        if (currentUserId == null || commentId == null) {
            return false;
        }
        if (likeStatusMap != null && likeStatusMap.containsKey(commentId)) {
            return Boolean.TRUE.equals(likeStatusMap.get(commentId));
        }
        try {
            Result<Boolean> likeResult = likeInteractionPort.isLiked(currentUserId, TARGET_TYPE_DYNAMIC_COMMENT, commentId);
            return likeResult != null && Boolean.TRUE.equals(likeResult.getData());
        } catch (Exception e) {
            return false;
        }
    }

    private int getLikeCount(String targetType, Integer targetId) {
        try {
            Result<Integer> result = likeInteractionPort.getLikeCount(targetType, targetId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            org.slf4j.LoggerFactory.getLogger(getClass()).error("获取点赞数失败: {}", e.getMessage());
        }
        throw new RuntimeException("获取点赞数失败");
    }
}
