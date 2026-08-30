package com.mybilibili.search.document;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.elasticsearch.annotations.Document;
import org.springframework.data.elasticsearch.annotations.Field;
import org.springframework.data.elasticsearch.annotations.FieldType;

import java.io.Serializable;
import java.util.Date;
import java.util.List;

@Data
@Document(indexName = "manuscripts")
public class ManuscriptDocument implements Serializable {

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