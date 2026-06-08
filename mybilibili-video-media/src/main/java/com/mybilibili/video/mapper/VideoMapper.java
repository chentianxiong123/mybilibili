package com.mybilibili.video.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.entity.VideoTag;
import org.apache.ibatis.annotations.*;

import java.util.List;
import java.util.Map;

@Mapper
public interface VideoMapper extends BaseMapper<Video> {

    @Select("SELECT * FROM videos WHERE manuscript_id = #{manuscriptId} ORDER BY video_order ASC")
    List<Video> selectByManuscriptId(Integer manuscriptId);

    @Select("<script>" +
            "SELECT * FROM videos WHERE manuscript_id IN " +
            "<foreach collection='manuscriptIds' item='manuscriptId' open='(' separator=',' close=')'>#{manuscriptId}</foreach> " +
            "ORDER BY manuscript_id ASC, video_order ASC" +
            "</script>")
    List<Video> selectByManuscriptIds(@Param("manuscriptIds") List<Integer> manuscriptIds);

    @Select("SELECT * FROM videos ORDER BY id DESC")
    List<Video> selectAll();

    @Select("SELECT * FROM videos WHERE process_status IN (1, 2, 3, 4) LIMIT 1")
    Video selectProcessing();

    @Select("<script>" +
            "SELECT * FROM videos WHERE 1=1 " +
            "<if test='keyword != null'> AND title LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='status != null'> AND status = #{status} </if>" +
            "ORDER BY upload_time DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Video> selectAdminList(@Param("keyword") String keyword, @Param("status") Integer status, @Param("offset") Integer offset, @Param("size") Integer size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM videos WHERE 1=1 " +
            "<if test='keyword != null'> AND title LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='status != null'> AND status = #{status} </if>" +
            "</script>")
    int countAdminList(@Param("keyword") String keyword, @Param("status") Integer status);

    @Select("SELECT * FROM videos WHERE id = #{id}")
    Video selectById(Integer id);

    @Insert("INSERT INTO videos (manuscript_id, video_order, title, play_url_hd, play_url_sd, play_url_ld, duration_seconds, upload_time, process_progress, process_stage, process_status, source_video_url, has_subtitle, has_summary) " +
            "VALUES (#{manuscriptId}, #{videoOrder}, #{title}, #{playUrlHd}, #{playUrlSd}, #{playUrlLd}, #{durationSeconds}, #{uploadTime}, #{processProgress}, #{processStage}, #{processStatus}, #{sourceVideoUrl}, #{hasSubtitle}, #{hasSummary})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(Video video);

    @Update("UPDATE videos SET video_order = #{videoOrder}, title = #{title}, description = #{description}, play_url_hd = #{playUrlHd}, play_url_sd = #{playUrlSd}, play_url_ld = #{playUrlLd}, duration_seconds = #{durationSeconds}, process_progress = #{processProgress}, process_stage = #{processStage}, process_status = #{processStatus}, process_error = #{processError}, source_video_url = #{sourceVideoUrl}, has_subtitle = #{hasSubtitle}, has_summary = #{hasSummary}, updated_at = NOW() WHERE id = #{id}")
    int updateById(Video video);

    @Update("UPDATE videos SET process_status = #{processStatus}, process_progress = #{processProgress}, process_stage = #{processStage}, process_error = #{processError}, updated_at = NOW() WHERE id = #{id}")
    int updateProcessState(@Param("id") Integer id,
                           @Param("processStatus") Integer processStatus,
                           @Param("processProgress") Integer processProgress,
                           @Param("processStage") String processStage,
                           @Param("processError") String processError);

    @Update("UPDATE videos SET process_status = #{processStatus}, process_progress = #{processProgress}, process_stage = #{processStage}, process_error = NULL, play_url_hd = #{playUrlHd}, play_url_sd = #{playUrlSd}, play_url_ld = #{playUrlLd}, updated_at = NOW() WHERE id = #{id}")
    int updateTranscodeResult(@Param("id") Integer id,
                              @Param("processStatus") Integer processStatus,
                              @Param("processProgress") Integer processProgress,
                              @Param("processStage") String processStage,
                              @Param("playUrlHd") String playUrlHd,
                              @Param("playUrlSd") String playUrlSd,
                              @Param("playUrlLd") String playUrlLd);

    @Update("UPDATE videos SET process_status = #{processStatus}, process_error = #{processError}, updated_at = NOW() WHERE id = #{id}")
    int updateProcessError(@Param("id") Integer id,
                           @Param("processStatus") Integer processStatus,
                           @Param("processError") String processError);

    @Update("UPDATE videos SET has_subtitle = #{hasSubtitle}, updated_at = NOW() WHERE id = #{id}")
    int updateHasSubtitle(@Param("id") Integer id, @Param("hasSubtitle") Integer hasSubtitle);

    @Delete("DELETE FROM videos WHERE id = #{id}")
    int deleteById(Integer id);

    @Delete("DELETE FROM videos WHERE manuscript_id = #{manuscriptId}")
    int deleteByManuscriptId(Integer manuscriptId);

    @Select("SELECT v.id FROM videos v JOIN manuscripts m ON v.manuscript_id = m.id WHERE m.user_id = #{userId}")
    List<Integer> selectVideoIdsByUserId(Integer userId);

    @Insert("INSERT INTO video_tags (video_id, tag_id) VALUES (#{videoId}, #{tagId})")
    int insertVideoTag(VideoTag videoTag);

    @Delete("<script>" +
            "DELETE FROM video_tags WHERE video_id IN " +
            "<foreach collection='videoIds' item='videoId' open='(' separator=',' close=')'>#{videoId}</foreach>" +
            "</script>")
    int deleteVideoTagsByVideoIds(@Param("videoIds") List<Integer> videoIds);

    @Select("SELECT t.name FROM tags t JOIN video_tags vt ON t.id = vt.tag_id WHERE vt.video_id = #{videoId}")
    List<Map<String, Object>> selectTagsByVideoId(@Param("videoId") Integer videoId);
}
