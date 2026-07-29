package com.mybilibili.interaction.mapper;

import com.mybilibili.common.entity.Share;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface ShareMapper {
    int insert(Share share);

    Integer countByManuscriptId(@Param("manuscriptId") Integer manuscriptId);

    Integer countByManuscriptIdAndChannel(@Param("manuscriptId") Integer manuscriptId, @Param("channel") String channel);
}
