package com.mybilibili.common.entity;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;

import java.util.Date;
import java.util.List;

@Data
@Document(collection = "subtitles")
public class Subtitle {

    @Id
    private String id;

    @Indexed
    private Integer videoId;

    private String language;

    private String languageName;

    private String format;

    private List<SubtitleItem> content;

    private Boolean isDefault;

    private Integer uploadedBy;

    private Date uploadTime;

    private Integer status;              // 0=待审核, 1=审核通过, 2=审核拒绝, 3=系统生成

    private String source;               // 来源：whisper-系统生成, user-用户上传, admin-管理员上传

    private Integer version;

    // 字幕审核相关字段
    private String uploadId;             // 上传批次ID，用于关联文件目录

    private Date reviewTime;             // 审核时间

    private Integer reviewerId;          // 审核管理员ID

    private String reviewReason;         // 审核备注/拒绝原因

    private String filePath;             // SRT 文件存储路径

    @Data
    public static class SubtitleItem {
        private Integer index;
        private Double startTime;
        private Double endTime;
        private String text;
    }
}
