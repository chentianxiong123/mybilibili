package com.mybilibili.interaction.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.common.entity.*;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.interaction.mapper.*;
import com.mybilibili.interaction.service.VideoInteractionService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.*;

@Slf4j
@Service
public class VideoInteractionServiceImpl implements VideoInteractionService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Autowired
    private ShareMapper shareMapper;

    @Autowired
    private FavoriteFolderMapper favoriteFolderMapper;

    @Autowired
    private FavoriteVideoMapper favoriteVideoMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String TARGET_TYPE_MANUSCRIPT = "MANUSCRIPT";

    @Override
    public boolean likeVideo(Integer userId, Integer manuscriptId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.LIKE.getCode());
        UserInteraction existing = userInteractionMapper.selectOne(wrapper);
        if (existing != null) {
            return false;
        }

        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType(TARGET_TYPE_MANUSCRIPT);
        interaction.setTargetId(manuscriptId);
        interaction.setInteractionType(InteractionType.LIKE.getCode());
        userInteractionMapper.insert(interaction);

        redisTemplate.opsForValue().increment("like:" + TARGET_TYPE_MANUSCRIPT + ":" + manuscriptId);

        manuscriptMapper.updateLikeCount(manuscriptId, 1);
        userMapper.updateLikedCount(userId, 1);

        return true;
    }

    @Override
    public boolean unlikeVideo(Integer userId, Integer manuscriptId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.LIKE.getCode());
        UserInteraction existing = userInteractionMapper.selectOne(wrapper);
        if (existing == null) {
            return false;
        }

        userInteractionMapper.delete(wrapper);
        redisTemplate.opsForValue().decrement("like:" + TARGET_TYPE_MANUSCRIPT + ":" + manuscriptId);

        manuscriptMapper.updateLikeCount(manuscriptId, -1);
        userMapper.updateLikedCount(userId, -1);

        return true;
    }

    @Override
    public boolean coinVideo(Integer userId, Integer manuscriptId, Integer coinCount) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        User user = userMapper.selectById(userId);
        if (user == null) {
            throw new BusinessException("用户不存在");
        }

        if (user.getCoinCount() < coinCount) {
            throw new BusinessException("硬币不足");
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.COIN.getCode());
        UserInteraction existing = userInteractionMapper.selectOne(wrapper);

        if (existing != null) {
            int oldCoinCount = 1;
            int coinDiff = coinCount - oldCoinCount;
            manuscriptMapper.updateCoinCount(manuscriptId, coinDiff);
            userMapper.updateCoinCount(userId, -coinDiff);
        } else {
            UserInteraction interaction = new UserInteraction();
            interaction.setUserId(userId);
            interaction.setTargetType(TARGET_TYPE_MANUSCRIPT);
            interaction.setTargetId(manuscriptId);
            interaction.setInteractionType(InteractionType.COIN.getCode());
            userInteractionMapper.insert(interaction);
            manuscriptMapper.updateCoinCount(manuscriptId, coinCount);
            userMapper.updateCoinCount(userId, -coinCount);
        }

        return true;
    }

    @Override
    public boolean collectVideo(Integer userId, Integer manuscriptId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
        UserInteraction existing = userInteractionMapper.selectOne(wrapper);
        if (existing != null) {
            return false;
        }

        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType(TARGET_TYPE_MANUSCRIPT);
        interaction.setTargetId(manuscriptId);
        interaction.setInteractionType(InteractionType.COLLECT.getCode());
        userInteractionMapper.insert(interaction);

        manuscriptMapper.updateCollectCount(manuscriptId, 1);

        return true;
    }

    @Override
    public boolean uncollectVideo(Integer userId, Integer manuscriptId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
        UserInteraction existing = userInteractionMapper.selectOne(wrapper);
        if (existing == null) {
            return false;
        }

        userInteractionMapper.delete(wrapper);
        manuscriptMapper.updateCollectCount(manuscriptId, -1);

        return true;
    }

    @Override
    public void shareVideo(Integer userId, Integer manuscriptId, String channel, String ipAddress) {
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        Share share = new Share();
        share.setUserId(userId);
        share.setManuscriptId(manuscriptId);
        share.setChannel(channel != null ? channel : "unknown");
        share.setIpAddress(ipAddress);
        shareMapper.insert(share);

        manuscriptMapper.updateShareCount(manuscriptId, 1);
    }

    @Override
    public Map<String, Object> getShareStatistics(Integer manuscriptId) {
        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        Map<String, Object> statistics = new HashMap<>();
        Integer totalShares = shareMapper.countByManuscriptId(manuscriptId);
        statistics.put("totalShares", totalShares);

        List<String> channels = Arrays.asList("wechat", "weibo", "qq", "link");
        Map<String, Integer> channelShares = new HashMap<>();
        for (String ch : channels) {
            Integer count = shareMapper.countByManuscriptIdAndChannel(manuscriptId, ch);
            channelShares.put(ch, count);
        }
        statistics.put("channelShares", channelShares);

        return statistics;
    }

    @Override
    public VideoInteractionStatus getInteractionStatus(Integer userId, Integer manuscriptId) {
        VideoInteractionStatus status = new VideoInteractionStatus();

        if (userId == null) {
            status.setLiked(false);
            status.setCollected(false);
            status.setCoinCount(0);
            return status;
        }

        LambdaQueryWrapper<UserInteraction> likeWrapper = new LambdaQueryWrapper<>();
        likeWrapper.eq(UserInteraction::getUserId, userId)
                   .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                   .eq(UserInteraction::getTargetId, manuscriptId)
                   .eq(UserInteraction::getInteractionType, InteractionType.LIKE.getCode());
        UserInteraction likeInteraction = userInteractionMapper.selectOne(likeWrapper);
        status.setLiked(likeInteraction != null);

        LambdaQueryWrapper<UserInteraction> collectWrapper = new LambdaQueryWrapper<>();
        collectWrapper.eq(UserInteraction::getUserId, userId)
                      .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                      .eq(UserInteraction::getTargetId, manuscriptId)
                      .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
        UserInteraction collectInteraction = userInteractionMapper.selectOne(collectWrapper);
        status.setCollected(collectInteraction != null);

        LambdaQueryWrapper<UserInteraction> coinWrapper = new LambdaQueryWrapper<>();
        coinWrapper.eq(UserInteraction::getUserId, userId)
                  .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                  .eq(UserInteraction::getTargetId, manuscriptId)
                  .eq(UserInteraction::getInteractionType, InteractionType.COIN.getCode());
        UserInteraction coinInteraction = userInteractionMapper.selectOne(coinWrapper);
        status.setCoinCount(coinInteraction != null ? 1 : 0);

        return status;
    }

    @Override
    public List<VideoVO> getLikedVideos(Integer userId) {
        if (userId == null) {
            return new ArrayList<>();
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getInteractionType, InteractionType.LIKE.getCode())
               .orderByDesc(UserInteraction::getCreatedAt);
        List<UserInteraction> interactions = userInteractionMapper.selectList(wrapper);

        List<VideoVO> videoVOs = new ArrayList<>();
        for (UserInteraction interaction : interactions) {
            Manuscript manuscript = manuscriptMapper.selectById(interaction.getTargetId());
            if (manuscript != null) {
                LambdaQueryWrapper<Video> videoWrapper = new LambdaQueryWrapper<>();
                videoWrapper.eq(Video::getManuscriptId, manuscript.getId());
                Video video = videoMapper.selectOne(videoWrapper);
                if (video != null) {
                    VideoVO vo = convertToVideoVO(video);
                    videoVOs.add(vo);
                }
            }
        }
        return videoVOs;
    }

    @Override
    public List<VideoVO> getCollectedVideos(Integer userId) {
        if (userId == null) {
            return new ArrayList<>();
        }

        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode())
               .orderByDesc(UserInteraction::getCreatedAt);
        List<UserInteraction> interactions = userInteractionMapper.selectList(wrapper);

        List<VideoVO> videoVOs = new ArrayList<>();
        for (UserInteraction interaction : interactions) {
            Manuscript manuscript = manuscriptMapper.selectById(interaction.getTargetId());
            if (manuscript != null) {
                LambdaQueryWrapper<Video> videoWrapper = new LambdaQueryWrapper<>();
                videoWrapper.eq(Video::getManuscriptId, manuscript.getId());
                Video video = videoMapper.selectOne(videoWrapper);
                if (video != null) {
                    VideoVO vo = convertToVideoVO(video);
                    videoVOs.add(vo);
                }
            }
        }
        return videoVOs;
    }

    @Override
    public List<FavoriteFolder> getFavoriteFolders(Integer userId) {
        if (userId == null) {
            return new ArrayList<>();
        }
        List<FavoriteFolder> folders = favoriteFolderMapper.findByUserId(userId);
        if (folders.isEmpty()) {
            FavoriteFolder defaultFolder = new FavoriteFolder();
            defaultFolder.setUserId(userId);
            defaultFolder.setName("默认收藏夹");
            defaultFolder.setCollectCount(0);
            favoriteFolderMapper.insert(defaultFolder);
            folders.add(defaultFolder);
        }
        return folders;
    }

    @Override
    public FavoriteFolder createFavoriteFolder(Integer userId, String name) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        if (name == null || name.trim().isEmpty()) {
            throw new BusinessException("收藏夹名称不能为空");
        }

        FavoriteFolder existingFolder = favoriteFolderMapper.findByUserIdAndName(userId, name);
        if (existingFolder != null) {
            throw new BusinessException("已存在同名收藏夹");
        }

        FavoriteFolder folder = new FavoriteFolder();
        folder.setUserId(userId);
        folder.setName(name);
        folder.setCollectCount(0);
        favoriteFolderMapper.insert(folder);
        return folder;
    }

    @Override
    public FavoriteFolder updateFavoriteFolder(Integer userId, Integer folderId, String name) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null) {
            throw new BusinessException("收藏夹不存在");
        }
        if (!folder.getUserId().equals(userId)) {
            throw new BusinessException("无权修改该收藏夹");
        }

        FavoriteFolder existingFolder = favoriteFolderMapper.findByUserIdAndName(userId, name);
        if (existingFolder != null && !existingFolder.getId().equals(folderId)) {
            throw new BusinessException("已存在同名收藏夹");
        }

        folder.setName(name);
        favoriteFolderMapper.update(folder);
        return folder;
    }

    @Override
    public boolean deleteFavoriteFolder(Integer userId, Integer folderId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null) {
            throw new BusinessException("收藏夹不存在");
        }
        if (!folder.getUserId().equals(userId)) {
            throw new BusinessException("无权删除该收藏夹");
        }

        List<ManuscriptFavorite> favoriteVideos = favoriteVideoMapper.findByFolderId(folderId);
        for (ManuscriptFavorite favoriteVideo : favoriteVideos) {
            favoriteVideoMapper.deleteById(favoriteVideo.getId());
            List<ManuscriptFavorite> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, favoriteVideo.getManuscriptId());
            if (remainingFavorites.isEmpty()) {
                manuscriptMapper.updateCollectCount(favoriteVideo.getManuscriptId(), -1);
            }
        }

        favoriteFolderMapper.deleteById(folderId);
        return true;
    }

    @Override
    public boolean addVideoToFavoriteFolders(Integer userId, Integer manuscriptId, List<Integer> folderIds) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        boolean success = false;

        LambdaQueryWrapper<UserInteraction> collectWrapper = new LambdaQueryWrapper<>();
        collectWrapper.eq(UserInteraction::getUserId, userId)
                      .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                      .eq(UserInteraction::getTargetId, manuscriptId)
                      .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
        UserInteraction existingCollection = userInteractionMapper.selectOne(collectWrapper);
        boolean isNewCollection = (existingCollection == null);

        for (Integer folderId : folderIds) {
            FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
            if (folder == null || !folder.getUserId().equals(userId)) {
                continue;
            }

            ManuscriptFavorite existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
            if (existing == null) {
                ManuscriptFavorite manuscriptFavorite = new ManuscriptFavorite();
                manuscriptFavorite.setFolderId(folderId);
                manuscriptFavorite.setManuscriptId(manuscriptId);
                favoriteVideoMapper.insert(manuscriptFavorite);
                favoriteFolderMapper.updateCollectCount(folderId, 1);
                success = true;
            }
        }

        if (success && isNewCollection) {
            UserInteraction interaction = new UserInteraction();
            interaction.setUserId(userId);
            interaction.setTargetType(TARGET_TYPE_MANUSCRIPT);
            interaction.setTargetId(manuscriptId);
            interaction.setInteractionType(InteractionType.COLLECT.getCode());
            userInteractionMapper.insert(interaction);
            manuscriptMapper.updateCollectCount(manuscriptId, 1);
        }

        return success;
    }

    @Override
    public boolean removeVideoFromFavoriteFolder(Integer userId, Integer manuscriptId, Integer folderId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null || !folder.getUserId().equals(userId)) {
            return false;
        }

        ManuscriptFavorite existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
        if (existing != null) {
            favoriteVideoMapper.deleteById(existing.getId());
            favoriteFolderMapper.updateCollectCount(folderId, -1);

            List<ManuscriptFavorite> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
            if (remainingFavorites.isEmpty()) {
                LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
                wrapper.eq(UserInteraction::getUserId, userId)
                       .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                       .eq(UserInteraction::getTargetId, manuscriptId)
                       .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
                userInteractionMapper.delete(wrapper);
                manuscriptMapper.updateCollectCount(manuscriptId, -1);
            }
            return true;
        }

        return false;
    }

    @Override
    public boolean removeManuscriptFromFavoriteFolder(Integer userId, Integer manuscriptId, Integer folderId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null || !folder.getUserId().equals(userId)) {
            return false;
        }

        ManuscriptFavorite existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
        if (existing != null) {
            favoriteVideoMapper.deleteById(existing.getId());
            favoriteFolderMapper.updateCollectCount(folderId, -1);

            List<ManuscriptFavorite> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
            if (remainingFavorites.isEmpty()) {
                LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
                wrapper.eq(UserInteraction::getUserId, userId)
                       .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
                       .eq(UserInteraction::getTargetId, manuscriptId)
                       .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
                userInteractionMapper.delete(wrapper);
                manuscriptMapper.updateCollectCount(manuscriptId, -1);
            }
            return true;
        }

        return false;
    }

    @Override
    public List<VideoVO> getFavoriteFolderVideos(Integer userId, Integer folderId, Integer page, Integer size) {
        if (userId == null) {
            return new ArrayList<>();
        }

        FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
        if (folder == null || !folder.getUserId().equals(userId)) {
            return new ArrayList<>();
        }

        int offset = (page - 1) * size;
        List<Manuscript> manuscripts = favoriteVideoMapper.findManuscriptsByFolderId(folderId, offset, size);

        List<VideoVO> videoVOs = new ArrayList<>();
        for (Manuscript manuscript : manuscripts) {
            VideoVO videoVO = new VideoVO();
            videoVO.setId(manuscript.getId());
            videoVO.setTitle(manuscript.getTitle());
            videoVO.setDescription(manuscript.getDescription());
            videoVO.setCoverUrl(manuscript.getCoverUrl());
            videoVO.setViewCount(manuscript.getViewCount());
            videoVO.setLikeCount(manuscript.getLikeCount());
            videoVO.setCoinCount(manuscript.getCoinCount());
            videoVO.setCollectCount(manuscript.getCollectCount());
            videoVO.setShareCount(manuscript.getShareCount());
            videoVO.setUploadTime(manuscript.getUploadTime());

            User user = userMapper.selectById(manuscript.getUserId());
            if (user != null) {
                VideoVO.UserInfo userInfo = new VideoVO.UserInfo();
                userInfo.setId(user.getId());
                userInfo.setName(user.getNickname() != null ? user.getNickname() : user.getUsername());
                userInfo.setAvatar(user.getAvatar());
                userInfo.setLevel(user.getLevel());
                videoVO.setUploader(userInfo);
            }

            videoVOs.add(videoVO);
        }

        return videoVOs;
    }

    @Override
    public List<FavoriteFolder> getVideoFavoriteFolders(Integer userId, Integer manuscriptId) {
        log.info("getVideoFavoriteFolders - userId: {}, manuscriptId: {}", userId, manuscriptId);
        
        if (userId == null) {
            log.warn("getVideoFavoriteFolders - userId is null");
            return new ArrayList<>();
        }

        if (manuscriptId == null) {
            log.warn("getVideoFavoriteFolders - manuscriptId is null");
            return new ArrayList<>();
        }

        List<FavoriteFolder> allFolders = getFavoriteFolders(userId);
        log.info("getVideoFavoriteFolders - allFolders size: {}, folders: {}", allFolders.size(), allFolders);
        
        List<ManuscriptFavorite> favoriteVideos = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
        log.info("getVideoFavoriteFolders - favoriteVideos size: {}, data: {}", favoriteVideos.size(), favoriteVideos);

        Map<Integer, FavoriteFolder> folderMap = new HashMap<>();
        for (FavoriteFolder folder : allFolders) {
            folderMap.put(folder.getId(), folder);
        }

        List<FavoriteFolder> result = new ArrayList<>();
        for (ManuscriptFavorite favoriteVideo : favoriteVideos) {
            log.info("getVideoFavoriteFolders - processing favoriteVideo: folderId={}, manuscriptId={}", 
                favoriteVideo.getFolderId(), favoriteVideo.getManuscriptId());
            FavoriteFolder folder = folderMap.get(favoriteVideo.getFolderId());
            if (folder != null) {
                result.add(folder);
                log.info("getVideoFavoriteFolders - found folder: {}", folder);
            } else {
                log.warn("getVideoFavoriteFolders - folder not found for folderId: {}", favoriteVideo.getFolderId());
            }
        }
        log.info("getVideoFavoriteFolders - result size: {}", result.size());
        return result;
    }

    @Override
    public boolean updateVideoFavoriteFolders(Integer userId, Integer manuscriptId, List<Integer> folderIds) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }

        if (manuscriptId == null) {
            throw new BusinessException("稿件不存在");
        }

        List<FavoriteFolder> currentFolders = getVideoFavoriteFolders(userId, manuscriptId);
        Set<Integer> currentFolderIds = new HashSet<>();
        for (FavoriteFolder folder : currentFolders) {
            currentFolderIds.add(folder.getId());
        }
        Set<Integer> newFolderIds = new HashSet<>(folderIds);

        boolean success = false;

        for (Integer folderId : currentFolderIds) {
            if (!newFolderIds.contains(folderId)) {
                if (removeManuscriptFromFavoriteFolder(userId, manuscriptId, folderId)) {
                    success = true;
                }
            }
        }

        for (Integer folderId : newFolderIds) {
            if (!currentFolderIds.contains(folderId)) {
                FavoriteFolder folder = favoriteFolderMapper.selectById(folderId);
                if (folder != null && folder.getUserId().equals(userId)) {
                    ManuscriptFavorite existing = favoriteVideoMapper.findByFolderIdAndManuscriptId(folderId, manuscriptId);
                    if (existing == null) {
                        ManuscriptFavorite manuscriptFavorite = new ManuscriptFavorite();
                        manuscriptFavorite.setFolderId(folderId);
                        manuscriptFavorite.setManuscriptId(manuscriptId);
                        favoriteVideoMapper.insert(manuscriptFavorite);
                        favoriteFolderMapper.updateCollectCount(folderId, 1);
                        success = true;
                    }
                }
            }
        }

        List<ManuscriptFavorite> remainingFavorites = favoriteVideoMapper.findByUserIdAndManuscriptId(userId, manuscriptId);
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, TARGET_TYPE_MANUSCRIPT)
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, InteractionType.COLLECT.getCode());
        UserInteraction existingCollection = userInteractionMapper.selectOne(wrapper);

        if (remainingFavorites.isEmpty() && existingCollection != null) {
            userInteractionMapper.delete(wrapper);
            manuscriptMapper.updateCollectCount(manuscriptId, -1);
        } else if (!remainingFavorites.isEmpty() && existingCollection == null) {
            UserInteraction interaction = new UserInteraction();
            interaction.setUserId(userId);
            interaction.setTargetType(TARGET_TYPE_MANUSCRIPT);
            interaction.setTargetId(manuscriptId);
            interaction.setInteractionType(InteractionType.COLLECT.getCode());
            userInteractionMapper.insert(interaction);
            manuscriptMapper.updateCollectCount(manuscriptId, 1);
        }

        if (!success && !currentFolderIds.equals(newFolderIds)) {
            success = true;
        }

        return success;
    }

    private Integer getManuscriptIdByVideoId(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            return null;
        }
        return video.getManuscriptId();
    }

    private VideoVO convertToVideoVO(Video video) {
        VideoVO videoVO = new VideoVO();
        videoVO.setId(video.getId());
        videoVO.setTitle(video.getTitle());
        videoVO.setPlayUrl(video.getPlayUrlHd() != null ? video.getPlayUrlHd() : video.getPlayUrlSd());

        Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
        if (manuscript != null) {
            videoVO.setDescription(manuscript.getDescription());
            videoVO.setCoverUrl(manuscript.getCoverUrl());
            videoVO.setViewCount(manuscript.getViewCount());
            videoVO.setLikeCount(manuscript.getLikeCount());
            videoVO.setCoinCount(manuscript.getCoinCount());
            videoVO.setCollectCount(manuscript.getCollectCount());
            videoVO.setShareCount(manuscript.getShareCount());
            videoVO.setUploadTime(manuscript.getUploadTime());

            User user = userMapper.selectById(manuscript.getUserId());
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
}
