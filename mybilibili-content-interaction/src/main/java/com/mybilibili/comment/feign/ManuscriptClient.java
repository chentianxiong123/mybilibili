package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.ManuscriptVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

@FeignClient(name = "mybilibili-video-media", contextId = "manuscriptClient")
public interface ManuscriptClient {

    @GetMapping("/manuscript/internal/{id}")
    Result<ManuscriptVO> getManuscriptById(@PathVariable("id") Integer manuscriptId);

    @PutMapping("/manuscript/internal/{id}/take-down")
    Result<?> takeDownManuscript(@PathVariable("id") Integer manuscriptId);
}
