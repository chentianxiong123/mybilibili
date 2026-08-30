package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Danmaku;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface DanmakuMapper {
    int insert(Danmaku danmaku);
    List<Danmaku> findByVideoId(Integer videoId);
    List<Danmaku> findByManuscriptId(Integer manuscriptId);
    List<Danmaku> findByVideoIdAndManuscriptId(@Param("videoId") Integer videoId, @Param("manuscriptId") Integer manuscriptId);
}
