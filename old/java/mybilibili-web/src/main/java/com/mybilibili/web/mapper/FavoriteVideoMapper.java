package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.FavoriteVideo;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FavoriteVideoMapper {
    // 根据收藏夹ID获取视频列表
    List<FavoriteVideo> findByFolderId(Integer folderId);

    // 根据收藏夹ID和稿件ID查找
    FavoriteVideo findByFolderIdAndManuscriptId(@Param("folderId") Integer folderId, @Param("manuscriptId") Integer manuscriptId);

    // 根据稿件ID获取所有收藏关系
    List<FavoriteVideo> findByManuscriptId(Integer manuscriptId);

    // 根据用户ID和稿件ID获取收藏关系
    List<FavoriteVideo> findByUserIdAndManuscriptId(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);

    // 插入收藏视频关系
    void insert(FavoriteVideo favoriteVideo);

    // 根据ID删除收藏视频关系
    void deleteById(Integer id);

    // 根据收藏夹ID获取稿件列表（带分页）
    List<Manuscript> findManuscriptsByFolderId(@Param("folderId") Integer folderId, @Param("offset") Integer offset, @Param("size") Integer size);
}
