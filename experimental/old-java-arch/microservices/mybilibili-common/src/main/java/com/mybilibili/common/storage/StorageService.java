package com.mybilibili.common.storage;

import java.io.InputStream;

public interface StorageService {

    String upload(String key, InputStream data, long size, String contentType);

    String upload(String key, InputStream data, String contentType);

    InputStream download(String key);

    String getPresignedUrl(String key, int expirySeconds);

    String getPublicUrl(String key);

    boolean exists(String key);

    void delete(String key);

    void deletePrefix(String prefix);

    void copy(String sourceKey, String targetKey);
}
