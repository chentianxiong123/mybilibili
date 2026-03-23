package com.mybilibili.web.service.impl;

import com.mybilibili.web.service.HotSearchService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;

/**
 * 热搜榜服务实现
 */
@Slf4j
@Service
public class HotSearchServiceImpl implements HotSearchService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    /**
     * 热搜排行 ZSet key
     */
    private static final String KEY_HOT_SEARCH_RANK = "hot_search:rank";

    /**
     * 热搜详情 Hash key 前缀
     */
    private static final String KEY_HOT_SEARCH_DETAIL = "hot_search:detail:%s";

    /**
     * 基础热度分数
     */
    private static final double BASE_SCORE = 10.0;

    /**
     * 时间衰减系数（每小时衰减 10%）
     */
    private static final double TIME_DECAY_FACTOR = 0.1;

    /**
     * Redis 过期时间：7天
     */
    private static final long EXPIRE_DAYS = 7;

    @Override
    public void incrementHotSearch(String keyword) {
        if (keyword == null || keyword.trim().isEmpty()) {
            return;
        }

        // 去除前后空格并转为小写（统一存储）
        keyword = keyword.trim().toLowerCase();

        try {
            long currentTimeMillis = System.currentTimeMillis();

            // 1. 更新 Sorted Set 中的热度分数
            double scoreIncrement = calculateScoreIncrement(currentTimeMillis);
            redisTemplate.opsForZSet().incrementScore(KEY_HOT_SEARCH_RANK, keyword, scoreIncrement);

            // 2. 更新 Hash 中的详情信息
            String detailKey = String.format(KEY_HOT_SEARCH_DETAIL, keyword);
            redisTemplate.opsForHash().increment(detailKey, "count", 1);
            redisTemplate.opsForHash().put(detailKey, "lastSearchTime", String.valueOf(currentTimeMillis));

            // 3. 如果是首次搜索，记录首次搜索时间
            Boolean exists = redisTemplate.hasKey(detailKey);
            if (exists == null || !exists) {
                redisTemplate.opsForHash().put(detailKey, "firstSearchTime", String.valueOf(currentTimeMillis));
            }

            // 4. 设置过期时间
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
            // 从 Sorted Set 中获取 Top10（按分数倒序）
            Set<org.springframework.data.redis.core.ZSetOperations.TypedTuple<String>> top10 =
                    redisTemplate.opsForZSet().reverseRangeWithScores(KEY_HOT_SEARCH_RANK, 0, 9);

            if (top10 == null || top10.isEmpty()) {
                return result;
            }

            int rank = 1;
            for (org.springframework.data.redis.core.ZSetOperations.TypedTuple<String> tuple : top10) {
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
            // 获取所有热搜关键词
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
                    // 如果首次搜索时间超过7天，删除该关键词
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

    /**
     * 计算热度增量（带时间衰减）
     *
     * @param currentTimeMillis 当前时间戳
     * @return 热度增量
     */
    private double calculateScoreIncrement(long currentTimeMillis) {
        // 将当前时间转换为小时数（从某个固定时间点开始）
        long hoursSinceEpoch = currentTimeMillis / (1000 * 60 * 60);

        // 时间衰减因子：越新的搜索权重越高
        // 使用对数衰减，避免早期数据分数过高
        double timeDecay = 1.0 / (1.0 + TIME_DECAY_FACTOR * Math.log1p(hoursSinceEpoch % 10000));

        return BASE_SCORE * timeDecay;
    }
}
