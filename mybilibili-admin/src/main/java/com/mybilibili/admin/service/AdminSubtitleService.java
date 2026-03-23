package com.mybilibili.admin.service;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.vo.Result;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.Map;

/**
 * 管理员字幕服务接口
 */
public interface AdminSubtitleService {

    /**
     * 获取带字幕信息的视频列表
     */
    Result<Map<String, Object>> getVideosWithSubtitleInfo(Integer page, Integer size);

    /**
     * 获取指定视频的所有字幕
     */
    List<Subtitle> getSubtitlesByVideoId(Integer videoId);

    /**
     * 管理员上传字幕（直接入库，无需审核）
     */
    Subtitle uploadSubtitle(Integer videoId, MultipartFile file, String language, String languageName, Integer uploadedBy);

    /**
     * SRT文件入库（将磁盘上的SRT文件导入MongoDB）
     */
    Subtitle importSrtToMongo(Integer videoId, String filePath, String language, String languageName, String source);

    /**
     * 审核通过字幕
     */
    void approveSubtitle(String subtitleId, Integer reviewerId);

    /**
     * 审核拒绝字幕
     */
    void rejectSubtitle(String subtitleId, Integer reviewerId, String reason);

    /**
     * 设为默认字幕
     */
    void setDefaultSubtitle(String subtitleId);

    /**
     * 删除字幕
     */
    void deleteSubtitle(String subtitleId);

    /**
     * 获取待审核字幕列表
     */
    List<Map<String, Object>> getPendingSubtitles();

    /**
     * 预览字幕内容
     */
    Map<String, Object> previewSubtitle(String subtitleId);

    /**
     * 扫描系统生成的字幕文件
     * @return 返回待入库的字幕文件列表 [{language: "zh-CN", fileName: "zh-CN.srt", status: "pending|imported"}]
     */
    List<Map<String, Object>> scanSystemSubtitles(Integer videoId);

    /**
     * 将系统字幕文件入库到MongoDB
     */
    Subtitle importSystemSubtitle(Integer videoId, String language);
}
