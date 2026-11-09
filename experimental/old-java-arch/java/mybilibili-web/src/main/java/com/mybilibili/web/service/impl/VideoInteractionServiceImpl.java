package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.*;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.mapper.*;
import com.mybilibili.web.service.DanmakuService;
import com.mybilibili.web.service.InteractionQueryService;
import com.mybilibili.web.service.LikeService;
import com.mybilibili.web.service.VideoInteractionService;
import com.mybilibili.web.mapper.MessageMapper;
import com.mybilibili.common.entity.Message;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class VideoInteractionServiceImpl implements VideoInteractionService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Autowired
    private LikeService likeService;

    @Autowired
    private InteractionQueryService interactionQueryService;

    @Autowired
    private CoinMapper coinMapper;

    @Autowired
    private CollectionMapper collectionMapper;

    @Autowired
    private ShareMapper shareMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private FavoriteFolderMapper favoriteFolderMapper;

    @Autowired
    private FavoriteVideoMapper favoriteVideoMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private DanmakuService danmakuService;

    @Autowired
    private MessageMapper messageMapper;

    private static final String TARGET_TYPE_VIDEO = "VIDEO";

    // 消息类型常量
    private static final int MESSAGE_TYPE_LIKE = 4;  // 收到的赞

    /**
     * 根据videoId获取manuscriptId
     */
    private Integer getManuscriptIdByVideoId(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            throw new RuntimeException("视频不存在");
        }
        return video.getManuscriptId();
    }

    @Override
    public boolean likeVideo(Integer userId, Integer videoId) {
        // 使用新的统一点赞服务
        boolean result = likeService.like(userId, TARGET_TYPE_VIDEO, videoId);
        if (result) {
            // 更新稿件点赞数
            Integer manuscriptId = getManuscriptIdByVideoId(videoId);
            manuscriptMapper.updateLikeCount(manuscriptId, 1);
            // 更新用户点赞数
            userMapper.updateLikedCount(userId, 1);

            // 发送点赞消息通知
            sendLikeMessage(userId, videoId);
        }
        return result;
    }

    @Override
    public boolean unlikeVideo(Integer userId, Integer videoId) {
        // 使用新的统一点赞服务
        boolean result = likeService.unlike(userId, TARGET_TYPE_VIDEO, videoId);
        if (result) {
            // 更新稿件点赞数
            Integer manuscriptId = getManuscriptIdByVideoId(videoId);
            manuscriptMapper.updateLikeCount(manuscriptId, -1);
            // 更新用户点赞数
            userMapper.updateLikedCount(userId, -1);
        }
        return result;
    }

    @Override
    public boolean coinVideo(Integer userId, Integer videoId, Integer coinCount) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 检查用户硬币数量
        User user = userMapper.findById(userId);
        if (user == null) {
            throw new RuntimeException("用户不存在");
        }

        if (user.getCoinCount() < coinCount) {
            throw new RuntimeException("硬币不足");
        }

        // 检查是否已经投币
        Coin existingCoin = coinMapper.findByUserAndManuscript(userId, manuscriptId);
        if (existingCoin != null) {
            // 更新投币数量
            int oldCoinCount = existingCoin.getCoinCount();
            int coinDiff = coinCount - oldCoinCount;
            coinMapper.update(userId, manuscriptId, coinCount);

            // 更新稿件投币数
            manuscriptMapper.updateCoinCount(manuscriptId, coinDiff);

            // 更新用户硬币数
            userMapper.updateCoinCount(userId, -coinDiff);
        } else {
            // 添加投币记录
            Coin coin = new Coin();
            coin.setUserId(userId);
            coin.setManuscriptId(manuscriptId);
            coin.setCoinCount(coinCount);
            coinMapper.insert(coin);

            // 更新稿件投币数
            manuscriptMapper.updateCoinCount(manuscriptId, coinCount);

            // 更新用户硬币数
            userMapper.updateCoinCount(userId, -coinCount);
        }

        return true;
    }

    @Override
    public boolean collectVideo(Integer userId, Integer videoId) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 检查是否已经收藏
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
        if (existing != null) {
            return false; // 已经收藏过
        }

        // 添加收藏记录到统一表
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType(TARGET_TYPE_VIDEO);
        interaction.setTargetId(videoId);
        interaction.setInteractionType(InteractionType.COLLECT.getCode());
        userInteractionMapper.insert(interaction);

        // 更新稿件收藏数
        manuscriptMapper.updateCollectCount(manuscriptId, 1);

        return true;
    }

    @Override
    public boolean uncollectVideo(Integer userId, Integer videoId) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 检查是否已经收藏
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
        if (existing == null) {
            return false; // 还没有收藏
        }

        // 删除收藏记录
        userInteractionMapper.delete(userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());

        // 更新稿件收藏数
        manuscriptMapper.updateCollectCount(manuscriptId, -1);

        return true;
    }

    @Override
    public void shareVideo(Integer userId, Integer videoId, String channel, String ipAddress) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 创建分享记录
        Share share = new Share();
        share.setUserId(userId);
        share.setManuscriptId(manuscriptId);
        share.setChannel(channel != null ? channel : "unknown");
        share.setIpAddress(ipAddress);
        shareMapper.insert(share);

        // 更新稿件分享数
        manuscriptMapper.updateShareCount(manuscriptId, 1);
    }

    @Override
    public void sendDanmaku(Integer userId, Integer videoId, String content, String time, String color, Integer mode) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);
        // 使用MongoDB存储弹幕
        danmakuService.sendDanmaku(userId, videoId, manuscriptId, content, time, color, mode);
    }

    @Override
    public List<?> getDanmakus(Integer videoId) {
        // 从MongoDB查询弹幕
        return danmakuService.getDanmakus(videoId);
    }

    @Override
    public VideoInteractionStatus getInteractionStatus(Integer userId, Integer videoId) {
        VideoInteractionStatus status = new VideoInteractionStatus();

        // 使用新的统一查询服务
        Map<String, Object> interactionStatus = interactionQueryService.getStatus(userId, TARGET_TYPE_VIDEO, videoId);
        status.setLiked((Boolean) interactionStatus.get("isLiked"));
        status.setCollected((Boolean) interactionStatus.get("isCollected"));

        // 检查投币数量（投币暂时还使用旧表）
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);
        Coin coin = coinMapper.findByUserAndManuscript(userId, manuscriptId);
        status.setCoinCount(coin != null ? coin.getCoinCount() : 0);

        return status;
    }

    @Override
    public List<VideoVO> getLikedVideos(Integer userId) {
        // 从统一交互表查询用户点赞的视频
        List<UserInteraction> interactions = userInteractionMapper.findByUserAndInteractionType(
                userId, TARGET_TYPE_VIDEO, InteractionType.LIKE.getCode());

        List<VideoVO> videoVOs = new ArrayList<>();
        for (UserInteraction interaction : interactions) {
            Video video = videoMapper.selectById(interaction.getTargetId());
            if (video != null) {
                VideoVO vo = convertToVideoVO(video);
                videoVOs.add(vo);
            }
        }
        return videoVOs;
    }

    @Override
    public List<VideoVO> getCollectedVideos(Integer userId) {
        // 从统一交互表查询用户收藏的视频
        List<UserInteraction> interactions = userInteractionMapper.findByUserAndInteractionType(
                userId, TARGET_TYPE_VIDEO, InteractionType.COLLECT.getCode());

        List<VideoVO> videoVOs = new ArrayList<>();
        for (UserInteraction interaction : interactions) {
            Video video = videoMapper.selectById(interaction.getTargetId());
            if (video != null) {
                VideoVO vo = convertToVideoVO(video);
                videoVOs.add(vo);
            }
        }
        return videoVOs;
    }

    private VideoVO convertToVideoVO(Video video) {
        VideoVO videoVO = new VideoVO();
        videoVO.setId(video.getId());
        videoVO.setTitle(video.getTitle());
        videoVO.setDescription(video.getDescription());
        videoVO.setCoverUrl(video.getCoverUrl());
        videoVO.setPlayUrl(video.getPlayUrl());
        videoVO.setDuration(DurationUtils.formatDuration(video.getDurationSeconds()));

        // 获取稿件信息
        Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
        if (manuscript != null) {
            videoVO.setViewCount(manuscript.getViewCount());
            videoVO.setLikeCount(manuscript.getLikeCount());
            videoVO.setCoinCount(manuscript.getCoinCount());
            videoVO.setCollectCount(manuscript.getCollectCount());
            videoVO.setShareCount(manuscript.getShareCount());
            videoVO.setUploadTime(manuscript.getUploadTime());

            // 设置用户信息
            User user = userMapper.findById(manuscript.getUserId());
            if (user != null) {
                VideoVO.UserInfo userInfo = new VideoVO.UserInfo();
                userInfo.setId(user.getId());
                userInfo.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
                userInfo.setAvatar(user.getAvatar());
                userInfo.setLevel(user.getLevel());
                videoVO.setUploader(userInfo);
            }
        }

        return videoVO;
    }

    @Override
    public List<FavoriteFolder> getFavoriteFolders(Integer userId) {
        // 获取用户的收藏夹列表
        List<FavoriteFolder> folders = favoriteFolderMapper.findByUserId(userId);
        // 如果没有收藏夹，创建一个默认收藏夹
        if (folders.isEmpty()) {
            FavoriteFolder defaultFolder = new FavoriteFolder();
            defaultFolder.setUserId(userId);
            defaultFolder.setName("默认收藏夹");
            defaultFolder.setVideoCount(0);
            favoriteFolderMapper.insert(defaultFolder);
            folders.add(defaultFolder);
        }
        return folders;
    }

    @Override
    public FavoriteFolder createFavoriteFolder(Integer userId, String name) {
        // 检查是否已存在同名收藏夹
        FavoriteFolder existingFolder = favoriteFolderMapper.findByUserIdAndName(userId, name);
        if (existingFolder != null) {
            throw new RuntimeException("已存在同名收藏夹");
        }

        // 创建新收藏夹
        FavoriteFolder folder = new FavoriteFolder();
        folder.setUserId(userId);
        folder.setName(name);
        folder.setVideoCount(0);
        favoriteFolderMapper.insert(folder);
        return folder;
    }

    @Override
    public FavoriteFolder updateFavoriteFolder(Integer userId, Integer folderId, String name) {
        // 检查收藏夹是否存在且属于当前用户
        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null) {
            throw new RuntimeException("收藏夹不存在");
        }
        if (!folder.getUserId().equals(userId)) {
            throw new RuntimeException("无权修改该收藏夹");
        }

        // 检查是否已存在同名收藏夹（排除当前收藏夹）
        FavoriteFolder existingFolder = favoriteFolderMapper.findByUserIdAndName(userId, name);
        if (existingFolder != null && !existingFolder.getId().equals(folderId)) {
            throw new RuntimeException("已存在同名收藏夹");
        }

        // 更新收藏夹名称
        folder.setName(name);
        favoriteFolderMapper.update(folder);
        return folder;
    }

    @Override
    public boolean deleteFavoriteFolder(Integer userId, Integer folderId) {
        // 检查收藏夹是否存在且属于当前用户
        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null) {
            throw new RuntimeException("收藏夹不存在");
        }
        if (!folder.getUserId().equals(userId)) {
            throw new RuntimeException("无权删除该收藏夹");
        }

        // 获取收藏夹中的所有视频
        List<FavoriteVideo> favoriteVideos = favoriteVideoMapper.findByFolderId(folderId);

        // 删除收藏夹中的所有视频关联
        for (FavoriteVideo favoriteVideo : favoriteVideos) {
            favoriteVideoMapper.deleteById(favoriteVideo.getId());

            // 检查用户是否还有其他收藏夹包含该稿件
            List<FavoriteVideo> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, favoriteVideo.getManuscriptId());
            if (remainingFavorites.isEmpty()) {
                // 获取该稿件的第一个视频ID
                List<Video> videos = videoMapper.selectByManuscriptId(favoriteVideo.getManuscriptId());
                if (!videos.isEmpty()) {
                    Integer videoId = videos.get(0).getId();
                    // 移除收藏记录
                    userInteractionMapper.delete(userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
                    // 更新稿件收藏数
                    manuscriptMapper.updateCollectCount(favoriteVideo.getManuscriptId(), -1);
                }
            }
        }

        // 删除收藏夹
        favoriteFolderMapper.deleteById(folderId);
        return true;
    }

    @Override
    public boolean addVideoToFavoriteFolders(Integer userId, Integer videoId, List<Integer> folderIds) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        boolean success = false;
        // 检查用户是否已经收藏了该稿件
        UserInteraction existingCollection = userInteractionMapper.findByUserAndTarget(
                userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
        boolean isNewCollection = (existingCollection == null);

        for (Integer folderId : folderIds) {
            // 检查收藏夹是否属于当前用户
            FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
            if (folder == null || !folder.getUserId().equals(userId)) {
                continue;
            }

            // 检查稿件是否已在收藏夹中
            FavoriteVideo existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
            if (existing == null) {
                // 添加到收藏夹
                FavoriteVideo favoriteVideo = new FavoriteVideo();
                favoriteVideo.setFolderId(folderId);
                favoriteVideo.setManuscriptId(manuscriptId);
                favoriteVideoMapper.insert(favoriteVideo);

                // 更新收藏夹视频数量
                favoriteFolderMapper.updateVideoCount(folderId, 1);
                success = true;
            }
        }

        // 同时更新稿件的收藏数，只有当用户第一次收藏稿件时才增加
        if (success && isNewCollection) {
            // 添加收藏记录到统一表
            UserInteraction interaction = new UserInteraction();
            interaction.setUserId(userId);
            interaction.setTargetType(TARGET_TYPE_VIDEO);
            interaction.setTargetId(videoId);
            interaction.setInteractionType(InteractionType.COLLECT.getCode());
            userInteractionMapper.insert(interaction);
            // 更新稿件收藏数
            manuscriptMapper.updateCollectCount(manuscriptId, 1);
        }

        return success;
    }

    @Override
    public boolean removeVideoFromFavoriteFolder(Integer userId, Integer videoId, Integer folderId) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 检查收藏夹是否属于当前用户
        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null || !folder.getUserId().equals(userId)) {
            return false;
        }

        // 检查稿件是否在收藏夹中
        FavoriteVideo existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
        if (existing != null) {
            // 从收藏夹移除
            favoriteVideoMapper.deleteById(existing.getId());

            // 更新收藏夹视频数量
            favoriteFolderMapper.updateVideoCount(folderId, -1);

            // 检查用户是否还有其他收藏夹包含该稿件
            List<FavoriteVideo> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
            if (remainingFavorites.isEmpty()) {
                // 如果用户没有其他收藏夹包含该稿件，移除收藏记录并更新稿件收藏数
                userInteractionMapper.delete(userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
                // 更新稿件的收藏数，确保不会出现负数
                manuscriptMapper.updateCollectCount(manuscriptId, -1);
            }

            return true;
        }

        return false;
    }

    @Override
    public List<VideoVO> getFavoriteFolderVideos(Integer userId, Integer folderId, Integer page, Integer size) {
        // 检查收藏夹是否属于当前用户
        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null || !folder.getUserId().equals(userId)) {
            return new ArrayList<>();
        }

        // 计算分页偏移量
        int offset = (page - 1) * size;

        // 查询收藏夹中的稿件列表
        List<Manuscript> manuscripts = favoriteVideoMapper.findManuscriptsByFolderId(folderId, offset, size);

        // 转换为 VideoVO 列表
        List<VideoVO> videoVOs = new ArrayList<>();
        for (Manuscript manuscript : manuscripts) {
            VideoVO videoVO = new VideoVO();
            videoVO.setId(manuscript.getId());
            videoVO.setTitle(manuscript.getTitle());
            videoVO.setDescription(manuscript.getDescription());
            videoVO.setCoverUrl(manuscript.getCoverUrl());
            // 获取稿件的第一个视频作为播放链接
            List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
            if (!videos.isEmpty()) {
                Video firstVideo = videos.get(0);
                videoVO.setPlayUrl(firstVideo.getPlayUrl());
            }
            videoVO.setViewCount(manuscript.getViewCount());
            videoVO.setLikeCount(manuscript.getLikeCount());
            videoVO.setCoinCount(manuscript.getCoinCount());
            videoVO.setCollectCount(manuscript.getCollectCount());
            videoVO.setShareCount(manuscript.getShareCount());
            // 统计评论数+回复数总量
            int commentCount = commentMapper.countByManuscriptId(manuscript.getId());
            int replyCount = replyMapper.countByManuscriptId(manuscript.getId());
            videoVO.setCommentCount(commentCount + replyCount);
            videoVO.setDanmakuCount(manuscript.getDanmakuCount());
            videoVO.setUploadTime(manuscript.getUploadTime());

            // 设置用户信息
            User user = userMapper.findById(manuscript.getUserId());
            if (user != null) {
                VideoVO.UserInfo userInfo = new VideoVO.UserInfo();
                userInfo.setId(user.getId());
                userInfo.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
                userInfo.setAvatar(user.getAvatar());
                userInfo.setLevel(user.getLevel());
                userInfo.setBio(user.getBio());
                userInfo.setSignature(user.getSignature());
                videoVO.setUploader(userInfo);
            }

            videoVOs.add(videoVO);
        }

        return videoVOs;
    }

    @Override
    public Map<String, Object> getShareStatistics(Integer videoId) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        Map<String, Object> statistics = new HashMap<>();

        // 获取总分享数
        Integer totalShares = shareMapper.countByManuscriptId(manuscriptId);
        statistics.put("totalShares", totalShares);

        // 获取各渠道分享数
        List<String> channels = Arrays.asList("wechat", "weibo", "qq", "link");
        Map<String, Integer> channelShares = new HashMap<>();
        for (String channel : channels) {
            Integer count = shareMapper.countByManuscriptIdAndChannel(manuscriptId, channel);
            channelShares.put(channel, count);
        }
        statistics.put("channelShares", channelShares);

        return statistics;
    }

    @Override
    public List<FavoriteFolder> getVideoFavoriteFolders(Integer userId, Integer videoId) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 获取用户的所有收藏夹
        List<FavoriteFolder> allFolders = getFavoriteFolders(userId);
        // 获取稿件在哪些收藏夹中（只查询当前用户的）
        List<FavoriteVideo> favoriteVideos = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
        // 构建收藏夹ID到收藏夹的映射
        Map<Integer, FavoriteFolder> folderMap = new HashMap<>();
        for (FavoriteFolder folder : allFolders) {
            folderMap.put(folder.getId(), folder);
        }
        // 筛选出稿件所在的收藏夹
        List<FavoriteFolder> result = new ArrayList<>();
        for (FavoriteVideo favoriteVideo : favoriteVideos) {
            FavoriteFolder folder = folderMap.get(favoriteVideo.getFolderId());
            if (folder != null) {
                result.add(folder);
            }
        }
        return result;
    }

    @Override
    public boolean updateVideoFavoriteFolders(Integer userId, Integer videoId, List<Integer> folderIds) {
        // 通过videoId找到manuscriptId
        Integer manuscriptId = getManuscriptIdByVideoId(videoId);

        // 获取稿件当前所在的收藏夹
        List<FavoriteFolder> currentFolders = getVideoFavoriteFolders(userId, videoId);
        // 构建当前收藏夹ID的集合
        Set<Integer> currentFolderIds = new HashSet<>();
        for (FavoriteFolder folder : currentFolders) {
            currentFolderIds.add(folder.getId());
        }
        // 构建新收藏夹ID的集合
        Set<Integer> newFolderIds = new HashSet<>(folderIds);

        boolean success = false;

        // 移除不在新列表中的收藏夹
        for (Integer folderId : currentFolderIds) {
            if (!newFolderIds.contains(folderId)) {
                if (removeVideoFromFavoriteFolder(userId, videoId, folderId)) {
                    success = true;
                }
            }
        }

        // 添加到新的收藏夹
        for (Integer folderId : newFolderIds) {
            if (!currentFolderIds.contains(folderId)) {
                // 检查收藏夹是否属于当前用户
                FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
                if (folder != null && folder.getUserId().equals(userId)) {
                    // 检查稿件是否已在收藏夹中
                    FavoriteVideo existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
                    if (existing == null) {
                        // 添加到收藏夹
                        FavoriteVideo favoriteVideo = new FavoriteVideo();
                        favoriteVideo.setFolderId(folderId);
                        favoriteVideo.setManuscriptId(manuscriptId);
                        favoriteVideoMapper.insert(favoriteVideo);

                        // 更新收藏夹视频数量
                        favoriteFolderMapper.updateVideoCount(folderId, 1);
                        success = true;
                    }
                }
            }
        }

        // 检查用户是否还有其他收藏夹包含该稿件
        List<FavoriteVideo> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
        // 检查用户是否已经收藏了该稿件
        UserInteraction existingCollection = userInteractionMapper.findByUserAndTarget(
                userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());

        if (remainingFavorites.isEmpty() && existingCollection != null) {
            // 如果用户没有其他收藏夹包含该稿件，但有收藏记录，移除收藏记录
            userInteractionMapper.delete(userId, TARGET_TYPE_VIDEO, videoId, InteractionType.COLLECT.getCode());
            // 更新稿件的收藏数，确保不会出现负数
            manuscriptMapper.updateCollectCount(manuscriptId, -1);
        } else if (!remainingFavorites.isEmpty() && existingCollection == null) {
            // 如果用户有其他收藏夹包含该稿件，但没有收藏记录，添加收藏记录
            UserInteraction interaction = new UserInteraction();
            interaction.setUserId(userId);
            interaction.setTargetType(TARGET_TYPE_VIDEO);
            interaction.setTargetId(videoId);
            interaction.setInteractionType(InteractionType.COLLECT.getCode());
            userInteractionMapper.insert(interaction);
            // 更新稿件的收藏数
            manuscriptMapper.updateCollectCount(manuscriptId, 1);
        }

        // 如果没有添加或移除操作，但收藏夹列表发生了变化，也返回成功
        if (!success && !currentFolderIds.equals(newFolderIds)) {
            success = true;
        }

        return success;
    }

    /**
     * 发送点赞消息通知
     */
    private void sendLikeMessage(Integer senderId, Integer videoId) {
        try {
            // 获取视频信息
            Video video = videoMapper.selectById(videoId);
            if (video == null) {
                return;
            }

            // 获取稿件作者ID
            Integer receiverId = video.getUserId();

            // 不给自己点赞发送消息
            if (receiverId == null || receiverId.equals(senderId)) {
                return;
            }

            Message message = new Message();
            message.setSenderId(senderId);
            message.setReceiverId(receiverId);
            message.setContent("赞了你的视频");
            message.setMessageType(MESSAGE_TYPE_LIKE);
            message.setTargetId(videoId);  // 设置视频ID
            message.setIsRead(false);
            message.setCreatedAt(new Date());
            messageMapper.insert(message);
        } catch (Exception e) {
            // 消息发送失败不影响点赞功能，记录日志即可
            System.err.println("发送点赞消息通知失败：" + e.getMessage());
        }
    }
}
