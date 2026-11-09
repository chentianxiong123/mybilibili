package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Notification;
import com.mybilibili.common.entity.NotificationSetting;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface NotificationMapper {
    // 插入通知
    int insert(Notification notification);

    // 根据用户ID查询通知列表
    List<Notification> selectByUserId(@Param("userId") Integer userId, @Param("offset") int offset, @Param("limit") int limit);

    // 根据用户ID查询未读通知数量
    int countUnreadByUserId(@Param("userId") Integer userId);

    // 标记通知为已读
    int updateReadStatus(@Param("id") Integer id, @Param("userId") Integer userId);

    // 批量标记通知为已读
    int batchUpdateReadStatus(@Param("ids") List<Integer> ids, @Param("userId") Integer userId);

    // 删除通知
    int delete(@Param("id") Integer id, @Param("userId") Integer userId);

    // 插入或更新通知设置
    int insertOrUpdateSetting(NotificationSetting setting);

    // 根据用户ID查询通知设置
    NotificationSetting selectSettingByUserId(@Param("userId") Integer userId);
}
