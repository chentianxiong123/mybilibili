package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.ManuscriptVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

@FeignClient(name = "mybilibili-video-media", contextId = "manuscriptClient")
public interface ManuscriptClient {

    @PutMapping("/manuscript/{id}/comment-count")
    Result<?> updateCommentCount(@PathVariable("id") Integer manuscriptId, @RequestParam("count") Integer count);

    @PostMapping("/manuscript/{id}/increment-comment")
    Result<?> incrementCommentCount(@PathVariable("id") Integer manuscriptId);

    @PostMapping("/manuscript/{id}/decrement-comment")
    Result<?> decrementCommentCount(@PathVariable("id") Integer manuscriptId);

    @GetMapping("/manuscript/internal/{id}")
    Result<ManuscriptVO> getManuscriptById(@PathVariable("id") Integer manuscriptId);

    @PutMapping("/manuscript/internal/{id}/take-down")
    Result<?> takeDownManuscript(@PathVariable("id") Integer manuscriptId);
}
