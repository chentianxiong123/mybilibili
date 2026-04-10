package com.mybilibili.search.service.impl;

import com.mybilibili.search.service.HotSearchService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ZSetOperations;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
public class HotSearchServiceImpl implements HotSearchService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_HOT_SEARCH_RANK = "hot_search:rank";
    private static final String KEY_HOT_SEARCH_DETAIL = "hot_search:detail:%s";
    private static final double BASE_SCORE = 10.0;
    private static final double TIME_DECAY_FACTOR = 0.1;
    private static final long EXPIRE_DAYS = 7;

    @Override
    public void incrementHotSearch(String keyword) {
        if (keyword == null || keyword.trim().isEmpty()) {
            return;
        }

        keyword = keyword.trim().toLowerCase();

        try {
            long currentTimeMillis = System.currentTimeMillis();
            double scoreIncrement = calculateScoreIncrement(currentTimeMillis);
            redisTemplate.opsForZSet().incrementScore(KEY_HOT_SEARCH_RANK, keyword, scoreIncrement);

            String detailKey = String.format(KEY_HOT_SEARCH_DETAIL, keyword);
            redisTemplate.opsForHash().increment(detailKey, "count", 1);
            redisTemplate.opsForHash().put(detailKey, "lastSearchTime", String.valueOf(currentTimeMillis));

            Boolean exists = redisTemplate.hasKey(detailKey);
            if (exists == null || !exists) {
                redisTemplate.opsForHash().put(detailKey, "firstSearchTime", String.valueOf(currentTimeMillis));
            }

            redisTemplate.expire(KEY_HOT_SEARCH_RANK, EXPIRE_DAYS, TimeUnit.DAYS);
            redisTemplate.expire(detailKey, EXPIRE_DAYS, TimeUnit.DAYS);

            log.debug("更新热搜关键词 '{}' 热度，增加分数: {}", keyword, scoreIncrement);
        } catch (Exception e) {
            log.error("更新热搜关键词 '{}' 热度失败: {}", keyword, e.getMessage());
        }
    }

    @Override
    public List<HotSearchVO> getHotSearchTop10() {
        List<HotSearchVO> result = new ArrayList<>();

        try {
            Set<ZSetOperations.TypedTuple<String>> top10 =
                    redisTemplate.opsForZSet().reverseRangeWithScores(KEY_HOT_SEARCH_RANK, 0, 9);

            if (top10 == null || top10.isEmpty()) {
                return result;
            }

            int rank = 1;
            for (ZSetOperations.TypedTuple<String> tuple : top10) {
                String keyword = tuple.getValue();
                Double score = tuple.getScore();

                if (keyword != null && score != null) {
                    HotSearchVO vo = new HotSearchVO();
                    vo.setKeyword(keyword);
                    vo.setScore(score);
                    vo.setRank(rank++);
                    result.add(vo);
                }
            }

            log.debug("获取热搜榜 Top10，共 {} 条", result.size());
        } catch (Exception e) {
            log.error("获取热搜榜 Top10 失败: {}", e.getMessage());
        }

        return result;
    }

    @Override
    public void cleanExpiredHotSearch() {
        try {
            Set<String> keywords = redisTemplate.opsForZSet().range(KEY_HOT_SEARCH_RANK, 0, -1);

            if (keywords == null || keywords.isEmpty()) {
                log.info("没有需要清理的热搜数据");
                return;
            }

            long currentTimeMillis = System.currentTimeMillis();
            long expireTimeMillis = EXPIRE_DAYS * 24 * 60 * 60 * 1000;
            int cleanedCount = 0;

            for (String keyword : keywords) {
                String detailKey = String.format(KEY_HOT_SEARCH_DETAIL, keyword);
                String firstSearchTimeStr = (String) redisTemplate.opsForHash().get(detailKey, "firstSearchTime");

                if (firstSearchTimeStr != null) {
                    long firstSearchTime = Long.parseLong(firstSearchTimeStr);
                    if (currentTimeMillis - firstSearchTime > expireTimeMillis) {
                        redisTemplate.opsForZSet().remove(KEY_HOT_SEARCH_RANK, keyword);
                        redisTemplate.delete(detailKey);
                        cleanedCount++;
                        log.debug("清理过期热搜关键词: {}", keyword);
                    }
                }
            }

            log.info("清理热搜数据完成，共清理 {} 条过期数据", cleanedCount);
        } catch (Exception e) {
            log.error("清理过期热搜数据失败: {}", e.getMessage());
        }
    }

    private double calculateScoreIncrement(long currentTimeMillis) {
        long hoursSinceEpoch = currentTimeMillis / (1000 * 60 * 60);
        double timeDecay = 1.0 / (1.0 + TIME_DECAY_FACTOR * Math.log1p(hoursSinceEpoch % 10000));
        return BASE_SCORE * timeDecay;
    }
}