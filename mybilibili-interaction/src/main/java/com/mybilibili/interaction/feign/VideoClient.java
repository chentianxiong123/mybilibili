package com.mybilibili.interaction.feign;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

@FeignClient(name = "mybilibili-video-media", contextId = "interactionVideoClient")
public interface VideoClient {
    
    @GetMapping("/video/{id}")
    Result<VideoVO> getVideoById(@PathVariable("id") Integer id);
}
