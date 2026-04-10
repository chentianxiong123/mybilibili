package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.ManuscriptFavorite;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FavoriteVideoMapper extends BaseMapper<ManuscriptFavorite> {
    int insert(ManuscriptFavorite manuscriptFavorite);
    int deleteById(Integer id);
    List<ManuscriptFavorite> findByFolderId(Integer folderId);
    ManuscriptFavorite findByFolderIdAndManuscriptId(@Param("folderId") Integer folderId, @Param("manuscriptId") Integer manuscriptId);
    List<ManuscriptFavorite> findByUserIdAndManuscriptId(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
    List<Manuscript> findManuscriptsByFolderId(@Param("folderId") Integer folderId, @Param("offset") Integer offset, @Param("size") Integer size);
}
