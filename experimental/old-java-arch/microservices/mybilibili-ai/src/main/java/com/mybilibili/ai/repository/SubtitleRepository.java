package com.mybilibili.ai.repository;

import com.mybilibili.common.entity.Subtitle;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface SubtitleRepository extends MongoRepository<Subtitle, String> {

    List<Subtitle> findByVideoId(Integer videoId);

    Optional<Subtitle> findFirstByVideoIdAndLanguage(Integer videoId, String language);

    List<Subtitle> findAllByVideoIdAndLanguage(Integer videoId, String language);

    List<Subtitle> findByVideoIdAndStatus(Integer videoId, Integer status);

    List<Subtitle> findByUploadedBy(Integer uploadedBy);

    List<Subtitle> findByStatus(Integer status);
}
