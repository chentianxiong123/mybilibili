package com.mybilibili.message.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.vo.ConversationVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface ConversationMapper extends BaseMapper<Conversation> {

    Conversation selectById(@Param("id") Long id);

    Conversation selectByUserAndTarget(@Param("userId") Integer userId, @Param("targetUserId") Integer targetUserId);

    List<ConversationVO> selectByUserId(@Param("userId") Integer userId);

    int update(Conversation conversation);

    int updateUnreadCount(@Param("id") Long id, @Param("unreadCount") Integer unreadCount);

    int deleteById(@Param("id") Long id);
}