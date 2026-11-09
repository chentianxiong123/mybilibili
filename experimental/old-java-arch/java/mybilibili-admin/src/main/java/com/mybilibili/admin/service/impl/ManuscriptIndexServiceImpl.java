package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.CategoryMapper;
import com.mybilibili.admin.mapper.ManuscriptMapper;
import com.mybilibili.admin.mapper.UserMapper;
import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.admin.repository.ManuscriptSearchRepository;
import com.mybilibili.admin.service.ManuscriptIndexService;
import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.common.entity.Category;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.Video;
import lombok.extern.slf4j.Slf4j;
import org.elasticsearch.index.query.QueryBuilders;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.data.elasticsearch.core.ElasticsearchRestTemplate;
import org.springframework.data.elasticsearch.core.IndexOperations;
import org.springframework.data.elasticsearch.core.document.Document;
import org.springframework.data.elasticsearch.core.mapping.IndexCoordinates;
import org.springframework.data.elasticsearch.core.query.IndexQuery;
import org.springframework.data.elasticsearch.core.query.IndexQueryBuilder;
import org.springframework.data.elasticsearch.core.query.NativeSearchQuery;
import org.springframework.data.elasticsearch.core.query.NativeSearchQueryBuilder;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "true", matchIfMissing = false)
public class ManuscriptIndexServiceImpl implements ManuscriptIndexService {

    @Autowired(required = false)
    private ManuscriptSearchRepository manuscriptSearchRepository;

    @Autowired(required = false)
    private ElasticsearchRestTemplate elasticsearchRestTemplate;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private CategoryMapper categoryMapper;

    private static final String INDEX_NAME = "manuscripts";

    @Override
    public Map<String, Object> bulkIndexAllPublished() {
        Map<String, Object> result = new HashMap<>();
        long startTime = System.currentTimeMillis();

        try {
            log.info("开始批量索引所有已上架稿件...");

            // 查询所有已上架的稿件（status = 3）
            List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PUBLISHED);

            if (manuscripts == null || manuscripts.isEmpty()) {
                result.put("success", true);
                result.put("message", "没有需要索引的稿件");
                result.put("indexedCount", 0);
                return result;
            }

            // 构建索引文档
            List<ManuscriptDocument> documents = manuscripts.stream()
                    .map(this::buildManuscriptDocument)
                    .filter(Objects::nonNull)
                    .collect(Collectors.toList());

            // 使用ElasticsearchRestTemplate批量保存
            List<IndexQuery> queries = documents.stream()
                    .map(doc -> new IndexQueryBuilder()
                            .withId(doc.getManuscriptId().toString())
                            .withObject(doc)
                            .build())
                    .collect(Collectors.toList());

            elasticsearchRestTemplate.bulkIndex(queries, IndexCoordinates.of(INDEX_NAME));

            long endTime = System.currentTimeMillis();
            log.info("批量索引完成，共索引 {} 个稿件，耗时 {} ms", documents.size(), (endTime - startTime));

            result.put("success", true);
            result.put("message", "批量索引成功");
            result.put("indexedCount", documents.size());
            result.put("timeMillis", (endTime - startTime));

        } catch (Exception e) {
            log.error("批量索引失败: {}", e.getMessage(), e);
            result.put("success", false);
            result.put("message", "批量索引失败: " + e.getMessage());
        }

