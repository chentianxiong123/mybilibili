package com.mybilibili.search.repository;

import com.mybilibili.search.document.ManuscriptDocument;
import org.springframework.data.elasticsearch.repository.ElasticsearchRepository;

public interface ManuscriptRepository extends ElasticsearchRepository<ManuscriptDocument, Integer> {
}
