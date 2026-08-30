package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.CreatorSettings;
import com.mybilibili.common.vo.CreatorSettingsVO;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface CreatorSettingsMapper {

    int insert(CreatorSettings creatorSettings);

    CreatorSettings selectById(Integer id);

    CreatorSettings selectByUserId(Integer userId);

    CreatorSettingsVO selectVOByUserId(Integer userId);

    int update(CreatorSettings creatorSettings);

    int deleteById(Integer id);

    int deleteByUserId(Integer userId);
}
