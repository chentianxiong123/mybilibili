package com.mybilibili.web.repository;

import com.mybilibili.common.document.ManuscriptDocument;
import org.springframework.data.elasticsearch.repository.ElasticsearchRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ManuscriptSearchRepository extends ElasticsearchRepository<ManuscriptDocument, Integer> {
}
