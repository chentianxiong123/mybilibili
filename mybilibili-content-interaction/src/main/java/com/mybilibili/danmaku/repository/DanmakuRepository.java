package com.mybilibili.danmaku.repository;

import com.mybilibili.danmaku.entity.Danmaku;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.time.LocalDateTime;
import java.util.List;

public interface DanmakuRepository extends MongoRepository<Danmaku, String> {
    List<Danmaku> findByVideoId(Integer videoId);

    List<Danmaku> findByVideoIdAndTimeBetween(Integer videoId, Double startTime, Double endTime);

    List<Danmaku> findByManuscriptId(Integer manuscriptId);

    long countByVideoId(Integer videoId);

    long countByManuscriptIdIn(List<Integer> manuscriptIds);

    List<Danmaku> findByManuscriptIdInAndCreateTimeBetween(List<Integer> manuscriptIds, LocalDateTime start, LocalDateTime end);

    Page<Danmaku> findByManuscriptIdIn(List<Integer> manuscriptIds, Pageable pageable);

    Page<Danmaku> findByManuscriptIdAndManuscriptIdIn(Integer manuscriptId, List<Integer> manuscriptIds, Pageable pageable);

    Page<Danmaku> findByVideoIdIn(List<Integer> videoIds, Pageable pageable);

    Page<Danmaku> findByVideoIdAndVideoIdIn(Integer videoId, List<Integer> videoIds, Pageable pageable);
}
