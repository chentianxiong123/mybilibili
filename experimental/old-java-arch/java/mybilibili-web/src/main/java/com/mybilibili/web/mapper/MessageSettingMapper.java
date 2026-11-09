package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.MessageSetting;
import com.mybilibili.common.vo.MessageSettingVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface MessageSettingMapper {

    int insert(MessageSetting messageSetting);

    MessageSetting selectById(Integer id);

    MessageSetting selectByUserId(Integer userId);

    MessageSettingVO selectVOByUserId(Integer userId);

    int update(MessageSetting messageSetting);

    int deleteById(Integer id);

    int deleteByUserId(Integer userId);
}
