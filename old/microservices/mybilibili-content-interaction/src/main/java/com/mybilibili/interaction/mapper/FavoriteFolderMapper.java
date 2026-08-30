package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.FavoriteFolder;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FavoriteFolderMapper extends BaseMapper<FavoriteFolder> {
    int insert(FavoriteFolder folder);
    int update(FavoriteFolder folder);
    int updateCollectCount(@Param("folderId") Integer folderId, @Param("increment") Integer increment);
    int deleteById(Integer id);
    FavoriteFolder selectById(Integer id);
    List<FavoriteFolder> findByUserId(Integer userId);
    FavoriteFolder findByUserIdAndName(@Param("userId") Integer userId, @Param("name") String name);
}
