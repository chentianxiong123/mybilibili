package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.mapper.ProhibitedWordMapper;
import com.mybilibili.comment.mapper.ReplyMapper;
import com.mybilibili.comment.service.CommentService;
import com.mybilibili.comment.service.LikeInteractionPort;
import com.mybilibili.comment.service.ProhibitedWordCacheService;
import com.mybilibili.comment.service.SpamPreventionService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.mq.ContentModerationMessage;
import com.mybilibili.mq.ContentReviewMQProducer;
import com.mybilibili.mq.ManuscriptCommentCountEvent;
import com.mybilibili.mq.UserEventMQProducer;
import com.mybilibili.mq.UserExperienceEvent;
import com.mybilibili.mq.UserNotificationEvent;
import com.mybilibili.mq.VideoMQProducer;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

@Service
public class CommentServiceImpl implements CommentService {

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private DynamicCommentMapper dynamicCommentMapper;

    @Autowired
    private UserClient userClient;

    @Autowired
    private LikeInteractionPort likeInteractionPort;

    @Autowired
    private UserEventMQProducer userEventMQProducer;

    @Autowired
    private ContentReviewMQProducer contentReviewMQProducer;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    @Autowired
    private ProhibitedWordCacheService prohibitedWordCacheService;

    @Autowired
    private SpamPreventionService spamPreventionService;

    private static final String TARGET_TYPE_COMMENT = "COMMENT";
    private static final String TARGET_TYPE_REPLY = "REPLY";
    private static final int MESSAGE_TYPE_REPLY = 2;
    private static final int COMMENT_EXPERIENCE = 5;
    private static final int REPLY_EXPERIENCE = 2;

    @Override
    public CommentVO addComment(Integer manuscriptId, Integer userId, String content) {
        if (spamPreventionService.isRateLimited(userId, "comment")) {
            throw new RuntimeException("发布太频繁，请稍后再试");
        }
        List<String> prohibitedWords = detectProhibitedWords(content);
        boolean hasProhibitedWords = !prohibitedWords.isEmpty();

        Comment comment = new Comment();
        comment.setManuscriptId(manuscriptId);
        comment.setUserId(userId);
        comment.setContent(content);
        comment.setLikeCount(0);
        comment.setReplyCount(0);
        comment.setStatus(hasProhibitedWords ? 1 : 0);  // 0-正常 1-已删除

        commentMapper.insert(comment);

        if (!hasProhibitedWords && manuscriptId != null) {
            videoMQProducer.sendManuscriptCommentCountEvent(ManuscriptCommentCountEvent.of(manuscriptId, 1));
        }

        if (!hasProhibitedWords) {
            spamPreventionService.recordAction(userId, "comment");
            contentReviewMQProducer.sendContentModerationMessage(
                    ContentModerationMessage.of(TARGET_TYPE_COMMENT, comment.getId(), content));
        }

        userEventMQProducer.sendExperienceEvent(UserExperienceEvent.of(
                userId, COMMENT_EXPERIENCE, "COMMENT", comment.getId(), "评论稿件"));

        CommentVO commentVO = buildCommentVO(comment, userId);
        commentVO.setHasProhibitedWords(hasProhibitedWords);
        if (hasProhibitedWords) {
            commentVO.setProhibitedWords(prohibitedWords);
        }
        return commentVO;
    }

    @Override
    public List<CommentVO> getCommentsByManuscriptIdOnly(Integer manuscriptId, Integer page, Integer size, Integer userId) {
        int offset = (page - 1) * size;
        List<Comment> comments = commentMapper.selectByManuscriptId(manuscriptId, offset, size, "new");

        if (comments.isEmpty()) {
            return new ArrayList<>();
        }

        List<Integer> commentIds = comments.stream().map(Comment::getId).collect(Collectors.toList());

        Map<Integer, Boolean> likeStatusMap = batchIsLiked(userId, TARGET_TYPE_COMMENT, commentIds);
        Map<Integer, Integer> likeCountMap = batchGetLikeCount(TARGET_TYPE_COMMENT, commentIds);

        return buildCommentThreadVOs(comments, userId, likeStatusMap, likeCountMap);
    }

