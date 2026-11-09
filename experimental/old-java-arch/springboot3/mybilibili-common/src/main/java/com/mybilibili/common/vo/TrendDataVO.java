package com.mybilibili.common.vo;

import lombok.Data;
import java.util.List;

@Data
public class TrendDataVO {
    private List<String> dates;
    private List<Integer> views;
    private List<Integer> likes;
    private List<Integer> followers;
    private List<Integer> comments;
    private List<Integer> danmaku;
    private List<Integer> coins;
    private List<Integer> collects;
    private List<Integer> shares;
}
