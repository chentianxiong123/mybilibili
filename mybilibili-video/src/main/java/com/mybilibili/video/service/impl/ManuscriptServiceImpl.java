package com.mybilibili.video.service.impl;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Tag;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.utils.UploadFilePathUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessMessage;
import com.mybilibili.video.feign.MessageClient;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.feign.VideoPipelineClient;
import com.mybilibili.video.mapper.CategoryMapper;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.mapper.TagMapper;
import com.mybilibili.video.mapper.VideoMapper;
import com.mybilibili.video.service.ManuscriptService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Slf4j
@Service
public class ManuscriptServiceImpl implements ManuscriptService {

    private static final int MANUSCRIPT_EXPERIENCE = 100;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private CategoryMapper categoryMapper;

    @Autowired
    private TagMapper tagMapper;

    @Autowired
    private UserClient userClient;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Autowired
    private MessageClient messageClient;

    @Autowired
    private VideoPipelineClient videoPipelineClient;

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
        log.info("稿件记录创建成功，manuscriptId: {}", manuscriptId);

        try {
            userClient.addExperience(userId, MANUSCRIPT_EXPERIENCE);
        } catch (Exception e) {
            log.warn("添加经验值失败: {}", e.getMessage());
        }

        uploadFilePathUtils.createManuscriptDirectory(manuscriptId);
        if (dto.getCover() != null && !dto.getCover().isEmpty()) {
            String coverPath = uploadFilePathUtils.getManuscriptCoverPath(manuscriptId);
            dto.getCover().transferTo(new File(coverPath));
            manuscript.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
            manuscriptMapper.updateById(manuscript);
            log.info("封面保存成功，manuscriptId: {}", manuscriptId);
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

        uploadFilePathUtils.createVideoDirectories(manuscriptId, videoId);
        log.info("视频目录创建成功，manuscriptId: {}, videoId: {}", manuscriptId, videoId);

        MultipartFile videoFile = dto.getVideo();
        int durationSeconds = 0;
        if (videoFile != null && !videoFile.isEmpty()) {
            String videoExt = getFileExtension(videoFile.getOriginalFilename());
            String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId, videoExt);
            log.info("准备保存视频文件到: {}", sourceVideoPath);
            try {
                videoFile.transferTo(new File(sourceVideoPath));
                log.info("视频文件保存成功: {}", sourceVideoPath);

                video.setSourceVideoUrl(uploadFilePathUtils.getVideoSourceUrl(manuscriptId, videoId, videoExt));

                durationSeconds = dto.getDurationSeconds() != null ? dto.getDurationSeconds() : 0;
                video.setDurationSeconds(durationSeconds);
                log.info("视频时长已设置: {} 秒", durationSeconds);
            } catch (Exception e) {
                log.error("视频文件保存失败: {}", sourceVideoPath, e);
                throw e;
            }
        } else {
            log.warn("视频文件为空，manuscriptId: {}, videoId: {}", manuscriptId, videoId);
        }

        videoMapper.updateById(video);

