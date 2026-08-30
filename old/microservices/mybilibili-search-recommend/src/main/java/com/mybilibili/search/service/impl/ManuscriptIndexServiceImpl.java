package com.mybilibili.search.service.impl;

import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.search.mapper.ManuscriptMapper;
import com.mybilibili.search.service.ManuscriptIndexService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.elasticsearch.client.elc.ElasticsearchTemplate;
import org.springframework.data.elasticsearch.core.IndexOperations;
import org.springframework.data.elasticsearch.core.query.Query;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class ManuscriptIndexServiceImpl implements ManuscriptIndexService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private ElasticsearchTemplate elasticsearchTemplate;

    @Override
    public int bulkIndex() {
        List<Map<String, Object>> rows = manuscriptMapper.selectPublishedManuscripts();
        if (rows.isEmpty()) {
            log.info("没有需要索引的已上架稿件");
            return 0;
        }

        List<ManuscriptDocument> documents = rows.stream()
                .map(this::convertToDocument)
                .filter(Objects::nonNull)
                .collect(Collectors.toList());

        elasticsearchTemplate.save(documents);
        log.info("批量索引完成，共索引 {} 条稿件", documents.size());
        return documents.size();
    }

    @Override
    public int incrementalIndex() {
        List<Map<String, Object>> allPublished = manuscriptMapper.selectPublishedManuscripts();
        if (allPublished.isEmpty()) {
            log.info("没有已上架稿件");
            return 0;
        }

        // 查询 ES 中已存在的稿件 ID
        Set<Integer> existingIds = getExistingIds();

        // 过滤出 ES 中不存在的稿件
        List<Map<String, Object>> missing = allPublished.stream()
                .filter(row -> {
                    Integer id = toInt(row.get("id"));
                    return id != null && !existingIds.contains(id);
                })
                .collect(Collectors.toList());

        if (missing.isEmpty()) {
            log.info("没有缺失的稿件需要索引");
            return 0;
        }

        List<ManuscriptDocument> documents = missing.stream()
                .map(this::convertToDocument)
                .filter(Objects::nonNull)
                .collect(Collectors.toList());

        elasticsearchTemplate.save(documents);
        log.info("增量索引完成，共索引 {} 条缺失稿件", documents.size());
        return documents.size();
    }

    @Override
    public int rebuildIndex() {
        IndexOperations indexOps = elasticsearchTemplate.indexOps(ManuscriptDocument.class);

        if (indexOps.exists()) {
            indexOps.delete();
        }
        indexOps.create();
        indexOps.putMapping();

        log.info("索引已重建，开始导入数据");
        return bulkIndex();
    }

    @Override
    public boolean indexOne(Integer manuscriptId) {
        List<Map<String, Object>> rows = manuscriptMapper.selectPublishedManuscriptById(manuscriptId);
        if (rows == null || rows.isEmpty()) {
            log.warn("稿件不存在或未上架: manuscriptId={}", manuscriptId);
            return false;
        }
        ManuscriptDocument doc = convertToDocument(rows.get(0));
        if (doc == null) {
            log.error("稿件转换失败: manuscriptId={}", manuscriptId);
            return false;
        }
        elasticsearchTemplate.save(doc);
        log.info("单条索引成功: manuscriptId={}", manuscriptId);
        return true;
    }

    @Override
    public boolean deleteOne(Integer manuscriptId) {
        try {
            String deleteId = manuscriptId.toString();
            elasticsearchTemplate.delete(deleteId, ManuscriptDocument.class);
            log.info("稿件索引删除成功: manuscriptId={}", manuscriptId);
            return true;
        } catch (Exception e) {
            log.error("稿件索引删除失败: manuscriptId={}, error={}", manuscriptId, e.getMessage());
            return false;
        }
    }

    private Set<Integer> getExistingIds() {
        Set<Integer> ids = new HashSet<>();
        try {
            Query query = Query.findAll();
            query.addFields("manuscriptId");
            elasticsearchTemplate.search(query, ManuscriptDocument.class)
                    .forEach(hit -> {
                        ManuscriptDocument doc = hit.getContent();
                        if (doc != null && doc.getManuscriptId() != null) {
                            ids.add(doc.getManuscriptId());
                        }
                    });
        } catch (Exception e) {
            log.warn("查询已有索引 ID 失败，将视为全部缺失: {}", e.getMessage());
        }
        return ids;
    }

    private ManuscriptDocument convertToDocument(Map<String, Object> row) {
        try {
            ManuscriptDocument doc = new ManuscriptDocument();
            doc.setManuscriptId(toInt(row.get("id")));
            doc.setTitle(toStr(row.get("title")));
            doc.setDescription(toStr(row.get("description")));
            doc.setUserId(toInt(row.get("user_id")));
            doc.setUserName(toStr(firstPresent(row, "userName", "user_name", "username")));
            doc.setUserAvatar(toStr(firstPresent(row, "userAvatar", "user_avatar", "avatar")));
            doc.setCategoryId(toInt(row.get("category_id")));
            doc.setCategoryName(toStr(row.get("categoryName")));
            doc.setViewCount(toInt(row.get("view_count")));
            doc.setLikeCount(toInt(row.get("like_count")));
            doc.setCommentCount(toInt(row.get("comment_count")));
            doc.setShareCount(toInt(row.get("share_count")));
            doc.setCollectCount(toInt(row.get("collect_count")));
            doc.setCoinCount(toInt(row.get("coin_count")));
            doc.setDurationSeconds(toInt(row.get("duration_seconds")));
            doc.setUploadTime(toDate(firstPresent(row, "upload_time", "uploadTime")));
            doc.setStatus(toInt(row.get("status")));
            doc.setCoverUrl(toStr(row.get("cover_url")));

            // 视频信息
            String videoTitles = toStr(row.get("videoTitles"));
            doc.setVideoTitles(videoTitles != null && !videoTitles.isEmpty() ? videoTitles.replace("||", " ") : null);

            String videoIdsStr = toStr(row.get("videoIds"));
            if (videoIdsStr != null && !videoIdsStr.isEmpty()) {
                doc.setVideoIds(Arrays.stream(videoIdsStr.split(","))
                        .map(String::trim)
                        .filter(s -> !s.isEmpty())
                        .map(Integer::parseInt)
                        .collect(Collectors.toList()));
            } else {
                doc.setVideoIds(Collections.emptyList());
            }

            doc.setVideoCount(toInt(row.get("videoCount")));

            // 标签（需要额外查询）
            if (doc.getManuscriptId() != null) {
                List<String> tags = manuscriptMapper.selectTagsByManuscriptId(doc.getManuscriptId());
                doc.setTags(tags != null ? tags : Collections.emptyList());
            }

            return doc;
        } catch (Exception e) {
            log.error("转换稿件文档失败: {}", e.getMessage(), e);
            return null;
        }
    }

    private Integer toInt(Object obj) {
        if (obj == null) return null;
        if (obj instanceof Number) return ((Number) obj).intValue();
        try { return Integer.parseInt(obj.toString()); } catch (NumberFormatException e) { return null; }
    }

    private String toStr(Object obj) {
        return obj != null ? obj.toString() : null;
    }

    private Date toDate(Object obj) {
        if (obj instanceof Date) return (Date) obj;
        if (obj instanceof LocalDateTime) {
            return Date.from(((LocalDateTime) obj).atZone(ZoneId.systemDefault()).toInstant());
        }
        if (obj instanceof LocalDate) {
            return Date.from(((LocalDate) obj).atStartOfDay(ZoneId.systemDefault()).toInstant());
        }
        if (obj instanceof Number) {
            return new Date(((Number) obj).longValue());
        }
        if (obj instanceof CharSequence) {
            String value = obj.toString().trim();
            if (value.isEmpty()) return null;
            try {
                return Date.from(java.time.OffsetDateTime.parse(value).toInstant());
            } catch (Exception ignored) {
                // Try common MySQL datetime format below.
            }
            try {
                return Date.from(LocalDateTime.parse(value.replace(' ', 'T'))
                        .atZone(ZoneId.systemDefault())
                        .toInstant());
            } catch (Exception ignored) {
                return null;
            }
        }
        return null;
    }

    private Object firstPresent(Map<String, Object> row, String... keys) {
        if (row == null || keys == null) return null;
        for (String key : keys) {
            if (row.containsKey(key)) {
                return row.get(key);
            }
        }
        return null;
    }
}
