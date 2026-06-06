package com.mybilibili.video.service.impl;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.dto.ManuscriptUpdateDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Tag;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.common.vo.ManuscriptEditVersionVO;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.TypeReference;
import com.mybilibili.mq.ManuscriptAnalyticsEvent;
import com.mybilibili.mq.UserEventMQProducer;
import com.mybilibili.mq.UserExperienceEvent;
import com.mybilibili.mq.UserNotificationEvent;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.video.entity.ManuscriptEditVersion;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.mapper.CategoryMapper;
import com.mybilibili.video.mapper.ManuscriptEditVersionMapper;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.mapper.TagMapper;
import com.mybilibili.video.mapper.VideoMapper;
import com.mybilibili.video.service.ManuscriptService;
import com.mybilibili.video.service.VideoProcessPort;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.transaction.support.TransactionSynchronization;
import org.springframework.transaction.support.TransactionSynchronizationManager;
import org.springframework.web.multipart.MultipartFile;

import java.net.URI;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collections;
import java.util.Date;
import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
public class ManuscriptServiceImpl implements ManuscriptService {

    private static final int MANUSCRIPT_EXPERIENCE = 100;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private ManuscriptEditVersionMapper manuscriptEditVersionMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private CategoryMapper categoryMapper;

    @Autowired
    private TagMapper tagMapper;

    @Autowired
    private UserClient userClient;

    @Autowired
    private StorageService storageService;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Autowired
    private UserEventMQProducer userEventMQProducer;

    @Autowired
    private VideoProcessPort videoProcessPort;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptVO uploadManuscript(ManuscriptUploadDTO dto, Integer userId) throws Exception {
        log.info("开始上传稿件，用户ID: {}, 标题: {}", userId, dto.getTitle());

        Manuscript manuscript = new Manuscript();
        manuscript.setTitle(dto.getTitle());
        manuscript.setDescription(dto.getDescription());
        manuscript.setUserId(userId);
        manuscript.setCategoryId(dto.getCategoryId());
        manuscript.setStatus(Manuscript.STATUS_PENDING_REVIEW);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_PENDING);
        manuscript.setUploadTime(new Date());

        manuscriptMapper.insert(manuscript);
        Integer manuscriptId = manuscript.getId();
        recordManuscriptStatusEvent(manuscript, null, Manuscript.STATUS_PENDING_REVIEW,
                "UPLOAD_SUBMITTED", "USER", userId, null);
        log.info("稿件记录创建成功，manuscriptId: {}", manuscriptId);

        userEventMQProducer.sendExperienceEvent(UserExperienceEvent.of(
                userId, MANUSCRIPT_EXPERIENCE, "MANUSCRIPT_UPLOAD", manuscriptId, "投稿"));

        if (dto.getCover() != null && !dto.getCover().isEmpty()) {
            String coverKey = StorageKeys.manuscriptCover(manuscriptId);
            String coverUrl = storageService.upload(coverKey, dto.getCover().getInputStream(), dto.getCover().getSize(), dto.getCover().getContentType());
            manuscript.setCoverUrl(coverUrl);
            manuscriptMapper.updateById(manuscript);
            log.info("封面保存到MinIO成功，manuscriptId: {}", manuscriptId);
        }

        List<Video> videoList = new ArrayList<>();
        if (dto.getVideos() != null && !dto.getVideos().isEmpty()) {
            for (int i = 0; i < dto.getVideos().size(); i++) {
                ManuscriptUploadDTO.VideoItemDTO videoItemDTO = dto.getVideos().get(i);
                Video video = createVideoFromDTO(videoItemDTO, manuscriptId, userId, dto.getCategoryId(), i);
                videoList.add(video);
            }
        }

        if (dto.getTags() != null && !dto.getTags().isEmpty()) {
            for (String tagName : dto.getTags()) {
                Tag tag = tagMapper.selectByName(tagName);
                if (tag == null) {
                    tag = new Tag();
                    tag.setName(tagName);
                    tagMapper.insert(tag);
                }
                if (!videoList.isEmpty()) {
                    VideoTag videoTag = new VideoTag();
                    videoTag.setVideoId(videoList.get(0).getId());
                    videoTag.setTagId(tag.getId());
                    videoMapper.insertVideoTag(videoTag);
                }
            }
        }

        int totalDurationSeconds = 0;
        for (Video video : videoList) {
            if (video.getDurationSeconds() != null) {
                totalDurationSeconds += video.getDurationSeconds();
            }
        }
        manuscript.setDurationSeconds(totalDurationSeconds);
        manuscript.setDuration(DurationUtils.formatDuration(totalDurationSeconds));
        manuscriptMapper.updateById(manuscript);
        log.info("稿件总时长已保存，manuscriptId: {}, durationSeconds: {}, duration: {}",
                manuscriptId, totalDurationSeconds, manuscript.getDuration());

