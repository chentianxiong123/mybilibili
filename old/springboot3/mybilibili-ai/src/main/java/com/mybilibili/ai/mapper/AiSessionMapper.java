package com.mybilibili.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.ai.entity.AiSession;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface AiSessionMapper extends BaseMapper<AiSession> {

    @Select("SELECT * FROM ai_sessions WHERE type = #{type} AND status = #{status} ORDER BY updated_at DESC")
    List<AiSession> selectByTypeAndStatus(@Param("type") String type, @Param("status") Integer status);

    @Select("SELECT COUNT(*) FROM ai_sessions WHERE type = #{type} AND status = #{status}")
    long countByTypeAndStatus(@Param("type") String type, @Param("status") Integer status);
}