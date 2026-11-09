package com.mybilibili.common.enums;

public enum MatchType {
    EXACT("精确匹配"),
    CONTAINS("包含匹配");

    private final String description;

    MatchType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}
