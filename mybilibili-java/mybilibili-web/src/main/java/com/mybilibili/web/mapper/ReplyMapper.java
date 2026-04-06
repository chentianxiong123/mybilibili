package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Reply;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface ReplyMapper {
    int insert(Reply reply);
    Reply selectById(Integer id);
    List<Reply> selectByCommentId(Integer commentId, int offset, int size);
    int updateLikeCount(Integer id, int count);
    int updateLikeCountDirect(@Param("id") Integer id, @Param("count") int count);
    int delete(Integer id);
    int countByCommentId(Integer commentId);

    // 根据稿件ID获取所有回复数量
    int countByManuscriptId(Integer manuscriptId);
}