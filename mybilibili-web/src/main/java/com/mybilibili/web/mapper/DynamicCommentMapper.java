package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.DynamicComment;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface DynamicCommentMapper {
    int insert(DynamicComment comment);
    
    int delete(@Param("id") Integer id);
    
    DynamicComment getById(@Param("id") Integer id);
    
    List<DynamicComment> getByDynamicId(@Param("dynamicId") Integer dynamicId, @Param("offset") Integer offset, @Param("limit") Integer limit);
    
    int countByDynamicId(@Param("dynamicId") Integer dynamicId);
    
    int updateLikeCount(@Param("id") Integer id, @Param("count") Integer count);
    
    int updateStatus(@Param("id") Integer id, @Param("status") Integer status);
    
    List<DynamicComment> getRepliesByParentId(@Param("parentId") Integer parentId);
}
