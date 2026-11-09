package com.mybilibili.video.service;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.mapper.VideoMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class VideoService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private UserClient userClient;

    public Result<VideoVO> getById(Integer id) {
        Video video = videoMapper.selectById(id);
        if (video == null) {
            return Result.error(404, "视频不存在");
        }
        VideoVO vo = convertToVideoVO(video);
        return Result.success(vo);
    }

    public Result<VideoVO> getVideoById(Integer id) {
        return getById(id);
    }

    public Result<List<Video>> getVideosByManuscriptId(Integer manuscriptId) {
        List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
        return Result.success(videos);
    }

    public Result<Map<String, Object>> getAdminVideoList(Integer page, Integer size, String keyword, Integer status) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;

        List<Video> videos = videoMapper.selectAdminList(keyword, status, offset, size);
        Integer total = videoMapper.countAdminList(keyword, status);

        Map<String, Object> data = new HashMap<>();
        data.put("list", videos);
        data.put("total", total);
        data.put("page", page);
        data.put("size", size);

        return Result.success(data);
    }

    public Result<?> deleteVideo(Integer id) {
        videoMapper.deleteById(id);
        return Result.success("删除成功", null);
    }

    public Result<?> deleteVideos(List<Integer> ids) {
        for (Integer id : ids) {
            videoMapper.deleteById(id);
        }
        return Result.success("批量删除成功", null);
    }

    public Result<List<Integer>> getManuscriptIdsByUserId(Integer userId) {
        List<Integer> ids = manuscriptMapper.selectIdsByUserId(userId);
        return Result.success(ids);
    }

    public Result<List<Integer>> getVideoIdsByUserId(Integer userId) {
        List<Integer> ids = videoMapper.selectVideoIdsByUserId(userId);
        return Result.success(ids);
    }

    private VideoVO convertToVideoVO(Video video) {
        VideoVO vo = new VideoVO();
        vo.setId(video.getId());
        vo.setTitle(video.getTitle());
        vo.setManuscriptId(video.getManuscriptId());
        vo.setPlayUrlHd(video.getPlayUrlHd());
        vo.setPlayUrlSd(video.getPlayUrlSd());
        vo.setPlayUrlLd(video.getPlayUrlLd());
        if (video.getDurationSeconds() != null) {
            vo.setDuration(String.valueOf(video.getDurationSeconds()));
        }
        vo.setUploadTime(video.getUploadTime());
        
        // 从稿件获取封面 URL
        if (video.getManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
            if (manuscript != null && manuscript.getCoverUrl() != null) {
                vo.setCoverUrl(manuscript.getCoverUrl());
            }
        }
        
        // 从稿件获取上传者信息
        if (video.getManuscriptId() != null) {
            Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
            if (manuscript != null && manuscript.getUserId() != null) {
                try {
                    Result<UserVO> userResult = userClient.getUserById(manuscript.getUserId());
                    if (userResult != null && userResult.getCode() == 200 && userResult.getData() != null) {
                        UserVO user = userResult.getData();
                        VideoVO.UserInfo uploader = new VideoVO.UserInfo();
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
                    // 忽略异常，继续返回视频信息
                }
            }
        }
        
        return vo;
    }
}