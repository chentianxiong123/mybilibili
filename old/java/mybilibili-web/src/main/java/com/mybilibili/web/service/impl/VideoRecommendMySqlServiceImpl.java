package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.service.VideoRecommendService;
import com.mybilibili.web.service.WatchHistoryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

/**
 * 视频推荐服务实现 - MySQL降级版本
 * 当Elasticsearch不可用时使用MySQL进行推荐
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "false", matchIfMissing = true)
public class VideoRecommendMySqlServiceImpl implements VideoRecommendService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private WatchHistoryService watchHistoryService;

    /**
     * 已上架稿件状态
     */
    private static final Integer PUBLISHED_STATUS = 3;

    /**
     * 默认推荐数量
     */
    private static final int DEFAULT_SIZE = 10;

    /**
     * 最大推荐数量
     */
    private static final int MAX_SIZE = 50;

    @Override
    public List<VideoRecommendVO> getRelatedVideos(Integer videoId, int size) {
        if (videoId == null) {
            return new ArrayList<>();
        }

        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 获取当前视频所属的稿件
            Manuscript currentManuscript = manuscriptMapper.selectById(videoId);
            if (currentManuscript == null) {
                log.warn("未找到视频对应的稿件: {}", videoId);
                return new ArrayList<>();
            }

            // 使用MySQL查询相关视频（同分类或相似标题）
            List<Manuscript> manuscripts = manuscriptMapper.selectRelatedManuscripts(
                    currentManuscript.getId(),
                    currentManuscript.getCategoryId(),
                    currentManuscript.getTitle(),
                    PUBLISHED_STATUS,
                    size
            );

            return manuscripts.stream()
                    .map(m -> convertToVO(m, "相关推荐"))
                    .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("获取相关视频失败，videoId: {}, error: {}", videoId, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    @Override
    public List<VideoRecommendVO> getHotVideos(Integer categoryId, int size) {
        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 使用MySQL按播放量排序查询热门视频
            List<Manuscript> manuscripts = manuscriptMapper.selectHotManuscripts(
                    categoryId,
                    PUBLISHED_STATUS,
                    size
            );

            int rank = 1;
            List<VideoRecommendVO> result = new ArrayList<>();
            for (Manuscript manuscript : manuscripts) {
                VideoRecommendVO vo = convertToVO(manuscript, "热门排行第" + rank + "名");
                result.add(vo);
                rank++;
            }

            return result;

        } catch (Exception e) {
            log.error("获取热门视频失败，categoryId: {}, error: {}", categoryId, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    @Override
    public List<VideoRecommendVO> getRecommendedVideosForUser(Integer userId, int size) {
        if (userId == null) {
            return new ArrayList<>();
        }

        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 获取用户浏览历史
            List<WatchHistoryVO> historyList = watchHistoryService.getWatchHistoryList(userId, 0, 20);

            if (historyList == null || historyList.isEmpty()) {
                // 没有浏览历史，返回热门视频
                return getHotVideos(null, size);
            }

            // 提取用户感兴趣的分类
            Set<Integer> interestedCategories = new HashSet<>();
            Set<Integer> watchedManuscriptIds = new HashSet<>();

            for (WatchHistoryVO history : historyList) {
                if (history.getVideo() != null) {
                    VideoVO video = history.getVideo();
                    if (video.getManuscriptId() != null) {
                        watchedManuscriptIds.add(video.getManuscriptId());
                    }
                    if (video.getCategoryId() != null) {
                        interestedCategories.add(video.getCategoryId());
                    }
                }
            }

            // 使用MySQL查询个性化推荐
            List<Manuscript> manuscripts = manuscriptMapper.selectRecommendedManuscripts(
                    new ArrayList<>(watchedManuscriptIds),
                    new ArrayList<>(interestedCategories),
                    PUBLISHED_STATUS,
                    size
            );

            return manuscripts.stream()
                    .map(m -> convertToVO(m, "猜你喜欢"))
                    .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("获取个性化推荐失败，userId: {}, error: {}", userId, e.getMessage(), e);
            // 出错时返回热门视频作为兜底
            return getHotVideos(null, size);
        }
    }

    /**
     * 转换为VO
     */
    private VideoRecommendVO convertToVO(Manuscript manuscript, String reason) {
        VideoRecommendVO vo = new VideoRecommendVO();

        vo.setManuscriptId(manuscript.getId());
        vo.setTitle(manuscript.getTitle());
        vo.setDescription(manuscript.getDescription());
        vo.setCoverUrl(manuscript.getCoverUrl());
        vo.setUserId(manuscript.getUserId());
        vo.setCategoryId(manuscript.getCategoryId());
        vo.setViewCount(manuscript.getViewCount());
        vo.setLikeCount(manuscript.getLikeCount());
        vo.setCommentCount(manuscript.getCommentCount());
        vo.setShareCount(manuscript.getShareCount());
        vo.setCollectCount(manuscript.getCollectCount());
        vo.setCoinCount(manuscript.getCoinCount());
        vo.setDurationSeconds(manuscript.getDurationSeconds());
        vo.setDuration(formatDuration(manuscript.getDurationSeconds()));
        vo.setUploadTime(manuscript.getUploadTime());
        vo.setRecommendReason(reason);
        vo.setScore(0.0);

        return vo;
    }

    /**
     * 格式化视频时长
     */
    private String formatDuration(Integer seconds) {
        if (seconds == null || seconds <= 0) {
            return "00:00";
        }

        int hours = seconds / 3600;
        int minutes = (seconds % 3600) / 60;
        int secs = seconds % 60;

        if (hours > 0) {
            return String.format("%02d:%02d:%02d", hours, minutes, secs);
        } else {
            return String.format("%02d:%02d", minutes, secs);
        }
    }
}
