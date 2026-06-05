package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

@FeignClient(name = "mybilibili-content-interaction", contextId = "dynamicClient", path = "/dynamic")
public interface DynamicClient {

    @PostMapping("/comment/count/{dynamicId}")
    Result<?> incrementCommentCount(@PathVariable("dynamicId") Integer dynamicId, @RequestParam("delta") int delta);
}
