package com.mybilibili.message.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.vo.*;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface MessageMapper extends BaseMapper<Message> {

    int insert(Message message);

    Message selectById(@Param("id") Long id);

    MessageVO selectVOById(@Param("id") Long id);

    List<MessageVO> selectByReceiverId(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    List<MessageVO> selectByConversationId(@Param("conversationId") Long conversationId, @Param("offset") Integer offset, @Param("size") Integer size);

    List<MessageVO> selectBetweenUsers(@Param("userId") Integer userId, @Param("targetUserId") Integer targetUserId, @Param("offset") Integer offset, @Param("size") Integer size);

    int updateReadStatus(@Param("id") Long id, @Param("isRead") Boolean isRead);

    int batchUpdateReadStatus(@Param("ids") List<Long> ids, @Param("isRead") Boolean isRead);

    int deleteById(@Param("id") Long id);

    int deleteByReceiverId(@Param("receiverId") Integer receiverId);

    int countUnreadByReceiverId(@Param("receiverId") Integer receiverId);

    List<ReplyMessageVO> selectReplies(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadReplies(@Param("receiverId") Integer receiverId);

    List<AtMessageVO> selectAtList(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadAt(@Param("receiverId") Integer receiverId);

    List<LikeMessageVO> selectLikes(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadLikes(@Param("receiverId") Integer receiverId);

    List<SystemNotificationMessageVO> selectSystemNotifications(@Param("receiverId") Integer receiverId, @Param("offset") Integer offset, @Param("size") Integer size);

    int countUnreadSystem(@Param("receiverId") Integer receiverId);

}
