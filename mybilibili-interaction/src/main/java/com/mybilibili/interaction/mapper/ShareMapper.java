package com.mybilibili.interaction.mapper;

import com.mybilibili.common.entity.Share;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface ShareMapper {
    @Insert("INSERT INTO shares (user_id, manuscript_id, channel, ip_address, create_time) VALUES (#{userId}, #{manuscriptId}, #{channel}, #{ipAddress}, NOW())")
    int insert(Share share);

    @Select("SELECT COUNT(*) FROM shares WHERE manuscript_id = #{manuscriptId}")
    Integer countByManuscriptId(@Param("manuscriptId") Integer manuscriptId);

    @Select("SELECT COUNT(*) FROM shares WHERE manuscript_id = #{manuscriptId} AND channel = #{channel}")
    Integer countByManuscriptIdAndChannel(@Param("manuscriptId") Integer manuscriptId, @Param("channel") String channel);
}
