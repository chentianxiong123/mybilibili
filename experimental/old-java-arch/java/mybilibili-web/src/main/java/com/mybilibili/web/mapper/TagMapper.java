package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Tag;
import org.apache.ibatis.annotations.Mapper;

import java.util.List;

@Mapper
public interface TagMapper {
    Tag selectByName(String name);
    int insert(Tag tag);
    List<Tag> selectByVideoId(Integer videoId);
}