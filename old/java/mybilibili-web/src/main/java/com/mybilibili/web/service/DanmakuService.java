package com.mybilibili.web.service;

import com.mybilibili.common.entity.DanmakuDocument;

import java.util.List;

public interface DanmakuService {

    DanmakuDocument sendDanmaku(Integer userId, Integer videoId, Integer manuscriptId, String content, String time, String color, Integer mode);

    List<DanmakuDocument> getDanmakus(Integer videoId);

    List<DanmakuDocument> getDanmakusByTimeRange(Integer videoId, Double startTime, Double endTime);

    long getDanmakuCount(Integer videoId);
}
