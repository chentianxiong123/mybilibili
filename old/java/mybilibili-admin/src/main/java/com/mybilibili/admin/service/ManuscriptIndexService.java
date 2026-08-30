package com.mybilibili.admin.service;

import java.util.Map;

public interface ManuscriptIndexService {

    /**
     * 批量索引所有已上架稿件
     *
     * @return 索引结果信息
     */
    Map<String, Object> bulkIndexAllPublished();

    /**
     * 重建索引（清空后重新导入）
     *
     * @return 索引结果信息
     */
    Map<String, Object> rebuildIndex();

    /**
     * 增量索引最近上架的稿件
     *
     * @param minutes 最近多少分钟内上架的稿件
     * @return 索引结果信息
     */
    Map<String, Object> incrementalIndex(int minutes);

    /**
     * 获取索引状态
     *
     * @return 索引状态信息
     */
    Map<String, Object> getIndexStatus();
}
