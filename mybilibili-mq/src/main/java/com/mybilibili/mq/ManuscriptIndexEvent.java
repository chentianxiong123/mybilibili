package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

/**
 * 稿件 ES 索引事件
 * UPSERT: 上架(自动) / 重新上架 → search-recommend 索引该稿件
 * DELETE: 下架 → search-recommend 删除该稿件的 ES 文档
 */
@Data
public class ManuscriptIndexEvent implements Serializable {

    private static final long serialVersionUID = 1L;

    public static final String OPERATION_UPSERT = "UPSERT";
    public static final String OPERATION_DELETE = "DELETE";

    private Integer manuscriptId;
    private String operation;
    private String trigger;  // AUTO_PUBLISH / OWNER_REPUBLISH / OWNER_UNPUBLISH / ADMIN_UNPUBLISH
}
