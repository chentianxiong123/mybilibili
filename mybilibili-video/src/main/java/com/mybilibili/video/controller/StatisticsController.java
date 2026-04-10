package com.mybilibili.video.controller;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.mapper.VideoMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.*;

@RestController
@RequestMapping("/statistics")
public class StatisticsController {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @GetMapping("/overview")
    public Result<Map<String, Object>> getOverviewStatistics() {
        Map<String, Object> data = new HashMap<>();

        // 稿件总数
        Long manuscriptCount = manuscriptMapper.selectCount(null);
        data.put("manuscriptCount", manuscriptCount);

        // 视频总数
        Long videoCount = videoMapper.selectCount(null);
        data.put("videoCount", videoCount);

        // 用户总数 (这里用稿件作者数代替，实际应该从user服务获取)
        Long userCount = manuscriptMapper.selectCount(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<Manuscript>()
                .select("DISTINCT user_id")
        );
        data.put("userCount", userCount);

        // 播放总数 (这里用稿件浏览量总和代替)
        Integer viewCount = manuscriptMapper.selectSumViewCount();
        data.put("viewCount", viewCount != null ? viewCount : 0);

        // 待审核稿件数
        Long pendingCount = manuscriptMapper.selectCount(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<Manuscript>()
                .eq("status", 0)
        );
        data.put("pendingManuscriptCount", pendingCount);

        return Result.success("获取成功", data);
    }

    @GetMapping("/manuscript/status")
    public Result<List<Map<String, Object>>> getManuscriptStatusStatistics() {
        List<Map<String, Object>> list = new ArrayList<>();

        // 按状态统计稿件数量
        int[] statuses = {0, 1, 2, 3, 4, 5, -1};
        String[] statusNames = {"待审核", "处理中", "待上架", "已上架", "审核拒绝", "处理失败", "已下架"};

        for (int i = 0; i < statuses.length; i++) {
            Long count = manuscriptMapper.selectCount(
                new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<Manuscript>()
                    .eq("status", statuses[i])
            );

            Map<String, Object> item = new HashMap<>();
            item.put("status", statuses[i]);
            item.put("statusName", statusNames[i]);
            item.put("count", count);
            list.add(item);
        }

        return Result.success("获取成功", list);
    }

    @GetMapping("/manuscript/recent")
    public Result<List<Map<String, Object>>> getRecentManuscripts(
            @RequestParam(defaultValue = "10") Integer limit) {

        List<Manuscript> manuscripts = manuscriptMapper.selectList(
            new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<Manuscript>()
                .orderByDesc("upload_time")
                .last("LIMIT " + limit)
        );

        List<Map<String, Object>> list = new ArrayList<>();
        for (Manuscript m : manuscripts) {
            Map<String, Object> item = new HashMap<>();
            item.put("id", m.getId());
            item.put("title", m.getTitle());
            item.put("status", m.getStatus());
            item.put("uploadTime", m.getUploadTime());
            list.add(item);
        }

        return Result.success("获取成功", list);
    }
}
