package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@FeignClient(name = "mybilibili-ai", contextId = "commentContentReviewClient", path = "/ai/review")
public interface ContentReviewClient {

    @PostMapping("/content")
    Result<List<String>> moderateContent(@RequestParam String content, @RequestParam(defaultValue = "COMMENT") String scene);

    @PostMapping("/comment")
    Result<Map<String, Object>> moderateComment(@RequestParam String content);

    @PostMapping("/reply")
    Result<Map<String, Object>> moderateReply(@RequestParam String content);

    @PostMapping("/report")
    Result<Map<String, Object>> moderateReport(@RequestParam String content);
}
