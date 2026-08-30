package com.mybilibili.common.enums;

public enum InteractionType {
    LIKE("LIKE", "点赞"),
    COLLECT("COLLECT", "收藏"),
    FOLLOW("FOLLOW", "关注"),
    COIN("COIN", "投币"),
    SHARE("SHARE", "分享");

    private final String code;
    private final String desc;

    InteractionType(String code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    public String getCode() {
        return code;
    }

    public String getDesc() {
        return desc;
    }

    public static InteractionType fromCode(String code) {
        for (InteractionType type : values()) {
            if (type.code.equals(code)) {
                return type;
            }
        }
        throw new IllegalArgumentException("Unknown interaction type: " + code);
    }
}
