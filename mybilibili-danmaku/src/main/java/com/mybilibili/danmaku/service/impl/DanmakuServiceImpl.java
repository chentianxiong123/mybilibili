package com.mybilibili.danmaku.service.impl;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.danmaku.dto.SendDanmakuDTO;
import com.mybilibili.danmaku.entity.Danmaku;
import com.mybilibili.danmaku.feign.VideoClient;
import com.mybilibili.danmaku.repository.DanmakuRepository;
import com.mybilibili.danmaku.service.DanmakuService;
import com.mybilibili.danmaku.vo.DanmakuVO;
import com.mybilibili.danmaku.websocket.DanmakuBroadcastService;
import com.mybilibili.mq.ManuscriptAnalyticsEvent;
import com.mybilibili.mq.VideoMQProducer;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;
import org.springframework.data.mongodb.core.MongoTemplate;
import org.springframework.data.mongodb.core.aggregation.Aggregation;
import org.springframework.data.mongodb.core.aggregation.AggregationResults;
import org.springframework.data.mongodb.core.query.Criteria;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.*;
import java.util.Objects;
import java.util.Set;
import java.util.stream.Collectors;

@Service
public class DanmakuServiceImpl implements DanmakuService {

    @Autowired
    private DanmakuRepository danmakuRepository;

    @Autowired
    private VideoClient videoClient;

    @Autowired
    private MongoTemplate mongoTemplate;

    @Autowired
    private DanmakuBroadcastService broadcastService;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Override
    public Result<List<DanmakuVO>> getDanmakus(Integer videoId) {
        List<Danmaku> danmakuList = danmakuRepository.findByVideoId(videoId);
        List<DanmakuVO> voList = danmakuList.stream()
                .map(this::convertToVO)
                .collect(Collectors.toList());
        return Result.success(voList);
    }

    @Override
    public Result<List<DanmakuVO>> getDanmakusByTimeRange(Integer videoId, Double startTime, Double endTime) {
        List<Danmaku> danmakuList = danmakuRepository.findByVideoIdAndTimeBetween(videoId, startTime, endTime);
        List<DanmakuVO> voList = danmakuList.stream()
                .map(this::convertToVO)
                .collect(Collectors.toList());
        return Result.success(voList);
    }

    @Override
    public Result<Void> sendDanmaku(SendDanmakuDTO dto, Integer userId) {
        if (userId == null) {
            throw new BusinessException("用户未登录");
        }

        Danmaku danmaku = new Danmaku();
        danmaku.setVideoId(dto.getVideoId());
        danmaku.setUserId(userId);
        danmaku.setContent(dto.getContent());
        danmaku.setTime(parseTime(dto.getTime()));
        danmaku.setColor(dto.getColor() != null ? dto.getColor() : "#ffffff");
        danmaku.setMode(dto.getMode() != null ? dto.getMode() : 0);
        danmaku.setStatus(0);
        danmaku.setCreateTime(LocalDateTime.now());

        Integer manuscriptOwnerId = resolveManuscript(danmaku, dto.getVideoId());

        danmakuRepository.save(danmaku);
        videoMQProducer.sendManuscriptAnalyticsEvent(ManuscriptAnalyticsEvent.metricIncrement(
                danmaku.getManuscriptId(),
                manuscriptOwnerId,
                ManuscriptAnalyticsEvent.METRIC_DANMAKU,
                1
        ));

        broadcastService.broadcast(danmaku.getVideoId(), Map.of(
                "type", "danmaku",
                "videoId", danmaku.getVideoId(),
                "userId", userId,
                "content", danmaku.getContent(),
                "time", danmaku.getTime(),
                "color", danmaku.getColor(),
                "mode", danmaku.getMode(),
                "timestamp", System.currentTimeMillis()
        ));

        return Result.<Void>success();
    }

