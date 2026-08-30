package com.mybilibili.ai.tool;

import java.util.Map;

public class StatsData {
    private String type;
    private Map<String, Object> data;
    private String chartType;
    private String title;

    public StatsData() {
    }

    public StatsData(String type, Map<String, Object> data, String chartType, String title) {
        this.type = type;
        this.data = data;
        this.chartType = chartType;
        this.title = title;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public Map<String, Object> getData() {
        return data;
    }

    public void setData(Map<String, Object> data) {
        this.data = data;
    }

    public String getChartType() {
        return chartType;
    }

    public void setChartType(String chartType) {
        this.chartType = chartType;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }
}