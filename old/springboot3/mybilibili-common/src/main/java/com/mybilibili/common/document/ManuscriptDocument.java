package com.mybilibili.common.document;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.elasticsearch.annotations.Document;
import org.springframework.data.elasticsearch.annotations.Field;
import org.springframework.data.elasticsearch.annotations.FieldType;

import java.util.Date;
import java.util.List;

/**
 * 稿件搜索文档
 * 以稿件为单位建立索引，包含稿件下所有视频的信息
 */
@Data
@Document(indexName = "manuscripts")
public class ManuscriptDocument {

    @Id
    private Integer manuscriptId;

    @Field(type = FieldType.Text, analyzer = "ik_max_word")
    private String title;

    @Field(type = FieldType.Text, analyzer = "ik_max_word")
    private String description;

    @Field(type = FieldType.Integer)
    private Integer userId;

    @Field(type = FieldType.Keyword)
    private String userName;

    @Field(type = FieldType.Integer)
    private Integer categoryId;

    @Field(type = FieldType.Keyword)
    private String categoryName;

    @Field(type = FieldType.Keyword)
    private List<String> tags;

    /**
     * 所有视频标题拼接，用于搜索
     */
    @Field(type = FieldType.Text, analyzer = "ik_max_word")
    private String videoTitles;

    @Field(type = FieldType.Integer)
    private List<Integer> videoIds;

    @Field(type = FieldType.Integer)
    private Integer videoCount;

    @Field(type = FieldType.Integer)
    private Integer viewCount;

    @Field(type = FieldType.Integer)
    private Integer likeCount;

    @Field(type = FieldType.Integer)
    private Integer commentCount;

    @Field(type = FieldType.Integer)
    private Integer shareCount;

    @Field(type = FieldType.Integer)
    private Integer collectCount;

    @Field(type = FieldType.Integer)
    private Integer coinCount;

    @Field(type = FieldType.Integer)
    private Integer durationSeconds;

    @Field(type = FieldType.Date)
    private Date uploadTime;

    @Field(type = FieldType.Integer)
    private Integer status;

    @Field(type = FieldType.Keyword)
    private String coverUrl;
}
