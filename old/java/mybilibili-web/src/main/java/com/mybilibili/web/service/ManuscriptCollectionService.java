package com.mybilibili.web.service;

import com.mybilibili.common.entity.ManuscriptCollection;
import com.mybilibili.common.vo.ManuscriptCollectionVO;

import java.util.List;

public interface ManuscriptCollectionService {

    /**
     * 创建合集
     *
     * @param collection 合集实体
     * @param userId     用户ID
     * @return 合集VO
     */
    ManuscriptCollectionVO createCollection(ManuscriptCollection collection, Integer userId);

    /**
     * 根据ID查询合集详情
     *
     * @param id 合集ID
     * @return 合集VO
     */
    ManuscriptCollectionVO getCollectionById(Integer id);

    /**
     * 根据ID查询合集详情（带当前用户ID）
     *
     * @param id            合集ID
     * @param currentUserId 当前用户ID
     * @return 合集VO
     */
    ManuscriptCollectionVO getCollectionById(Integer id, Integer currentUserId);

    /**
     * 查询用户的合集列表
     *
     * @param userId 用户ID
     * @return 合集VO列表
     */
    List<ManuscriptCollectionVO> getCollectionsByUserId(Integer userId);

    /**
     * 查询用户的合集列表（带状态筛选）
     *
     * @param userId 用户ID
     * @param status 状态
     * @return 合集VO列表
     */
    List<ManuscriptCollectionVO> getCollectionsByUserIdAndStatus(Integer userId, Integer status);

    /**
     * 更新合集
     *
     * @param id         合集ID
     * @param collection 合集实体
     * @param userId     用户ID
     * @return 更新后的合集VO
     */
    ManuscriptCollectionVO updateCollection(Integer id, ManuscriptCollection collection, Integer userId);

    /**
     * 删除合集
     *
     * @param id     合集ID
     * @param userId 用户ID
     */
    void deleteCollection(Integer id, Integer userId);

    /**
     * 添加稿件到合集
     *
     * @param collectionId  合集ID
     * @param manuscriptId  稿件ID
     * @param userId        用户ID
     * @return 是否添加成功
     */
    boolean addManuscriptToCollection(Integer collectionId, Integer manuscriptId, Integer userId);

    /**
     * 从合集中移除稿件
     *
     * @param collectionId 合集ID
     * @param manuscriptId 稿件ID
     * @param userId       用户ID
     * @return 是否移除成功
     */
    boolean removeManuscriptFromCollection(Integer collectionId, Integer manuscriptId, Integer userId);

    /**
     * 调整稿件在合集中的顺序
     *
     * @param collectionId 合集ID
     * @param manuscriptId 稿件ID
     * @param newOrder     新顺序
     * @param userId       用户ID
     * @return 是否调整成功
     */
    boolean updateManuscriptOrder(Integer collectionId, Integer manuscriptId, Integer newOrder, Integer userId);

    /**
     * 批量调整稿件顺序
     *
     * @param collectionId   合集ID
     * @param manuscriptIds  稿件ID列表（按新顺序排列）
     * @param userId         用户ID
     * @return 是否调整成功
     */
    boolean batchUpdateManuscriptOrder(Integer collectionId, List<Integer> manuscriptIds, Integer userId);
}
