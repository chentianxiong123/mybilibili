package com.mybilibili.admin.service;

import com.mybilibili.common.entity.Subtitle;

/**
 * 字幕转换服务
 * 提供 SRT 字幕文件转 MongoDB 的独立接口
 */
public interface SubtitleConvertService {

    /**
     * 将 SRT 文件内容解析并保存到 MongoDB
     *
     * @param videoId      视频ID
     * @param srtContent   SRT 文件内容
     * @param language     语言代码，如 "zh-CN"
     * @param languageName 语言名称，如 "中文"
     * @param source       来源标识，如 "whisper", "user", "admin"
     * @param uploadedBy   上传者ID，0 表示系统生成
     * @return 保存后的字幕对象
     */
    Subtitle saveSrtToMongo(Integer videoId, String srtContent, String language,
                            String languageName, String source, Integer uploadedBy);

    /**
     * 从文件路径读取 SRT 并保存到 MongoDB
     *
     * @param videoId      视频ID
     * @param srtFilePath  SRT 文件路径
     * @param language     语言代码
     * @param languageName 语言名称
     * @param source       来源标识
     * @param uploadedBy   上传者ID
     * @return 保存后的字幕对象
     */
    Subtitle saveSrtFileToMongo(Integer videoId, String srtFilePath, String language,
                                String languageName, String source, Integer uploadedBy);

    /**
     * 解析 SRT 内容（不保存，仅返回解析结果）
     *
     * @param srtContent SRT 文件内容
     * @return 解析后的字幕条目列表
     */
    java.util.List<Subtitle.SubtitleItem> parseSrtContent(String srtContent);

    /**
     * 从 MongoDB 删除指定视频的字幕
     *
     * @param videoId  视频ID
     * @param language 语言代码
     */
    void deleteSubtitleFromMongo(Integer videoId, String language);

    /**
     * 查询字幕是否存在
     *
     * @param videoId  视频ID
     * @param language 语言代码
     * @return 是否存在
     */
    boolean existsSubtitle(Integer videoId, String language);
}