        return result;
    }

    @Override
    public Map<String, Object> rebuildIndex() {
        Map<String, Object> result = new HashMap<>();
        long startTime = System.currentTimeMillis();

        try {
            log.info("开始重建索引...");

            // 1. 清空所有索引
            IndexOperations indexOperations = elasticsearchRestTemplate.indexOps(IndexCoordinates.of(INDEX_NAME));
            if (indexOperations.exists()) {
                indexOperations.delete();
            }
            indexOperations.create();
            indexOperations.putMapping(indexOperations.createMapping(ManuscriptDocument.class));
            log.info("已清空并重建索引");

            // 2. 重新导入所有已上架稿件
            List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PUBLISHED);

            if (manuscripts == null || manuscripts.isEmpty()) {
                result.put("success", true);
                result.put("message", "索引已重建，没有需要导入的稿件");
                result.put("indexedCount", 0);
                return result;
            }

            List<ManuscriptDocument> documents = manuscripts.stream()
                    .map(this::buildManuscriptDocument)
                    .filter(Objects::nonNull)
                    .collect(Collectors.toList());

            // 使用ElasticsearchRestTemplate批量保存
            List<IndexQuery> queries = documents.stream()
                    .map(doc -> new IndexQueryBuilder()
                            .withId(doc.getManuscriptId().toString())
                            .withObject(doc)
                            .build())
                    .collect(Collectors.toList());

            elasticsearchRestTemplate.bulkIndex(queries, IndexCoordinates.of(INDEX_NAME));

            long endTime = System.currentTimeMillis();
            log.info("重建索引完成，共索引 {} 个稿件，耗时 {} ms", documents.size(), (endTime - startTime));

            result.put("success", true);
            result.put("message", "重建索引成功");
            result.put("indexedCount", documents.size());
            result.put("timeMillis", (endTime - startTime));

        } catch (Exception e) {
            log.error("重建索引失败: {}", e.getMessage(), e);
            result.put("success", false);
            result.put("message", "重建索引失败: " + e.getMessage());
        }

        return result;
    }

    @Override
    public Map<String, Object> incrementalIndex(int minutes) {
        Map<String, Object> result = new HashMap<>();
        long startTime = System.currentTimeMillis();

        try {
            log.info("开始增量索引，时间范围：最近 {} 分钟", minutes);

            // 查询最近上架的稿件
            List<Manuscript> manuscripts = manuscriptMapper.selectRecentlyPublished(minutes);

            if (manuscripts == null || manuscripts.isEmpty()) {
                result.put("success", true);
                result.put("message", "指定时间范围内没有新上架的稿件");
                result.put("indexedCount", 0);
                return result;
            }

            // 构建索引文档
            List<ManuscriptDocument> documents = manuscripts.stream()
                    .map(this::buildManuscriptDocument)
                    .filter(Objects::nonNull)
                    .collect(Collectors.toList());

            // 使用ElasticsearchRestTemplate批量保存
            List<IndexQuery> queries = documents.stream()
                    .map(doc -> new IndexQueryBuilder()
                            .withId(doc.getManuscriptId().toString())
                            .withObject(doc)
                            .build())
                    .collect(Collectors.toList());

            elasticsearchRestTemplate.bulkIndex(queries, IndexCoordinates.of(INDEX_NAME));

            long endTime = System.currentTimeMillis();
            log.info("增量索引完成，共索引 {} 个稿件，耗时 {} ms", documents.size(), (endTime - startTime));

            result.put("success", true);
            result.put("message", "增量索引成功");
            result.put("indexedCount", documents.size());
            result.put("timeMillis", (endTime - startTime));

        } catch (Exception e) {
            log.error("增量索引失败: {}", e.getMessage(), e);
            result.put("success", false);
            result.put("message", "增量索引失败: " + e.getMessage());
        }

        return result;
    }

    @Override
    public Map<String, Object> getIndexStatus() {
        Map<String, Object> result = new HashMap<>();

        try {
            // 使用ElasticsearchRestTemplate查询索引中的文档数量
            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                    .withQuery(QueryBuilders.matchAllQuery())
                    .build();
            long count = elasticsearchRestTemplate.count(searchQuery, ManuscriptDocument.class, IndexCoordinates.of(INDEX_NAME));
            
            result.put("indexedCount", count);
            result.put("indexName", INDEX_NAME);
            result.put("status", "active");
        } catch (Exception e) {
            log.error("获取索引状态失败: {}", e.getMessage(), e);
            result.put("indexedCount", 0);
            result.put("indexName", INDEX_NAME);
            result.put("status", "error");
            result.put("error", e.getMessage());
        }

        return result;
    }

    /**
     * 构建稿件索引文档
     *
     * @param manuscript 稿件实体
     * @return ManuscriptDocument
     */
    private ManuscriptDocument buildManuscriptDocument(Manuscript manuscript) {
        if (manuscript == null || manuscript.getId() == null) {
            return null;
        }

        try {
            ManuscriptDocument document = new ManuscriptDocument();
            document.setManuscriptId(manuscript.getId());
            document.setTitle(manuscript.getTitle());
            document.setDescription(manuscript.getDescription());
            document.setCoverUrl(manuscript.getCoverUrl());
            document.setUserId(manuscript.getUserId());
            document.setCategoryId(manuscript.getCategoryId());
            document.setViewCount(manuscript.getViewCount());
            document.setLikeCount(manuscript.getLikeCount());
            document.setCommentCount(manuscript.getCommentCount());
            document.setShareCount(manuscript.getShareCount());
            document.setCollectCount(manuscript.getCollectCount());
            document.setCoinCount(manuscript.getCoinCount());
            document.setDurationSeconds(manuscript.getDurationSeconds());
            document.setUploadTime(manuscript.getUploadTime());
            document.setStatus(manuscript.getStatus());

            // 查询用户名
            if (manuscript.getUserId() != null) {
                User user = userMapper.selectById(manuscript.getUserId());
                if (user != null) {
                    document.setUserName(user.getUsername());
                }
            }

            // 查询分类名
            if (manuscript.getCategoryId() != null) {
                Category category = categoryMapper.selectById(manuscript.getCategoryId());
                if (category != null) {
                    document.setCategoryName(category.getName());
                }
            }

            // 查询该稿件下的所有视频
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            if (videos != null && !videos.isEmpty()) {
                // 视频ID列表
                List<Integer> videoIds = videos.stream()
                        .map(Video::getId)
                        .filter(Objects::nonNull)
                        .collect(Collectors.toList());
                document.setVideoIds(videoIds);
                document.setVideoCount(videos.size());

                // 视频标题拼接（用于搜索）
                String videoTitles = videos.stream()
                        .map(Video::getTitle)
                        .filter(Objects::nonNull)
                        .collect(Collectors.joining(" "));
                document.setVideoTitles(videoTitles);

                // 收集标签（从所有视频中收集）
                Set<String> tagSet = new HashSet<>();
                for (Video video : videos) {
                    if (video.getId() != null) {
                        List<String> tags = videoMapper.selectTagsByVideoId(video.getId());
                        if (tags != null) {
                            tagSet.addAll(tags);
                        }
                    }
                }
                document.setTags(new ArrayList<>(tagSet));
            } else {
                document.setVideoIds(new ArrayList<>());
                document.setVideoCount(0);
                document.setVideoTitles("");
                document.setTags(new ArrayList<>());
            }

            return document;
        } catch (Exception e) {
            log.error("构建稿件索引文档失败，manuscriptId: {}, error: {}", manuscript.getId(), e.getMessage());
            return null;
        }
    }
}
