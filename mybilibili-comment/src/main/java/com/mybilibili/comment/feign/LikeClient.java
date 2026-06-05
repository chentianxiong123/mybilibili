package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@FeignClient(name = "mybilibili-content-interaction", contextId = "likeClient", path = "/interaction")
public interface LikeClient {

    @PostMapping("/like")
    Result<?> like(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId);

    @DeleteMapping("/like")
    Result<?> unlike(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId);

    @GetMapping("/status")
    Result<Boolean> isLiked(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId);

    @PostMapping("/batch/status")
    Result<Map<Integer, Boolean>> batchIsLiked(@RequestParam Integer userId, @RequestParam String targetType, @RequestBody List<Integer> targetIds);

    @GetMapping("/count")
    Result<Integer> getLikeCount(@RequestParam String targetType, @RequestParam Integer targetId);

    @PostMapping("/batch/count")
    Result<Map<Integer, Integer>> batchGetLikeCount(@RequestParam String targetType, @RequestBody List<Integer> targetIds);
}