    private Integer resolveManuscript(Danmaku danmaku, Integer videoId) {
        Result<VideoVO> videoResult = videoClient.getVideoById(videoId);
        if (videoResult == null || videoResult.getCode() != 200 || videoResult.getData() == null) {
            throw new BusinessException("视频不存在");
        }
        VideoVO video = videoResult.getData();
        if (video.getManuscriptId() == null) {
            throw new BusinessException("视频缺少稿件关联");
        }
        danmaku.setManuscriptId(video.getManuscriptId());
        if (video.getUploader() != null && video.getUploader().getId() != null) {
            return video.getUploader().getId();
        }
        Result<ManuscriptVO> manuscriptResult = videoClient.getManuscriptById(video.getManuscriptId());
        if (manuscriptResult == null || manuscriptResult.getCode() != 200 || manuscriptResult.getData() == null
                || manuscriptResult.getData().getUserId() == null) {
            throw new BusinessException("稿件作者不存在");
        }
        return manuscriptResult.getData().getUserId();
    }

    @Override
    public Result<Void> deleteDanmaku(String id, Integer userId) {
        if (userId == null) {
            throw new BusinessException("用户未登录");
        }

        Danmaku danmaku = danmakuRepository.findById(id).orElse(null);
        if (danmaku == null) {
            throw new BusinessException("弹幕不存在");
        }

        if (!danmaku.getUserId().equals(userId)) {
            throw new BusinessException("无权删除该弹幕");
        }

        danmakuRepository.deleteById(id);
        return Result.<Void>success();
    }

    @Override
    public Result<Long> getDanmakuCount(Integer videoId) {
        long count = danmakuRepository.countByVideoId(videoId);
        return Result.success(count);
    }

    @Override
    public Map<Integer, Long> getDanmakuCountByManuscriptIds(List<Integer> manuscriptIds) {
        if (manuscriptIds == null || manuscriptIds.isEmpty()) {
            return Collections.emptyMap();
        }

        Aggregation aggregation = Aggregation.newAggregation(
                Aggregation.match(Criteria.where("manuscriptId").in(manuscriptIds)),
                Aggregation.group("manuscriptId").count().as("count")
        );

        AggregationResults<Map> results = mongoTemplate.aggregate(aggregation, "danmakus", Map.class);
        Map<Integer, Long> countMap = new HashMap<>();
        for (Map result : results.getMappedResults()) {
            Object id = result.get("_id");
            Number count = (Number) result.get("count");
            if (id != null) {
                countMap.put(((Number) id).intValue(), count.longValue());
            }
        }
        // Fill 0 for manuscripts without danmaku
        for (Integer mid : manuscriptIds) {
            countMap.putIfAbsent(mid, 0L);
        }
        return countMap;
    }

    @Override
    public Map<String, Integer> getDanmakuTrend(List<Integer> manuscriptIds, String startDate, String endDate) {
        if (manuscriptIds == null || manuscriptIds.isEmpty()) {
            return Collections.emptyMap();
        }

        LocalDateTime start = LocalDate.parse(startDate).atStartOfDay();
        LocalDateTime end = LocalDate.parse(endDate).plusDays(1).atStartOfDay();

        Aggregation aggregation = Aggregation.newAggregation(
                Aggregation.match(Criteria.where("manuscriptId").in(manuscriptIds)
                        .and("createTime").gte(start).lt(end)),
                Aggregation.project()
                        .andExpression("dateToString('%Y-%m-%d', createTime)").as("dateStr"),
                Aggregation.group("dateStr").count().as("count")
        );

        AggregationResults<Map> results = mongoTemplate.aggregate(aggregation, "danmakus", Map.class);
        Map<String, Integer> trendMap = new HashMap<>();
        for (Map result : results.getMappedResults()) {
            String date = (String) result.get("_id");
            Number count = (Number) result.get("count");
            if (date != null) {
                trendMap.put(date, count.intValue());
            }
        }
        return trendMap;
    }

    private Double parseTime(String time) {
        try {
            return Double.parseDouble(time);
        } catch (NumberFormatException e) {
            String[] parts = time.split(":");
            if (parts.length == 2) {
                int minutes = Integer.parseInt(parts[0]);
                double seconds = Double.parseDouble(parts[1]);
                return minutes * 60 + seconds;
            }
            return 0.0;
        }
    }

    private DanmakuVO convertToVO(Danmaku danmaku) {
        DanmakuVO vo = new DanmakuVO();
        BeanUtils.copyProperties(danmaku, vo);
        return vo;
    }

