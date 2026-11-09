package com.mybilibili.web.repository;

import com.mybilibili.common.entity.DanmakuDocument;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface DanmakuRepository extends MongoRepository<DanmakuDocument, String> {

    List<DanmakuDocument> findByVideoId(Integer videoId);

    List<DanmakuDocument> findByVideoIdAndTimeBetween(Integer videoId, Double startTime, Double endTime);

    List<DanmakuDocument> findByManuscriptId(Integer manuscriptId);

    long countByVideoId(Integer videoId);
}
