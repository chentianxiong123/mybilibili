package com.mybilibili.user.task;

import com.mybilibili.user.mapper.UserMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;

@Slf4j
@Component
public class DailyCoinTask {

    @Autowired
    private UserMapper userMapper;

    private static final int DAILY_COIN_AMOUNT = 1;

    @Scheduled(cron = "0 0 0 * * ?")
    @Transactional
    public void addDailyCoins() {
        log.info("开始执行每日硬币发放任务，日期: {}", LocalDate.now());
        
        try {
            int updatedCount = userMapper.addDailyCoins(DAILY_COIN_AMOUNT);
            log.info("每日硬币发放完成，共发放 {} 个硬币给 {} 个用户", DAILY_COIN_AMOUNT, updatedCount);
        } catch (Exception e) {
            log.error("每日硬币发放失败", e);
            throw e;
        }
    }
}
