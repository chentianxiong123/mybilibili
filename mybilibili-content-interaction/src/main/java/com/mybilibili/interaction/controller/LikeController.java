package com.mybilibili.interaction.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.service.InteractionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/interaction")
public class LikeController {

    @Autowired
    private InteractionService interactionService;

    @PostMapping("/like")
    public Result<?> like(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId) {
        return interactionService.like(userId, targetType, targetId);
    }

    @DeleteMapping("/like")
    public Result<?> unlike(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId) {
        return interactionService.unlike(userId, targetType, targetId);
    }

    @GetMapping("/status")
    public Result<Boolean> isLiked(@RequestParam Integer userId, @RequestParam String targetType, @RequestParam Integer targetId) {
        return interactionService.checkLike(userId, targetType, targetId);
    }

    @PostMapping("/batch/status")
    public Result<Map<Integer, Boolean>> batchIsLiked(@RequestParam Integer userId, @RequestParam String targetType, @RequestBody List<Integer> targetIds) {
        Map<Integer, Boolean> result = new HashMap<>();
        for (Integer targetId : targetIds) {
            Result<Boolean> checkResult = interactionService.checkLike(userId, targetType, targetId);
            result.put(targetId, checkResult.getData());
        }
        return Result.success(result);
    }

    @GetMapping("/count")
    public Result<Integer> getLikeCount(@RequestParam String targetType, @RequestParam Integer targetId) {
        int count = interactionService.getLikeCount(targetType, targetId);
        return Result.success(count);
    }

    @PostMapping("/batch/count")
    public Result<Map<Integer, Integer>> batchGetLikeCount(@RequestParam String targetType, @RequestBody List<Integer> targetIds) {
        Map<Integer, Integer> result = new HashMap<>();
        for (Integer targetId : targetIds) {
            int count = interactionService.getLikeCount(targetType, targetId);
            result.put(targetId, count);
        }
        return Result.success(result);
    }
}