    @Override
    public CommentVO addCommentByManuscriptId(Integer manuscriptId, Integer userId, String content) {
        if (spamPreventionService.isRateLimited(userId, "comment")) {
            throw new RuntimeException("发布太频繁，请稍后再试");
        }
        List<String> prohibitedWords = detectProhibitedWords(content);
        boolean hasProhibitedWords = !prohibitedWords.isEmpty();

        Comment comment = new Comment();
        comment.setManuscriptId(manuscriptId);
        comment.setUserId(userId);
        comment.setContent(content);
        comment.setLikeCount(0);
        comment.setReplyCount(0);
        comment.setStatus(hasProhibitedWords ? 1 : 0);  // 0-正常 1-已删除

        commentMapper.insert(comment);

        if (!hasProhibitedWords && manuscriptId != null) {
            videoMQProducer.sendManuscriptCommentCountEvent(ManuscriptCommentCountEvent.of(manuscriptId, 1));
        }

        if (!hasProhibitedWords) {
            spamPreventionService.recordAction(userId, "comment");
            contentReviewMQProducer.sendContentModerationMessage(
                    ContentModerationMessage.of(TARGET_TYPE_COMMENT, comment.getId(), content));
        }

        userEventMQProducer.sendExperienceEvent(UserExperienceEvent.of(
                userId, COMMENT_EXPERIENCE, "COMMENT", comment.getId(), "评论稿件"));

        CommentVO commentVO = buildCommentVO(comment, userId);
        commentVO.setHasProhibitedWords(hasProhibitedWords);
        if (hasProhibitedWords) {
            commentVO.setProhibitedWords(prohibitedWords);
        }
        return commentVO;
    }

    @Override
    public List<CommentVO> getCommentsByManuscriptId(Integer manuscriptId, Integer page, Integer size, Integer userId, String sort) {
        int offset = (page - 1) * size;
        List<Comment> comments = commentMapper.selectByManuscriptId(manuscriptId, offset, size, sort);

        if (comments.isEmpty()) {
            return new ArrayList<>();
        }

        List<Integer> commentIds = comments.stream().map(Comment::getId).collect(Collectors.toList());

        Map<Integer, Boolean> likeStatusMap = batchIsLiked(userId, TARGET_TYPE_COMMENT, commentIds);
        Map<Integer, Integer> likeCountMap = batchGetLikeCount(TARGET_TYPE_COMMENT, commentIds);

        return buildCommentThreadVOs(comments, userId, likeStatusMap, likeCountMap);
    }

    @Override
    public void deleteComment(Integer commentId, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        if (!comment.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此评论");
        }
        commentMapper.deleteById(commentId);
    }

    @Override
    public ReplyVO addReply(Integer commentId, Integer userId, String content, Integer replyToUserId) {
        if (spamPreventionService.isRateLimited(userId, "reply")) {
            throw new RuntimeException("发布太频繁，请稍后再试");
        }
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        List<String> prohibitedWords = detectProhibitedWords(content);
        boolean hasProhibitedWords = !prohibitedWords.isEmpty();

        if (replyToUserId != null) {
            UserVO targetUser = getUserById(replyToUserId);
            if (targetUser != null) {
                content = "@" + targetUser.getNickname() + "：" + content;
            }
        }

        Reply reply = new Reply();
        reply.setCommentId(commentId);
        reply.setUserId(userId);
        reply.setReplyToUserId(replyToUserId);
        reply.setContent(content);
        reply.setLikeCount(0);
        reply.setStatus(hasProhibitedWords ? "REMOVED" : "NORMAL");

        replyMapper.insert(reply);
        if (!hasProhibitedWords) {
            commentMapper.updateReplyCount(commentId, 1);
            spamPreventionService.recordAction(userId, "reply");
            contentReviewMQProducer.sendContentModerationMessage(
                    ContentModerationMessage.of(TARGET_TYPE_REPLY, reply.getId(), content));
        }

        userEventMQProducer.sendExperienceEvent(UserExperienceEvent.of(
                userId, REPLY_EXPERIENCE, "REPLY", reply.getId(), "回复评论"));

        if (replyToUserId != null && !replyToUserId.equals(userId)) {
            sendReplyMessage(userId, replyToUserId, content, commentId);
        }

        ReplyVO replyVO = buildReplyVO(reply, userId);
        replyVO.setHasProhibitedWords(hasProhibitedWords);
        if (hasProhibitedWords) {
            replyVO.setProhibitedWords(prohibitedWords);
        }
        return replyVO;
    }