    @Override
    public Map<String, Object> getCreatorDanmakuList(Integer userId, Integer videoId, Integer page, Integer size) {
        List<Integer> userVideoIds = getUserVideoIds(userId);
        
        if (userVideoIds.isEmpty()) {
            Map<String, Object> emptyResult = new HashMap<>();
            emptyResult.put("list", Collections.emptyList());
            emptyResult.put("total", 0);
            emptyResult.put("page", page);
            emptyResult.put("size", size);
            return emptyResult;
        }

        PageRequest pageRequest = PageRequest.of(page - 1, size, Sort.by(Sort.Direction.DESC, "createTime"));
        Page<Danmaku> danmakuPage;

        if (videoId != null) {
            danmakuPage = danmakuRepository.findByVideoIdAndVideoIdIn(videoId, userVideoIds, pageRequest);
        } else {
            danmakuPage = danmakuRepository.findByVideoIdIn(userVideoIds, pageRequest);
        }

        // 获取所有需要的稿件ID和视频ID
        Set<Integer> manuscriptIds = danmakuPage.getContent().stream()
                .map(Danmaku::getManuscriptId)
                .filter(Objects::nonNull)
                .collect(Collectors.toSet());
        Set<Integer> videoIds = danmakuPage.getContent().stream()
                .map(Danmaku::getVideoId)
                .filter(Objects::nonNull)
                .collect(Collectors.toSet());

        // 批量获取稿件名称
        Map<Integer, String> manuscriptNameMap = new HashMap<>();
        for (Integer mid : manuscriptIds) {
            try {
                Result<ManuscriptVO> result = videoClient.getManuscriptById(mid);
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    manuscriptNameMap.put(mid, result.getData().getTitle());
                }
            } catch (Exception e) {
                // 忽略错误
            }
        }

        // 批量获取视频名称和顺序
        Map<Integer, String> videoNameMap = new HashMap<>();
        Map<Integer, Integer> videoOrderMap = new HashMap<>();
        for (Integer vid : videoIds) {
            try {
                Result<VideoVO> result = videoClient.getVideoById(vid);
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    videoNameMap.put(vid, result.getData().getTitle());
                    videoOrderMap.put(vid, result.getData().getVideoOrder());
                }
            } catch (Exception e) {
                // 忽略错误
            }
        }

        List<Map<String, Object>> list = danmakuPage.getContent().stream().map(danmaku -> {
            Map<String, Object> item = new HashMap<>();
            item.put("id", danmaku.getId());
            item.put("manuscriptId", danmaku.getManuscriptId());
            item.put("manuscriptName", manuscriptNameMap.getOrDefault(danmaku.getManuscriptId(), "未知稿件"));
            item.put("videoName", videoNameMap.getOrDefault(danmaku.getVideoId(), "未知视频"));
            item.put("videoOrder", videoOrderMap.getOrDefault(danmaku.getVideoId(), 1));
            item.put("userId", danmaku.getUserId());
            item.put("content", danmaku.getContent());
            item.put("time", danmaku.getTime());
            item.put("color", danmaku.getColor());
            item.put("mode", danmaku.getMode());
            item.put("createTime", danmaku.getCreateTime());
            return item;
        }).collect(Collectors.toList());

        Map<String, Object> result = new HashMap<>();
        result.put("list", list);
        result.put("total", danmakuPage.getTotalElements());
        result.put("page", page);
        result.put("size", size);
        return result;
    }

    @Override
    public void deleteDanmakuByCreator(String danmakuId, Integer userId) {
        if (userId == null) {
            throw new BusinessException("用户未登录");
        }

        Danmaku danmaku = danmakuRepository.findById(danmakuId).orElse(null);
        if (danmaku == null) {
            throw new BusinessException("弹幕不存在");
        }

        List<Integer> userVideoIds = getUserVideoIds(userId);
        if (!userVideoIds.contains(danmaku.getVideoId())) {
            throw new BusinessException("无权删除该弹幕");
        }

        danmakuRepository.deleteById(danmakuId);
    }

    private List<Integer> getUserVideoIds(Integer userId) {
        try {
            Result<List<Integer>> result = videoClient.getVideoIdsByUserId(userId);
            if (result != null && result.getCode() == 200 && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
        return Collections.emptyList();
    }
}
