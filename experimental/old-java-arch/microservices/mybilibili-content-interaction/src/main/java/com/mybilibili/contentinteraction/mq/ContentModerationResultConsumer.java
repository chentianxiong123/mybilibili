package com.mybilibili.contentinteraction.mq;

import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.mapper.ReplyMapper;
import com.mybilibili.comment.service.DynamicInteractionPort;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.mq.ContentModerationMessage;
import com.mybilibili.mq.ContentModerationResultEvent;
import com.mybilibili.mq.ManuscriptCommentCountEvent;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoMQProducer;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_CONTENT_MODERATION_RESULT,
        consumerGroup = MQConstants.GROUP_CONTENT_MODERATION_RESULT
)
public class ContentModerationResultConsumer implements RocketMQListener<ContentModerationResultEvent> {

    private final CommentMapper commentMapper;
    private final ReplyMapper replyMapper;
    private final DynamicCommentMapper dynamicCommentMapper;
    private final VideoMQProducer videoMQProducer;
    private final DynamicInteractionPort dynamicInteractionPort;

    public ContentModerationResultConsumer(CommentMapper commentMapper,
                                           ReplyMapper replyMapper,
                                           DynamicCommentMapper dynamicCommentMapper,
                                           VideoMQProducer videoMQProducer,
                                           DynamicInteractionPort dynamicInteractionPort) {
        this.commentMapper = commentMapper;
        this.replyMapper = replyMapper;
        this.dynamicCommentMapper = dynamicCommentMapper;
        this.videoMQProducer = videoMQProducer;
        this.dynamicInteractionPort = dynamicInteractionPort;
    }

    @Override
    public void onMessage(ContentModerationResultEvent event) {
        if (event == null || event.getTargetType() == null || event.getTargetId() == null || event.isPassed()) {
            return;
        }

        switch (event.getTargetType()) {
            case ContentModerationMessage.TARGET_COMMENT -> removeComment(event.getTargetId());
            case ContentModerationMessage.TARGET_REPLY -> removeReply(event.getTargetId());
            case ContentModerationMessage.TARGET_DYNAMIC_COMMENT -> removeDynamicComment(event.getTargetId());
            default -> log.warn("忽略未知内容异步审核结果类型: {}", event.getTargetType());
        }
    }

    private void removeComment(Integer commentId) {
        Comment comment = commentMapper.selectById(commentId);
        if (comment == null || comment.getStatus() != null && comment.getStatus() != 0) {
            return;
        }
        comment.setStatus(1);
        commentMapper.updateById(comment);
        if (comment.getManuscriptId() != null) {
            videoMQProducer.sendManuscriptCommentCountEvent(
                    ManuscriptCommentCountEvent.of(comment.getManuscriptId(), -1));
        }
    }

    private void removeReply(Integer replyId) {
        Reply reply = replyMapper.selectById(replyId);
        if (reply == null || !"NORMAL".equals(reply.getStatus())) {
            return;
        }
        reply.setStatus("REMOVED");
        replyMapper.updateById(reply);
        if (reply.getCommentId() != null) {
            commentMapper.updateReplyCount(reply.getCommentId(), -1);
        }
    }

    private void removeDynamicComment(Integer commentId) {
        DynamicComment comment = dynamicCommentMapper.selectById(commentId);
        if (comment == null || comment.getStatus() != null && comment.getStatus() != 0) {
            return;
        }
        comment.setStatus(1);
        dynamicCommentMapper.updateById(comment);
        if (comment.getParentId() == null && comment.getDynamicId() != null) {
            try {
                dynamicInteractionPort.incrementCommentCount(comment.getDynamicId(), -1);
            } catch (Exception e) {
                log.debug("异步审核回滚动态评论数失败: dynamicId={}, error={}", comment.getDynamicId(), e.getMessage());
            }
        }
    }
}
