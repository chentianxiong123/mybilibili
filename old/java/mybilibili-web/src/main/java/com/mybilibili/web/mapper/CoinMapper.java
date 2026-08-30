package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Coin;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface CoinMapper {
    int insert(Coin coin);
    int update(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId, @Param("coinCount") Integer coinCount);
    Coin findByUserAndManuscript(@Param("userId") Integer userId, @Param("manuscriptId") Integer manuscriptId);
}
