package com.mybilibili.web.service;

import com.mybilibili.common.entity.FavoriteFolder;
import com.mybilibili.common.vo.VideoVO;

import java.util.List;
import java.util.Map;

public interface VideoInteractionService {
    // 点赞视频
    boolean likeVideo(Integer userId, Integer videoId);
    
    // 取消点赞
    boolean unlikeVideo(Integer userId, Integer videoId);
    
    // 投币
    boolean coinVideo(Integer userId, Integer videoId, Integer coinCount);
    
    // 收藏视频
    boolean collectVideo(Integer userId, Integer videoId);
    
    // 取消收藏
    boolean uncollectVideo(Integer userId, Integer videoId);
    
    // 分享视频
    void shareVideo(Integer userId, Integer videoId, String channel, String ipAddress);
    
    // 发送弹幕
    void sendDanmaku(Integer userId, Integer videoId, String content, String time, String color, Integer mode);
    
    // 获取视频弹幕
    List<?> getDanmakus(Integer videoId);
    
    // 获取用户对视频的互动状态
    VideoInteractionStatus getInteractionStatus(Integer userId, Integer videoId);
    
    // 获取用户点赞的视频
    List<VideoVO> getLikedVideos(Integer userId);
    
    // 获取用户收藏的视频
    List<VideoVO> getCollectedVideos(Integer userId);
    
    // 获取视频分享统计
    Map<String, Object> getShareStatistics(Integer videoId);
    
    // 获取用户收藏夹列表
    List<FavoriteFolder> getFavoriteFolders(Integer userId);
    
    // 创建收藏夹
    FavoriteFolder createFavoriteFolder(Integer userId, String name);

    // 更新收藏夹
    FavoriteFolder updateFavoriteFolder(Integer userId, Integer folderId, String name);

    // 删除收藏夹
    boolean deleteFavoriteFolder(Integer userId, Integer folderId);

    // 添加视频到收藏夹
    boolean addVideoToFavoriteFolders(Integer userId, Integer videoId, List<Integer> folderIds);
    
    // 从收藏夹移除视频
    boolean removeVideoFromFavoriteFolder(Integer userId, Integer videoId, Integer folderId);
    
    // 获取收藏夹视频列表
    List<VideoVO> getFavoriteFolderVideos(Integer userId, Integer folderId, Integer page, Integer size);
    
    // 获取视频在哪些收藏夹中
    List<FavoriteFolder> getVideoFavoriteFolders(Integer userId, Integer videoId);
    
    // 更新视频的收藏夹
    boolean updateVideoFavoriteFolders(Integer userId, Integer videoId, List<Integer> folderIds);
    
    // 互动状态类
    class VideoInteractionStatus {
        private boolean isLiked;
        private boolean isCollected;
        private Integer coinCount;
        
        // getters and setters
        public boolean isLiked() {
            return isLiked;
        }
        
        public void setLiked(boolean liked) {
            isLiked = liked;
        }
        
        public boolean isCollected() {
            return isCollected;
        }
        
        public void setCollected(boolean collected) {
            isCollected = collected;
        }
        
        public Integer getCoinCount() {
            return coinCount;
        }
        
        public void setCoinCount(Integer coinCount) {
            this.coinCount = coinCount;
        }
    }
}
