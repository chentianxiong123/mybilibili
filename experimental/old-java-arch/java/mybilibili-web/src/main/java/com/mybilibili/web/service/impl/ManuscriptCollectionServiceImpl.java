package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.ManuscriptCollection;
import com.mybilibili.common.entity.ManuscriptCollectionRelation;
import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.utils.DurationUtils;
import com.mybilibili.common.vo.ManuscriptCollectionVO;
import com.mybilibili.web.mapper.CommentMapper;
import com.mybilibili.web.mapper.ManuscriptCollectionMapper;
import com.mybilibili.web.mapper.ManuscriptCollectionRelationMapper;
import com.mybilibili.web.mapper.ReplyMapper;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.ManuscriptCollectionService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;

@Slf4j
@Service
public class ManuscriptCollectionServiceImpl implements ManuscriptCollectionService {

    @Autowired
    private ManuscriptCollectionMapper collectionMapper;

    @Autowired
    private ManuscriptCollectionRelationMapper relationMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptCollectionVO createCollection(ManuscriptCollection collection, Integer userId) {
        // 设置合集基本信息
        collection.setUserId(userId);
        collection.setManuscriptCount(0);
        collection.setViewCount(0);
        collection.setStatus(ManuscriptCollection.STATUS_PUBLIC);
        collection.setCreatedAt(new Date());
        collection.setUpdatedAt(new Date());

        // 如果没有设置封面，使用默认封面
        if (collection.getCoverUrl() == null || collection.getCoverUrl().isEmpty()) {
            collection.setCoverUrl("/default/collection-cover.jpg");
        }

        collectionMapper.insert(collection);

        return buildCollectionVO(collection, null, null);
    }

    @Override
    public ManuscriptCollectionVO getCollectionById(Integer id) {
        return getCollectionById(id, null);
    }

    @Override
    public ManuscriptCollectionVO getCollectionById(Integer id, Integer currentUserId) {
        ManuscriptCollection collection = collectionMapper.selectById(id);
        if (collection == null) {
            return null;
        }

        // 查询合集中的稿件列表
        List<Manuscript> manuscripts = relationMapper.selectManuscriptsByCollectionId(id);

        return buildCollectionVO(collection, manuscripts, currentUserId);
    }

    @Override
    public List<ManuscriptCollectionVO> getCollectionsByUserId(Integer userId) {
        List<ManuscriptCollection> collections = collectionMapper.selectByUserId(userId);
        List<ManuscriptCollectionVO> result = new ArrayList<>();

        for (ManuscriptCollection collection : collections) {
            // 查询合集中的稿件列表
            List<Manuscript> manuscripts = relationMapper.selectManuscriptsByCollectionId(collection.getId());
            result.add(buildCollectionVO(collection, manuscripts, null));
        }

        return result;
    }

