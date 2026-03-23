package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Collection;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface CollectionMapper {
    int insert(Collection collection);
    int delete(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
    Collection findByUserAndManuscript(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
}
