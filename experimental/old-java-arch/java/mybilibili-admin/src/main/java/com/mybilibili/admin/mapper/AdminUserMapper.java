package com.mybilibili.admin.mapper;

import com.mybilibili.admin.entity.AdminUser;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface AdminUserMapper {
    @Select("SELECT * FROM admin_users WHERE username = #{username}")
    AdminUser selectByUsername(String username);
    
    @Insert("INSERT INTO admin_users (username, password, admin_level) VALUES (#{username}, #{password}, #{adminLevel})")
    int insert(AdminUser adminUser);
    
    @Select("SELECT * FROM admin_users")
    List<AdminUser> selectAll();
    
    @Select("SELECT * FROM admin_users WHERE id = #{id}")
    AdminUser selectById(Integer id);
}