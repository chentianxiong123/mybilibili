package com.mybilibili.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.ai.entity.AiConversation;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

public interface AiConversationMapper extends BaseMapper<AiConversation> {

    @Select("SELECT * FROM ai_conversations WHERE user_id = #{userId} ORDER BY updated_at DESC")
    List<AiConversation> selectByUserId(@Param("userId") Integer userId);
}