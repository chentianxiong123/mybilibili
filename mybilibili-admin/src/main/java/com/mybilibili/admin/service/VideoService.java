package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;

import java.util.List;
import java.util.Map;

public interface VideoService {
    Result<?> getVideoList(Integer page, Integer size, String keyword, Integer status);
    Result<?> getVideoById(Integer id);
    Result<?> deleteVideo(Integer id);
    Result<?> deleteVideos(List<Integer> ids);
}