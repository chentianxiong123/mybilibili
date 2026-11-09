package com.mybilibili.comment.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.Report;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface ReportMapper extends BaseMapper<Report> {

    @Select("<script>" +
            "SELECT r.*, u.username as reporterName, u.avatar as reporterAvatar " +
            "FROM reports r LEFT JOIN users u ON r.reporter_id = u.id " +
            "WHERE 1=1 " +
            "<if test='status != null'> AND r.status = #{status} </if>" +
            "<if test='targetType != null'> AND r.target_type = #{targetType} </if>" +
            "ORDER BY r.created_at DESC LIMIT #{offset}, #{size}" +
            "</script>")
    List<Report> selectReportList(@Param("status") String status,
                                  @Param("targetType") String targetType,
                                  @Param("offset") int offset,
                                  @Param("size") int size);

    @Select("<script>" +
            "SELECT COUNT(*) FROM reports r WHERE 1=1 " +
            "<if test='status != null'> AND r.status = #{status} </if>" +
            "<if test='targetType != null'> AND r.target_type = #{targetType} </if>" +
            "</script>")
    int countReportList(@Param("status") String status,
                        @Param("targetType") String targetType);
}