    @Override
    public List<ManuscriptCollectionVO> getCollectionsByUserIdAndStatus(Integer userId, Integer status) {
        List<ManuscriptCollection> collections = collectionMapper.selectByUserIdAndStatus(userId, status);
        List<ManuscriptCollectionVO> result = new ArrayList<>();

        for (ManuscriptCollection collection : collections) {
            // 查询合集中的稿件列表
            List<Manuscript> manuscripts = relationMapper.selectManuscriptsByCollectionId(collection.getId());
            result.add(buildCollectionVO(collection, manuscripts, null));
        }

        return result;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ManuscriptCollectionVO updateCollection(Integer id, ManuscriptCollection collection, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection existingCollection = collectionMapper.selectById(id);
        if (existingCollection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!existingCollection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限修改此合集");
        }

        // 更新合集信息
        collection.setId(id);
        collection.setUpdatedAt(new Date());
        collectionMapper.update(collection);

        // 重新查询更新后的合集
        ManuscriptCollection updatedCollection = collectionMapper.selectById(id);
        List<Manuscript> manuscripts = relationMapper.selectManuscriptsByCollectionId(id);

        return buildCollectionVO(updatedCollection, manuscripts, null);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void deleteCollection(Integer id, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection collection = collectionMapper.selectById(id);
        if (collection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!collection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限删除此合集");
        }

        // 删除合集中的所有关联关系
        relationMapper.deleteByCollectionId(id);

        // 删除合集
        collectionMapper.delete(id);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean addManuscriptToCollection(Integer collectionId, Integer manuscriptId, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!collection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限操作此合集");
        }

        // 检查稿件是否存在且属于当前用户
        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            throw new RuntimeException("稿件不存在");
        }

        if (!manuscript.getUserId().equals(userId)) {
            throw new RuntimeException("只能添加自己的稿件到合集");
        }

        // 检查稿件是否已经在当前合集中
        int existingCount = relationMapper.countByManuscriptAndCollection(manuscriptId, collectionId);
        if (existingCount > 0) {
            throw new RuntimeException("该稿件已在当前合集中");
        }

        // 获取当前合集中最大的顺序号
        int maxOrder = relationMapper.selectMaxOrderByCollectionId(collectionId);
        int newOrder = maxOrder + 1;

        // 创建关联关系
        ManuscriptCollectionRelation relation = new ManuscriptCollectionRelation();
        relation.setManuscriptId(manuscriptId);
        relation.setCollectionId(collectionId);
        relation.setCollectionOrder(newOrder);
        relation.setCreatedAt(new Date());

        int result = relationMapper.insert(relation);
        if (result > 0) {
            // 更新合集稿件数量
            collectionMapper.updateManuscriptCount(collectionId, 1);
        }

        return result > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean removeManuscriptFromCollection(Integer collectionId, Integer manuscriptId, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!collection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限操作此合集");
        }

        // 检查稿件是否在合集中
        ManuscriptCollectionRelation relation = relationMapper.selectByManuscriptAndCollection(manuscriptId, collectionId);
        if (relation == null) {
            throw new RuntimeException("该稿件不在当前合集中");
        }

        int currentOrder = relation.getCollectionOrder();

        // 从合集中移除稿件
        int result = relationMapper.deleteByManuscriptAndCollection(manuscriptId, collectionId);
        if (result > 0) {
            // 更新合集稿件数量
            collectionMapper.updateManuscriptCount(collectionId, -1);
            // 重新排序后面的稿件
            relationMapper.shiftOrdersAfterRemove(collectionId, currentOrder);
        }

        return result > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean updateManuscriptOrder(Integer collectionId, Integer manuscriptId, Integer newOrder, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!collection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限操作此合集");
        }

        // 检查稿件是否在合集中
        ManuscriptCollectionRelation relation = relationMapper.selectByManuscriptAndCollection(manuscriptId, collectionId);
        if (relation == null) {
            throw new RuntimeException("该稿件不在当前合集中");
        }

        // 更新稿件顺序
        int result = relationMapper.updateOrderByManuscriptAndCollection(manuscriptId, collectionId, newOrder);
        return result > 0;
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public boolean batchUpdateManuscriptOrder(Integer collectionId, List<Integer> manuscriptIds, Integer userId) {
        // 检查合集是否存在
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            throw new RuntimeException("合集不存在");
        }

        // 检查权限
        if (!collection.getUserId().equals(userId)) {
            throw new RuntimeException("没有权限操作此合集");
        }

        // 批量更新稿件顺序
        for (int i = 0; i < manuscriptIds.size(); i++) {
            Integer manuscriptId = manuscriptIds.get(i);
            relationMapper.updateOrderByManuscriptAndCollection(manuscriptId, collectionId, i);
        }

        return true;
    }

    /**
     * 构建ManuscriptCollectionVO
     */
    private ManuscriptCollectionVO buildCollectionVO(ManuscriptCollection collection,
                                                      List<Manuscript> manuscripts,
                                                      Integer currentUserId) {
        ManuscriptCollectionVO vo = new ManuscriptCollectionVO();
        vo.setId(collection.getId());
        vo.setTitle(collection.getTitle());
        vo.setDescription(collection.getDescription());
        vo.setCoverUrl(collection.getCoverUrl());
        vo.setUserId(collection.getUserId());
        vo.setManuscriptCount(collection.getManuscriptCount());
        vo.setViewCount(collection.getViewCount());
        vo.setStatus(collection.getStatus());
        vo.setCreatedAt(collection.getCreatedAt());
        vo.setUpdatedAt(collection.getUpdatedAt());

        // 设置创建者信息
        ManuscriptCollectionVO.UserInfo userInfo = new ManuscriptCollectionVO.UserInfo();
        User user = userMapper.findById(collection.getUserId());
        if (user != null) {
            userInfo.setId(user.getId());
            userInfo.setName(user.getNickname());
            userInfo.setAvatar(user.getAvatar());
            userInfo.setLevel(user.getLevel());
        }
        vo.setCreator(userInfo);

        // 设置稿件列表
        List<ManuscriptCollectionVO.ManuscriptItemVO> manuscriptVOs = new ArrayList<>();
        if (manuscripts != null && !manuscripts.isEmpty()) {
            // 获取每个稿件在合集中的顺序
            List<ManuscriptCollectionRelation> relations = relationMapper.selectByCollectionId(collection.getId());
            for (Manuscript manuscript : manuscripts) {
                ManuscriptCollectionVO.ManuscriptItemVO manuscriptVO = new ManuscriptCollectionVO.ManuscriptItemVO();
                manuscriptVO.setId(manuscript.getId());
                manuscriptVO.setTitle(manuscript.getTitle());
                manuscriptVO.setDescription(manuscript.getDescription());
                manuscriptVO.setCoverUrl(manuscript.getCoverUrl());
                manuscriptVO.setViewCount(manuscript.getViewCount());
                manuscriptVO.setLikeCount(manuscript.getLikeCount());
                // 统计评论数+回复数总量
                int commentCount = commentMapper.countByManuscriptId(manuscript.getId());
                int replyCount = replyMapper.countByManuscriptId(manuscript.getId());
                manuscriptVO.setCommentCount(commentCount + replyCount);
                
                // 从关联关系中获取顺序
                for (ManuscriptCollectionRelation relation : relations) {
                    if (relation.getManuscriptId().equals(manuscript.getId())) {
                        manuscriptVO.setCollectionOrder(relation.getCollectionOrder());
                        break;
                    }
                }
                
                manuscriptVO.setUploadTime(manuscript.getUploadTime());
                // 获取视频时长（查询视频列表）
                List<Video> videos = videoMapper.selectByManuscriptId(manuscript.getId());
                if (videos != null && !videos.isEmpty()) {
                    manuscriptVO.setDuration(DurationUtils.formatDuration(videos.get(0).getDurationSeconds()));
                }
                manuscriptVOs.add(manuscriptVO);
            }
        }
        vo.setManuscripts(manuscriptVOs);

        return vo;
    }
}
