package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface VideoMapper {
    int insert(Video video);
    int insertVideoTag(VideoTag videoTag);
    Video selectById(Integer id);
    List<Video> selectByUserId(Integer userId);
    List<Video> selectByUserIdOrderByViewCount(Integer userId);
    List<Video> selectByUserIdOrderByCollectCount(Integer userId);
    List<Video> selectByCategoryId(Integer categoryId);
    List<Video> selectRecommended();
    List<Video> selectHot();
    List<Video> selectList(int offset, int size);
    List<Video> selectAll();
    List<Video> selectByStatus(@Param("status") Integer status);
    List<Video> selectByManuscriptId(@Param("manuscriptId") Integer manuscriptId);
    List<Video> selectRecentlyPublished(@Param("status") Integer status, @Param("minutes") int minutes);
    int updateViewCount(Integer id);
    int updateLikeCount(Integer id, int count);
    int updateCoinCount(Integer id, int count);
    int updateCollectCount(Integer id, int count);
    int updateShareCount(Integer id, int count);
    int updateDanmakuCount(Integer id, int count);
    int update(Video video);
    int delete(Integer id);
    int deleteVideoTags(Integer videoId);
    int countCollections(Integer id);

    /**
     * 查询所有视频ID
     *
     * @return 视频ID列表
     */
    List<Integer> selectAllVideoIds();
}