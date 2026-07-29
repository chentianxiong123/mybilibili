package com.mybilibili.search.job;

import com.mybilibili.search.service.HotSearchService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class HotSearchCleanJob {

    @Autowired
    private HotSearchService hotSearchService;

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