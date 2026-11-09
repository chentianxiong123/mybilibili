package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.ManuscriptCollectionRelation;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface ManuscriptCollectionRelationMapper {

    @Insert("INSERT INTO manuscript_collection_relations (manuscript_id, collection_id, collection_order, created_at) " +
            "VALUES (#{manuscriptId}, #{collectionId}, #{collectionOrder}, #{createdAt})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(ManuscriptCollectionRelation relation);

    @Delete("DELETE FROM manuscript_collection_relations WHERE id = #{id}")
    int deleteById(Integer id);

    @Delete("DELETE FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId} AND collection_id = #{collectionId}")
    int deleteByManuscriptAndCollection(@Param("manuscriptId") Integer manuscriptId, @Param("collectionId") Integer collectionId);

    @Delete("DELETE FROM manuscript_collection_relations WHERE collection_id = #{collectionId}")
    int deleteByCollectionId(Integer collectionId);

    @Delete("DELETE FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId}")
    int deleteByManuscriptId(Integer manuscriptId);

    @Select("SELECT * FROM manuscript_collection_relations WHERE id = #{id}")
    ManuscriptCollectionRelation selectById(Integer id);

    @Select("SELECT * FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId} AND collection_id = #{collectionId}")
    ManuscriptCollectionRelation selectByManuscriptAndCollection(@Param("manuscriptId") Integer manuscriptId, @Param("collectionId") Integer collectionId);

    @Select("SELECT * FROM manuscript_collection_relations WHERE collection_id = #{collectionId} ORDER BY collection_order ASC")
    List<ManuscriptCollectionRelation> selectByCollectionId(Integer collectionId);

    @Select("SELECT * FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId}")
    List<ManuscriptCollectionRelation> selectByManuscriptId(Integer manuscriptId);

    @Select("SELECT COUNT(*) FROM manuscript_collection_relations WHERE collection_id = #{collectionId}")
    int countByCollectionId(Integer collectionId);

    @Select("SELECT COUNT(*) FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId} AND collection_id = #{collectionId}")
    int countByManuscriptAndCollection(@Param("manuscriptId") Integer manuscriptId, @Param("collectionId") Integer collectionId);

    @Update("UPDATE manuscript_collection_relations SET " +
            "collection_order = #{collectionOrder} " +
            "WHERE id = #{id}")
    int updateOrder(ManuscriptCollectionRelation relation);

    @Update("UPDATE manuscript_collection_relations SET " +
            "collection_order = #{order} " +
            "WHERE manuscript_id = #{manuscriptId} AND collection_id = #{collectionId}")
    int updateOrderByManuscriptAndCollection(@Param("manuscriptId") Integer manuscriptId, 
                                              @Param("collectionId") Integer collectionId, 
                                              @Param("order") Integer order);

    @Select("SELECT COALESCE(MAX(collection_order), -1) FROM manuscript_collection_relations WHERE collection_id = #{collectionId}")
    int selectMaxOrderByCollectionId(Integer collectionId);

    @Update("UPDATE manuscript_collection_relations SET " +
            "collection_order = collection_order - 1 " +
            "WHERE collection_id = #{collectionId} AND collection_order > #{order}")
    int shiftOrdersAfterRemove(@Param("collectionId") Integer collectionId, @Param("order") Integer order);

    /**
     * 查询合集中的稿件列表（按顺序）
     */
    @Select("SELECT m.* FROM manuscripts m " +
            "INNER JOIN manuscript_collection_relations mcr ON m.id = mcr.manuscript_id " +
            "WHERE mcr.collection_id = #{collectionId} " +
            "ORDER BY mcr.collection_order ASC")
    List<Manuscript> selectManuscriptsByCollectionId(Integer collectionId);

    /**
     * 查询稿件所属的所有合集ID列表
     */
    @Select("SELECT collection_id FROM manuscript_collection_relations WHERE manuscript_id = #{manuscriptId}")
    List<Integer> selectCollectionIdsByManuscriptId(Integer manuscriptId);
}
