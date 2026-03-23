package com.mybilibili.web.job;

import com.mybilibili.web.service.HotSearchService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * 热搜榜清理定时任务
 * 每天凌晨清理过期的热搜数据
 */
@Slf4j
@Component
public class HotSearchCleanJob {

    @Autowired
    private HotSearchService hotSearchService;

    /**
     * 每天凌晨3点执行清理任务
     * cron表达式：秒 分 时 日 月 周
     */
    @Scheduled(cron = "0 0 3 * * ?")
    public void cleanExpiredHotSearch() {
        log.info("开始执行热搜榜清理定时任务");
        try {
            hotSearchService.cleanExpiredHotSearch();
            log.info("热搜榜清理定时任务执行完成");
        } catch (Exception e) {
            log.error("热搜榜清理定时任务执行失败: {}", e.getMessage(), e);
        }
    }
}
