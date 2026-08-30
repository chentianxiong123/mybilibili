package com.mybilibili.web.service;

import java.util.List;

/**
 * 热搜榜服务接口
 */
public interface HotSearchService {

    /**
     * 增加搜索关键词热度
     *
     * @param keyword 搜索关键词
     */
    void incrementHotSearch(String keyword);

    /**
     * 获取热搜榜 Top10
     *
     * @return 热搜关键词列表
     */
    List<HotSearchVO> getHotSearchTop10();

    /**
     * 清理过期的热搜数据
     */
    void cleanExpiredHotSearch();

    /**
     * 热搜榜VO
     */
    class HotSearchVO {
        private String keyword;
        private Double score;
        private Integer rank;

        public HotSearchVO() {
        }

        public HotSearchVO(String keyword, Double score, Integer rank) {
            this.keyword = keyword;
            this.score = score;
            this.rank = rank;
        }

        public String getKeyword() {
            return keyword;
        }

        public void setKeyword(String keyword) {
            this.keyword = keyword;
        }

        public Double getScore() {
            return score;
        }

        public void setScore(Double score) {
            this.score = score;
        }

        public Integer getRank() {
            return rank;
        }

        public void setRank(Integer rank) {
            this.rank = rank;
        }
    }
}
