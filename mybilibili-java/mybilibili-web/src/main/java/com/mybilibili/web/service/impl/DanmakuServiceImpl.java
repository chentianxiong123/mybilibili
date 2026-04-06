package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.DanmakuDocument;
import com.mybilibili.web.repository.DanmakuRepository;
import com.mybilibili.web.service.DanmakuService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class DanmakuServiceImpl implements DanmakuService {

    @Autowired
    private DanmakuRepository danmakuRepository;

    @Override
    public DanmakuDocument sendDanmaku(Integer userId, Integer videoId, Integer manuscriptId, String content, String time, String color, Integer mode) {
        DanmakuDocument danmaku = new DanmakuDocument();
        danmaku.setUserId(userId);
        danmaku.setVideoId(videoId);
        danmaku.setManuscriptId(manuscriptId);
        danmaku.setContent(content);
        danmaku.setTime(parseTime(time));
        danmaku.setColor(color != null ? color : "#ffffff");
        danmaku.setMode(mode != null ? mode : 0);
        return danmakuRepository.save(danmaku);
    }

    @Override
    public List<DanmakuDocument> getDanmakus(Integer videoId) {
        return danmakuRepository.findByVideoId(videoId);
    }

    @Override
    public List<DanmakuDocument> getDanmakusByTimeRange(Integer videoId, Double startTime, Double endTime) {
        return danmakuRepository.findByVideoIdAndTimeBetween(videoId, startTime, endTime);
    }

    @Override
    public long getDanmakuCount(Integer videoId) {
        return danmakuRepository.countByVideoId(videoId);
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
}
