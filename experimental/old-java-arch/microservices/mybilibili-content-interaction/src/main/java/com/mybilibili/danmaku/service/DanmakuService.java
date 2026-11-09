package com.mybilibili.danmaku.service;

import com.mybilibili.common.vo.Result;
import com.mybilibili.danmaku.dto.SendDanmakuDTO;
import com.mybilibili.danmaku.vo.DanmakuVO;

import java.util.List;
import java.util.Map;

public interface DanmakuService {

    Result<List<DanmakuVO>> getDanmakus(Integer videoId);

    Result<List<DanmakuVO>> getDanmakusByTimeRange(Integer videoId, Double startTime, Double endTime);

    Result<Void> sendDanmaku(SendDanmakuDTO dto, Integer userId);

    Result<Void> deleteDanmaku(String id, Integer userId);

    Result<Long> getDanmakuCount(Integer videoId);

    Map<String, Object> getCreatorDanmakuList(Integer userId, Integer videoId, Integer page, Integer size);

    void deleteDanmakuByCreator(String danmakuId, Integer userId);

    Map<Integer, Long> getDanmakuCountByManuscriptIds(List<Integer> manuscriptIds);

    Map<String, Integer> getDanmakuTrend(List<Integer> manuscriptIds, String startDate, String endDate);
}
