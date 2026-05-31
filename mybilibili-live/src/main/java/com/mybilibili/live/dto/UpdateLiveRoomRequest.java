package com.mybilibili.live.dto;

import lombok.Data;

@Data
public class UpdateLiveRoomRequest {
    private String coverUrl;
    private String roomName;
    private String category;
}
