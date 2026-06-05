package com.mybilibili.search.service;

public interface ManuscriptIndexService {
    int bulkIndex();
    int incrementalIndex();
    int rebuildIndex();
}