        return video;
    }

    private String getFileExtension(String filename) {
        if (filename == null || !filename.contains(".")) {
            return ".mp4";
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
                videoVO.setPlayUrl(video.getPlayUrlHd() != null ? video.getPlayUrlHd() : video.getPlayUrlSd());
                videoVO.setPlayUrlHd(video.getPlayUrlHd());
                videoVO.setPlayUrlSd(video.getPlayUrlSd());
                videoVO.setPlayUrlLd(video.getPlayUrlLd());
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

        return vo;
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserId(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserId(userId);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserIdWithPaging(Integer userId, Integer status, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserIdWithPaging(userId, status, offset, size);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
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
        
        return manuscripts.stream().map(this::convertToVOWithUploader).collect(Collectors.toList());
    }

    @Override
    public List<ManuscriptVO> getRecommendedManuscripts(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectRecommended(0, 20);
        return manuscripts.stream().map(this::convertToVOWithUploader).collect(Collectors.toList());
    }

    @Override
    public Map<String, Object> getManuscriptsByCategoryId(Integer categoryId, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 20;
        Integer offset = (page - 1) * size;

        List<Manuscript> manuscripts = manuscriptMapper.selectByCategoryId(categoryId, offset, size);
        Integer total = manuscriptMapper.countByCategoryId(categoryId);

        Map<String, Object> result = new HashMap<>();
        result.put("list", manuscripts.stream().map(this::convertToVOWithUploader).collect(Collectors.toList()));
        result.put("total", total);
        result.put("page", page);
        result.put("size", size);
        result.put("totalPages", (int) Math.ceil((double) total / size));

        return result;
    }

    @Override
    public List<ManuscriptVO> getHotManuscripts(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectHot(0, 20);
        return manuscripts.stream().map(this::convertToVOWithUploader).collect(Collectors.toList());
    }

    private ManuscriptVO convertToVO(Manuscript manuscript) {
        if (manuscript == null) return null;
        ManuscriptVO vo = new ManuscriptVO();
        BeanUtils.copyProperties(manuscript, vo);
        List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
        vo.setVideos(convertVideosToItems(videos));
        return vo;
    }

    private ManuscriptVO convertToVOWithUploader(Manuscript manuscript) {
        ManuscriptVO vo = convertToVO(manuscript);
        if (vo != null && vo.getUserId() != null) {
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
                    uploader.setFollowing(user.getFollowingCount() != null && user.getFollowingCount() > 0);
                    vo.setUploader(uploader);
                }
            } catch (Exception e) {
            }
        }
        return vo;
    }

    @Override
    public ManuscriptVO updateManuscript(Integer id, Manuscript manuscript) {
        Manuscript existing = manuscriptMapper.selectById(id);
        if (existing == null) {
            throw new BusinessException("稿件不存在");
        }
        manuscriptMapper.updateById(manuscript);
        return convertToVO(manuscriptMapper.selectById(id));
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
            item.setPlayUrl(video.getPlayUrlHd() != null ? video.getPlayUrlHd() : video.getPlayUrlSd());
            item.setPlayUrlHd(video.getPlayUrlHd());
            item.setPlayUrlSd(video.getPlayUrlSd());
            item.setPlayUrlLd(video.getPlayUrlLd());
            item.setDurationSeconds(video.getDurationSeconds());
            item.setProcessStatus(video.getProcessStatus());
            items.add(item);
        }
        return items;
    }

    @Override
    public List<ManuscriptVO> getPendingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PENDING_REVIEW);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
    }

    @Override
    public List<ManuscriptVO> getProcessingManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_PROCESSING);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
    }

    @Override
    public List<ManuscriptVO> getReadyManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectByStatus(Manuscript.STATUS_READY_TO_PUBLISH);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
    }

    @Override
    public List<ManuscriptVO> getAllManuscripts() {
        List<Manuscript> manuscripts = manuscriptMapper.selectList(null);
        return manuscripts.stream().map(this::convertToVO).collect(Collectors.toList());
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
        return result;
    }

    @Override
    public boolean approveManuscript(Integer manuscriptId, Integer reviewerId, String reason) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);

        if (manuscriptMapper.updateById(manuscript) > 0) {
            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            for (Video video : videos) {
                VideoProcessMessage message = VideoProcessMessage.of(
                    manuscriptId,
                    video.getId(),
                    manuscript.getUserId(),
                    VideoProcessMessage.PROCESS_TYPE_TRANSCODE
                );
                videoMQProducer.sendVideoProcessMessage(message);
                log.info("审核通过，发送视频转码消息，manuscriptId: {}, videoId: {}", manuscriptId, video.getId());
            }
            return true;
        }
        return false;
    }

    @Override
    public boolean approveManuscriptWithProcess(Integer manuscriptId, Integer reviewerId, String reason, boolean autoProcess) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);

        if (manuscriptMapper.updateById(manuscript) <= 0) {
            return false;
        }
        
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        
        for (Video video : videos) {
            try {
                videoPipelineClient.submitPipelineTask(manuscriptId, video.getId(), manuscript.getUserId());
                log.info("审核通过并提交全流程任务，manuscriptId: {}, videoId: {}", manuscriptId, video.getId());
            } catch (Exception e) {
                log.error("提交全流程任务失败，manuscriptId: {}, videoId: {}", manuscriptId, video.getId(), e);
            }
        }
        return true;
    }

    @Override
    public boolean rejectManuscript(Integer manuscriptId, Integer reviewerId, String reason) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        manuscript.setStatus(Manuscript.STATUS_REJECTED);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_REJECTED);
        manuscript.setReviewTime(new Date());
        manuscript.setReviewerId(reviewerId);
        manuscript.setReviewReason(reason);
        return manuscriptMapper.updateById(manuscript) > 0;
    }

    @Override
    public boolean publishManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        manuscript.setStatus(Manuscript.STATUS_PUBLISHED);
        boolean updated = manuscriptMapper.updateById(manuscript) > 0;

        if (updated) {
            try {
                String content = "您的稿件《" + manuscript.getTitle() + "》已通过审核并成功上架啦！";
                messageClient.sendSystemNotification(manuscript.getUserId(), "稿件上架通知", content);
            } catch (Exception e) {
                log.error("发送稿件上架通知失败", e);
            }
        }

        return updated;
    }

    @Override
    public boolean unpublishManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        manuscript.setStatus(Manuscript.STATUS_UNPUBLISHED);
        return manuscriptMapper.updateById(manuscript) > 0;
    }

    @Override
    public boolean publishManuscriptByOwner(Integer manuscriptId, Integer userId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null || !manuscript.getUserId().equals(userId)) {
            return false;
        }
        return manuscriptMapper.updateStatusById(manuscriptId, Manuscript.STATUS_PUBLISHED) > 0;
    }

    @Override
    public boolean unpublishManuscriptByOwner(Integer manuscriptId, Integer userId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        System.out.println("=== 下架稿件调试 ===");
        System.out.println("manuscriptId: " + manuscriptId);
        System.out.println("userId: " + userId);
        System.out.println("manuscript: " + manuscript);
        if (manuscript != null) {
            System.out.println("manuscript.userId: " + manuscript.getUserId());
            System.out.println("当前status: " + manuscript.getStatus());
        }
        if (manuscript == null || !manuscript.getUserId().equals(userId)) {
            System.out.println("稿件不存在或无权操作");
            return false;
        }
        int result = manuscriptMapper.updateStatusById(manuscriptId, Manuscript.STATUS_UNPUBLISHED);
        System.out.println("更新结果: " + result);
        System.out.println("=====================");
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
            item.setPlayUrl(video.getPlayUrlHd() != null ? video.getPlayUrlHd() : video.getPlayUrlSd());
            item.setPlayUrlHd(video.getPlayUrlHd());
            item.setPlayUrlSd(video.getPlayUrlSd());
            item.setPlayUrlLd(video.getPlayUrlLd());
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
    public boolean retryManuscript(Integer manuscriptId) {
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            return false;
        }
        manuscript.setStatus(Manuscript.STATUS_PROCESSING);
        return manuscriptMapper.updateById(manuscript) > 0;
    }

    @Override
    public boolean manualTranscode(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_TRANSCODING);
        return videoMapper.updateById(video) > 0;
    }

    @Override
    public boolean manualExtractAudio(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_AUDIO_EXTRACTING);
        return videoMapper.updateById(video) > 0;
    }

    @Override
    public boolean manualGenerateSubtitle(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_SUBTITLE_GENERATING);
        return videoMapper.updateById(video) > 0;
    }

    @Override
    public boolean manualAiSummary(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_AI_SUMMARIZING);
        return videoMapper.updateById(video) > 0;
    }

    @Override
    public boolean manualProcessAll(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return false;
        }
        video.setProcessStatus(Video.PROCESS_STATUS_TRANSCODING);
        return videoMapper.updateById(video) > 0;
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
    public Map<String, Object> getVideoProcessStatus(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return null;
        }
        Map<String, Object> result = new HashMap<>();
        result.put("videoId", videoId);
        result.put("processStatus", video.getProcessStatus());
        result.put("processError", video.getProcessError());
        result.put("hasSubtitle", video.getHasSubtitle());
        result.put("hasSummary", video.getHasSummary());
        result.put("playUrl", video.getPlayUrlHd());
        result.put("sourceVideoUrl", video.getSourceVideoUrl());
        return result;
    }

    @Override
    public Map<String, Object> getVideoSourceUrl(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return null;
        }
        String sourceUrl = "/uploads/manuscripts/" + video.getManuscriptId() + "/videos/" + videoId + "/source/video.mp4";
        Map<String, Object> result = new HashMap<>();
        result.put("videoId", videoId);
        result.put("sourceUrl", sourceUrl);
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
        if (manuscriptId != null) {
            manuscriptMapper.incrementViewCount(manuscriptId);
        }
    }
}