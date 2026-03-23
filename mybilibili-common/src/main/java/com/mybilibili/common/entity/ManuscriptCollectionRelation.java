package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class ManuscriptCollectionRelation {
    private Integer id;
    private Integer manuscriptId;
    private Integer collectionId;
    private Integer collectionOrder;
    private Date createdAt;
    
    // 非数据库字段
    private Manuscript manuscript;  // 关联的稿件
    private ManuscriptCollection collection;  // 关联的合集
}
