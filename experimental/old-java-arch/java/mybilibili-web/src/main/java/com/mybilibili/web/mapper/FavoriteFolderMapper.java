package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.FavoriteFolder;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FavoriteFolderMapper {
    // 根据用户ID获取收藏夹列表
    List<FavoriteFolder> findByUserId(Integer userId);
    
    // 根据用户ID和名称查找收藏夹
    FavoriteFolder findByUserIdAndName(@Param("userId") Integer userId, @Param("name") String name);
    
    // 更新收藏夹视频数量
    void updateVideoCount(@Param("folderId") Integer folderId, @Param("increment") Integer increment);
    
    // 插入收藏夹
    void insert(FavoriteFolder folder);
    
    // 根据ID查询收藏夹
    FavoriteFolder selectById(Integer id);

    // 更新收藏夹
    void update(FavoriteFolder folder);

    // 根据ID删除收藏夹
    void deleteById(Integer id);
}