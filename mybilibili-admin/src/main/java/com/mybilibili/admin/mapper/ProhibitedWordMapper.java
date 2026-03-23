package com.mybilibili.admin.mapper;

import com.mybilibili.common.entity.ProhibitedWord;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface ProhibitedWordMapper {

    List<ProhibitedWord> selectByKeyword(@Param("offset") Integer offset, @Param("size") Integer size, @Param("keyword") String keyword);

    int countByKeyword(@Param("keyword") String keyword);

    ProhibitedWord selectById(@Param("id") Integer id);

    ProhibitedWord selectByWord(@Param("word") String word);

    int insert(ProhibitedWord word);

    int update(ProhibitedWord word);

    int deleteById(@Param("id") Integer id);

    List<ProhibitedWord> selectAllEnabled();
}
