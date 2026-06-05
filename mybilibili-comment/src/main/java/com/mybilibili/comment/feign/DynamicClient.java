package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

/**
 * Legacy standalone comment-service client.
 * <p>
 * In mybilibili-content-interaction runtime this call is handled by
 * CommentDynamicClientLocalAdapter and must not be used as a new internal Feign path.
 */
@Deprecated(since = "2026-06-05", forRemoval = false)
@FeignClient(name = "mybilibili-content-interaction", contextId = "dynamicClient", path = "/dynamic")
public interface DynamicClient {

    @PostMapping("/comment/count/{dynamicId}")
    Result<?> incrementCommentCount(@PathVariable("dynamicId") Integer dynamicId, @RequestParam("delta") int delta);
}
