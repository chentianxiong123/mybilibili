package com.mybilibili.search.service;

public interface ManuscriptIndexService {
    int bulkIndex();
    int incrementalIndex();
    int rebuildIndex();
    boolean indexOne(Integer manuscriptId);
    boolean deleteOne(Integer manuscriptId);
}
