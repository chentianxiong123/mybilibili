package com.mybilibili.message.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.MessageSetting;
import com.mybilibili.common.vo.MessageSettingVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface MessageSettingMapper extends BaseMapper<MessageSetting> {

    MessageSetting selectById(@Param("id") Long id);

    MessageSetting selectByUserId(@Param("userId") Integer userId);

    MessageSettingVO selectVOByUserId(@Param("userId") Integer userId);

    int update(MessageSetting messageSetting);
}