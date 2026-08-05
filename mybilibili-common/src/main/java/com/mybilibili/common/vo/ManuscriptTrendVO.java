package com.mybilibili.common.vo;

import lombok.Data;
import java.util.List;

@Data
public class ManuscriptTrendVO {
    private List<String> dates;
    private List<Integer> views;
    private List<Integer> danmaku;
    private List<String> titles;
}