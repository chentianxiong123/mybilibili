package com.mybilibili.ai.mapper;

import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface ManuscriptMapper {

    @Select("SELECT * FROM manuscripts WHERE id = #{id}")
    Manuscript selectById(Integer id);
}
