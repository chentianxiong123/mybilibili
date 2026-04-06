package com.mybilibili.web.service;

import java.util.List;

/**
 * 标签迁移服务接口
 * 用于将 MySQL 中的标签数据迁移到 Redis
 */
public interface TagMigrationService {

    /**
     * 迁移单个视频的标签到 Redis
     *
     * @param videoId 视频ID
     * @return 是否成功
     */
    boolean migrateVideoTags(Integer videoId);

    /**
     * 迁移多个视频的标签到 Redis
     *
     * @param videoIds 视频ID列表
     * @return 成功迁移的数量
     */
    int migrateVideoTagsBatch(List<Integer> videoIds);

    /**
     * 迁移所有视频的标签到 Redis
     *
     * @return 成功迁移的数量
     */
    int migrateAllVideoTags();

    /**
     * 验证迁移结果
     *
     * @param videoId 视频ID
     * @return 验证结果（MySQL标签 vs Redis标签）
     */
    TagMigrationResult verifyMigration(Integer videoId);

    /**
     * 清空 Redis 中的所有标签数据
     *
     * @return 是否成功
     */
    boolean clearAllRedisTags();

    /**
     * 迁移结果
     */
    class TagMigrationResult {
        private Integer videoId;
        private List<String> mysqlTags;
        private List<String> redisTags;
        private boolean matched;
        private String message;

        public Integer getVideoId() {
            return videoId;
        }

        public void setVideoId(Integer videoId) {
            this.videoId = videoId;
        }

        public List<String> getMysqlTags() {
            return mysqlTags;
        }

        public void setMysqlTags(List<String> mysqlTags) {
            this.mysqlTags = mysqlTags;
        }

        public List<String> getRedisTags() {
            return redisTags;
        }

        public void setRedisTags(List<String> redisTags) {
            this.redisTags = redisTags;
        }

        public boolean isMatched() {
            return matched;
        }

        public void setMatched(boolean matched) {
            this.matched = matched;
        }

        public String getMessage() {
            return message;
        }

        public void setMessage(String message) {
            this.message = message;
        }
    }
}
