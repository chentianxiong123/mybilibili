package com.mybilibili.interaction.service;

import com.mybilibili.common.vo.ManuscriptCollectionVO;
import java.util.List;

public interface CollectionService {
    List<ManuscriptCollectionVO> getCollectionsByUserId(Integer userId);
    ManuscriptCollectionVO getCollectionById(Integer collectionId, Integer currentUserId);
    ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status);
    ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status, List<Integer> manuscriptIds);
    ManuscriptCollectionVO createCollection(String name, String description, Integer userId, Integer status, List<Integer> manuscriptIds, String coverUrl);
    ManuscriptCollectionVO updateCollection(Integer collectionId, String name, String description, Integer userId, Integer status);
    void deleteCollection(Integer collectionId, Integer userId);
    void addManuscriptToCollection(Integer collectionId, Integer manuscriptId, Integer userId, Integer order);
    void removeManuscriptFromCollection(Integer collectionId, Integer manuscriptId, Integer userId);
    List<ManuscriptCollectionVO.ManuscriptItemVO> getCollectionManuscripts(Integer collectionId);
}
