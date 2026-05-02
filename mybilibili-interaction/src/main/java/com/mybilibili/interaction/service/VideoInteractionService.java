package com.mybilibili.interaction.service;

import com.mybilibili.common.entity.FavoriteFolder;
import com.mybilibili.common.vo.VideoVO;

import java.util.List;
import java.util.Map;

public interface VideoInteractionService {

    boolean likeVideo(Integer userId, Integer manuscriptId);

    boolean unlikeVideo(Integer userId, Integer manuscriptId);

    boolean coinVideo(Integer userId, Integer manuscriptId, Integer coinCount);

    boolean collectVideo(Integer userId, Integer manuscriptId);

    boolean uncollectVideo(Integer userId, Integer manuscriptId);

    void shareVideo(Integer userId, Integer manuscriptId, String channel, String ipAddress);

    Map<String, Object> getShareStatistics(Integer manuscriptId);

    VideoInteractionStatus getInteractionStatus(Integer userId, Integer manuscriptId);

    List<VideoVO> getLikedVideos(Integer userId);

    List<VideoVO> getCollectedVideos(Integer userId);

    List<FavoriteFolder> getFavoriteFolders(Integer userId);

    FavoriteFolder createFavoriteFolder(Integer userId, String name);

    FavoriteFolder updateFavoriteFolder(Integer userId, Integer folderId, String name);

    boolean deleteFavoriteFolder(Integer userId, Integer folderId);

    boolean addVideoToFavoriteFolders(Integer userId, Integer manuscriptId, List<Integer> folderIds);

    boolean removeVideoFromFavoriteFolder(Integer userId, Integer manuscriptId, Integer folderId);

    boolean removeManuscriptFromFavoriteFolder(Integer userId, Integer manuscriptId, Integer folderId);

    List<VideoVO> getFavoriteFolderVideos(Integer userId, Integer folderId, Integer page, Integer size);

    List<FavoriteFolder> getVideoFavoriteFolders(Integer userId, Integer manuscriptId);

    boolean updateVideoFavoriteFolders(Integer userId, Integer manuscriptId, List<Integer> folderIds);

    class VideoInteractionStatus {
        private boolean isLiked;
        private boolean isCollected;
        private boolean isShared;
        private Integer coinCount;

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

        public boolean isShared() {
            return isShared;
        }

        public void setShared(boolean shared) {
            isShared = shared;
        }

        public Integer getCoinCount() {
            return coinCount;
        }

        public void setCoinCount(Integer coinCount) {
            this.coinCount = coinCount;
        }
    }
}
