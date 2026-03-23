package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.UserDynamic;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface UserDynamicMapper {
    int insert(UserDynamic userDynamic);
    
    List<UserDynamic> getByUserId(@Param("userId") Integer userId, @Param("offset") Integer offset, @Param("limit") Integer limit);
    
    List<UserDynamic> getByUserIds(@Param("userIds") List<Integer> userIds, @Param("offset") Integer offset, @Param("limit") Integer limit);
    
    List<UserDynamic> getAllDynamics(@Param("offset") Integer offset, @Param("limit") Integer limit);
    
    UserDynamic getById(Integer id);
    
    int updateLikeCount(@Param("id") Integer id, @Param("count") Integer count);
    
    int updateCommentCount(@Param("id") Integer id, @Param("count") Integer count);
    
    int updateShareCount(@Param("id") Integer id, @Param("count") Integer count);
    
    int updateStatus(@Param("id") Integer id, @Param("status") Integer status);
    
    int deleteById(@Param("id") Integer id);
    
    int countAllDynamics();
    
    int countByUserIds(@Param("userIds") List<Integer> userIds);
    
    int countByUserId(@Param("userId") Integer userId);
}
