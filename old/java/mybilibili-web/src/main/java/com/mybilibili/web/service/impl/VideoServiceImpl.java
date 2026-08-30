package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.VideoUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Tag;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.mapper.CategoryMapper;
import com.mybilibili.web.mapper.CommentMapper;
import com.mybilibili.web.mapper.ReplyMapper;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.TagMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.FollowService;
import com.mybilibili.web.service.VideoService;
import com.mybilibili.web.utils.FFmpegUtils;
import com.mybilibili.web.utils.UploadFilePathUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

@Service
public class VideoServiceImpl implements VideoService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private CategoryMapper categoryMapper;

    @Autowired
    private TagMapper tagMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Autowired
    private FFmpegUtils ffmpegUtils;

    @Autowired
    private FollowService followService;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Override
    public VideoVO uploadVideo(VideoUploadDTO videoUploadDTO, Integer userId) throws Exception {
        // 1. 先创建稿件记录
        Manuscript manuscript = new Manuscript();
        manuscript.setTitle(videoUploadDTO.getTitle());
        manuscript.setDescription(videoUploadDTO.getDescription());
        manuscript.setUserId(userId);
        manuscript.setCategoryId(videoUploadDTO.getCategoryId());
        manuscript.setStatus(Manuscript.STATUS_PENDING_REVIEW);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_PENDING);
        manuscript.setProcessProgress(0);
        manuscript.setUploadTime(new Date());
        manuscriptMapper.insert(manuscript);

        Integer manuscriptId = manuscript.getId();

        // 2. 创建视频记录，关联稿件ID
        Video tempVideo = new Video();
        tempVideo.setManuscriptId(manuscriptId);
        tempVideo.setVideoOrder(1); // 单视频，排序为1
        tempVideo.setTitle(videoUploadDTO.getTitle());
        tempVideo.setDescription(videoUploadDTO.getDescription());
        tempVideo.setUserId(userId);
        tempVideo.setCategoryId(videoUploadDTO.getCategoryId());
        tempVideo.setDurationSeconds(0);
        tempVideo.setStatus(Video.STATUS_PENDING_REVIEW);
        tempVideo.setReviewStatus(Video.REVIEW_STATUS_PENDING);
        tempVideo.setProcessProgress(0);
        tempVideo.setUploadTime(new Date());
        videoMapper.insert(tempVideo);

        Integer videoId = tempVideo.getId();

        // 创建稿件目录
        uploadFilePathUtils.createManuscriptDirectory(manuscriptId);

        // 创建视频目录
        uploadFilePathUtils.createVideoDirectories(manuscriptId, videoId);

        // 上传原始视频到 source 目录
        MultipartFile videoFile = videoUploadDTO.getVideo();
        String videoExt = videoFile.getOriginalFilename().substring(videoFile.getOriginalFilename().lastIndexOf("."));
        String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId, videoExt);
        videoFile.transferTo(new File(sourceVideoPath));

        // 上传封面到稿件目录
        MultipartFile coverFile = videoUploadDTO.getCover();
        String coverPath = uploadFilePathUtils.getManuscriptCoverPath(manuscriptId);
        coverFile.transferTo(new File(coverPath));

        // 更新视频信息 - 使用转码后的路径（审核通过后会有）
        Video video = new Video();
        video.setId(videoId);
        video.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
        videoMapper.update(video);

        // 更新稿件封面
        manuscript.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
        manuscriptMapper.update(manuscript);

        if (videoUploadDTO.getTags() != null && !videoUploadDTO.getTags().isEmpty()) {
            for (String tagName : videoUploadDTO.getTags()) {
                Tag tag = tagMapper.selectByName(tagName);
                if (tag == null) {
                    tag = new Tag();
                    tag.setName(tagName);
                    tagMapper.insert(tag);
                }
                VideoTag videoTag = new VideoTag();
                videoTag.setVideoId(videoId);
                videoTag.setTagId(tag.getId());
                videoMapper.insertVideoTag(videoTag);
            }
        }

        // 构建返回VO，包含稿件信息
        VideoVO videoVO = new VideoVO();
        videoVO.setId(videoId);
        videoVO.setManuscriptId(manuscriptId);
        videoVO.setTitle(tempVideo.getTitle());
        videoVO.setDescription(tempVideo.getDescription());
        videoVO.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
        videoVO.setStatus(Video.STATUS_PENDING_REVIEW);

        // 设置稿件信息
        VideoVO.ManuscriptInfo manuscriptInfo = new VideoVO.ManuscriptInfo();
        manuscriptInfo.setId(manuscriptId);
        manuscriptInfo.setTitle(manuscript.getTitle());
        manuscriptInfo.setDescription(manuscript.getDescription());
        manuscriptInfo.setCoverUrl(manuscript.getCoverUrl());
        manuscriptInfo.setStatus(manuscript.getStatus());
        manuscriptInfo.setReviewStatus(manuscript.getReviewStatus());
        manuscriptInfo.setUploadTime(manuscript.getUploadTime());
        videoVO.setManuscript(manuscriptInfo);

        return videoVO;
    }

    @Override
    public VideoVO getVideoById(Integer id) {
        return getVideoById(id, null);
    }

    @Override
    public VideoVO getVideoById(Integer id, Integer currentUserId) {
        return getVideoById(id, currentUserId, false);
    }

    @Override
    public VideoVO getVideoById(Integer id, Integer currentUserId, boolean includeManuscript) {
        Video video = videoMapper.selectById(id);
        if (video == null) {
            return null;
        }
        VideoVO videoVO = buildVideoVO(video, currentUserId);

        // 如果需要包含稿件信息
        if (includeManuscript && video.getManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
            if (manuscript != null) {
                videoVO.setManuscriptId(manuscript.getId());
                VideoVO.ManuscriptInfo manuscriptInfo = new VideoVO.ManuscriptInfo();
                manuscriptInfo.setId(manuscript.getId());
                manuscriptInfo.setTitle(manuscript.getTitle());
                manuscriptInfo.setDescription(manuscript.getDescription());
                manuscriptInfo.setCoverUrl(manuscript.getCoverUrl());
                manuscriptInfo.setStatus(manuscript.getStatus());
                manuscriptInfo.setReviewStatus(manuscript.getReviewStatus());
                manuscriptInfo.setUploadTime(manuscript.getUploadTime());
                videoVO.setManuscript(manuscriptInfo);
            }
        }

        return videoVO;
    }

    @Override
    public List<VideoVO> getVideosByUserId(Integer userId) {
        // 改为查询用户的稿件，并转换为VideoVO格式
        return getManuscriptsByUserId(userId, "latest");
    }

    @Override
    public List<VideoVO> getVideosByUserId(Integer userId, String sort) {
        System.out.println("【调试】VideoService.getVideosByUserId 被调用，用户ID: " + userId + ", 排序: " + sort);
        // 改为查询用户的稿件
        return getManuscriptsByUserId(userId, sort);
    }

    /**
     * 获取用户的稿件列表（转换为VideoVO格式以兼容前端）
     */
    private List<VideoVO> getManuscriptsByUserId(Integer userId, String sort) {
        List<Manuscript> manuscripts = manuscriptMapper.selectByUserId(userId);

        // 根据排序方式排序
        switch (sort) {
            case "views":
                manuscripts.sort((a, b) -> (b.getViewCount() != null ? b.getViewCount() : 0) -
                        (a.getViewCount() != null ? a.getViewCount() : 0));
                break;
            case "collects":
                manuscripts.sort((a, b) -> (b.getCollectCount() != null ? b.getCollectCount() : 0) -
                        (a.getCollectCount() != null ? a.getCollectCount() : 0));
                break;
            case "latest":
            default:
                // 默认按上传时间倒序
                manuscripts.sort((a, b) -> b.getUploadTime().compareTo(a.getUploadTime()));
                break;
        }

        System.out.println("【调试】查询到稿件数量: " + manuscripts.size());

        // 将稿件转换为VideoVO格式（兼容前端）
        List<VideoVO> result = new ArrayList<>();
        for (Manuscript manuscript : manuscripts) {
            // 获取稿件的第一个视频
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            Video firstVideo = videos.isEmpty() ? null : videos.get(0);

            VideoVO vo = new VideoVO();
            vo.setId(firstVideo != null ? firstVideo.getId() : manuscript.getId());
            vo.setTitle(manuscript.getTitle());
            vo.setDescription(manuscript.getDescription());
            vo.setCoverUrl(manuscript.getCoverUrl());
            vo.setViewCount(manuscript.getViewCount());
            vo.setLikeCount(manuscript.getLikeCount());
            vo.setCoinCount(manuscript.getCoinCount());
            vo.setCollectCount(manuscript.getCollectCount());
            vo.setShareCount(manuscript.getShareCount());
            // 统计评论数+回复数总量
            vo.setCommentCount(getTotalCommentCount(manuscript.getId()));
            vo.setUploadTime(manuscript.getUploadTime());

            // 如果有视频，设置播放地址
            if (firstVideo != null) {
                vo.setPlayUrl(firstVideo.getPlayUrl());
                vo.setPlayUrlHd(firstVideo.getPlayUrlHd());
                vo.setPlayUrlSd(firstVideo.getPlayUrlSd());
                vo.setPlayUrlLd(firstVideo.getPlayUrlLd());
                // 使用 duration_seconds 格式化时长
                Integer durationSeconds = firstVideo.getDurationSeconds();
                if (durationSeconds != null && durationSeconds > 0) {
                    int minutes = durationSeconds / 60;
                    int seconds = durationSeconds % 60;
                    vo.setDuration(String.format("%02d:%02d", minutes, seconds));
                } else {
                    vo.setDuration("00:00");
                }
            }

            // 设置稿件信息
            vo.setManuscriptId(manuscript.getId());
            VideoVO.ManuscriptInfo manuscriptInfo = new VideoVO.ManuscriptInfo();
            manuscriptInfo.setId(manuscript.getId());
            manuscriptInfo.setTitle(manuscript.getTitle());
            manuscriptInfo.setDescription(manuscript.getDescription());
            manuscriptInfo.setCoverUrl(manuscript.getCoverUrl());
            manuscriptInfo.setStatus(manuscript.getStatus());
            manuscriptInfo.setUploadTime(manuscript.getUploadTime());
            vo.setManuscript(manuscriptInfo);

            result.add(vo);
        }

        return result;
    }

    @Override
    public List<VideoVO> getVideosByCategoryId(Integer categoryId) {
        // 改为查询该分类下的稿件
        List<Manuscript> manuscripts = manuscriptMapper.selectByCategoryId(categoryId);

        // 将稿件转换为VideoVO格式（兼容前端）
        List<VideoVO> result = new ArrayList<>();
        for (Manuscript manuscript : manuscripts) {
            // 获取稿件的第一个视频
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            Video firstVideo = videos.isEmpty() ? null : videos.get(0);

            VideoVO vo = new VideoVO();
            vo.setId(firstVideo != null ? firstVideo.getId() : manuscript.getId());
            vo.setTitle(manuscript.getTitle());
            vo.setDescription(manuscript.getDescription());
            vo.setCoverUrl(manuscript.getCoverUrl());
            vo.setViewCount(manuscript.getViewCount());
            vo.setLikeCount(manuscript.getLikeCount());
            vo.setCoinCount(manuscript.getCoinCount());
            vo.setCollectCount(manuscript.getCollectCount());
            vo.setShareCount(manuscript.getShareCount());
            // 统计评论数+回复数总量
            vo.setCommentCount(getTotalCommentCount(manuscript.getId()));
            vo.setUploadTime(manuscript.getUploadTime());
            vo.setCategoryId(manuscript.getCategoryId());

            // 如果有视频，设置播放地址
            if (firstVideo != null) {
                vo.setPlayUrl(firstVideo.getPlayUrl());
                vo.setPlayUrlHd(firstVideo.getPlayUrlHd());
                vo.setPlayUrlSd(firstVideo.getPlayUrlSd());
                vo.setPlayUrlLd(firstVideo.getPlayUrlLd());
                // 使用 duration_seconds 格式化时长
                Integer durationSeconds = firstVideo.getDurationSeconds();
                if (durationSeconds != null && durationSeconds > 0) {
                    int minutes = durationSeconds / 60;
                    int seconds = durationSeconds % 60;
                    vo.setDuration(String.format("%02d:%02d", minutes, seconds));
                } else {
                    vo.setDuration("00:00");
                }
            }

            // 设置稿件信息
            vo.setManuscriptId(manuscript.getId());
            VideoVO.ManuscriptInfo manuscriptInfo = new VideoVO.ManuscriptInfo();
            manuscriptInfo.setId(manuscript.getId());
            manuscriptInfo.setTitle(manuscript.getTitle());
            manuscriptInfo.setDescription(manuscript.getDescription());
            manuscriptInfo.setCoverUrl(manuscript.getCoverUrl());
            manuscriptInfo.setStatus(manuscript.getStatus());
            manuscriptInfo.setUploadTime(manuscript.getUploadTime());
            vo.setManuscript(manuscriptInfo);

            result.add(vo);
        }

        return result;
    }

    @Override
    public List<VideoVO> getRecommendedVideos() {
        List<Video> videos = videoMapper.selectRecommended();
        // 去重逻辑
        List<Video> uniqueVideos = new ArrayList<>();
        Set<Integer> videoIds = new HashSet<>();
        for (Video video : videos) {
            if (!videoIds.contains(video.getId())) {
                videoIds.add(video.getId());
                uniqueVideos.add(video);
            }
        }
        return buildVideoVOList(uniqueVideos);
    }

    @Override
    public List<VideoVO> getHotVideos() {
        List<Video> videos = videoMapper.selectHot();
        return buildVideoVOList(videos);
    }

    @Override
    public void updateViewCount(Integer id) {
        videoMapper.updateViewCount(id);
    }

    @Override
    public List<VideoVO> getVideoList(Integer page, Integer size) {
        int offset = (page - 1) * size;
        List<Video> videos = videoMapper.selectList(offset, size);
        return buildVideoVOList(videos);
    }

    @Override
    public VideoVO updateVideo(Integer id, VideoUploadDTO videoUploadDTO, Integer userId) throws Exception {
        // 检查视频是否存在
        Video video = videoMapper.selectById(id);
        if (video == null) {
            throw new RuntimeException("视频不存在");
        }

        // 检查是否是视频的上传者
        if (!video.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限修改此视频");
        }

        // 更新视频信息
        if (videoUploadDTO.getTitle() != null) {
            video.setTitle(videoUploadDTO.getTitle());
        }
        if (videoUploadDTO.getDescription() != null) {
            video.setDescription(videoUploadDTO.getDescription());
        }
        if (videoUploadDTO.getCategoryId() != null) {
            video.setCategoryId(videoUploadDTO.getCategoryId());
        }

        // 获取稿件ID
        Integer manuscriptId = video.getManuscriptId();

        // 处理封面更新（更新到稿件目录）
        if (videoUploadDTO.getCover() != null && !videoUploadDTO.getCover().isEmpty()) {
            String coverPath = uploadFilePathUtils.getManuscriptCoverPath(manuscriptId);
            videoUploadDTO.getCover().transferTo(new File(coverPath));
            video.setCoverUrl(uploadFilePathUtils.getManuscriptCoverUrl(manuscriptId));
        }

        // 处理视频更新
        if (videoUploadDTO.getVideo() != null && !videoUploadDTO.getVideo().isEmpty()) {
            String videoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, id);
            videoUploadDTO.getVideo().transferTo(new File(videoPath));
            video.setPlayUrl(uploadFilePathUtils.getVideoSourcePath(manuscriptId, id));
        }

        // 保存更新
        videoMapper.update(video);

        // 处理标签更新
        if (videoUploadDTO.getTags() != null) {
            // 删除旧标签关联
            videoMapper.deleteVideoTags(id);
            // 添加新标签关联
            for (String tagName : videoUploadDTO.getTags()) {
                Tag tag = tagMapper.selectByName(tagName);
                if (tag == null) {
                    tag = new Tag();
                    tag.setName(tagName);
                    tagMapper.insert(tag);
                }
                VideoTag videoTag = new VideoTag();
                videoTag.setVideoId(id);
                videoTag.setTagId(tag.getId());
                videoMapper.insertVideoTag(videoTag);
            }
        }

        // 注意：索引更新已由admin端处理，web端不再直接操作ES索引
        // 稿件审核通过后会自动同步到ES

        return buildVideoVO(video);
    }

    @Override
    public void deleteVideo(Integer id, Integer userId) throws Exception {
        // 检查视频是否存在
        Video video = videoMapper.selectById(id);
        if (video == null) {
            throw new RuntimeException("视频不存在");
        }

        // 检查是否是视频的上传者
        if (!video.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此视频");
        }

        // 获取稿件ID
        Integer manuscriptId = video.getManuscriptId();

        // 删除视频标签关联
        videoMapper.deleteVideoTags(id);
        // 删除视频
        videoMapper.delete(id);

        // 删除视频目录（如果有关联稿件）
        if (manuscriptId != null) {
            uploadFilePathUtils.deleteVideoDirectory(manuscriptId, id);
        }
    }

    private VideoVO buildVideoVO(Video video) {
        return buildVideoVO(video, null);
    }

    /**
     * 设置视频基本信息（兼容旧数据）
     */
    private void setVideoBasicInfo(VideoVO videoVO, Video video) {
        videoVO.setTitle(video.getTitle());
        videoVO.setDescription(video.getDescription());
        videoVO.setCoverUrl(video.getCoverUrl());
        videoVO.setUploadTime(video.getUploadTime());
        videoVO.setViewCount(video.getViewCount());
        videoVO.setLikeCount(video.getLikeCount());
        videoVO.setCoinCount(video.getCoinCount());
        // 使用数据库查询计算收藏数
        int collectCount = videoMapper.countCollections(video.getId());
        videoVO.setCollectCount(collectCount);
        videoVO.setShareCount(video.getShareCount());
        // 统计评论数+回复数总量
        if (video.getManuscriptId() != null) {
            videoVO.setCommentCount(getTotalCommentCount(video.getManuscriptId()));
        } else {
            videoVO.setCommentCount(video.getCommentCount());
        }
        videoVO.setDanmakuCount(video.getDanmakuCount());
    }

    private VideoVO buildVideoVO(Video video, Integer currentUserId) {
        VideoVO videoVO = new VideoVO();
        videoVO.setId(video.getId());
        
        // 如果有稿件ID，查询稿件获取完整信息
        if (video.getManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
            if (manuscript != null) {
                // 稿件级别的基本信息
                videoVO.setTitle(manuscript.getTitle());
                videoVO.setDescription(manuscript.getDescription());
                videoVO.setCoverUrl(manuscript.getCoverUrl());
                videoVO.setUploadTime(manuscript.getUploadTime());
                
                // 稿件级别的统计信息（关键改动：所有互动数据来自稿件）
                videoVO.setViewCount(manuscript.getViewCount());
                videoVO.setLikeCount(manuscript.getLikeCount());
                videoVO.setCoinCount(manuscript.getCoinCount());
                videoVO.setCollectCount(manuscript.getCollectCount());
                videoVO.setShareCount(manuscript.getShareCount());
                // 统计评论数+回复数总量
                videoVO.setCommentCount(getTotalCommentCount(manuscript.getId()));
                videoVO.setDanmakuCount(manuscript.getDanmakuCount());
                
                // 设置稿件ID和稿件信息
                videoVO.setManuscriptId(manuscript.getId());
                VideoVO.ManuscriptInfo manuscriptInfo = new VideoVO.ManuscriptInfo();
                manuscriptInfo.setId(manuscript.getId());
                manuscriptInfo.setTitle(manuscript.getTitle());
                manuscriptInfo.setDescription(manuscript.getDescription());
                manuscriptInfo.setCoverUrl(manuscript.getCoverUrl());
                manuscriptInfo.setStatus(manuscript.getStatus());
                manuscriptInfo.setUploadTime(manuscript.getUploadTime());
                videoVO.setManuscript(manuscriptInfo);
            } else {
                // 稿件不存在，使用video表数据
                setVideoBasicInfo(videoVO, video);
            }
        } else {
            // 兼容旧数据（没有关联稿件）
            setVideoBasicInfo(videoVO, video);
        }
        
        // 视频级别的信息（播放地址、时长等）
        videoVO.setPlayUrl(video.getPlayUrl());
        videoVO.setPlayUrlHd(video.getPlayUrlHd());
        videoVO.setPlayUrlSd(video.getPlayUrlSd());
        videoVO.setPlayUrlLd(video.getPlayUrlLd());
        videoVO.setCategoryId(video.getCategoryId());
        // 使用 duration_seconds 格式化时长
        Integer durationSeconds = video.getDurationSeconds();
        if (durationSeconds != null && durationSeconds > 0) {
            int minutes = durationSeconds / 60;
            int seconds = durationSeconds % 60;
            videoVO.setDuration(String.format("%02d:%02d", minutes, seconds));
        } else {
            videoVO.setDuration("00:00");
        }
        videoVO.setStatus(video.getStatus());

        // 设置分类名称
        if (video.getCategoryId() != null) {
            String categoryName = categoryMapper.selectById(video.getCategoryId()).getName();
            videoVO.setCategoryName(categoryName);
        }

        // 设置标签
        List<Tag> tags = tagMapper.selectByVideoId(video.getId());
        List<String> tagNames = new ArrayList<>();
        for (Tag tag : tags) {
            tagNames.add(tag.getName());
        }
        videoVO.setTags(tagNames);

        // 设置上传者信息
        VideoVO.UserInfo userInfo = new VideoVO.UserInfo();
        com.mybilibili.common.entity.User user = userMapper.findById(video.getUserId());
        System.out.println("【调试】buildVideoVO - 视频上传者ID: " + video.getUserId() + ", 当前用户ID: " + currentUserId);
        if (user != null) {
            userInfo.setId(user.getId());
            userInfo.setName(user.getNickname());
            userInfo.setAvatar(user.getAvatar());
            userInfo.setLevel(user.getLevel());
            userInfo.setBio(user.getBio());
            userInfo.setSignature(user.getSignature());
            System.out.println("【调试】用户 " + user.getId() + " 的bio: " + user.getBio() + ", signature: " + user.getSignature());
            // 实时计算粉丝数
            int followerCount = followService.getFollowerCount(user.getId());
            userInfo.setFollowerCount(followerCount);
            System.out.println("【调试】用户 " + user.getId() + " 的粉丝数: " + followerCount);
            // 检查当前用户是否关注了该用户
            boolean isFollowing = false;
            if (currentUserId != null) {
                if (!currentUserId.equals(user.getId())) {
                    isFollowing = followService.isFollowing(currentUserId, user.getId());
                    System.out.println("【调试】检查关注状态 - 关注者: " + currentUserId + ", 被关注者: " + user.getId() + ", 结果: " + isFollowing);
                } else {
                    System.out.println("【调试】跳过关注检查 - 不能关注自己");
                }
            } else {
                System.out.println("【调试】未登录用户，isFollowing设置为false");
            }
            userInfo.setFollowing(isFollowing);
            System.out.println("【调试】最终设置 following: " + isFollowing);
        }
        videoVO.setUploader(userInfo);

        return videoVO;
    }

    private List<VideoVO> buildVideoVOList(List<Video> videos) {
        List<VideoVO> videoVOs = new ArrayList<>();
        for (Video video : videos) {
            videoVOs.add(buildVideoVO(video));
        }
        return videoVOs;
    }

    @Override
    public void transcodeAllVideos() {
        List<Video> videos = videoMapper.selectAll();
        System.out.println("开始批量转码，共 " + videos.size() + " 个视频");
        
        for (Video video : videos) {
            try {
                transcodeVideo(video.getId());
            } catch (Exception e) {
                System.err.println("视频 " + video.getId() + " 转码失败: " + e.getMessage());
            }
        }
    }

    @Override
    public void transcodeVideo(Integer id) {
        Video video = videoMapper.selectById(id);
        if (video == null) {
            throw new RuntimeException("视频不存在: " + id);
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            throw new RuntimeException("视频未关联稿件: " + id);
        }

        String playUrl = video.getPlayUrl();
        if (playUrl == null || playUrl.isEmpty()) {
            throw new RuntimeException("视频URL为空: " + id);
        }

        // 获取视频源文件路径
        String absoluteVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, id);
        // 获取转码输出目录
        String outputDir = uploadFilePathUtils.getVideoTranscodedDir(manuscriptId, id);
        // 确保输出目录存在
        uploadFilePathUtils.ensureDirectoryExists(outputDir);

        System.out.println("开始转码视频: videoId=" + video.getId() + ", manuscriptId=" + manuscriptId + ", 输入路径=" + absoluteVideoPath + ", 输出目录=" + outputDir);

        ffmpegUtils.transcodeVideo(absoluteVideoPath, outputDir, video.getId(), new FFmpegUtils.VideoTranscodeCallback() {
            @Override
            public void onTranscodeComplete(String hdPath, String sdPath, String ldPath) {
                try {
                    // 使用新的URL格式
                    String hdUrl = uploadFilePathUtils.getHdVideoUrl(manuscriptId, video.getId());
                    String sdUrl = uploadFilePathUtils.getSdVideoUrl(manuscriptId, video.getId());
                    String ldUrl = uploadFilePathUtils.getLdVideoUrl(manuscriptId, video.getId());

                    Video updateVideo = new Video();
                    updateVideo.setId(video.getId());
                    updateVideo.setPlayUrl(hdUrl);
                    updateVideo.setPlayUrlHd(hdUrl);
                    updateVideo.setPlayUrlSd(sdUrl);
                    updateVideo.setPlayUrlLd(ldUrl);
                    updateVideo.setStatus(Video.STATUS_PUBLISHED);

                    int durationSeconds = ffmpegUtils.getVideoDuration(hdPath);
                    updateVideo.setDurationSeconds(durationSeconds);

                    videoMapper.update(updateVideo);
                    
                    // 更新稿件总时长
                    updateManuscriptDuration(video.getManuscriptId());
                    System.out.println("视频转码完成: " + video.getId() + ", HD: " + hdUrl + ", SD: " + sdUrl + ", LD: " + ldUrl);

                    // 更新关联稿件的状态
                    Manuscript updateManuscript = new Manuscript();
                    updateManuscript.setId(manuscriptId);
                    updateManuscript.setStatus(Manuscript.STATUS_PUBLISHED);
                    manuscriptMapper.update(updateManuscript);
                    System.out.println("稿件状态更新为已上架: manuscriptId=" + manuscriptId);
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }

            @Override
            public void onTranscodeError(String errorMessage) {
                System.err.println("视频转码失败: " + video.getId() + ", 错误: " + errorMessage);
                try {
                    Video updateVideo = new Video();
                    updateVideo.setId(video.getId());
                    updateVideo.setStatus(Video.STATUS_PROCESS_FAILED);
                    videoMapper.update(updateVideo);

                    // 更新关联稿件的状态为处理失败
                    Manuscript updateManuscript = new Manuscript();
                    updateManuscript.setId(manuscriptId);
                    updateManuscript.setStatus(Manuscript.STATUS_PROCESS_FAILED);
                    manuscriptMapper.update(updateManuscript);
                    System.out.println("稿件状态更新为处理失败: manuscriptId=" + manuscriptId);
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }
        });
    }

    @Override
    public VideoVO getVideoByManuscriptId(Integer manuscriptId, Integer p, Integer currentUserId) {
        // 1. 查询稿件下的所有视频
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        if (videos.isEmpty()) {
            return null;
        }

        // 2. 按videoOrder排序
        videos.sort((a, b) -> {
            int orderA = a.getVideoOrder() != null ? a.getVideoOrder() : 0;
            int orderB = b.getVideoOrder() != null ? b.getVideoOrder() : 0;
            return orderA - orderB;
        });

        // 3. 获取第p个视频（默认第一个，p从1开始）
        int videoIndex = Math.min(Math.max(p - 1, 0), videos.size() - 1);
        Video currentVideo = videos.get(videoIndex);

        // 4. 构建VO
        VideoVO vo = buildVideoVO(currentVideo, currentUserId);

        // 5. 添加稿件下的所有视频列表（用于分P显示）
        List<VideoVO.VideoItemVO> videoItems = new ArrayList<>();
        for (int i = 0; i < videos.size(); i++) {
            Video video = videos.get(i);
            VideoVO.VideoItemVO item = new VideoVO.VideoItemVO();
            item.setId(video.getId());
            item.setTitle(video.getTitle());
            item.setVideoOrder(video.getVideoOrder() != null ? video.getVideoOrder() : i);
            // 使用 duration_seconds 格式化时长
            Integer durationSeconds = video.getDurationSeconds();
            if (durationSeconds != null && durationSeconds > 0) {
                int minutes = durationSeconds / 60;
                int seconds = durationSeconds % 60;
                item.setDuration(String.format("%02d:%02d", minutes, seconds));
            } else {
                item.setDuration("00:00");
            }
            item.setPlayUrl(video.getPlayUrl());
            videoItems.add(item);
        }
        vo.setManuscriptVideos(videoItems);
        vo.setCurrentVideoIndex(videoIndex);
        vo.setTotalVideos(videos.size());

        // 6. 更新稿件的观看次数
        manuscriptMapper.updateViewCount(manuscriptId, 1);

        return vo;
    }

    /**
     * 获取稿件的评论+回复总数
     */
    private int getTotalCommentCount(Integer manuscriptId) {
        int commentCount = commentMapper.countByManuscriptId(manuscriptId);
        int replyCount = replyMapper.countByManuscriptId(manuscriptId);
        return commentCount + replyCount;
    }

    /**
     * 更新稿件总时长
     * @param manuscriptId 稿件ID
     */
    private void updateManuscriptDuration(Integer manuscriptId) {
        if (manuscriptId == null) {
            return;
        }

        try {
            // 获取稿件下所有视频
            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            if (videos == null || videos.isEmpty()) {
                return;
            }

            // 计算总时长
            int totalDuration = videos.stream()
                    .mapToInt(v -> v.getDurationSeconds() != null ? v.getDurationSeconds() : 0)
                    .sum();

            // 更新稿件时长
            manuscriptMapper.updateDuration(manuscriptId, totalDuration);
            System.out.println("更新稿件 " + manuscriptId + " 总时长: " + totalDuration + " 秒");
        } catch (Exception e) {
            System.err.println("更新稿件时长失败: " + manuscriptId + ", 错误: " + e.getMessage());
        }
    }
}