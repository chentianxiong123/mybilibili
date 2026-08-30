package com.mybilibili.video.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.video.entity.ManuscriptEditVersion;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

@Mapper
public interface ManuscriptEditVersionMapper extends BaseMapper<ManuscriptEditVersion> {

    @Select("SELECT * FROM manuscript_edit_versions WHERE manuscript_id = #{manuscriptId} ORDER BY created_at DESC LIMIT 1")
    ManuscriptEditVersion selectLatestByManuscriptId(@Param("manuscriptId") Integer manuscriptId);

    @Update("UPDATE manuscript_edit_versions SET status = #{status}, reviewer_id = #{reviewerId}, review_reason = #{reviewReason}, reviewed_at = NOW(), updated_at = NOW() WHERE id = #{id}")
    int updateReviewResult(@Param("id") Long id,
                           @Param("status") String status,
                           @Param("reviewerId") Integer reviewerId,
                           @Param("reviewReason") String reviewReason);
}
