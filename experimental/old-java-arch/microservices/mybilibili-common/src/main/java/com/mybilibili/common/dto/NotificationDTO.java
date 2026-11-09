package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

@Data
public class NotificationDTO {
    @NotNull(message = "用户ID不能为空")
    private Integer userId;

    @NotNull(message = "通知类型不能为空")
    private Integer type;

    @NotBlank(message = "标题不能为空")
    @Size(max = 100, message = "标题最多100个字符")
    private String title;

    @NotBlank(message = "内容不能为空")
    @Size(max = 1000, message = "内容最多1000个字符")
    private String content;

    private String relatedId;
}
