package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.ManuscriptCollectionRelation;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface ManuscriptCollectionRelationMapper extends BaseMapper<ManuscriptCollectionRelation> {

    @Select("SELECT * FROM manuscript_collection_relations WHERE collection_id = #{collectionId} ORDER BY collection_order ASC")
    List<ManuscriptCollectionRelation> selectByCollectionId(@Param("collectionId") Integer collectionId);

    @Insert("INSERT INTO manuscript_collection_relations (collection_id, manuscript_id, collection_order) VALUES (#{collectionId}, #{manuscriptId}, #{collectionOrder})")
    int insert(ManuscriptCollectionRelation relation);

    @Delete("DELETE FROM manuscript_collection_relations WHERE collection_id = #{collectionId} AND manuscript_id = #{manuscriptId}")
    int deleteByCollectionAndManuscript(@Param("collectionId") Integer collectionId, @Param("manuscriptId") Integer manuscriptId);

    @Delete("DELETE FROM manuscript_collection_relations WHERE collection_id = #{collectionId}")
    int deleteByCollectionId(@Param("collectionId") Integer collectionId);
}
