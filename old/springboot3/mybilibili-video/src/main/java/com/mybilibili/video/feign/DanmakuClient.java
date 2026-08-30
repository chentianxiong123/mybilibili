package com.mybilibili.video.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;

import java.util.List;
import java.util.Map;

@FeignClient(name = "mybilibili-danmaku", contextId = "danmakuStats")
public interface DanmakuClient {

    @PostMapping("/danmaku/batch-count")
    Result<Map<Integer, Long>> getDanmakuCountByManuscriptIds(@RequestBody List<Integer> manuscriptIds);

    @PostMapping("/danmaku/trend")
    Result<Map<String, Integer>> getDanmakuTrend(
            @RequestBody List<Integer> manuscriptIds,
            @RequestParam String startDate,
            @RequestParam String endDate);
}