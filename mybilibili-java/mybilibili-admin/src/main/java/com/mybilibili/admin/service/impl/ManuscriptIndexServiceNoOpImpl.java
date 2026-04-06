package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.service.ManuscriptIndexService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.Map;

@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "false", matchIfMissing = true)
public class ManuscriptIndexServiceNoOpImpl implements ManuscriptIndexService {

    private static final String MSG = "Elasticsearch未启用，索引功能不可用";

    @Override
    public Map<String, Object> bulkIndexAllPublished() {
        log.warn(MSG);
        Map<String, Object> result = new HashMap<>();
        result.put("success", false);
        result.put("message", MSG);
        return result;
    }

    @Override
    public Map<String, Object> rebuildIndex() {
        log.warn(MSG);
        Map<String, Object> result = new HashMap<>();
        result.put("success", false);
        result.put("message", MSG);
        return result;
    }

    @Override
    public Map<String, Object> incrementalIndex(int minutes) {
        log.warn(MSG);
        Map<String, Object> result = new HashMap<>();
        result.put("success", false);
        result.put("message", MSG);
        return result;
    }

    @Override
    public Map<String, Object> getIndexStatus() {
        Map<String, Object> result = new HashMap<>();
        result.put("indexedCount", 0);
        result.put("indexName", "manuscripts");
        result.put("status", "disabled");
        result.put("message", MSG);
        return result;
    }
}