        return buildManuscriptVO(manuscript, videoList, dto.getTags());
    }

    private Video createVideoFromDTO(ManuscriptUploadDTO.VideoItemDTO dto, Integer manuscriptId,
                                     Integer userId, Integer categoryId, int order) throws Exception {
        Video video = new Video();
        video.setManuscriptId(manuscriptId);
        video.setVideoOrder(order);
        video.setTitle(dto.getTitle());
        video.setDurationSeconds(0);
        video.setProcessProgress(0);
        video.setProcessStatus(Video.PROCESS_STATUS_PENDING);
        video.setUploadTime(new Date());

        videoMapper.insert(video);
        Integer videoId = video.getId();
        log.info("视频记录创建成功，videoId: {}", videoId);

        MultipartFile videoFile = dto.getVideo();
        if (videoFile != null && !videoFile.isEmpty()) {
            String videoExt = getFileExtension(videoFile.getOriginalFilename());
            String videoKey = StorageKeys.videoSource(manuscriptId, videoId, videoExt);
            String videoUrl = storageService.upload(videoKey, videoFile.getInputStream(), videoFile.getSize(), videoFile.getContentType());
            video.setSourceVideoUrl(videoUrl);
            video.setPlayUrlHd(videoUrl);
            int durationSeconds = dto.getDurationSeconds() != null ? dto.getDurationSeconds() : 0;
            video.setDurationSeconds(durationSeconds);
            log.info("视频保存到MinIO成功: {}", videoKey);
        } else {
            log.warn("视频文件为空，manuscriptId: {}, videoId: {}", manuscriptId, videoId);
        }

        videoMapper.updateById(video);

        return video;
    }

    private String getFileExtension(String filename) {
        if (filename == null || !filename.contains(".")) {
            throw new BusinessException("视频文件缺少扩展名");
        }
        return filename.substring(filename.lastIndexOf("."));
    }

    private ManuscriptVO buildManuscriptVO(Manuscript manuscript, List<Video> videos, List<String> tags) {
        ManuscriptVO vo = new ManuscriptVO();
        BeanUtils.copyProperties(manuscript, vo);

        int totalDurationSeconds = 0;
        if (videos != null && !videos.isEmpty()) {
            for (Video video : videos) {
                if (video.getDurationSeconds() != null) {
                    totalDurationSeconds += video.getDurationSeconds();
                }
            }
        }
        vo.setDurationSeconds(totalDurationSeconds);
        vo.setDuration(DurationUtils.formatDuration(totalDurationSeconds));

        if (manuscript.getCategoryId() != null) {
            try {
                String categoryName = categoryMapper.selectById(manuscript.getCategoryId()).getName();
                vo.setCategoryName(categoryName);
            } catch (Exception e) {
                log.warn("获取分类名称失败: {}", manuscript.getCategoryId());
            }
        }

        if (videos != null && !videos.isEmpty()) {
            List<ManuscriptVO.VideoItemVO> videoVOs = new ArrayList<>();
            for (Video video : videos) {
                ManuscriptVO.VideoItemVO videoVO = new ManuscriptVO.VideoItemVO();
                videoVO.setId(video.getId());
                videoVO.setTitle(video.getTitle());
                videoVO.setDescription(video.getDescription());
                videoVO.setVideoOrder(video.getVideoOrder());
                videoVO.setDuration(DurationUtils.formatDuration(video.getDurationSeconds()));
                videoVO.setDurationSeconds(video.getDurationSeconds());
                videoVO.setPlayUrl(video.getPlayUrlHd());
                videoVO.setPlayUrlHd(video.getPlayUrlHd());
                videoVO.setPlayUrlSd(video.getPlayUrlSd());
                videoVO.setPlayUrlLd(video.getPlayUrlLd());
                videoVO.setSourceVideoUrl(video.getSourceVideoUrl());
                videoVOs.add(videoVO);
            }
            videoVOs.sort((a, b) -> a.getVideoOrder() - b.getVideoOrder());
            vo.setVideos(videoVOs);
        }

        if (vo.getUserId() != null) {
            try {
                Result<UserVO> userResult = userClient.getUserById(vo.getUserId());
                if (userResult != null && userResult.getCode() == 200 && userResult.getData() != null) {
                    UserVO user = userResult.getData();
                    ManuscriptVO.UserInfo uploader = new ManuscriptVO.UserInfo();
                    uploader.setId(user.getId());
                    uploader.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
                    uploader.setAvatar(user.getAvatar());
                    uploader.setLevel(user.getLevel());
                    uploader.setBio(user.getSignature());
                    uploader.setSignature(user.getSignature());
                    uploader.setFollowerCount(user.getFollowerCount());
                    uploader.setFollowingCount(user.getFollowingCount());
                    uploader.setLikedCount(user.getTotalLikeCount());
                    uploader.setFollowing(user.getFollowingCount() != null && user.getFollowingCount() > 0);
                    vo.setUploader(uploader);
                }
            } catch (Exception e) {
            }
        }

        return vo;
    }

    @Override
    public ManuscriptVO getManuscriptById(Integer id) {
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            return null;
        }
        return convertToVO(manuscript);
    }

    @Override
    public ManuscriptVO getManuscriptWithVideos(Integer id, Integer currentUserId) {
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            return null;
        }
        ManuscriptVO vo = convertToVO(manuscript);
        List<Video> videos = videoMapper.selectByManuscriptId(id);
        vo.setVideos(convertVideosToItems(videos));

        if (vo.getUserId() != null) {
            try {
                Result<UserVO> userResult = userClient.getUserById(vo.getUserId());
                if (userResult != null && userResult.getCode() == 200 && userResult.getData() != null) {
                    UserVO user = userResult.getData();
                    ManuscriptVO.UserInfo uploader = new ManuscriptVO.UserInfo();
                    uploader.setId(user.getId());
                    uploader.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
                    uploader.setAvatar(user.getAvatar());
                    uploader.setLevel(user.getLevel());
                    uploader.setBio(user.getSignature());
                    uploader.setSignature(user.getSignature());
                    uploader.setFollowerCount(user.getFollowerCount());
                    uploader.setFollowingCount(user.getFollowingCount());
                    uploader.setLikedCount(user.getTotalLikeCount());

                    // 正确查询关注状态: 如果当前用户已登录且不是自己关注自己
                    if (currentUserId != null && !currentUserId.equals(vo.getUserId())) {
                        Result<Map<String, Object>> followResult = userClient.checkFollowStatus(vo.getUserId(), currentUserId);
                        if (followResult != null && followResult.getCode() == 200 && followResult.getData() != null) {
                            Boolean isFollowing = (Boolean) followResult.getData().get("following");
                            uploader.setFollowing(isFollowing != null ? isFollowing : false);
                        } else {
                            uploader.setFollowing(false);
                        }
                    } else {
                        // 未登录或自己看自己的稿件，不显示关注状态
                        uploader.setFollowing(false);
                    }
                    vo.setUploader(uploader);
                }
            } catch (Exception e) {
                log.error("获取作者信息失败", e);
            }
        }

        try {
            if (vo.getVideos() != null && !vo.getVideos().isEmpty()) {
                Set<String> tags = new TreeSet<>();
                for (ManuscriptVO.VideoItemVO v : vo.getVideos()) {
                    List<Map<String, Object>> tagRows = videoMapper.selectTagsByVideoId(v.getId());
                    if (tagRows != null) {
                        for (Map<String, Object> row : tagRows) {
                            Object nameObj = row.get("name");
                            if (nameObj != null) {
                                tags.add(String.valueOf(nameObj));
                            }
                        }
                    }
                }
                vo.setTags(new ArrayList<>(tags));
            } else {
                vo.setTags(Collections.emptyList());
            }
        } catch (Exception e) {
            log.warn("获取稿件标签失败: {}", e.getMessage());
            vo.setTags(Collections.emptyList());
        }

        return vo;
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserId(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserId(userId);
        return convertToVOs(manuscripts);
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserIdWithPaging(Integer userId, Integer status, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserIdWithPaging(userId, status, offset, size);
        return convertToVOs(manuscripts);
    }

    @Override
    public Integer countManuscriptsByUserIdAndStatus(Integer userId, Integer status) {
        if (status == null) {
            return manuscriptMapper.selectByUserId(userId).size();
        }
        return manuscriptMapper.countByUserIdAndStatus(userId, status);
    }

    @Override
    public Map<String, Integer> getManuscriptStatsByUserId(Integer userId) {
        Map<String, Integer> stats = new HashMap<>();
        stats.put("draft", manuscriptMapper.countPendingByUserId(userId));
        stats.put("processing", manuscriptMapper.countProcessingByUserId(userId));
        stats.put("ready", manuscriptMapper.countReadyByUserId(userId));
        stats.put("published", manuscriptMapper.countPublishedByUserId(userId));
        stats.put("rejected", manuscriptMapper.countRejectedByUserId(userId));
        stats.put("unpublished", manuscriptMapper.countUnpublishedByUserId(userId));
        return stats;
    }

    @Override
    public List<ManuscriptVO> searchUserManuscripts(Integer userId, String keyword, String sort) {
        if (keyword == null || keyword.trim().isEmpty()) {
            return new ArrayList<>();
        }
        
        String searchKeyword = "%" + keyword.trim() + "%";
        List<Manuscript> manuscripts = manuscriptMapper.searchByUserIdAndKeyword(userId, searchKeyword);
        
        if ("play_count".equals(sort)) {
            manuscripts.sort((a, b) -> (b.getViewCount() != null ? b.getViewCount() : 0) - (a.getViewCount() != null ? a.getViewCount() : 0));
        } else if ("favorite_count".equals(sort)) {
            manuscripts.sort((a, b) -> (b.getCollectCount() != null ? b.getCollectCount() : 0) - (a.getCollectCount() != null ? a.getCollectCount() : 0));
        } else {
            manuscripts.sort((a, b) -> {
                if (a.getUploadTime() == null) return 1;
                if (b.getUploadTime() == null) return -1;
                return b.getUploadTime().compareTo(a.getUploadTime());
            });
        }
        
        return convertToVOsWithUploaders(manuscripts);
    }

    @Override
    public List<ManuscriptVO> getRecommendedManuscripts(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectRecommended(0, 60);
        Collections.shuffle(manuscripts);
        if (manuscripts.size() > 20) {
            manuscripts = manuscripts.subList(0, 20);
        }
        return convertToVOsWithUploaders(manuscripts);
    }

    @Override
    public Map<String, Object> getManuscriptsByCategoryId(Integer categoryId, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 20;
        Integer offset = (page - 1) * size;

        List<Manuscript> manuscripts = manuscriptMapper.selectByCategoryId(categoryId, offset, size);
        Integer total = manuscriptMapper.countByCategoryId(categoryId);

        Map<String, Object> result = new HashMap<>();
        result.put("list", convertToVOsWithUploaders(manuscripts));
        result.put("total", total);
        result.put("page", page);
        result.put("size", size);
        result.put("totalPages", (int) Math.ceil((double) total / size));

        return result;
    }

    @Override
    public List<ManuscriptVO> getHotManuscripts(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectHot(0, 20);
        return convertToVOsWithUploaders(manuscripts);
    }

    private ManuscriptVO convertToVO(Manuscript manuscript) {
        if (manuscript == null) return null;
        return convertToVO(manuscript, videoMapper.selectByManuscriptId(manuscript.getId()));
    }

    private ManuscriptVO convertToVO(Manuscript manuscript, List<Video> videos) {
        if (manuscript == null) return null;
        ManuscriptVO vo = new ManuscriptVO();
        BeanUtils.copyProperties(manuscript, vo);
        vo.setVideos(convertVideosToItems(videos == null ? Collections.emptyList() : videos));
        return vo;
    }

    private List<ManuscriptVO> convertToVOsWithUploaders(List<Manuscript> manuscripts) {
        if (manuscripts == null || manuscripts.isEmpty()) {
            return Collections.emptyList();
        }

        List<ManuscriptVO> vos = convertToVOs(manuscripts);

        LinkedHashSet<Integer> userIds = vos.stream()
                .filter(Objects::nonNull)
                .map(ManuscriptVO::getUserId)
                .filter(Objects::nonNull)
                .collect(Collectors.toCollection(LinkedHashSet::new));

        Map<Integer, UserVO> usersById = fetchUsersByIds(userIds);
        if (!usersById.isEmpty()) {
            for (ManuscriptVO vo : vos) {
                if (vo != null && vo.getUserId() != null) {
                    attachUploader(vo, usersById.get(vo.getUserId()));
                }
            }
        }

        return vos;
    }

    private List<ManuscriptVO> convertToVOs(List<Manuscript> manuscripts) {
        if (manuscripts == null || manuscripts.isEmpty()) {
            return Collections.emptyList();
        }

        Map<Integer, List<Video>> videosByManuscriptId = fetchVideosByManuscriptIds(manuscripts);
        return manuscripts.stream()
                .map(manuscript -> convertToVO(
                        manuscript,
                        videosByManuscriptId.getOrDefault(manuscript.getId(), Collections.emptyList())))
                .collect(Collectors.toList());
    }

    private Map<Integer, List<Video>> fetchVideosByManuscriptIds(List<Manuscript> manuscripts) {
        LinkedHashSet<Integer> manuscriptIds = manuscripts.stream()
                .filter(Objects::nonNull)
                .map(Manuscript::getId)
                .filter(Objects::nonNull)
                .collect(Collectors.toCollection(LinkedHashSet::new));
        if (manuscriptIds.isEmpty()) {
            return Collections.emptyMap();
        }

        List<Video> videos = videoMapper.selectByManuscriptIds(new ArrayList<>(manuscriptIds));
        Map<Integer, List<Video>> videosByManuscriptId = new HashMap<>();
        for (Video video : videos) {
            if (video != null && video.getManuscriptId() != null) {
                videosByManuscriptId
                        .computeIfAbsent(video.getManuscriptId(), key -> new ArrayList<>())
                        .add(video);
            }
        }
        return videosByManuscriptId;
    }

    private Map<Integer, UserVO> fetchUsersByIds(Set<Integer> userIds) {
        if (userIds == null || userIds.isEmpty()) {
            return Collections.emptyMap();
        }
        try {
            Result<List<UserVO>> result = userClient.getUsersByIds(new ArrayList<>(userIds));
            if (result == null || result.getCode() == null || result.getCode() != 200 || result.getData() == null) {
                return Collections.emptyMap();
            }
            Map<Integer, UserVO> usersById = new HashMap<>();
            for (UserVO user : result.getData()) {
                if (user != null && user.getId() != null) {
                    usersById.put(user.getId(), user);
                }
            }
            return usersById;
        } catch (Exception e) {
            log.warn("批量获取作者信息失败: {}", e.getMessage());
            return Collections.emptyMap();
        }
    }

    private void attachUploader(ManuscriptVO vo, UserVO user) {
        if (vo == null || user == null) {
            return;
        }
        ManuscriptVO.UserInfo uploader = new ManuscriptVO.UserInfo();
        uploader.setId(user.getId());
        uploader.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
        uploader.setAvatar(user.getAvatar());
        uploader.setLevel(user.getLevel());
        uploader.setBio(user.getSignature());
        uploader.setSignature(user.getSignature());
        uploader.setFollowerCount(user.getFollowerCount());
        uploader.setFollowingCount(user.getFollowingCount());
        uploader.setLikedCount(user.getTotalLikeCount());
        uploader.setFollowing(user.getFollowingCount() != null && user.getFollowingCount() > 0);
        vo.setUploader(uploader);
    }

    @Override
    public ManuscriptVO updateManuscript(Integer id, Manuscript manuscript) {
        Manuscript existing = manuscriptMapper.selectById(id);
        if (existing == null) {
            throw new BusinessException("稿件不存在");
        }
        if (manuscript.getTitle() != null) {
            existing.setTitle(manuscript.getTitle());
        }
        if (manuscript.getDescription() != null) {
            existing.setDescription(manuscript.getDescription());
        }
        if (manuscript.getCoverUrl() != null) {
            existing.setCoverUrl(manuscript.getCoverUrl());
        }
        if (manuscript.getCategoryId() != null) {
            existing.setCategoryId(manuscript.getCategoryId());
        }
        if (manuscript.getDuration() != null) {
            existing.setDuration(manuscript.getDuration());
        }
        if (manuscript.getDurationSeconds() != null) {
            existing.setDurationSeconds(manuscript.getDurationSeconds());
        }
        if (manuscript.getStatus() != null) {
            existing.setStatus(manuscript.getStatus());
        }
        if (manuscript.getReviewStatus() != null) {
            existing.setReviewStatus(manuscript.getReviewStatus());
        }
        if (manuscript.getReviewReason() != null) {
            existing.setReviewReason(manuscript.getReviewReason());
        }
        if (manuscript.getReviewTime() != null) {
            existing.setReviewTime(manuscript.getReviewTime());
        }
        if (manuscript.getReviewerId() != null) {
            existing.setReviewerId(manuscript.getReviewerId());
        }
        manuscriptMapper.updateById(existing);
        return convertToVO(manuscriptMapper.selectById(id));
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptVO updateManuscriptByOwner(Integer id, Integer userId, ManuscriptUpdateDTO dto) throws Exception {
        if (userId == null) {
            throw new BusinessException("用户未登录");
        }
        Manuscript existing = manuscriptMapper.selectById(id);
        if (existing == null) {
            throw new BusinessException("稿件不存在");
        }
        if (!existing.getUserId().equals(userId)) {
            throw new BusinessException("没有权限编辑此稿件");
        }

        Map<String, Object> beforeSnapshot = buildEditSnapshot(existing);
        Integer fromStatus = existing.getStatus();

        existing.setTitle(dto.getTitle().trim());
        existing.setDescription(dto.getDescription());
        existing.setCategoryId(dto.getCategoryId());
        existing.setStatus(Manuscript.STATUS_PENDING_REVIEW);
        existing.setReviewStatus(Manuscript.REVIEW_STATUS_PENDING);
        existing.setReviewReason(null);
        existing.setReviewTime(null);
        existing.setReviewerId(null);

        if (dto.getCover() != null && !dto.getCover().isEmpty()) {
            String coverKey = StorageKeys.manuscriptCover(id);
            String coverUrl = storageService.upload(coverKey, dto.getCover().getInputStream(), dto.getCover().getSize(), dto.getCover().getContentType());
            existing.setCoverUrl(coverUrl);
            log.info("稿件封面更新成功，manuscriptId: {}", id);
        }

        manuscriptMapper.updateById(existing);
        recordManuscriptStatusEvent(existing, fromStatus, Manuscript.STATUS_PENDING_REVIEW,
                "OWNER_EDIT_SUBMITTED", "USER", userId, "稿件编辑后重新进入审核");
        replaceManuscriptTags(id, dto.getTags());
        replaceManuscriptVideos(existing, dto.getVideos());
        Manuscript afterManuscript = manuscriptMapper.selectById(id);
        Map<String, Object> afterSnapshot = buildEditSnapshot(afterManuscript);
        recordEditVersion(id, userId, beforeSnapshot, afterSnapshot);
        return getManuscriptWithVideos(id, userId);
    }

    private void recordEditVersion(Integer manuscriptId, Integer userId, Map<String, Object> beforeSnapshot, Map<String, Object> afterSnapshot) {
        ManuscriptEditVersion version = new ManuscriptEditVersion();
        version.setManuscriptId(manuscriptId);
        version.setUserId(userId);
        version.setBeforeSnapshot(JSON.toJSONString(beforeSnapshot));
        version.setAfterSnapshot(JSON.toJSONString(afterSnapshot));
        version.setChangedFields(JSON.toJSONString(diffEditFields(beforeSnapshot, afterSnapshot)));
        version.setStatus(ManuscriptEditVersion.STATUS_PENDING);
        manuscriptEditVersionMapper.insert(version);
    }

    private Map<String, Object> buildEditSnapshot(Manuscript manuscript) {
        Map<String, Object> snapshot = new HashMap<>();
        if (manuscript == null) {
            return snapshot;
        }
        List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
        snapshot.put("title", manuscript.getTitle());
        snapshot.put("description", manuscript.getDescription());
        snapshot.put("categoryId", manuscript.getCategoryId());
        snapshot.put("categoryName", resolveCategoryName(manuscript.getCategoryId()));
        snapshot.put("coverUrl", manuscript.getCoverUrl());
        snapshot.put("tags", collectTags(videos));
        snapshot.put("videos", videos.stream()
                .map(this::buildVideoEditSnapshot)
                .collect(Collectors.toList()));
        return snapshot;
    }

    private Map<String, Object> buildVideoEditSnapshot(Video video) {
        Map<String, Object> snapshot = new HashMap<>();
        snapshot.put("id", video.getId());
        snapshot.put("title", video.getTitle());
        snapshot.put("videoOrder", video.getVideoOrder());
        snapshot.put("sourceVideoUrl", video.getSourceVideoUrl());
        snapshot.put("durationSeconds", video.getDurationSeconds());
        return snapshot;
    }

    private String resolveCategoryName(Integer categoryId) {
        if (categoryId == null) {
            return null;
        }
        try {
            return categoryMapper.selectById(categoryId).getName();
        } catch (Exception e) {
            return null;
        }
    }

    private List<String> collectTags(List<Video> videos) {
        if (videos == null || videos.isEmpty()) {
            return Collections.emptyList();
        }
        Set<String> tags = new TreeSet<>();
        for (Video video : videos) {
            List<Map<String, Object>> rows = videoMapper.selectTagsByVideoId(video.getId());
            if (rows == null) {
                continue;
            }
            for (Map<String, Object> row : rows) {
                Object name = row.get("name");
                if (name != null) {
                    tags.add(String.valueOf(name));
                }
            }
        }
        return new ArrayList<>(tags);
    }

    @SuppressWarnings("unchecked")
    private List<String> diffEditFields(Map<String, Object> beforeSnapshot, Map<String, Object> afterSnapshot) {
        List<String> fields = new ArrayList<>();
        if (!Objects.equals(beforeSnapshot.get("title"), afterSnapshot.get("title"))) {
            fields.add("title");
        }
        if (!Objects.equals(beforeSnapshot.get("description"), afterSnapshot.get("description"))) {
            fields.add("description");
        }
        if (!Objects.equals(beforeSnapshot.get("categoryId"), afterSnapshot.get("categoryId"))) {
            fields.add("category");
        }
        if (!Objects.equals(beforeSnapshot.get("coverUrl"), afterSnapshot.get("coverUrl"))) {
            fields.add("cover");
        }
        if (!Objects.equals(beforeSnapshot.get("tags"), afterSnapshot.get("tags"))) {
            fields.add("tags");
        }
        if (!Objects.equals(beforeSnapshot.get("videos"), afterSnapshot.get("videos"))) {
            fields.add("videos");
        }
        return fields;
    }

    private void replaceManuscriptTags(Integer manuscriptId, List<String> tags) {
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        if (videos == null || videos.isEmpty()) {
            return;
        }
        List<Integer> videoIds = videos.stream()
                .map(Video::getId)
                .collect(Collectors.toList());
        videoMapper.deleteVideoTagsByVideoIds(videoIds);

        List<String> normalizedTags = normalizeTags(tags);
        if (normalizedTags.isEmpty()) {
            return;
        }
        Integer firstVideoId = videos.get(0).getId();
        for (String tagName : normalizedTags) {
            Tag tag = tagMapper.selectByName(tagName);
            if (tag == null) {
                tag = new Tag();
                tag.setName(tagName);
                tagMapper.insert(tag);
            }
            VideoTag videoTag = new VideoTag();
            videoTag.setVideoId(firstVideoId);
            videoTag.setTagId(tag.getId());
            videoMapper.insertVideoTag(videoTag);
        }
    }

    private List<String> normalizeTags(List<String> tags) {
        if (tags == null || tags.isEmpty()) {
            return Collections.emptyList();
        }
        return tags.stream()
                .filter(tag -> tag != null && !tag.isBlank())
                .map(String::trim)
                .collect(Collectors.collectingAndThen(Collectors.toCollection(LinkedHashSet::new), ArrayList::new));
    }

    private void replaceManuscriptVideos(Manuscript manuscript, List<ManuscriptUpdateDTO.VideoUpdateDTO> updates) throws Exception {
        if (manuscript == null || updates == null || updates.isEmpty()) {
            return;
        }
        List<Video> existingVideos = videoMapper.selectByManuscriptId(manuscript.getId());
        Map<Integer, Video> existingVideoMap = existingVideos.stream()
                .collect(Collectors.toMap(Video::getId, video -> video, (a, b) -> a, java.util.LinkedHashMap::new));
        for (ManuscriptUpdateDTO.VideoUpdateDTO update : updates) {
            if (update == null || update.getId() == null) {
                continue;
            }
            Video video = existingVideoMap.get(update.getId());
            if (video == null) {
                throw new BusinessException("稿件分P不存在");
            }
            if (update.getTitle() != null) {
                video.setTitle(update.getTitle().trim());
            }
            if (update.getVideoOrder() != null) {
                video.setVideoOrder(update.getVideoOrder());
            }
            String oldSourceKey = null;
            String newVideoKey = null;
            if (update.getFile() != null && !update.getFile().isEmpty()) {
                String videoExt = getFileExtension(update.getFile().getOriginalFilename());
                oldSourceKey = extractObjectKey(video.getSourceVideoUrl());
                newVideoKey = StorageKeys.videoSource(manuscript.getId(), video.getId(), videoExt);
                String videoUrl = storageService.upload(newVideoKey, update.getFile().getInputStream(), update.getFile().getSize(), update.getFile().getContentType());
                video.setSourceVideoUrl(videoUrl);
                video.setPlayUrlHd(videoUrl);
                video.setPlayUrlSd(null);
                video.setPlayUrlLd(null);
                video.setDurationSeconds(update.getDurationSeconds() != null ? update.getDurationSeconds() : 0);
                video.setProcessStatus(Video.PROCESS_STATUS_PENDING);
                video.setProcessProgress(0);
                video.setProcessStage("PENDING");
                video.setProcessError(null);
                video.setHasSubtitle(0);
                video.setHasSummary(0);
            }
            videoMapper.updateById(video);
            if (newVideoKey != null) {
                cleanupReplacedVideoObjects(manuscript.getId(), video.getId(), oldSourceKey, newVideoKey);
            }
        }
    }

    private void cleanupReplacedVideoObjects(Integer manuscriptId, Integer videoId, String oldSourceKey, String newSourceKey) {
        try {
            if (oldSourceKey != null && !oldSourceKey.equals(newSourceKey)) {
                storageService.delete(oldSourceKey);
            }
            String videoPrefix = "manuscripts/%d/videos/%d/".formatted(manuscriptId, videoId);
            storageService.deletePrefix(videoPrefix + "transcoded/");
            storageService.deletePrefix(videoPrefix + "audio/");
            storageService.deletePrefix(videoPrefix + "subtitles/");
            storageService.deletePrefix(videoPrefix + "summary/");
            log.info("稿件分P旧处理产物已清理，manuscriptId: {}, videoId: {}", manuscriptId, videoId);
        } catch (Exception e) {
            log.warn("清理稿件分P旧处理产物失败，manuscriptId: {}, videoId: {}", manuscriptId, videoId, e);
        }
    }

    private String extractObjectKey(String url) {
        if (url == null || url.isBlank()) {
            return null;
        }
        String value = url.trim();
        if (!value.contains("://")) {
            return value.startsWith("manuscripts/") ? value : null;
        }
        try {
            String path = URLDecoder.decode(URI.create(value).getPath(), StandardCharsets.UTF_8);
            int index = path.indexOf("/manuscripts/");
            if (index < 0) {
                return null;
            }
            return path.substring(index + 1);
        } catch (Exception e) {
            log.warn("解析对象存储Key失败: {}", url);
            return null;
        }
    }

    @Override
    public void deleteManuscript(Integer id, Integer userId) {
        Manuscript existing = manuscriptMapper.selectById(id);
        if (existing == null) {
            throw new BusinessException("稿件不存在");
        }
        if (!existing.getUserId().equals(userId)) {
            throw new BusinessException("没有权限删除此稿件");
        }
        videoMapper.deleteByManuscriptId(id);
        manuscriptMapper.deleteById(id);
    }

    @Override
    public int fixAllManuscriptDurations() {
        List<Manuscript> manuscripts = manuscriptMapper.selectRecommended(0, 1000);
        int fixedCount = 0;
        
        for (Manuscript manuscript : manuscripts) {
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            
            if (!videos.isEmpty()) {
                int totalDurationSeconds = 0;
                for (Video video : videos) {
                    if (video.getDurationSeconds() != null) {
                        totalDurationSeconds += video.getDurationSeconds();
                    }
                }
                
                String duration = formatDuration(totalDurationSeconds);
                
                System.out.println("Fixing manuscript " + manuscript.getId() + ": " + totalDurationSeconds + " seconds, duration: " + duration);
                
                manuscript.setDurationSeconds(totalDurationSeconds);
                manuscript.setDuration(duration);
                
                int result = manuscriptMapper.updateById(manuscript);
                System.out.println("Update result for manuscript " + manuscript.getId() + ": " + result);
                
                if (result > 0) {
                    fixedCount++;
                }
            }
        }
        
        System.out.println("Fixed " + fixedCount + " manuscripts");
        return fixedCount;
    }
    
    private String formatDuration(int seconds) {
        if (seconds <= 0) return "0:00";
        int mins = seconds / 60;
        int secs = seconds % 60;
        return mins + ":" + (secs < 10 ? "0" : "") + secs;
    }

    private List<ManuscriptVO.VideoItemVO> convertVideosToItems(List<Video> videos) {
        List<ManuscriptVO.VideoItemVO> items = new ArrayList<>();
        for (Video video : videos) {
            ManuscriptVO.VideoItemVO item = new ManuscriptVO.VideoItemVO();
            item.setId(video.getId());
            item.setVideoOrder(video.getVideoOrder());
            item.setTitle(video.getTitle());
            item.setPlayUrl(video.getPlayUrlHd());
            item.setPlayUrlHd(video.getPlayUrlHd());
            item.setPlayUrlSd(video.getPlayUrlSd());
            item.setPlayUrlLd(video.getPlayUrlLd());
            item.setSourceVideoUrl(video.getSourceVideoUrl());
            item.setDurationSeconds(video.getDurationSeconds());
            item.setProcessStatus(video.getProcessStatus());
            items.add(item);
        }
        return items;
    }

    @Override
    public List<ManuscriptVO> getPendingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PENDING_REVIEW);
        return convertToVOs(manuscripts);
    }

    @Override
    public List<ManuscriptVO> getProcessingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PROCESSING);
        return convertToVOs(manuscripts);
    }

    @Override
    public List<ManuscriptVO> getReadyManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_READY_TO_PUBLISH);
        return convertToVOs(manuscripts);
    }

    @Override
    public List<ManuscriptVO> getAllManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectList(null);
        return convertToVOs(manuscripts);
    }

    @Override
    public Map<String, Object> getManuscriptDetail(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return null;
        }
        ManuscriptVO vo = convertToVO(manuscript);
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        vo.setVideos(convertVideosToItems(videos));

        Map<String, Object> result = new HashMap<>();
        result.put("manuscript", vo);
        result.put("videos", videos);
        result.put("editVersion", convertEditVersionVO(manuscriptEditVersionMapper.selectLatestByManuscriptId(manuscriptId)));
        return result;
    }

    private ManuscriptEditVersionVO convertEditVersionVO(ManuscriptEditVersion version) {
        if (version == null) {
            return null;
        }
        ManuscriptEditVersionVO vo = new ManuscriptEditVersionVO();
        vo.setId(version.getId());
        vo.setManuscriptId(version.getManuscriptId());
        vo.setUserId(version.getUserId());
        vo.setStatus(version.getStatus());
        vo.setReviewReason(version.getReviewReason());
        vo.setReviewerId(version.getReviewerId());
        vo.setReviewedAt(version.getReviewedAt());
        vo.setCreatedAt(version.getCreatedAt());
        vo.setChangedFields(parseStringList(version.getChangedFields()));
        vo.setBeforeSnapshot(parseObjectMap(version.getBeforeSnapshot()));
        vo.setAfterSnapshot(parseObjectMap(version.getAfterSnapshot()));
        return vo;
    }

    private Map<String, Object> parseObjectMap(String json) {
        if (json == null || json.isBlank()) {
            return Collections.emptyMap();
        }
        try {
            return JSON.parseObject(json, new TypeReference<Map<String, Object>>() {});
        } catch (Exception e) {
            return Collections.emptyMap();
        }
    }

    private List<String> parseStringList(String json) {
        if (json == null || json.isBlank()) {
            return Collections.emptyList();
        }
        try {
            return JSON.parseArray(json, String.class);
        } catch (Exception e) {
            return Collections.emptyList();
        }
    }

    private void recordManuscriptStatusEvent(Manuscript manuscript,
                                             Integer fromStatus,
                                             Integer toStatus,
                                             String action,
                                             String operatorType,
                                             Integer operatorId,
                                             String reason) {
        if (manuscript == null || manuscript.getId() == null || toStatus == null) {
            throw new BusinessException("稿件状态流水缺少必要字段");
        }
        ManuscriptAnalyticsEvent event = ManuscriptAnalyticsEvent.statusChange(
                manuscript.getId(),
                manuscript.getUserId(),
                fromStatus,
                toStatus,
                action,
                operatorType,
                operatorId,
                reason
        );
        publishAnalyticsEventAfterCommit(event);
    }

    private void publishAnalyticsEventAfterCommit(ManuscriptAnalyticsEvent event) {
        if (TransactionSynchronizationManager.isSynchronizationActive()) {
            TransactionSynchronizationManager.registerSynchronization(new TransactionSynchronization() {
                @Override
                public void afterCommit() {
                    publishAnalyticsEventSafely(event);
                }
            });
            return;
        }
        publishAnalyticsEventSafely(event);
    }

    private void publishAnalyticsEventSafely(ManuscriptAnalyticsEvent event) {
        try {
            videoMQProducer.sendManuscriptAnalyticsEvent(event);
        } catch (Exception e) {
            log.warn("Failed to publish manuscript analytics event: type={}, manuscriptId={}",
                    event == null ? null : event.getEventType(),
                    event == null ? null : event.getManuscriptId(),
                    e);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean approveManuscript(Integer manuscriptId, Integer reviewerId, String reason) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);

        if (manuscriptMapper.updateById(manuscript) > 0) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_PROCESSING,
                    "ADMIN_APPROVE", "ADMIN", reviewerId, reason);
            updateLatestEditVersionStatus(manuscriptId, ManuscriptEditVersion.STATUS_APPROVED, reviewerId, reason);
            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            for (Video video : videos) {
                videoProcessPort.triggerAutoProcess(video.getId(), manuscriptId, manuscript.getUserId());
                log.info("审核通过，提交视频全流程MQ任务，manuscriptId: {}, videoId: {}", manuscriptId, video.getId());
            }
            return true;
        }
        return false;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean approveManuscriptWithProcess(Integer manuscriptId, Integer reviewerId, String reason, boolean autoProcess) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);

        if (manuscriptMapper.updateById(manuscript) <= 0) {
            return false;
        }
        recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_PROCESSING,
                "ADMIN_APPROVE_WITH_PROCESS", "ADMIN", reviewerId, reason);
        updateLatestEditVersionStatus(manuscriptId, ManuscriptEditVersion.STATUS_APPROVED, reviewerId, reason);
        
        if (!autoProcess) {
            return true;
        }

        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);

        for (Video video : videos) {
            try {
                videoProcessPort.triggerAutoProcess(video.getId(), manuscriptId, manuscript.getUserId());
                log.info("审核通过并提交视频全流程MQ任务，manuscriptId: {}, videoId: {}", manuscriptId, video.getId());
            } catch (Exception e) {
                log.error("提交视频全流程MQ任务失败，manuscriptId: {}, videoId: {}", manuscriptId, video.getId(), e);
            }
        }
        return true;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean rejectManuscript(Integer manuscriptId, Integer reviewerId, String reason) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        manuscript.setStatus(Manuscript.STATUS_REJECTED);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_REJECTED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);
        boolean updated = manuscriptMapper.updateById(manuscript) > 0;
        if (updated) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_REJECTED,
                    "ADMIN_REJECT", "ADMIN", reviewerId, reason);
            updateLatestEditVersionStatus(manuscriptId, ManuscriptEditVersion.STATUS_REJECTED, reviewerId, reason);
        }
        return updated;
    }

    private void updateLatestEditVersionStatus(Integer manuscriptId, String status, Integer reviewerId, String reviewReason) {
        ManuscriptEditVersion latest = manuscriptEditVersionMapper.selectLatestByManuscriptId(manuscriptId);
        if (latest != null && ManuscriptEditVersion.STATUS_PENDING.equals(latest.getStatus())) {
            manuscriptEditVersionMapper.updateReviewResult(latest.getId(), status, reviewerId, reviewReason);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean publishManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        manuscript.setStatus(Manuscript.STATUS_PUBLISHED);
        boolean updated = manuscriptMapper.updateById(manuscript) > 0;

        if (updated) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_PUBLISHED,
                    "ADMIN_PUBLISH", "ADMIN", null, null);
            String content = "您的稿件《" + manuscript.getTitle() + "》已通过审核并成功上架啦！";
            userEventMQProducer.sendNotificationEvent(
                    UserNotificationEvent.system(manuscript.getUserId(), "稿件上架通知", content));
        }

        return updated;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean unpublishManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        manuscript.setStatus(Manuscript.STATUS_UNPUBLISHED);
        boolean updated = manuscriptMapper.updateById(manuscript) > 0;
        if (updated) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_UNPUBLISHED,
                    "ADMIN_UNPUBLISH", "ADMIN", null, null);
        }
        return updated;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean publishManuscriptByOwner(Integer manuscriptId, Integer userId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null || !manuscript.getUserId().equals(userId)) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        boolean updated = manuscriptMapper.updateStatusById(manuscriptId, Manuscript.STATUS_PUBLISHED) > 0;
        if (updated) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_PUBLISHED,
                    "OWNER_PUBLISH", "USER", userId, null);
        }
        return updated;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean unpublishManuscriptByOwner(Integer manuscriptId, Integer userId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null || !manuscript.getUserId().equals(userId)) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        int result = manuscriptMapper.updateStatusById(manuscriptId, Manuscript.STATUS_UNPUBLISHED);
        if (result > 0) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_UNPUBLISHED,
                    "OWNER_UNPUBLISH", "USER", userId, null);
        }
        return result > 0;
    }

    @Override
    public List<ManuscriptVO.VideoItemVO> getManuscriptVideos(Integer manuscriptId) {
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        return videos.stream().map(video -> {
            ManuscriptVO.VideoItemVO item = new ManuscriptVO.VideoItemVO();
            item.setId(video.getId());
            item.setVideoOrder(video.getVideoOrder());
            item.setTitle(video.getTitle());
            item.setDescription(video.getDescription());
            item.setPlayUrl(video.getPlayUrlHd());
            item.setPlayUrlHd(video.getPlayUrlHd());
            item.setPlayUrlSd(video.getPlayUrlSd());
            item.setPlayUrlLd(video.getPlayUrlLd());
            item.setSourceVideoUrl(video.getSourceVideoUrl());
            item.setDurationSeconds(video.getDurationSeconds());
            item.setStatus(video.getStatus());
            item.setProcessStatus(video.getProcessStatus());
            return item;
        }).collect(Collectors.toList());
    }

    @Override
    public Map<String, Object> getManuscriptStatistics() {
        Map<String, Object> stats = new HashMap<>();
        stats.put("total", manuscriptMapper.selectList(null).size());
        stats.put("pending", manuscriptMapper.selectByStatus(Manuscript.STATUS_PENDING_REVIEW).size());
        stats.put("processing", manuscriptMapper.selectByStatus(Manuscript.STATUS_PROCESSING).size());
        stats.put("ready", manuscriptMapper.selectByStatus(Manuscript.STATUS_READY_TO_PUBLISH).size());
        stats.put("published", manuscriptMapper.selectByStatus(Manuscript.STATUS_PUBLISHED).size());
        stats.put("rejected", manuscriptMapper.selectByStatus(Manuscript.STATUS_REJECTED).size());
        stats.put("failed", manuscriptMapper.selectByStatus(Manuscript.STATUS_PROCESS_FAILED).size());
        return stats;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean retryManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        Integer fromStatus = manuscript.getStatus();
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        boolean updated = manuscriptMapper.updateById(manuscript) > 0;
        if (updated) {
            recordManuscriptStatusEvent(manuscript, fromStatus, Manuscript.STATUS_PROCESSING,
                    "ADMIN_RETRY", "ADMIN", null, null);
        }
        return updated;
    }

    @Override
    public boolean manualProcessAll(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_TRANSCODING);
        boolean updated = videoMapper.updateById(video) > 0;
        if (updated) {
            videoProcessPort.triggerAutoProcess(videoId, video.getManuscriptId(), null);
        }
        return updated;
    }

    @Override
    public boolean resetVideoStatus(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_PENDING);
        return videoMapper.updateById(video) > 0;
    }

    @Override
    public Map<String, Object> getVideoSourceUrl(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return null;
        }
        Map<String, Object> result = new HashMap<>();
        result.put("videoId", videoId);
        result.put("sourceUrl", video.getSourceVideoUrl());
        result.put("title", video.getTitle());
        result.put("durationSeconds", video.getDurationSeconds());
        return result;
    }

    @Override
    public Video getVideoById(Integer videoId) {
        return videoMapper.selectById(videoId);
    }

    @Override
    public void incrementViewCount(Integer manuscriptId) {
        incrementViewCount(manuscriptId, null);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void incrementViewCount(Integer manuscriptId, String viewerKey) {
        if (manuscriptId == null) return;
        if (viewerKey != null) {
            String redisKey = "view:dedup:" + manuscriptId + ":" + viewerKey;
            Boolean isNew = redisTemplate.opsForValue().setIfAbsent(redisKey, "1", 30, TimeUnit.MINUTES);
            if (!Boolean.TRUE.equals(isNew)) {
                return;
            }
        }
        manuscriptMapper.incrementViewCount(manuscriptId);
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript != null) {
            publishAnalyticsEventAfterCommit(ManuscriptAnalyticsEvent.viewIncrement(manuscriptId, manuscript.getUserId()));
        }
    }

    @Override
    public void updateCommentCount(Integer manuscriptId, Integer count) {
        if (manuscriptId != null && count != null) {
            manuscriptMapper.updateCommentCount(manuscriptId, count);
        }
    }

    @Override
    public void incrementCommentCount(Integer manuscriptId) {
        if (manuscriptId != null) {
            manuscriptMapper.incrementCommentCount(manuscriptId);
            Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
            if (manuscript != null) {
                publishAnalyticsEventAfterCommit(ManuscriptAnalyticsEvent.metricIncrement(
                        manuscriptId,
                        manuscript.getUserId(),
                        ManuscriptAnalyticsEvent.METRIC_COMMENT,
                        1
                ));
            }
        }
    }

    @Override
    public void decrementCommentCount(Integer manuscriptId) {
        if (manuscriptId != null) {
            manuscriptMapper.decrementCommentCount(manuscriptId);
        }
    }
}
