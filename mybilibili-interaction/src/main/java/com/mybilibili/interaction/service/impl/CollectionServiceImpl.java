package com.mybilibili.interaction.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.ManuscriptCollection;
import com.mybilibili.common.entity.ManuscriptCollectionRelation;
import com.mybilibili.common.vo.ManuscriptCollectionVO;
import com.mybilibili.interaction.mapper.ManuscriptCollectionMapper;
import com.mybilibili.interaction.mapper.ManuscriptCollectionRelationMapper;
import com.mybilibili.interaction.mapper.ManuscriptMapper;
import com.mybilibili.interaction.service.CollectionService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
public class CollectionServiceImpl implements CollectionService {

    @Autowired
    private ManuscriptCollectionMapper collectionMapper;

    @Autowired
    private ManuscriptCollectionRelationMapper relationMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Override
    public List<ManuscriptCollectionVO> getCollectionsByUserId(Integer userId) {
        List<ManuscriptCollection> collections = collectionMapper.selectByUserId(userId);
        return collections.stream().map(this::convertToVO).collect(Collectors.toList());
    }

    @Override
    public ManuscriptCollectionVO getCollectionById(Integer collectionId, Integer currentUserId) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            return null;
        }
        ManuscriptCollectionVO vo = convertToVO(collection);
        List<ManuscriptCollectionRelation> relations = relationMapper.selectByCollectionId(collectionId);
        List<ManuscriptCollectionVO.ManuscriptItemVO> items = new ArrayList<>();
        for (ManuscriptCollectionRelation relation : relations) {
            Manuscript manuscript = manuscriptMapper.selectBasicFieldsById(relation.getManuscriptId());
            if (manuscript != null) {
                ManuscriptCollectionVO.ManuscriptItemVO item = new ManuscriptCollectionVO.ManuscriptItemVO();
                item.setId(manuscript.getId());
                item.setTitle(manuscript.getTitle());
                item.setDescription(manuscript.getDescription());
                item.setCoverUrl(manuscript.getCoverUrl());
                item.setViewCount(manuscript.getViewCount());
                item.setLikeCount(manuscript.getLikeCount());
                item.setCollectionOrder(relation.getCollectionOrder());
                items.add(item);
            }
        }
        vo.setManuscripts(items);
        return vo;
    }

    @Override
    public ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status) {
        return createCollection(name, description, userId, status, null, null);
    }

    @Override
    public ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status, List<Integer> manuscriptIds) {
        return createCollection(name, description, userId, status, manuscriptIds, null);
    }

    @Override
    @Transactional
    public ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status, List<Integer> manuscriptIds, String coverUrl) {
        log.info("创建合集 - name: {}, userId: {}, status: {}, manuscriptIds: {}, coverUrl: {}", name, userId, status, manuscriptIds, coverUrl);
        ManuscriptCollection collection = new ManuscriptCollection();
        collection.setTitle(name);
        collection.setDescription(description);
        collection.setUserId(userId);
        collection.setManuscriptCount(0);
        collection.setStatus(status != null ? status : ManuscriptCollection.STATUS_PUBLIC);
        collection.setCoverUrl(coverUrl);
        collectionMapper.insert(collection);
        log.info("合集插入成功 - id: {}", collection.getId());
        
        if (manuscriptIds != null && !manuscriptIds.isEmpty()) {
            log.info("开始插入 {} 个稿件关联", manuscriptIds.size());
            for (int i = 0; i < manuscriptIds.size(); i++) {
                ManuscriptCollectionRelation relation = new ManuscriptCollectionRelation();
                relation.setCollectionId(collection.getId());
                relation.setManuscriptId(manuscriptIds.get(i));
                relation.setCollectionOrder(i);
                int result = relationMapper.insert(relation);
                log.info("插入关联结果: {}", result);
            }
            collection.setManuscriptCount(manuscriptIds.size());
            collectionMapper.update(collection);
            log.info("更新合集稿件数量完成");
        } else {
            log.info("manuscriptIds 为空或null，不插入关联记录");
        }
        
        return convertToVO(collection);
    }

    @Override
    @Transactional
    public ManuscriptCollectionVO updateCollection(Integer collectionId, String name, String description, Integer userId, Integer status) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null || !collection.getUserId().equals(userId)) {
            return null;
        }
        if (name != null && !name.isEmpty()) {
            collection.setTitle(name);
        }
        if (description != null) {
            collection.setDescription(description);
        }
        if (status != null) {
            collection.setStatus(status);
        }
        collectionMapper.update(collection);
        return convertToVO(collection);
    }

    @Override
    @Transactional
    public void deleteCollection(Integer collectionId, Integer userId) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection != null && collection.getUserId().equals(userId)) {
            relationMapper.deleteByCollectionId(collectionId);
            collectionMapper.deleteById(collectionId);
        }
    }

    @Override
    @Transactional
    public void addManuscriptToCollection(Integer collectionId, Integer manuscriptId, Integer userId, Integer order) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null || !collection.getUserId().equals(userId)) {
            return;
        }
        ManuscriptCollectionRelation relation = new ManuscriptCollectionRelation();
        relation.setCollectionId(collectionId);
        relation.setManuscriptId(manuscriptId);
        relation.setCollectionOrder(order != null ? order : 0);
        relationMapper.insert(relation);
        collection.setManuscriptCount(collection.getManuscriptCount() + 1);
        collectionMapper.update(collection);
    }

    @Override
    @Transactional
    public void removeManuscriptFromCollection(Integer collectionId, Integer manuscriptId, Integer userId) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null || !collection.getUserId().equals(userId)) {
            return;
        }
        relationMapper.deleteByCollectionAndManuscript(collectionId, manuscriptId);
        collection.setManuscriptCount(Math.max(0, collection.getManuscriptCount() - 1));
        collectionMapper.update(collection);
    }

    @Override
    public List<ManuscriptCollectionVO.ManuscriptItemVO> getCollectionManuscripts(Integer collectionId) {
        ManuscriptCollection collection = collectionMapper.selectById(collectionId);
        if (collection == null) {
            return new ArrayList<>();
        }
        List<ManuscriptCollectionRelation> relations = relationMapper.selectByCollectionId(collectionId);
        List<ManuscriptCollectionVO.ManuscriptItemVO> items = new ArrayList<>();
        for (ManuscriptCollectionRelation relation : relations) {
            Manuscript manuscript = manuscriptMapper.selectBasicFieldsById(relation.getManuscriptId());
            if (manuscript != null) {
                ManuscriptCollectionVO.ManuscriptItemVO item = new ManuscriptCollectionVO.ManuscriptItemVO();
                item.setId(manuscript.getId());
                item.setTitle(manuscript.getTitle());
                item.setDescription(manuscript.getDescription());
                item.setCoverUrl(manuscript.getCoverUrl());
                item.setViewCount(manuscript.getViewCount());
                item.setLikeCount(manuscript.getLikeCount());
                item.setCollectionOrder(relation.getCollectionOrder());
                items.add(item);
            }
        }
        return items;
    }

    private ManuscriptCollectionVO convertToVO(ManuscriptCollection collection) {
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
        return vo;
    }
}
