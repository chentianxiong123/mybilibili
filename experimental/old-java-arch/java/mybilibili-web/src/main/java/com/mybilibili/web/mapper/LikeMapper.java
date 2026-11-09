package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Like;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface LikeMapper {
    int insert(Like like);
    int delete(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
    Like findByUserAndManuscript(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
}
