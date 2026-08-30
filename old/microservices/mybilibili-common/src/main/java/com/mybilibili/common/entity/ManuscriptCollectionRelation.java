package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("manuscript_collection_relations")
public class ManuscriptCollectionRelation {
    @TableId(type = IdType.AUTO)
    private Integer id;
    private Integer manuscriptId;
    private Integer collectionId;
    private Integer collectionOrder;
    private Date createdAt;
    
    // 非数据库字段
    private Manuscript manuscript;  // 关联的稿件
    private ManuscriptCollection collection;  // 关联的合集
}
