package com.mybilibili.video.studio;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class StudioAssetUploadResponse {
    private String mediaId;
    private String objectKey;
    private String url;
    private String fileName;
    private String contentType;
    private Long size;
}