    @Override
    public List<ReplyVO> getRepliesByCommentId(Integer commentId, Integer page, Integer size, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }

        int offset = (page - 1) * size;
        List<Reply> replies = replyMapper.selectByCommentId(commentId, offset, size);

        if (replies.isEmpty()) {
            return new ArrayList<>();
        }

        List<Integer> replyIds = replies.stream().map(Reply::getId).collect(Collectors.toList());
        Map<Integer, Boolean> likeStatusMap = batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);
        Map<Integer, Integer> likeCountMap = batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);
        Map<Integer, UserVO> usersById = fetchUsersByIds(extractReplyUserIds(replies));

        List<ReplyVO> replyVOs = new ArrayList<>();
        for (Reply reply : replies) {
            replyVOs.add(buildReplyVO(reply, userId, likeStatusMap, likeCountMap, usersById));
        }
        return replyVOs;
    }

    @Override
    public void deleteReply(Integer replyId, Integer userId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }
        if (!reply.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此回复");
        }
        replyMapper.deleteById(replyId);
        commentMapper.updateReplyCount(reply.getCommentId(), -1);
    }

    @Override
    public void likeComment(Integer commentId, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        Result<?> result = likeInteractionPort.like(userId, TARGET_TYPE_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_COMMENT, commentId);
        commentMapper.updateLikeCountDirect(commentId, newLikeCount);
        userEventMQProducer.sendNotificationEvent(
                UserNotificationEvent.commentLike(userId, comment.getUserId(), commentId, comment.getContent()));
    }

    @Override
    public void unlikeComment(Integer commentId, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        Result<?> result = likeInteractionPort.unlike(userId, TARGET_TYPE_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "取消点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_COMMENT, commentId);
        commentMapper.updateLikeCountDirect(commentId, newLikeCount);
    }

    @Override
    public void likeReply(Integer replyId, Integer userId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }
        Result<?> result = likeInteractionPort.like(userId, TARGET_TYPE_REPLY, replyId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_REPLY, replyId);
        replyMapper.updateLikeCountDirect(replyId, newLikeCount);
        userEventMQProducer.sendNotificationEvent(
                UserNotificationEvent.commentLike(userId, reply.getUserId(), reply.getCommentId(), reply.getContent()));
    }

    @Override
    public void unlikeReply(Integer replyId, Integer userId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }
        Result<?> result = likeInteractionPort.unlike(userId, TARGET_TYPE_REPLY, replyId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "取消点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_REPLY, replyId);
        replyMapper.updateLikeCountDirect(replyId, newLikeCount);
    }

    @Override
    public Map<Integer, Boolean> batchIsLiked(Integer userId, String targetType, List<Integer> targetIds) {
        if (userId == null || targetIds == null || targetIds.isEmpty()) {
            return new HashMap<>();
        }
        try {
            Result<Map<Integer, Boolean>> result = likeInteractionPort.batchIsLiked(userId, targetType, targetIds);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
        }
        return new HashMap<>();
    }

    @Override
    public Map<Integer, Integer> batchGetLikeCount(String targetType, List<Integer> targetIds) {
        if (targetIds == null || targetIds.isEmpty()) {
            return new HashMap<>();
        }
        try {
            Result<Map<Integer, Integer>> result = likeInteractionPort.batchGetLikeCount(targetType, targetIds);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
        }
        return new HashMap<>();
    }

    @Override
    public int getLikeCount(String targetType, Integer targetId) {
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

    @Override
    public boolean isLiked(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            return false;
        }
        try {
            Result<Boolean> result = likeInteractionPort.isLiked(userId, targetType, targetId);
            if (result != null && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
        }
        return false;
    }

    private CommentVO buildCommentVO(Comment comment, Integer userId) {
        Map<Integer, Boolean> likeStatusMap = new HashMap<>();
        Map<Integer, Integer> likeCountMap = new HashMap<>();

        likeStatusMap.put(comment.getId(), isLiked(userId, TARGET_TYPE_COMMENT, comment.getId()));
        likeCountMap.put(comment.getId(), getLikeCount(TARGET_TYPE_COMMENT, comment.getId()));

        return buildCommentVO(comment, userId, likeStatusMap, likeCountMap);
    }

    private CommentVO buildCommentVO(Comment comment, Integer userId,
                                     Map<Integer, Boolean> likeStatusMap,
                                     Map<Integer, Integer> likeCountMap) {
        return buildCommentVO(comment, userId, likeStatusMap, likeCountMap, Collections.emptyMap());
    }

    private CommentVO buildCommentVO(Comment comment, Integer userId,
                                     Map<Integer, Boolean> likeStatusMap,
                                     Map<Integer, Integer> likeCountMap,
                                     Map<Integer, UserVO> usersById) {
        CommentVO commentVO = new CommentVO();
        commentVO.setId(comment.getId());
        commentVO.setVideoId(comment.getManuscriptId());
        commentVO.setUserId(comment.getUserId());
        commentVO.setContent(comment.getContent());
        commentVO.setLikeCount(comment.getLikeCount() != null ? comment.getLikeCount() : 0);
        commentVO.setReplyCount(comment.getReplyCount());
        commentVO.setCreateTime(comment.getCreatedAt());
        commentVO.setManuscriptId(comment.getManuscriptId());

        UserVO user = resolveUser(comment.getUserId(), usersById);
        if (user != null) {
            commentVO.setUserName(user.getNickname());
            commentVO.setUserAvatar(user.getAvatar());
            commentVO.setUserLevel(user.getLevel());
        }

        commentVO.setLiked(likeStatusMap.getOrDefault(comment.getId(), false));

        return commentVO;
    }

    private ReplyVO buildReplyVO(Reply reply, Integer userId) {
        Map<Integer, Boolean> likeStatusMap = new HashMap<>();
        Map<Integer, Integer> likeCountMap = new HashMap<>();

        likeStatusMap.put(reply.getId(), isLiked(userId, TARGET_TYPE_REPLY, reply.getId()));
        likeCountMap.put(reply.getId(), getLikeCount(TARGET_TYPE_REPLY, reply.getId()));

        return buildReplyVO(reply, userId, likeStatusMap, likeCountMap);
    }

    private ReplyVO buildReplyVO(Reply reply, Integer userId,
                                  Map<Integer, Boolean> likeStatusMap,
                                  Map<Integer, Integer> likeCountMap) {
        return buildReplyVO(reply, userId, likeStatusMap, likeCountMap, Collections.emptyMap());
    }

    private ReplyVO buildReplyVO(Reply reply, Integer userId,
                                  Map<Integer, Boolean> likeStatusMap,
                                  Map<Integer, Integer> likeCountMap,
                                  Map<Integer, UserVO> usersById) {
        ReplyVO replyVO = new ReplyVO();
        replyVO.setId(reply.getId());
        replyVO.setCommentId(reply.getCommentId());
        replyVO.setUserId(reply.getUserId());
        replyVO.setContent(reply.getContent());
        replyVO.setLikeCount(reply.getLikeCount() != null ? reply.getLikeCount() : 0);
        replyVO.setCreateTime(reply.getCreatedAt());

        UserVO user = resolveUser(reply.getUserId(), usersById);
        if (user != null) {
            replyVO.setUserName(user.getNickname());
            replyVO.setUserAvatar(user.getAvatar());
            replyVO.setUserLevel(user.getLevel());
        }

        if (reply.getReplyToUserId() != null) {
            UserVO targetUser = resolveUser(reply.getReplyToUserId(), usersById);
            if (targetUser != null) {
                replyVO.setReplyToUserName(targetUser.getNickname());
            }
        }

        replyVO.setLiked(likeStatusMap.getOrDefault(reply.getId(), false));

        return replyVO;
    }

    private List<CommentVO> buildCommentThreadVOs(List<Comment> comments,
                                                  Integer userId,
                                                  Map<Integer, Boolean> commentLikeStatusMap,
                                                  Map<Integer, Integer> commentLikeCountMap) {
        Map<Integer, List<Reply>> repliesByCommentId = new HashMap<>();
        List<Reply> allReplies = new ArrayList<>();
        LinkedHashSet<Integer> userIds = new LinkedHashSet<>();

        for (Comment comment : comments) {
            if (comment.getUserId() != null) {
                userIds.add(comment.getUserId());
            }
            List<Reply> replies = replyMapper.selectByCommentId(comment.getId(), 0, 3);
            repliesByCommentId.put(comment.getId(), replies);
            allReplies.addAll(replies);
            userIds.addAll(extractReplyUserIds(replies));
        }

        List<Integer> replyIds = allReplies.stream().map(Reply::getId).collect(Collectors.toList());
        Map<Integer, Boolean> replyLikeStatusMap = batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);
        Map<Integer, Integer> replyLikeCountMap = batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);
        Map<Integer, UserVO> usersById = fetchUsersByIds(userIds);

        List<CommentVO> commentVOs = new ArrayList<>();
        for (Comment comment : comments) {
            CommentVO commentVO = buildCommentVO(comment, userId, commentLikeStatusMap, commentLikeCountMap, usersById);
            List<ReplyVO> replyVOs = new ArrayList<>();
            for (Reply reply : repliesByCommentId.getOrDefault(comment.getId(), Collections.emptyList())) {
                replyVOs.add(buildReplyVO(reply, userId, replyLikeStatusMap, replyLikeCountMap, usersById));
            }
            commentVO.setReplies(replyVOs);
            commentVOs.add(commentVO);
        }

        return commentVOs;
    }

    private Set<Integer> extractReplyUserIds(List<Reply> replies) {
        LinkedHashSet<Integer> userIds = new LinkedHashSet<>();
        if (replies == null) {
            return userIds;
        }
        for (Reply reply : replies) {
            if (reply.getUserId() != null) {
                userIds.add(reply.getUserId());
            }
            if (reply.getReplyToUserId() != null) {
                userIds.add(reply.getReplyToUserId());
            }
        }
        return userIds;
    }

    private UserVO resolveUser(Integer userId, Map<Integer, UserVO> usersById) {
        if (userId == null) {
            return null;
        }
        UserVO user = usersById == null ? null : usersById.get(userId);
        return user != null ? user : getUserById(userId);
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
            return Collections.emptyMap();
        }
    }

    private List<String> detectProhibitedWords(String content) {
        return new java.util.ArrayList<>(prohibitedWordCacheService.check(content));
    }

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
        }
        return null;
    }

    private void sendReplyMessage(Integer senderId, Integer receiverId, String content, Integer commentId) {
        try {
            Comment comment = commentMapper.selectById(commentId);
            Integer manuscriptId = null;
            if (comment != null) {
                manuscriptId = comment.getManuscriptId();
            }
            userEventMQProducer.sendNotificationEvent(UserNotificationEvent.reply(
                    senderId, receiverId, "回复了你：" + content, MESSAGE_TYPE_REPLY, manuscriptId, commentId));
        } catch (Exception e) {
        }
    }

    @Override
    public Result<?> getAdminCommentList(Integer page, Integer size, String keyword, Integer manuscriptId) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        int offset = (page - 1) * size;

        List<Comment> comments = commentMapper.selectAdminList(keyword, manuscriptId, offset, size);
        int total = commentMapper.countAdminList(keyword, manuscriptId);

        Map<String, Object> data = new HashMap<>();
        data.put("list", comments);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);

        return Result.success(data);
    }

    @Override
    public Result<?> getAdminCommentById(Integer id) {
        Comment comment = commentMapper.selectById(id);
        if (comment == null) {
            return Result.error("评论不存在");
        }
        return Result.success(comment);
    }

    @Override
    public void deleteCommentByAdmin(Integer id) {
        commentMapper.deleteById(id);
    }

    @Override
    public void updateCommentStatusByAdmin(Integer id, Integer status) {
        Comment comment = commentMapper.selectById(id);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        comment.setStatus(status == 0 ? 0 : 1);  // 0-正常 1-已删除
        commentMapper.updateById(comment);
    }

    @Override
    public Result<?> getPendingList(String contentType, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        int offset = (page - 1) * size;

        List<Map<String, Object>> list = new ArrayList<>();
        int total = 0;

        if (contentType == null || "COMMENT".equals(contentType)) {
            List<Comment> comments = commentMapper.selectPendingComments(offset, size);
            for (Comment comment : comments) {
                Map<String, Object> map = new HashMap<>();
                map.put("id", comment.getId());
                map.put("content", comment.getContent());
                map.put("status", comment.getStatus());
                map.put("createdAt", comment.getCreatedAt());
                map.put("targetType", "COMMENT");
                map.put("targetId", comment.getId());
                UserVO user = getUserById(comment.getUserId());
                if (user != null) {
                    map.put("userId", user.getId());
                    map.put("userName", user.getNickname());
                    map.put("userAvatar", user.getAvatar());
                }
                list.add(map);
            }
            total += commentMapper.countPendingComments();
        }

        if (contentType == null || "REPLY".equals(contentType)) {
            List<Reply> replies = replyMapper.selectPendingReplies(offset, size);
            for (Reply reply : replies) {
                Map<String, Object> map = new HashMap<>();
                map.put("id", reply.getId());
                map.put("content", reply.getContent());
                map.put("status", reply.getStatus());
                map.put("createdAt", reply.getCreatedAt());
                map.put("targetType", "REPLY");
                map.put("targetId", reply.getId());
                map.put("commentId", reply.getCommentId());
                UserVO user = getUserById(reply.getUserId());
                if (user != null) {
                    map.put("userId", user.getId());
                    map.put("userName", user.getNickname());
                    map.put("userAvatar", user.getAvatar());
                }
                list.add(map);
            }
            total += replyMapper.countPendingReplies();
        }

        list.sort((a, b) -> {
            Object timeA = a.get("createdAt");
            Object timeB = b.get("createdAt");
            if (timeA == null || timeB == null) return 0;
            return timeB.toString().compareTo(timeA.toString());
        });

        Map<String, Object> data = new HashMap<>();
        data.put("list", list);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);

        return Result.success(data);
    }

    @Override
    public Result<?> getAllContent(String contentType, String status, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        int offset = (page - 1) * size;

        List<Map<String, Object>> list = new ArrayList<>();
        int total = 0;

        Integer commentStatus = null;
        if ("NORMAL".equals(status)) {
            commentStatus = 0;
        } else if ("REMOVED".equals(status)) {
            commentStatus = 1;
        }

        if (contentType == null || "COMMENT".equals(contentType)) {
            List<Comment> comments = commentMapper.selectAllComments(commentStatus, offset, size);
            for (Comment comment : comments) {
                Map<String, Object> map = new HashMap<>();
                map.put("id", comment.getId());
                map.put("content", comment.getContent());
                map.put("status", comment.getStatus() == 0 ? "NORMAL" : "REMOVED");
                map.put("createdAt", comment.getCreatedAt());
                map.put("targetType", "COMMENT");
                map.put("targetId", comment.getId());
                UserVO user = getUserById(comment.getUserId());
                if (user != null) {
                    map.put("userId", user.getId());
                    map.put("userName", user.getNickname());
                    map.put("userAvatar", user.getAvatar());
                }
                list.add(map);
            }
            total += commentMapper.countAllComments(commentStatus);
        }

        if (contentType == null || "REPLY".equals(contentType)) {
            String replyStatus = null;
            if ("NORMAL".equals(status)) {
                replyStatus = "NORMAL";
            } else if ("REMOVED".equals(status)) {
                replyStatus = "REMOVED";
            }
            List<Reply> replies = replyMapper.selectAllReplies(replyStatus, offset, size);
            for (Reply reply : replies) {
                Map<String, Object> map = new HashMap<>();
                map.put("id", reply.getId());
                map.put("content", reply.getContent());
                map.put("status", reply.getStatus());
                map.put("createdAt", reply.getCreatedAt());
                map.put("targetType", "REPLY");
                map.put("targetId", reply.getId());
                map.put("commentId", reply.getCommentId());
                UserVO user = getUserById(reply.getUserId());
                if (user != null) {
                    map.put("userId", user.getId());
                    map.put("userName", user.getNickname());
                    map.put("userAvatar", user.getAvatar());
                }
                list.add(map);
            }
            total += replyMapper.countAllReplies(replyStatus);
        }

        if (contentType == null || "DYNAMIC_COMMENT".equals(contentType)) {
            List<DynamicComment> dynamicComments = dynamicCommentMapper.selectAllDynamicComments(commentStatus, offset, size);
            for (DynamicComment dc : dynamicComments) {
                Map<String, Object> map = new HashMap<>();
                map.put("id", dc.getId());
                map.put("content", dc.getContent());
                map.put("status", dc.getStatus() == 0 ? "NORMAL" : "REMOVED");
                map.put("createdAt", dc.getCreatedAt());
                map.put("targetType", "DYNAMIC_COMMENT");
                map.put("targetId", dc.getId());
                map.put("dynamicId", dc.getDynamicId());
                UserVO user = getUserById(dc.getUserId());
                if (user != null) {
                    map.put("userId", user.getId());
                    map.put("userName", user.getNickname());
                    map.put("userAvatar", user.getAvatar());
                }
                list.add(map);
            }
            total += dynamicCommentMapper.countAllDynamicComments(commentStatus);
        }

        list.sort((a, b) -> {
            Object timeA = a.get("createdAt");
            Object timeB = b.get("createdAt");
            if (timeA == null || timeB == null) return 0;
            return timeB.toString().compareTo(timeA.toString());
        });

        Map<String, Object> data = new HashMap<>();
        data.put("list", list);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);

        return Result.success(data);
    }

    @Override
    public Result<?> restoreContent(String type, Integer id) {
        try {
            if ("COMMENT".equals(type)) {
                Comment comment = commentMapper.selectById(id);
                if (comment == null) {
                    return Result.error("评论不存在");
                }
                comment.setStatus(0);
                commentMapper.updateById(comment);
            } else if ("REPLY".equals(type)) {
                Reply reply = replyMapper.selectById(id);
                if (reply == null) {
                    return Result.error("回复不存在");
                }
                reply.setStatus("NORMAL");
                replyMapper.updateById(reply);
            } else {
                return Result.error("无效的内容类型");
            }
            return Result.success("恢复成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteContent(String type, Integer id) {
        try {
            if ("COMMENT".equals(type)) {
                Comment comment = commentMapper.selectById(id);
                if (comment == null) {
                    return Result.error("评论不存在");
                }
                comment.setStatus(1);
                commentMapper.updateById(comment);
            } else if ("REPLY".equals(type)) {
                Reply reply = replyMapper.selectById(id);
                if (reply == null) {
                    return Result.error("回复不存在");
                }
                reply.setStatus("REMOVED");
                replyMapper.updateById(reply);
            } else if ("DYNAMIC_COMMENT".equals(type)) {
                DynamicComment dc = dynamicCommentMapper.selectById(id);
                if (dc == null) {
                    return Result.error("动态评论不存在");
                }
                dc.setStatus(1);
                dynamicCommentMapper.updateById(dc);
            } else {
                return Result.error("无效的内容类型");
            }
            return Result.success("下架成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> batchProcess(String action, List<Map<String, Object>> items) {
        if (items == null || items.isEmpty()) {
            return Result.error("没有选择任何内容");
        }

        int successCount = 0;
        int failCount = 0;

        for (Map<String, Object> item : items) {
            String type = (String) item.get("type");
            Integer id = (Integer) item.get("id");

            if (type == null || id == null) {
                failCount++;
                continue;
            }

            try {
                if ("restore".equals(action)) {
                    restoreContent(type, id);
                    successCount++;
                } else if ("delete".equals(action)) {
                    deleteContent(type, id);
                    successCount++;
                }
            } catch (Exception e) {
                failCount++;
            }
        }

        Map<String, Object> result = new HashMap<>();
        result.put("successCount", successCount);
        result.put("failCount", failCount);
        return Result.success("处理完成，成功：" + successCount + "，失败：" + failCount, result);
    }
}
