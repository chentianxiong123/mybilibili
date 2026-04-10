package com.mybilibili.comment.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@FeignClient(name = "mybilibili-ai", path = "/ai")
public interface ContentReviewClient {

    @PostMapping("/review/content")
    Result<List<String>> detectProhibitedWords(@RequestParam String content);
}