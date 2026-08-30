package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Tag;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.web.mapper.CategoryMapper;
import com.mybilibili.web.mapper.CommentMapper;
import com.mybilibili.web.mapper.ReplyMapper;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.TagMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.FollowService;
import com.mybilibili.web.service.ManuscriptService;
import com.mybilibili.web.service.RandomRecommendService;
import com.mybilibili.web.utils.FFmpegUtils;
import com.mybilibili.web.utils.UploadFilePathUtils;
import lombok.extern.slf4j.Slf4j;
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

@Slf4j
@Service
public class ManuscriptServiceImpl implements ManuscriptService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private CategoryMapper categoryMapper;

    @Autowired
    private TagMapper tagMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private FollowService followService;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Autowired
    private FFmpegUtils ffmpegUtils;

    @Autowired
    private RandomRecommendService randomRecommendService;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptVO uploadManuscript(ManuscriptUploadDTO dto, Integer userId) throws Exception {
        // 1. 创建稿件记录
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

        // 2. 创建稿件目录并保存封面到稿件目录
        uploadFilePathUtils.createManuscriptDirectory(manuscriptId);
        if (dto.getCover() != null && !dto.getCover().isEmpty()) {
            String coverPath = uploadFilePathUtils.getManuscriptCoverPath(manuscriptId);
            dto.getCover().transferTo(new File(coverPath));
            manuscript.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
            manuscriptMapper.update(manuscript);
        }

        // 3. 处理视频列表（支持多视频分P）
        List<Video> videoList = new ArrayList<>();
        if (dto.getVideos() != null && !dto.getVideos().isEmpty()) {
            for (int i = 0; i < dto.getVideos().size(); i++) {
                ManuscriptUploadDTO.VideoItemDTO videoItemDTO = dto.getVideos().get(i);
                Video video = createVideoFromDTO(videoItemDTO, manuscriptId, userId, dto.getCategoryId(), i);
                videoList.add(video);
            }
        }

        // 4. 处理标签
        if (dto.getTags() != null && !dto.getTags().isEmpty()) {
            for (String tagName : dto.getTags()) {
                Tag tag = tagMapper.selectByName(tagName);
                if (tag == null) {
                    tag = new Tag();
                    tag.setName(tagName);
                    tagMapper.insert(tag);
                }
                // 标签关联到稿件的第一个视频（或者可以创建新的关联表）
                if (!videoList.isEmpty()) {
                    VideoTag videoTag = new VideoTag();
                    videoTag.setVideoId(videoList.get(0).getId());
                    videoTag.setTagId(tag.getId());
                    videoMapper.insertVideoTag(videoTag);
                }
            }
        }

        // 5. 计算并保存稿件总时长
        int totalDurationSeconds = 0;
        for (Video video : videoList) {
            if (video.getDurationSeconds() != null) {
                totalDurationSeconds += video.getDurationSeconds();
            }
        }
        manuscript.setDurationSeconds(totalDurationSeconds);
        manuscript.setDuration(DurationUtils.formatDuration(totalDurationSeconds));
        manuscriptMapper.update(manuscript);
        log.info("稿件总时长已保存，manuscriptId: {}, durationSeconds: {}, duration: {}",
                manuscriptId, totalDurationSeconds, manuscript.getDuration());

        // 6. 构建返回VO
        return buildManuscriptVO(manuscript, videoList, dto.getTags());
    }

    /**
     * 根据DTO创建视频记录
     */
    private Video createVideoFromDTO(ManuscriptUploadDTO.VideoItemDTO dto, Integer manuscriptId,
                                     Integer userId, Integer categoryId, int order) throws Exception {
        // 创建视频记录
        Video video = new Video();
        video.setManuscriptId(manuscriptId);
        video.setVideoOrder(order);
        video.setTitle(dto.getTitle());
        video.setDurationSeconds(0);
        video.setStatus(Video.STATUS_PENDING_REVIEW);
        video.setReviewStatus(Video.REVIEW_STATUS_PENDING);
        video.setProcessProgress(0);
        video.setUploadTime(new Date());

        videoMapper.insert(video);
        Integer videoId = video.getId();
        log.info("创建视频记录成功，videoId: {}", videoId);

        // 为每个视频创建目录
        uploadFilePathUtils.createVideoDirectories(manuscriptId, videoId);
        log.info("创建视频目录成功，manuscriptId: {}, videoId: {}", manuscriptId, videoId);

        // 上传视频文件到source目录
        MultipartFile videoFile = dto.getVideo();
        int durationSeconds = 0;
        if (videoFile != null && !videoFile.isEmpty()) {
            String videoExt = getFileExtension(videoFile.getOriginalFilename());
            String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId, videoExt);
            log.info("准备保存视频文件到: {}", sourceVideoPath);
            try {
                videoFile.transferTo(new File(sourceVideoPath));
                log.info("视频文件保存成功: {}", sourceVideoPath);

                // 设置源视频URL
                video.setSourceVideoUrl(uploadFilePathUtils.getVideoSourceUrl(manuscriptId, videoId, videoExt));

                // 提取视频时长
                try {
                    durationSeconds = ffmpegUtils.getVideoDuration(sourceVideoPath);
                    video.setDurationSeconds(durationSeconds);
                    log.info("视频时长: {}秒", durationSeconds);
                } catch (Exception e) {
                    log.warn("提取视频时长失败: {}", sourceVideoPath, e);
                    video.setDurationSeconds(0);
                }
            } catch (Exception e) {
                log.error("视频文件保存失败: {}", sourceVideoPath, e);
                throw e;
            }
        } else {
            log.warn("视频文件为空，manuscriptId: {}, videoId: {}", manuscriptId, videoId);
        }

        // 更新视频时长和源视频URL
        videoMapper.update(video);

        return video;
    }

    /**
     * 获取文件扩展名
     */
    private String getFileExtension(String filename) {
        if (filename == null || !filename.contains(".")) {
            return ".mp4";
        }
        return filename.substring(filename.lastIndexOf("."));
    }

    @Override
    public ManuscriptVO getManuscriptById(Integer id) {
        return getManuscriptById(id, null);
    }

    @Override
    public ManuscriptVO getManuscriptById(Integer id, Integer currentUserId) {
        // 查询稿件
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            return null;
        }

        // 查询关联的视频列表
        List<Video> videos = videoMapper.selectByManuscriptId(id);

        // 查询标签
        List<String> tags = new ArrayList<>();
        if (!videos.isEmpty()) {
            List<Tag> tagList = tagMapper.selectByVideoId(videos.get(0).getId());
            for (Tag tag : tagList) {
                tags.add(tag.getName());
            }
        }

        return buildManuscriptVO(manuscript, videos, tags, currentUserId);
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserId(Integer userId) {
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserId(userId);
        List<ManuscriptVO> result = new ArrayList<>();

        for (Manuscript manuscript : manuscripts) {
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            result.add(buildManuscriptVO(manuscript, videos, null));
        }

        return result;
    }

    @Override
    public List<ManuscriptVO> getManuscriptsByUserIdWithPaging(Integer userId, Integer status, Integer page, Integer size) {
        // 计算偏移量
        int offset = (page - 1) * size;
        
        // 分页查询稿件列表
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserIdWithPaging(userId, status, offset, size);
        List<ManuscriptVO> result = new ArrayList<>();

        for (Manuscript manuscript : manuscripts) {
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            result.add(buildManuscriptVO(manuscript, videos, null));
        }

        return result;
    }

    @Override
    public Integer countManuscriptsByUserIdAndStatus(Integer userId, Integer status) {
        return manuscriptMapper.countByUserIdAndStatus(userId, status);
    }

    @Override
    public Map<String, Integer> getManuscriptStatsByUserId(Integer userId) {
        // 初始化所有状态的计数
        Map<String, Integer> stats = new HashMap<>();
        stats.put("total", 0);
        stats.put("pendingReview", 0);    // 待审核
        stats.put("processing", 0);        // 处理中
        stats.put("readyToPublish", 0);    // 待上架
        stats.put("published", 0);         // 已上架
        stats.put("rejected", 0);          // 审核拒绝
        stats.put("processFailed", 0);     // 处理失败
        stats.put("unpublished", 0);       // 已下架

        // 查询各状态稿件数量
        List<Map<String, Object>> statusCounts = manuscriptMapper.countByUserIdGroupByStatus(userId);
        
        int total = 0;
        for (Map<String, Object> item : statusCounts) {
            Integer status = (Integer) item.get("status");
            Long count = (Long) item.get("count");
            int countInt = count.intValue();
            total += countInt;

            // 根据状态映射到对应的字段
            if (status != null) {
                switch (status) {
                    case Manuscript.STATUS_PENDING_REVIEW:
                        stats.put("pendingReview", countInt);
                        break;
                    case Manuscript.STATUS_PROCESSING:
                        stats.put("processing", countInt);
                        break;
                    case Manuscript.STATUS_READY_TO_PUBLISH:
                        stats.put("readyToPublish", countInt);
                        break;
                    case Manuscript.STATUS_PUBLISHED:
                        stats.put("published", countInt);
                        break;
                    case Manuscript.STATUS_REJECTED:
                        stats.put("rejected", countInt);
                        break;
                    case Manuscript.STATUS_PROCESS_FAILED:
                        stats.put("processFailed", countInt);
                        break;
                    case Manuscript.STATUS_UNPUBLISHED:
                        stats.put("unpublished", countInt);
                        break;
                }
            }
        }
        
        stats.put("total", total);
        return stats;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptVO updateManuscript(Integer id, Manuscript manuscript) throws Exception {
        // 检查稿件是否存在
        Manuscript existingManuscript = manuscriptMapper.selectById(id);
        if (existingManuscript == null) {
            throw new RuntimeException("稿件不存在");
        }

        // 检查是否是稿件的上传者
        if (!existingManuscript.getUserId().equals(manuscript.getUserId())) {
            throw new RuntimeException("没有权限修改此稿件");
        }

        // 更新稿件信息
        manuscript.setId(id);
        manuscript.setUpdatedAt(new Date());
        manuscriptMapper.update(manuscript);

        // 重新查询更新后的稿件
        Manuscript updatedManuscript = manuscriptMapper.selectById(id);
        List<Video> videos = videoMapper.selectByManuscriptId(id);

        return buildManuscriptVO(updatedManuscript, videos, null);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteManuscript(Integer id, Integer userId) throws Exception {
        // 检查稿件是否存在
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            throw new RuntimeException("稿件不存在");
        }

        // 检查是否是稿件的上传者
        if (!manuscript.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此稿件");
        }

        // 删除关联的视频
        List<Video> videos = videoMapper.selectByManuscriptId(id);
        for (Video video : videos) {
            // 删除视频标签关联
            videoMapper.deleteVideoTags(video.getId());
            // 删除视频记录
            videoMapper.delete(video.getId());
        }

        // 删除稿件目录（包含所有视频文件）
        uploadFilePathUtils.deleteManuscriptDirectory(id);

        // 删除稿件
        manuscriptMapper.delete(id);
    }

    @Override
    public boolean updateManuscriptStatus(Integer id, Integer status) {
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            return false;
        }

        manuscript.setStatus(status);
        manuscript.setUpdatedAt(new Date());
        int result = manuscriptMapper.update(manuscript);
        return result > 0;
    }

    @Override
    public ManuscriptVO getManuscriptWithVideos(Integer id) {
        Manuscript manuscript = manuscriptMapper.selectById(id);
        if (manuscript == null) {
            return null;
        }

        // 获取稿件下的所有视频
        List<Video> videos = videoMapper.selectByManuscriptId(id);

        // 构建ManuscriptVO
        ManuscriptVO vo = buildManuscriptVO(manuscript, videos, null);

        // 添加视频列表到VO
        if (videos != null && !videos.isEmpty()) {
            List<ManuscriptVO.VideoItemVO> videoItems = new ArrayList<>();
            for (Video video : videos) {
                ManuscriptVO.VideoItemVO item = new ManuscriptVO.VideoItemVO();
                item.setId(video.getId());
                item.setTitle(video.getTitle());
                item.setVideoOrder(video.getVideoOrder());
                item.setDuration(DurationUtils.formatDuration(video.getDurationSeconds()));
                item.setPlayUrl(video.getPlayUrl());
                videoItems.add(item);
            }
            // 按videoOrder排序
            videoItems.sort((a, b) -> a.getVideoOrder() - b.getVideoOrder());
            vo.setVideos(videoItems);
        }

        return vo;
    }

    @Override
    public List<ManuscriptVO> getRecommendedManuscripts(Integer userId) {
        // 使用加权随机推荐算法获取推荐稿件
        List<Manuscript> manuscripts = randomRecommendService.getRandomRecommendedManuscripts(userId, 50);

        List<ManuscriptVO> result = new ArrayList<>();
        for (Manuscript manuscript : manuscripts) {
            // 获取稿件的第一个视频用于获取播放地址
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            Video firstVideo = videos.isEmpty() ? null : videos.get(0);

            ManuscriptVO vo = buildManuscriptVO(manuscript, videos, null);

            // 如果有视频，设置第一个视频的ID和播放地址
            if (firstVideo != null) {
                vo.setFirstVideoId(firstVideo.getId());
                vo.setFirstVideoPlayUrl(firstVideo.getPlayUrl());
            }

            result.add(vo);
        }

        return result;
    }

    /**
     * 构建ManuscriptVO（不带当前用户ID）
     */
    private ManuscriptVO buildManuscriptVO(Manuscript manuscript, List<Video> videos, List<String> tags) {
        return buildManuscriptVO(manuscript, videos, tags, null);
    }

    /**
     * 构建ManuscriptVO
     */
    private ManuscriptVO buildManuscriptVO(Manuscript manuscript, List<Video> videos,
                                           List<String> tags, Integer currentUserId) {
        ManuscriptVO vo = new ManuscriptVO();
        vo.setId(manuscript.getId());
        vo.setTitle(manuscript.getTitle());
        vo.setDescription(manuscript.getDescription());
        vo.setCoverUrl(manuscript.getCoverUrl());
        vo.setUserId(manuscript.getUserId());
        vo.setCategoryId(manuscript.getCategoryId());
        vo.setViewCount(manuscript.getViewCount());
        vo.setLikeCount(manuscript.getLikeCount());
        vo.setCoinCount(manuscript.getCoinCount());
        vo.setCollectCount(manuscript.getCollectCount());
        vo.setShareCount(manuscript.getShareCount());
        // 统计评论数+回复数总量
        int commentCount = commentMapper.countByManuscriptId(manuscript.getId());
        int replyCount = replyMapper.countByManuscriptId(manuscript.getId());
        vo.setCommentCount(commentCount + replyCount);
        vo.setDanmakuCount(manuscript.getDanmakuCount());
        vo.setStatus(manuscript.getStatus());
        vo.setReviewStatus(manuscript.getReviewStatus());
        vo.setReviewReason(manuscript.getReviewReason());
        vo.setReviewTime(manuscript.getReviewTime());
        vo.setReviewerId(manuscript.getReviewerId());
        vo.setUploadTime(manuscript.getUploadTime());
        vo.setUpdatedAt(manuscript.getUpdatedAt());

        // 计算稿件总时长
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

        // 设置分类名称
        if (manuscript.getCategoryId() != null) {
            try {
                String categoryName = categoryMapper.selectById(manuscript.getCategoryId()).getName();
                vo.setCategoryName(categoryName);
            } catch (Exception e) {
                log.warn("获取分类名称失败: {}", manuscript.getCategoryId());
            }
        }

        // 设置视频列表
        if (videos != null && !videos.isEmpty()) {
            List<ManuscriptVO.VideoItemVO> videoVOs = new ArrayList<>();
            for (Video video : videos) {
                ManuscriptVO.VideoItemVO videoVO = new ManuscriptVO.VideoItemVO();
                videoVO.setId(video.getId());
                videoVO.setTitle(video.getTitle());
                videoVO.setDescription(video.getDescription());
                videoVO.setPlayUrl(video.getPlayUrl());
                videoVO.setPlayUrlHd(video.getPlayUrlHd());
                videoVO.setPlayUrlSd(video.getPlayUrlSd());
                videoVO.setPlayUrlLd(video.getPlayUrlLd());
                videoVO.setDuration(DurationUtils.formatDuration(video.getDurationSeconds()));
                videoVO.setVideoOrder(video.getVideoOrder());
                videoVO.setStatus(video.getStatus());
                videoVOs.add(videoVO);
            }
            vo.setVideos(videoVOs);
        }

        // 设置标签
        vo.setTags(tags);

        // 设置上传者信息
        ManuscriptVO.UserInfo userInfo = new ManuscriptVO.UserInfo();
        com.mybilibili.common.entity.User user = userMapper.findById(manuscript.getUserId());
        if (user != null) {
            userInfo.setId(user.getId());
            userInfo.setName(user.getNickname());
            userInfo.setAvatar(user.getAvatar());
            userInfo.setLevel(user.getLevel());
            userInfo.setBio(user.getBio());
            userInfo.setSignature(user.getSignature());

            // 实时计算粉丝数
            int followerCount = followService.getFollowerCount(user.getId());
            userInfo.setFollowerCount(followerCount);

            // 检查当前用户是否关注了该用户
            boolean isFollowing = false;
            if (currentUserId != null && !currentUserId.equals(user.getId())) {
                isFollowing = followService.isFollowing(currentUserId, user.getId());
            }
            userInfo.setFollowing(isFollowing);
        }
        vo.setUploader(userInfo);

        return vo;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptVO recalculateDuration(Integer manuscriptId) {
        // 1. 查询稿件
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            throw new RuntimeException("稿件不存在");
        }

        // 2. 查询稿件下的所有视频
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);

        // 3. 计算总时长
        int totalDurationSeconds = 0;
        for (Video video : videos) {
            if (video.getDurationSeconds() != null) {
                totalDurationSeconds += video.getDurationSeconds();
            }
        }

        // 4. 更新稿件时长
        manuscript.setDurationSeconds(totalDurationSeconds);
        manuscript.setDuration(DurationUtils.formatDuration(totalDurationSeconds));
        manuscriptMapper.update(manuscript);

        log.info("稿件时长已重新计算，manuscriptId: {}, durationSeconds: {}, duration: {}",
                manuscriptId, totalDurationSeconds, manuscript.getDuration());

        // 5. 构建并返回VO
        return buildManuscriptVO(manuscript, videos, null);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public int fixAllManuscriptDurations() {
        // 1. 查询所有稿件
        List<Manuscript> manuscripts = manuscriptMapper.selectAll();
        int fixedCount = 0;

        for (Manuscript manuscript : manuscripts) {
            try {
                // 查询稿件下的所有视频
                List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());

                // 计算总时长
                int totalDurationSeconds = 0;
                for (Video video : videos) {
                    if (video.getDurationSeconds() != null) {
                        totalDurationSeconds += video.getDurationSeconds();
                    }
                }

                // 更新稿件时长
                manuscript.setDurationSeconds(totalDurationSeconds);
                manuscript.setDuration(DurationUtils.formatDuration(totalDurationSeconds));
                manuscriptMapper.update(manuscript);

                fixedCount++;
                log.info("稿件时长已修复，manuscriptId: {}, durationSeconds: {}",
                        manuscript.getId(), totalDurationSeconds);
            } catch (Exception e) {
                log.error("修复稿件时长失败，manuscriptId: {}", manuscript.getId(), e);
            }
        }

        log.info("共修复 {} 个稿件的时长", fixedCount);
        return fixedCount;
    }
}
