package com.mybilibili.admin.mapper;

import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface ShareMapper {
    
    @Delete("DELETE FROM shares WHERE video_id = #{videoId}")
    int deleteByVideoId(Integer videoId);
    
    @Select("SELECT COUNT(*) FROM shares WHERE video_id = #{videoId}")
    int countByVideoId(Integer videoId);
}
