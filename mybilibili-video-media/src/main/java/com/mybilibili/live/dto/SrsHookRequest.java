package com.mybilibili.live.dto;

import lombok.Data;

@Data
public class SrsHookRequest {
    private String action;
    private String stream;
    private String app;
    private String vhost;
}
