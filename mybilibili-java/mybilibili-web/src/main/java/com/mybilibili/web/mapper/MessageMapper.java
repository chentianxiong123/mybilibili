package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Message;
import com.mybilibili.common.vo.MessageVO;
import com.mybilibili.common.vo.ReplyMessageVO;
import com.mybilibili.common.vo.AtMessageVO;
import com.mybilibili.common.vo.LikeMessageVO;
import com.mybilibili.common.vo.SystemNotificationMessageVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface MessageMapper {

    int insert(Message message);

    Message selectById(Integer id);

    MessageVO selectVOById(Integer id);

    List<MessageVO> selectByReceiverId(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    List<MessageVO> selectByConversationId(@Param("conversationId") Integer conversationId, @Param("offset") Integer offset, @Param("size") Integer size);

    List<MessageVO> selectBetweenUsers(@Param("userId") Integer userId, @Param("targetUserId") Integer targetUserId, @Param("offset") Integer offset, @Param("size") Integer size);

    int updateReadStatus(@Param("id") Integer id, @Param("isRead") Boolean isRead);

    int batchUpdateReadStatus(@Param("ids") List<Integer> ids, @Param("isRead") Boolean isRead);

    int deleteById(Integer id);

    int deleteByReceiverId(Integer receiverId);

    int countUnreadByReceiverId(Integer receiverId);

    // 回复我的消息
    List<ReplyMessageVO> selectReplies(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadReplies(Integer receiverId);

    // @我的消息
    List<AtMessageVO> selectAtList(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadAt(Integer receiverId);

    // 收到的赞
    List<LikeMessageVO> selectLikes(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadLikes(Integer receiverId);

    // 系统通知
    List<SystemNotificationMessageVO> selectSystemNotifications(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadSystem(Integer receiverId);
}
