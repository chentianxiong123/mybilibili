package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class MessageSettingDTO {
    private Boolean privateMessageNotify;
    private Boolean replyNotify;
    private Boolean atNotify;
    private Boolean likeNotify;
    private Boolean systemNotify;
}
