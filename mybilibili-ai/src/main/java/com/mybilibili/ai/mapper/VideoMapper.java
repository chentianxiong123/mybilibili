package com.mybilibili.ai.mapper;

import com.mybilibili.common.entity.Video;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface VideoMapper {

    @Select("SELECT * FROM videos WHERE id = #{id}")
    Video selectById(Integer id);

    @Select("SELECT * FROM videos WHERE manuscript_id = #{manuscriptId} ORDER BY video_order ASC")
    List<Video> selectByManuscriptId(Integer manuscriptId);

    @Select("SELECT * FROM videos ORDER BY id DESC")
    List<Video> selectAll();

    @Select("SELECT * FROM videos WHERE process_status IN (1, 2, 3, 4) LIMIT 1")
    Video selectProcessing();

    @Select("SELECT * FROM videos WHERE process_status = #{processStatus}")
    List<Video> selectByProcessStatus(@Param("processStatus") Integer processStatus);

    @Update("UPDATE videos SET process_status = #{processStatus} WHERE id = #{id}")
    int updateProcessStatus(@Param("id") Integer id, @Param("processStatus") Integer processStatus);

    @Update("UPDATE videos SET process_status = #{processStatus}, play_url_hd = #{playUrlHd}, play_url_sd = #{playUrlSd}, play_url_ld = #{playUrlLd} WHERE id = #{id}")
    int updateTranscodeResult(@Param("id") Integer id,
                              @Param("processStatus") Integer processStatus,
                              @Param("playUrlHd") String playUrlHd,
                              @Param("playUrlSd") String playUrlSd,
                              @Param("playUrlLd") String playUrlLd);

    @Update("UPDATE videos SET has_summary = #{hasSummary} WHERE id = #{id}")
    int updateHasSummary(@Param("id") Integer id, @Param("hasSummary") Integer hasSummary);

    @Update("UPDATE videos SET has_subtitle = #{hasSubtitle} WHERE id = #{id}")
    int updateHasSubtitle(@Param("id") Integer id, @Param("hasSubtitle") Integer hasSubtitle);

    @Update("UPDATE videos SET process_status = #{processStatus}, process_error = #{processError} WHERE id = #{id}")
    int updateProcessError(@Param("id") Integer id, @Param("processStatus") Integer processStatus, @Param("processError") String processError);
}
