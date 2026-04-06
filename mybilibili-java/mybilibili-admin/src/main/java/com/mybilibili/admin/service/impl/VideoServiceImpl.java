package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.*;
import com.mybilibili.admin.service.VideoService;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.service.VideoIndexService;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class VideoServiceImpl implements VideoService {

    private static final Logger logger = LoggerFactory.getLogger(VideoServiceImpl.class);

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private DanmakuMapper danmakuMapper;

    @Autowired
    private LikeMapper likeMapper;

    @Autowired
    private CollectionMapper collectionMapper;

    @Autowired
    private CoinMapper coinMapper;

    @Autowired
    private WatchHistoryMapper watchHistoryMapper;

    @Autowired
    private FavoriteVideoMapper favoriteVideoMapper;

    @Autowired
    private VideoTagMapper videoTagMapper;

    @Autowired(required = false)
    private VideoIndexService videoIndexService;

    @Override
    public Result<?> getVideoList(Integer page, Integer size, String keyword, Integer status) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<Video> videos = videoMapper.selectVideosByKeyword(offset, size, keyword, status);
            int total = videoMapper.countVideosByKeyword(keyword, status);

            List<VideoVO> videoVOs = new ArrayList<>();
            for (Video video : videos) {
                VideoVO videoVO = new VideoVO();
                BeanUtils.copyProperties(video, videoVO);
                
                // 添加稿件信息
                if (video.getManuscriptId() != null) {
                    com.mybilibili.common.entity.Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
                    if (manuscript != null) {
                        videoVO.setManuscriptTitle(manuscript.getTitle());
                    }
                }
                
                videoVOs.add(videoVO);
            }

            Map<String, Object> data = new HashMap<>();
            data.put("list", videoVOs);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取视频列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getVideoById(Integer id) {
        try {
            Video video = videoMapper.selectById(id);
            if (video == null) {
                return Result.error("视频不存在");
            }
            VideoVO videoVO = new VideoVO();
            BeanUtils.copyProperties(video, videoVO);
            return Result.success("获取视频详情成功", videoVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> deleteVideo(Integer id) {
        logger.info("【删除视频】开始删除视频，ID: {}", id);
        try {
            Video video = videoMapper.selectById(id);
            if (video == null) {
                logger.warn("【删除视频】视频不存在，ID: {}", id);
                return Result.error("视频不存在");
            }
            logger.info("【删除视频】找到视频，标题: {}", video.getTitle());

            // 删除视频索引（如果视频已上架）
            if (videoIndexService != null && video.getStatus() != null && video.getStatus() == Video.STATUS_PUBLISHED) {
                logger.info("【删除视频】删除视频索引，视频ID: {}", id);
                videoIndexService.deleteVideo(id);
            }

            // 先删除所有关联数据
            logger.info("【删除视频】开始删除关联数据，视频ID: {}", id);
            deleteRelatedData(id);
            logger.info("【删除视频】关联数据删除完成，视频ID: {}", id);

            // 再删除视频
            logger.info("【删除视频】开始删除视频记录，ID: {}", id);
            int result = videoMapper.delete(id);
            logger.info("【删除视频】删除视频记录结果，ID: {}, 影响行数: {}", id, result);

            if (result > 0) {
                logger.info("【删除视频】删除视频成功，ID: {}", id);
                return Result.success("删除视频成功", null);
            } else {
                logger.warn("【删除视频】删除视频失败，ID: {}", id);
                return Result.error("删除视频失败");
            }
        } catch (Exception e) {
            logger.error("【删除视频】删除视频异常，ID: {}, 错误: {}", id, e.getMessage(), e);
            return Result.error("删除视频失败: " + e.getMessage());
        }
    }

    @Override
    @Transactional
    public Result<?> deleteVideos(List<Integer> ids) {
        logger.info("【批量删除视频】开始批量删除视频，ID列表: {}", ids);
        try {
            if (ids == null || ids.isEmpty()) {
                logger.warn("【批量删除视频】视频ID列表为空");
                return Result.error("视频ID列表不能为空");
            }

            int successCount = 0;
            int failCount = 0;
            List<String> failMessages = new ArrayList<>();

            for (Integer id : ids) {
                logger.info("【批量删除视频】处理视频ID: {}", id);
                try {
                    Video video = videoMapper.selectById(id);
                    if (video == null) {
                        logger.warn("【批量删除视频】视频不存在，ID: {}", id);
                        failCount++;
                        failMessages.add("视频ID " + id + ": 视频不存在");
                        continue;
                    }

                    // 删除视频索引（如果视频已上架）
                    if (videoIndexService != null && video.getStatus() != null && video.getStatus() == Video.STATUS_PUBLISHED) {
                        logger.info("【批量删除视频】删除视频索引，视频ID: {}", id);
                        videoIndexService.deleteVideo(id);
                    }

                    // 先删除所有关联数据
                    logger.info("【批量删除视频】开始删除关联数据，视频ID: {}", id);
                    deleteRelatedData(id);
                    logger.info("【批量删除视频】关联数据删除完成，视频ID: {}", id);

                    // 再删除视频
                    logger.info("【批量删除视频】开始删除视频记录，ID: {}", id);
                    int result = videoMapper.delete(id);
                    logger.info("【批量删除视频】删除视频记录结果，ID: {}, 影响行数: {}", id, result);

                    if (result > 0) {
                        successCount++;
                        logger.info("【批量删除视频】删除视频成功，ID: {}", id);
                    } else {
                        failCount++;
                        failMessages.add("视频ID " + id + ": 删除失败");
                        logger.warn("【批量删除视频】删除视频失败，ID: {}", id);
                    }
                } catch (Exception e) {
                    failCount++;
                    failMessages.add("视频ID " + id + ": " + e.getMessage());
                    logger.error("【批量删除视频】删除视频异常，ID: {}, 错误: {}", id, e.getMessage(), e);
                }
            }

            Map<String, Object> data = new HashMap<>();
            data.put("successCount", successCount);
            data.put("failCount", failCount);
            data.put("failMessages", failMessages);

            logger.info("【批量删除视频】批量删除完成，成功: {}, 失败: {}, 失败详情: {}", successCount, failCount, failMessages);

            if (successCount > 0 && failCount == 0) {
                return Result.success("批量删除成功，共删除 " + successCount + " 个视频", data);
            } else if (successCount > 0) {
                return Result.success("批量删除部分成功，成功 " + successCount + " 个，失败 " + failCount + " 个", data);
            } else {
                return Result.error("批量删除失败，全部 " + failCount + " 个视频删除失败: " + String.join(", ", failMessages));
            }
        } catch (Exception e) {
            logger.error("【批量删除视频】批量删除异常，错误: {}", e.getMessage(), e);
            return Result.error(e.getMessage());
        }
    }

    /**
     * 删除视频的所有关联数据
     */
    private void deleteRelatedData(Integer videoId) {
        logger.info("【删除关联数据】开始删除视频关联数据，视频ID: {}", videoId);

        // 删除评论
        int commentCount = commentMapper.countByVideoId(videoId);
        if (commentCount > 0) {
            logger.info("【删除关联数据】删除评论，视频ID: {}, 数量: {}", videoId, commentCount);
            int deleted = commentMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】评论删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无评论需要删除，视频ID: {}", videoId);
        }

        // 删除弹幕
        int danmakuCount = danmakuMapper.countByVideoId(videoId);
        if (danmakuCount > 0) {
            logger.info("【删除关联数据】删除弹幕，视频ID: {}, 数量: {}", videoId, danmakuCount);
            int deleted = danmakuMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】弹幕删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无弹幕需要删除，视频ID: {}", videoId);
        }

        // 删除点赞记录
        int likeCount = likeMapper.countByVideoId(videoId);
        if (likeCount > 0) {
            logger.info("【删除关联数据】删除点赞，视频ID: {}, 数量: {}", videoId, likeCount);
            int deleted = likeMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】点赞删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无点赞需要删除，视频ID: {}", videoId);
        }

        // 删除收藏记录
        int collectionCount = collectionMapper.countByVideoId(videoId);
        if (collectionCount > 0) {
            logger.info("【删除关联数据】删除收藏，视频ID: {}, 数量: {}", videoId, collectionCount);
            int deleted = collectionMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】收藏删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无收藏需要删除，视频ID: {}", videoId);
        }

        // 删除投币记录
        int coinCount = coinMapper.countByVideoId(videoId);
        if (coinCount > 0) {
            logger.info("【删除关联数据】删除投币，视频ID: {}, 数量: {}", videoId, coinCount);
            int deleted = coinMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】投币删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无投币需要删除，视频ID: {}", videoId);
        }

        // 删除观看历史
        int watchHistoryCount = watchHistoryMapper.countByVideoId(videoId);
        if (watchHistoryCount > 0) {
            logger.info("【删除关联数据】删除观看历史，视频ID: {}, 数量: {}", videoId, watchHistoryCount);
            int deleted = watchHistoryMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】观看历史删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无观看历史需要删除，视频ID: {}", videoId);
        }

        // 删除收藏夹关联
        int favoriteVideoCount = favoriteVideoMapper.countByVideoId(videoId);
        if (favoriteVideoCount > 0) {
            logger.info("【删除关联数据】删除收藏夹关联，视频ID: {}, 数量: {}", videoId, favoriteVideoCount);
            int deleted = favoriteVideoMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】收藏夹关联删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无收藏夹关联需要删除，视频ID: {}", videoId);
        }

        // 删除视频标签关联
        int videoTagCount = videoTagMapper.countByVideoId(videoId);
        if (videoTagCount > 0) {
            logger.info("【删除关联数据】删除标签关联，视频ID: {}, 数量: {}", videoId, videoTagCount);
            int deleted = videoTagMapper.deleteByVideoId(videoId);
            logger.info("【删除关联数据】标签关联删除完成，视频ID: {}, 删除行数: {}", videoId, deleted);
        } else {
            logger.info("【删除关联数据】无标签关联需要删除，视频ID: {}", videoId);
        }

        logger.info("【删除关联数据】关联数据删除完成，视频ID: {}", videoId);
    }
}
