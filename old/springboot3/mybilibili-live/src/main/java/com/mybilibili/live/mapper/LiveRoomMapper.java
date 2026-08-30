package com.mybilibili.live.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.live.entity.LiveRoom;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

public interface LiveRoomMapper extends BaseMapper<LiveRoom> {

    @Select("SELECT * FROM live_rooms WHERE user_id = #{userId}")
    LiveRoom selectByUserId(@Param("userId") Integer userId);

    @Select("SELECT * FROM live_rooms WHERE stream_key = #{streamKey}")
    LiveRoom selectByStreamKey(@Param("streamKey") String streamKey);

    @Select("SELECT * FROM live_rooms WHERE status = 'live' ORDER BY viewer_count DESC")
    List<LiveRoom> selectLiveList();
}