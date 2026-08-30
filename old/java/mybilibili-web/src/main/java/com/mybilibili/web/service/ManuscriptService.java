package com.mybilibili.web.service;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.vo.ManuscriptVO;

import java.util.List;
import java.util.Map;

public interface ManuscriptService {

    /**
     * 上传稿件（支持单视频和多视频分P）
     *
     * @param dto    稿件上传DTO
     * @param userId 用户ID
     * @return 稿件VO
     * @throws Exception 上传异常
     */
    ManuscriptVO uploadManuscript(ManuscriptUploadDTO dto, Integer userId) throws Exception;

    /**
     * 查询稿件详情（关联查询视频列表）
     *
     * @param id 稿件ID
     * @return 稿件VO
     */
    ManuscriptVO getManuscriptById(Integer id);

    /**
     * 查询稿件详情（关联查询视频列表，带当前用户ID）
     *
     * @param id           稿件ID
     * @param currentUserId 当前用户ID
     * @return 稿件VO
     */
    ManuscriptVO getManuscriptById(Integer id, Integer currentUserId);

    /**
     * 查询用户稿件列表
     *
     * @param userId 用户ID
     * @return 稿件VO列表
     */
    List<ManuscriptVO> getManuscriptsByUserId(Integer userId);

    /**
     * 分页查询用户稿件列表（支持状态筛选）
     *
     * @param userId   用户ID
     * @param status   状态（可选）
     * @param page     页码（从1开始）
     * @param size     每页数量
     * @return 稿件VO列表
     */
    List<ManuscriptVO> getManuscriptsByUserIdWithPaging(Integer userId, Integer status, Integer page, Integer size);

    /**
     * 统计用户稿件数量（支持状态筛选）
     *
     * @param userId 用户ID
     * @param status 状态（可选）
     * @return 稿件数量
     */
    Integer countManuscriptsByUserIdAndStatus(Integer userId, Integer status);

    /**
     * 获取用户稿件统计（各状态数量）
     *
     * @param userId 用户ID
     * @return 各状态稿件数量统计
     */
    Map<String, Integer> getManuscriptStatsByUserId(Integer userId);

    /**
     * 更新稿件
     *
     * @param id         稿件ID
     * @param manuscript 稿件实体
     * @return 更新后的稿件VO
     * @throws Exception 更新异常
     */
    ManuscriptVO updateManuscript(Integer id, Manuscript manuscript) throws Exception;

    /**
     * 删除稿件
     *
     * @param id     稿件ID
     * @param userId 用户ID
     * @throws Exception 删除异常
     */
    void deleteManuscript(Integer id, Integer userId) throws Exception;

    /**
     * 更新稿件状态
     *
     * @param id     稿件ID
     * @param status 状态值
     * @return 是否更新成功
     */
    boolean updateManuscriptStatus(Integer id, Integer status);

    /**
     * 获取稿件详情（包含视频列表）
     *
     * @param id 稿件ID
     * @return 稿件VO（包含videos列表）
     */
    ManuscriptVO getManuscriptWithVideos(Integer id);

    /**
     * 获取推荐稿件列表
     *
     * @param userId 用户ID（可为null，表示未登录用户）
     * @return 推荐稿件VO列表
     */
    List<ManuscriptVO> getRecommendedManuscripts(Integer userId);

    /**
     * 重新计算并更新稿件时长
     * 用于修复历史数据或补充缺失的时长信息
     *
     * @param manuscriptId 稿件ID
     * @return 更新后的稿件VO
     */
    ManuscriptVO recalculateDuration(Integer manuscriptId);

    /**
     * 批量修复所有稿件的时长
     * 用于系统升级后修复历史数据
     *
     * @return 修复的稿件数量
     */
    int fixAllManuscriptDurations();
}
