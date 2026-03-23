package com.mybilibili.web.service;

import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.vo.MessageVO;
import com.mybilibili.common.vo.ReplyMessageVO;
import com.mybilibili.common.vo.AtMessageVO;
import com.mybilibili.common.vo.LikeMessageVO;
import com.mybilibili.common.vo.SystemNotificationMessageVO;

import java.util.List;
import java.util.Map;

public interface MessageService {

    void sendMessage(Message message);

    MessageVO sendMessage(Integer senderId, SendMessageDTO dto);

    MessageVO getMessageById(Integer id);

    List<MessageVO> getMessagesByUserId(Integer userId, Integer page, Integer size);

    List<MessageVO> getMessagesByConversationId(Integer conversationId, Integer page, Integer size);

    List<MessageVO> getMessagesBetweenUsers(Integer userId, Integer targetUserId, Integer page, Integer size);

    void markMessageAsRead(Integer id);

    void batchMarkMessagesAsRead(List<Integer> ids);

    void deleteMessage(Integer id);

    void clearMessagesByUserId(Integer userId);

    int getUnreadMessageCount(Integer userId);

    // 回复我的消息
    List<ReplyMessageVO> getReplies(Integer userId, Integer page, Integer size);

    // @我的消息
    List<AtMessageVO> getAtList(Integer userId, Integer page, Integer size);

    // 收到的赞
    List<LikeMessageVO> getLikes(Integer userId, Integer page, Integer size);

    // 系统通知
    List<SystemNotificationMessageVO> getSystemNotifications(Integer userId, Integer page, Integer size);

    // 未读消息数统计
    Map<String, Integer> getUnreadCounts(Integer userId);
}
