package com.mybilibili.search.service;

import java.util.List;

public interface HotSearchService {

    void incrementHotSearch(String keyword);

    List<HotSearchVO> getHotSearchTop10();

    void cleanExpiredHotSearch();

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