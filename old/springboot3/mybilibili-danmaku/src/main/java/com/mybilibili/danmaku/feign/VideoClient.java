package com.mybilibili.danmaku.feign;

import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;

import java.util.List;

@FeignClient(name = "mybilibili-video", contextId = "videoClient")
public interface VideoClient {

    @GetMapping("/video/{id}")
    Result<VideoVO> getVideoById(@PathVariable("id") Integer id);

    @GetMapping("/video/user/{userId}/ids")
    Result<List<Integer>> getManuscriptIdsByUserId(@PathVariable("userId") Integer userId);

    @GetMapping("/video/user/{userId}/video-ids")
    Result<List<Integer>> getVideoIdsByUserId(@PathVariable("userId") Integer userId);

    @GetMapping("/video/manuscript/{id}")
    Result<ManuscriptVO> getManuscriptById(@PathVariable("id") Integer id);
}