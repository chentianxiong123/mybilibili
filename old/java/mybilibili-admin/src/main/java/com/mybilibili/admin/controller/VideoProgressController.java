package com.mybilibili.admin.controller;

import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.service.VideoProcessSupport;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/videos")
@Tag(name = "视频进度查询", description = "视频处理进度查询")
public class VideoProgressController {

    @Autowired
    private VideoMapper videoMapper;

    @GetMapping("/progress/{videoId}")
    @Operation(summary = "获取处理进度", description = "获取视频处理进度")
    public Result<VideoProcessSupport.ProcessProgress> getProgress(@PathVariable Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return Result.error("视频不存在");
        }

        VideoProcessSupport.ProcessProgress progress = new VideoProcessSupport.ProcessProgress();
        progress.setProgress(video.getProcessProgress() != null ? video.getProcessProgress() : 0);
        progress.setStage(video.getProcessStage());
        progress.setStatus(getStatusText(video.getStatus()));
        progress.setHasSubtitle(video.getHasSubtitle() != null && video.getHasSubtitle() == 1);
        progress.setHasSummary(video.getHasSummary() != null && video.getHasSummary() == 1);

        return Result.success(progress);
    }

    private String getStatusText(Integer status) {
        if (status == null) return "未知";
        switch (status) {
            case 0: return "待审核";
            case 1: return "处理中";
            case 2: return "待上架";
            case 3: return "已上架";
            case 4: return "审核拒绝";
            case 5: return "处理失败";
            case -1: return "已下架";
            default: return "未知";
        }
    }
}
