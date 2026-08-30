package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Share;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface ShareMapper {
    @Insert("INSERT INTO shares (user_id, manuscript_id, channel, ip_address, create_time) VALUES (#{userId}, #{manuscriptId}, #{channel}, #{ipAddress}, NOW())")
    int insert(Share share);

    @Select("SELECT * FROM shares WHERE manuscript_id = #{manuscriptId}")
    List<Share> findByManuscriptId(Integer manuscriptId);

    @Select("SELECT COUNT(*) FROM shares WHERE manuscript_id = #{manuscriptId}")
    Integer countByManuscriptId(Integer manuscriptId);

    @Select("SELECT COUNT(*) FROM shares WHERE manuscript_id = #{manuscriptId} AND channel = #{channel}")
    Integer countByManuscriptIdAndChannel(Integer manuscriptId, String channel);
}
