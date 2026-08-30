package com.mybilibili.web.mapper;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

@Mapper
public interface ProhibitedWordMapper {

    @Select("SELECT word, match_type FROM prohibited_word WHERE is_enabled = 1")
    List<Map<String, Object>> selectAllEnabled();
}
