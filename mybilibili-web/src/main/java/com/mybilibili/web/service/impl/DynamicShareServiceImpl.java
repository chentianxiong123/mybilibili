package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.DynamicShare;
import com.mybilibili.common.entity.UserDynamic;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.mapper.DynamicShareMapper;
import com.mybilibili.web.mapper.UserDynamicMapper;
import com.mybilibili.web.service.DynamicShareService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class DynamicShareServiceImpl implements DynamicShareService {

    @Autowired
    private DynamicShareMapper dynamicShareMapper;

    @Autowired
    private UserDynamicMapper userDynamicMapper;

    @Override
    @Transactional
    public Result<?> shareDynamic(Integer userId, Integer dynamicId, String content) {
        try {
            // 检查原动态是否存在
            UserDynamic originalDynamic = userDynamicMapper.getById(dynamicId);
            if (originalDynamic == null) {
                return Result.error("动态不存在");
            }

            // 检查是否已转发
            DynamicShare existingShare = dynamicShareMapper.findByDynamicAndUser(dynamicId, userId);
            if (existingShare != null) {
                return Result.error("已经转发过了");
            }

            // 创建转发记录
            DynamicShare share = new DynamicShare();
            share.setDynamicId(dynamicId);
            share.setUserId(userId);
            share.setContent(content);
            dynamicShareMapper.insert(share);

            // 更新原动态的转发数
            int newShareCount = originalDynamic.getShareCount() + 1;
            userDynamicMapper.updateShareCount(dynamicId, newShareCount);

            return Result.success("转发成功", share);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicShare>> getShareList(Integer dynamicId) {
        try {
            List<DynamicShare> shareList = dynamicShareMapper.findByDynamicId(dynamicId);
            return Result.success("获取成功", shareList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<List<DynamicShare>> getUserShareList(Integer userId) {
        try {
            List<DynamicShare> shareList = dynamicShareMapper.findByUserId(userId);
            return Result.success("获取成功", shareList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public boolean isShared(Integer userId, Integer dynamicId) {
        DynamicShare share = dynamicShareMapper.findByDynamicAndUser(dynamicId, userId);
        return share != null;
    }

    @Override
    @Transactional
    public Result<?> cancelShare(Integer userId, Integer shareId) {
        try {
            // 获取转发记录
            DynamicShare share = dynamicShareMapper.findByDynamicAndUser(shareId, userId);
            if (share == null) {
                return Result.error("转发记录不存在");
            }

            // 删除转发记录
            dynamicShareMapper.deleteById(shareId);

            // 更新原动态的转发数
            UserDynamic originalDynamic = userDynamicMapper.getById(share.getDynamicId());
            if (originalDynamic != null) {
                int newShareCount = Math.max(0, originalDynamic.getShareCount() - 1);
                userDynamicMapper.updateShareCount(share.getDynamicId(), newShareCount);
            }

            return Result.success("取消转发成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
