package com.mybilibili.comment.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.ProhibitedWord;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface ProhibitedWordMapper extends BaseMapper<ProhibitedWord> {

    @Select("<script>" +
            "SELECT * FROM prohibited_words WHERE 1=1 " +
            "<if test='keyword != null'> AND word LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='category != null'> AND category = #{category} </if>" +
            "<if test='isEnabled != null'> AND is_enabled = #{isEnabled} </if>" +
            "ORDER BY created_at DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<ProhibitedWord> selectList(@Param("keyword") String keyword,
                                    @Param("category") String category,
                                    @Param("isEnabled") Integer isEnabled,
                                    @Param("offset") int offset,
                                    @Param("size") int size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM prohibited_words WHERE 1=1 " +
            "<if test='keyword != null'> AND word LIKE CONCAT('%', #{keyword}, '%') </if>" +
            "<if test='category != null'> AND category = #{category} </if>" +
            "<if test='isEnabled != null'> AND is_enabled = #{isEnabled} </if>" +
            "</script>")
    int countList(@Param("keyword") String keyword,
                  @Param("category") String category,
                  @Param("isEnabled") Integer isEnabled);

    @Select("SELECT * FROM prohibited_words WHERE word = #{word} LIMIT 1")
    ProhibitedWord selectByWord(@Param("word") String word);

    @Select("SELECT * FROM prohibited_words WHERE is_enabled = 1")
    List<ProhibitedWord> selectAllEnabled();
}