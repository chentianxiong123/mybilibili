package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.DynamicShare;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface DynamicShareMapper {
    
    /**
     * 插入转发记录
     */
    int insert(DynamicShare share);
    
    /**
     * 根据动态ID查询转发列表
     */
    List<DynamicShare> findByDynamicId(@Param("dynamicId") Integer dynamicId);
    
    /**
     * 根据用户ID查询转发列表
     */
    List<DynamicShare> findByUserId(@Param("userId") Integer userId);
    
    /**
     * 检查用户是否已转发该动态
     */
    DynamicShare findByDynamicAndUser(@Param("dynamicId") Integer dynamicId, @Param("userId") Integer userId);
    
    /**
     * 删除转发记录
     */
    int deleteById(@Param("id") Integer id);
    
    /**
     * 统计动态的转发数
     */
    int countByDynamicId(@Param("dynamicId") Integer dynamicId);
}
