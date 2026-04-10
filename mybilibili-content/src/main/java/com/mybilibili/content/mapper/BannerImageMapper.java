package com.mybilibili.content.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.BannerImage;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface BannerImageMapper extends BaseMapper<BannerImage> {

    @Select("SELECT * FROM banner_images WHERE type = #{type} AND status = 1 AND (start_time IS NULL OR start_time <= NOW()) AND (end_time IS NULL OR end_time >= NOW()) ORDER BY sort_order ASC")
    List<BannerImage> selectByType(@Param("type") Integer type);

    @Select("SELECT * FROM banner_images WHERE type = #{type} AND category_id = #{categoryId} AND status = 1 AND (start_time IS NULL OR start_time <= NOW()) AND (end_time IS NULL OR end_time >= NOW()) ORDER BY sort_order ASC")
    List<BannerImage> selectByTypeAndCategory(@Param("type") Integer type, @Param("categoryId") Integer categoryId);
}
