package com.mybilibili.common.vo;

public class InteractionResponse {
    private boolean status;    // 当前状态：true=已点赞/已收藏/已关注, false=未点赞/未收藏/未关注
    private int count;         // 当前总数
    private String action;     // 执行的操作：like/unlike/collect/uncollect/follow/unfollow

    public InteractionResponse() {
    }

    public InteractionResponse(boolean status, int count, String action) {
        this.status = status;
        this.count = count;
        this.action = action;
    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public int getCount() {
        return count;
    }

    public void setCount(int count) {
        this.count = count;
    }

    public String getAction() {
        return action;
    }

    public void setAction(String action) {
        this.action = action;
    }
}
