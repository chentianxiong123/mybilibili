package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("favorite_manuscripts")
public class ManuscriptFavorite {
    private Integer id;
    private Integer folderId;
    private Integer manuscriptId;
    private Date createdAt;
}
