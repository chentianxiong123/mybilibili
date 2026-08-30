package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Manuscript;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

@Mapper
public interface ManuscriptMapper extends BaseMapper<Manuscript> {

    @Select("SELECT id, title, description, cover_url, user_id, category_id, duration, view_count, like_count, coin_count, collect_count, share_count, comment_count, danmaku_count, status, review_status, review_reason, review_time, reviewer_id, upload_time, updated_at FROM manuscripts WHERE id = #{id}")
    Manuscript selectBasicFieldsById(@Param("id") Integer id);

    @Update("UPDATE manuscripts SET like_count = like_count + #{count} WHERE id = #{id}")
    int updateLikeCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET coin_count = coin_count + #{count} WHERE id = #{id}")
    int updateCoinCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET collect_count = collect_count + #{count} WHERE id = #{id}")
    int updateCollectCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE manuscripts SET share_count = share_count + #{count} WHERE id = #{id}")
    int updateShareCount(@Param("id") Integer id, @Param("count") Integer count);
}
