package com.mybilibili.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.ai.entity.AiChatMessage;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

public interface AiChatMessageMapper extends BaseMapper<AiChatMessage> {

    @Select("SELECT * FROM ai_chat_messages WHERE session_id = #{sessionId} ORDER BY created_at ASC")
    List<AiChatMessage> selectBySessionId(@Param("sessionId") Long sessionId);
}
