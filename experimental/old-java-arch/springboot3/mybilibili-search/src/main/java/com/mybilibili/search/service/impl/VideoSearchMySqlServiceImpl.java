package com.mybilibili.search.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.VideoSearchVO;
import com.mybilibili.search.mapper.ManuscriptMapper;
import com.mybilibili.search.service.HotSearchService;
import com.mybilibili.search.service.VideoSearchService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "false", matchIfMissing = true)
public class VideoSearchMySqlServiceImpl implements VideoSearchService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private HotSearchService hotSearchService;

    private static final Integer PUBLISHED_STATUS = 3;
    private static final String SORT_RELEVANCE = "relevance";
    private static final String SORT_TIME = "time";
    private static final String SORT_HOT = "hot";

    @Override
    public Page<VideoSearchVO> search(String keyword, Integer categoryId, String tag, Integer userId,
                                       String sort, int page, int size) {
        try {
            log.info("使用MySQL搜索，keyword: {}", keyword);

            if (StringUtils.hasText(keyword)) {
                try {
                    hotSearchService.incrementHotSearch(keyword);
                } catch (Exception e) {
                    log.warn("更新热搜榜失败，keyword: {}, error: {}", keyword, e.getMessage());
                }
            }

            String orderBy = buildOrderBy(sort);
            int offset = Math.max(0, (page - 1) * size);

            List<Manuscript> manuscripts = manuscriptMapper.searchByKeyword(
                    keyword, categoryId, userId, PUBLISHED_STATUS, orderBy, offset, size);

            long total = manuscriptMapper.countSearchByKeyword(keyword, categoryId, userId, PUBLISHED_STATUS);

            List<VideoSearchVO> voList = manuscripts.stream()
                    .map(this::convertToVO)
                    .collect(Collectors.toList());

            Pageable pageable = PageRequest.of(Math.max(0, page - 1), size);
            return new PageImpl<>(voList, pageable, total);

        } catch (Exception e) {
            log.error("MySQL搜索视频失败，keyword: {}, error: {}", keyword, e.getMessage(), e);
            return Page.empty();
        }
    }

    @Override
    public List<String> suggest(String keyword) {
        if (!StringUtils.hasText(keyword)) {
            return new ArrayList<>();
        }

        try {
            List<Manuscript> manuscripts = manuscriptMapper.suggestByKeyword(keyword, PUBLISHED_STATUS, 10);
            return manuscripts.stream()
                    .map(Manuscript::getTitle)
                    .distinct()
                    .limit(10)
                    .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("MySQL获取搜索建议失败，keyword: {}, error: {}", keyword, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    private String buildOrderBy(String sort) {
        if (!StringUtils.hasText(sort)) {
            sort = SORT_RELEVANCE;
        }

        switch (sort) {
            case SORT_TIME:
                return "upload_time DESC";
            case SORT_HOT:
                return "view_count DESC";
            case SORT_RELEVANCE:
            default:
                return "upload_time DESC";
        }
    }

    private VideoSearchVO convertToVO(Manuscript manuscript) {
        try {
            VideoSearchVO vo = new VideoSearchVO();

            vo.setManuscriptId(manuscript.getId());
            vo.setTitle(manuscript.getTitle());
            vo.setDescription(manuscript.getDescription());
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
            vo.setStatus(manuscript.getStatus());
            vo.setCoverUrl(manuscript.getCoverUrl());

            return vo;
        } catch (Exception e) {
            log.error("转换VO失败: {}", e.getMessage(), e);
            return null;
        }
    }

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