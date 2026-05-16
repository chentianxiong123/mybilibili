package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.ContentReviewClient;
import com.mybilibili.comment.feign.LikeClient;
import com.mybilibili.comment.feign.ManuscriptClient;
import com.mybilibili.comment.feign.MessageClient;
import com.mybilibili.comment.feign.UserClient;
import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.mapper.ProhibitedWordMapper;
import com.mybilibili.comment.mapper.ReplyMapper;
import com.mybilibili.comment.service.CommentService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.vo.CommentVO;
import com.mybilibili.common.vo.ReplyVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
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
    private LikeClient likeClient;

    @Autowired
    private ContentReviewClient contentReviewClient;

    @Autowired
    private MessageClient messageClient;

    @Autowired
    private ManuscriptClient manuscriptClient;

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    private static final String TARGET_TYPE_COMMENT = "COMMENT";
    private static final String TARGET_TYPE_REPLY = "REPLY";
    private static final int MESSAGE_TYPE_REPLY = 2;
    private static final int COMMENT_EXPERIENCE = 5;
    private static final int REPLY_EXPERIENCE = 2;

    @Override
    public CommentVO addComment(Integer manuscriptId, Integer userId, String content) {
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
            try {
                manuscriptClient.incrementCommentCount(manuscriptId);
            } catch (Exception e) {
                org.slf4j.LoggerFactory.getLogger(getClass()).error("更新稿件评论数失败, manuscriptId={}, error={}", manuscriptId, e.getMessage());
            }
        }

        System.out.println("========== 准备添加经验值，userId=" + userId + ", experience=" + COMMENT_EXPERIENCE + " ==========");
        try {
            Result<?> result = userClient.addExperience(userId, COMMENT_EXPERIENCE);
            System.out.println("========== 经验值调用返回: " + result + " ==========");
            if (result != null && result.getCode() == 200) {
                org.slf4j.LoggerFactory.getLogger(getClass()).info("用户 {} 评论成功，经验值 +{}", userId, COMMENT_EXPERIENCE);
            } else {
                org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加经验值失败，响应: {}", userId, result);
            }
        } catch (Exception e) {
            System.out.println("========== 经验值调用异常: " + e.getMessage() + " ==========");
            org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加经验值失败，原因: {}", userId, e.getMessage());
        }

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

        List<CommentVO> commentVOs = new ArrayList<>();
        for (Comment comment : comments) {
            CommentVO commentVO = buildCommentVO(comment, userId, likeStatusMap, likeCountMap);

            List<Reply> replies = replyMapper.selectByCommentId(comment.getId(), 0, 3);
            List<ReplyVO> replyVOs = new ArrayList<>();

            if (!replies.isEmpty()) {
                List<Integer> replyIds = replies.stream().map(Reply::getId).collect(Collectors.toList());
                Map<Integer, Boolean> replyLikeStatusMap = batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);
                Map<Integer, Integer> replyLikeCountMap = batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);

                for (Reply reply : replies) {
                    replyVOs.add(buildReplyVO(reply, userId, replyLikeStatusMap, replyLikeCountMap));
                }
            }
            commentVO.setReplies(replyVOs);
            commentVOs.add(commentVO);
        }

        return commentVOs;
    }

    @Override
    public CommentVO addCommentByManuscriptId(Integer manuscriptId, Integer userId, String content) {
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
            try {
                manuscriptClient.incrementCommentCount(manuscriptId);
            } catch (Exception e) {
                org.slf4j.LoggerFactory.getLogger(getClass()).error("更新稿件评论数失败, manuscriptId={}, error={}", manuscriptId, e.getMessage());
            }
        }

        System.out.println("========== 准备添加经验值，userId=" + userId + ", experience=" + COMMENT_EXPERIENCE + " ==========");
        try {
            Result<?> result = userClient.addExperience(userId, COMMENT_EXPERIENCE);
            System.out.println("========== 经验值调用返回: " + result + " ==========");
            if (result != null && result.getCode() == 200) {
                System.out.println("========== 经验值添加成功 ==========");
            } else {
                System.out.println("========== 经验值添加失败，响应: " + result + " ==========");
            }
        } catch (Exception e) {
            System.out.println("========== 经验值调用异常: " + e.getMessage() + " ==========");
        }

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

        List<CommentVO> commentVOs = new ArrayList<>();
        for (Comment comment : comments) {
            CommentVO commentVO = buildCommentVO(comment, userId, likeStatusMap, likeCountMap);

            List<Reply> replies = replyMapper.selectByCommentId(comment.getId(), 0, 3);
            List<ReplyVO> replyVOs = new ArrayList<>();

            if (!replies.isEmpty()) {
                List<Integer> replyIds = replies.stream().map(Reply::getId).collect(Collectors.toList());
                Map<Integer, Boolean> replyLikeStatusMap = batchIsLiked(userId, TARGET_TYPE_REPLY, replyIds);
                Map<Integer, Integer> replyLikeCountMap = batchGetLikeCount(TARGET_TYPE_REPLY, replyIds);

                for (Reply reply : replies) {
                    replyVOs.add(buildReplyVO(reply, userId, replyLikeStatusMap, replyLikeCountMap));
                }
            }
            commentVO.setReplies(replyVOs);
            commentVOs.add(commentVO);
        }

        return commentVOs;
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
        }

        System.out.println("========== 准备添加回复经验值，userId=" + userId + ", experience=" + REPLY_EXPERIENCE + " ==========");
        try {
            Result<?> result = userClient.addExperience(userId, REPLY_EXPERIENCE);
            System.out.println("========== 回复经验值调用返回: " + result + " ==========");
            if (result != null && result.getCode() == 200) {
                org.slf4j.LoggerFactory.getLogger(getClass()).info("用户 {} 回复成功，经验值 +{}", userId, REPLY_EXPERIENCE);
            } else {
                org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加回复经验值失败，响应: {}", userId, result);
            }
        } catch (Exception e) {
            System.out.println("========== 回复经验值调用异常: " + e.getMessage() + " ==========");
            org.slf4j.LoggerFactory.getLogger(getClass()).error("用户 {} 添加回复经验值失败，原因: {}", userId, e.getMessage());
        }

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

        List<ReplyVO> replyVOs = new ArrayList<>();
        for (Reply reply : replies) {
            replyVOs.add(buildReplyVO(reply, userId, likeStatusMap, likeCountMap));
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
        Result<?> result = likeClient.like(userId, TARGET_TYPE_COMMENT, commentId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_COMMENT, commentId);
        commentMapper.updateLikeCountDirect(commentId, newLikeCount);
        try {
            messageClient.sendCommentLikeNotification(userId, comment.getUserId(), commentId, comment.getContent());
        } catch (Exception e) {
            org.slf4j.LoggerFactory.getLogger(getClass()).warn("发送评论点赞通知失败: {}", e.getMessage());
        }
    }

    @Override
    public void unlikeComment(Integer commentId, Integer userId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null) {
            throw new RuntimeException("评论不存在");
        }
        Result<?> result = likeClient.unlike(userId, TARGET_TYPE_COMMENT, commentId);
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
        Result<?> result = likeClient.like(userId, TARGET_TYPE_REPLY, replyId);
        if (result == null || result.getCode() != 200) {
            throw new RuntimeException(result != null ? result.getMessage() : "点赞失败");
        }
        int newLikeCount = getLikeCount(TARGET_TYPE_REPLY, replyId);
        replyMapper.updateLikeCountDirect(replyId, newLikeCount);
        try {
            messageClient.sendCommentLikeNotification(userId, reply.getUserId(), reply.getCommentId(), reply.getContent());
        } catch (Exception e) {
            org.slf4j.LoggerFactory.getLogger(getClass()).warn("发送回复点赞通知失败: {}", e.getMessage());
        }
    }

    @Override
    public void unlikeReply(Integer replyId, Integer userId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null) {
            throw new RuntimeException("回复不存在");
        }
        Result<?> result = likeClient.unlike(userId, TARGET_TYPE_REPLY, replyId);
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
            Result<Map<Integer, Boolean>> result = likeClient.batchIsLiked(userId, targetType, targetIds);
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
            Result<Map<Integer, Integer>> result = likeClient.batchGetLikeCount(targetType, targetIds);
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
            Result<Integer> result = likeClient.getLikeCount(targetType, targetId);
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
            Result<Boolean> result = likeClient.isLiked(userId, targetType, targetId);
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
        CommentVO commentVO = new CommentVO();
        commentVO.setId(comment.getId());
        commentVO.setVideoId(comment.getManuscriptId());
        commentVO.setUserId(comment.getUserId());
        commentVO.setContent(comment.getContent());
        commentVO.setLikeCount(comment.getLikeCount() != null ? comment.getLikeCount() : 0);
        commentVO.setReplyCount(comment.getReplyCount());
        commentVO.setCreateTime(comment.getCreatedAt());
        commentVO.setManuscriptId(comment.getManuscriptId());

        UserVO user = getUserById(comment.getUserId());
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
        ReplyVO replyVO = new ReplyVO();
        replyVO.setId(reply.getId());
        replyVO.setCommentId(reply.getCommentId());
        replyVO.setUserId(reply.getUserId());
        replyVO.setContent(reply.getContent());
        replyVO.setLikeCount(reply.getLikeCount() != null ? reply.getLikeCount() : 0);
        replyVO.setCreateTime(reply.getCreatedAt());

        UserVO user = getUserById(reply.getUserId());
        if (user != null) {
            replyVO.setUserName(user.getNickname());
            replyVO.setUserAvatar(user.getAvatar());
            replyVO.setUserLevel(user.getLevel());
        }

        if (reply.getReplyToUserId() != null) {
            UserVO targetUser = getUserById(reply.getReplyToUserId());
            if (targetUser != null) {
                replyVO.setReplyToUserName(targetUser.getNickname());
            }
        }

        replyVO.setLiked(likeStatusMap.getOrDefault(reply.getId(), false));

        return replyVO;
    }

    private List<String> detectProhibitedWords(String content) {
        List<ProhibitedWord> prohibitedWords = prohibitedWordMapper.selectAllEnabled();
        if (prohibitedWords == null || prohibitedWords.isEmpty()) {
            return new ArrayList<>();
        }

        List<String> found = new ArrayList<>();
        for (ProhibitedWord pw : prohibitedWords) {
            if (pw.getMatchType() == null || "CONTAINS".equals(pw.getMatchType())) {
                if (content.contains(pw.getWord())) {
                    found.add(pw.getWord());
                }
            } else if ("EXACT".equals(pw.getMatchType())) {
                if (content.equals(pw.getWord())) {
                    found.add(pw.getWord());
                }
            }
        }
        return found;
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
            messageClient.sendReplyNotification(senderId, receiverId, "回复了你：" + content, MESSAGE_TYPE_REPLY, manuscriptId, commentId);
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