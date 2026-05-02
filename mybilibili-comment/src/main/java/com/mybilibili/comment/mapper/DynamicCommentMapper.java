package com.mybilibili.comment.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.DynamicComment;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Update;

import java.util.List;

@Mapper
public interface DynamicCommentMapper extends BaseMapper<DynamicComment> {

    List<DynamicComment> selectByDynamicId(Integer dynamicId, int offset, int size);

    List<DynamicComment> selectRepliesByParentId(Integer parentId);

    int countByDynamicId(Integer dynamicId);

    List<DynamicComment> selectAllDynamicComments(@Param("status") Integer status, @Param("offset") int offset, @Param("size") int size);

    int countAllDynamicComments(@Param("status") Integer status);

    @Update("UPDATE dynamic_comments SET like_count = #{count} WHERE id = #{id}")
    int updateLikeCountDirect(@Param("id") Integer id, @Param("count") Integer count);
}