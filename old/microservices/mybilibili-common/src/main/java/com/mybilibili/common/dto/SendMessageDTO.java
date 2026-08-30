package com.mybilibili.common.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

@Data
public class SendMessageDTO {
    @NotNull(message = "接收者ID不能为空")
    private Integer receiverId;

    @NotBlank(message = "消息内容不能为空")
    @Size(max = 500, message = "消息内容最多500个字符")
    private String content;

    private Integer messageType;
    private Integer targetId;
    private String mediaUrl;
    private Integer commentId;
}
