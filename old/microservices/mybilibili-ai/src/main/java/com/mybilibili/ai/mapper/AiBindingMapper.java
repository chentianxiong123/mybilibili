package com.mybilibili.ai.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.ai.entity.AiBinding;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import java.util.List;

@Mapper
public interface AiBindingMapper extends BaseMapper<AiBinding> {

    @Select("SELECT * FROM ai_bindings WHERE feature = #{feature}")
    AiBinding selectByFeature(String feature);
}
