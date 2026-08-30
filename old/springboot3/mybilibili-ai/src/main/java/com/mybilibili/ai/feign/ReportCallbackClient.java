package com.mybilibili.ai.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;

import java.util.Map;

@FeignClient(name = "mybilibili-comment", path = "/admin/report")
public interface ReportCallbackClient {

    @PutMapping("/ai-review-result")
    Result<?> updateAiReviewResult(@RequestBody Map<String, Object> result);
}