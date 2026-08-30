package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.vo.ConversationVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface ConversationMapper {

    int insert(Conversation conversation);

    Conversation selectById(Integer id);

    Conversation selectByUserAndTarget(@Param("userId") Integer userId, @Param("targetUserId") Integer targetUserId);

    List<ConversationVO> selectByUserId(@Param("userId") Integer userId);

    int update(Conversation conversation);

    int updateUnreadCount(@Param("id") Integer id, @Param("unreadCount") Integer unreadCount);

    int deleteById(Integer id);

    int deleteByUserId(Integer userId);
}
