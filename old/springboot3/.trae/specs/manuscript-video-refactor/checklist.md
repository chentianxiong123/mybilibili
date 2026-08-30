# 稿件视频概念重构检查清单

## 阶段一：数据库清理检查

- [x] videos表已删除废弃字段user_id和category_id
- [x] comments表已删除冗余字段target_type和target_id
- [x] favorite_folders表已重命名字段video_count为collect_count
- [x] user_interactions表数据已从VIDEO迁移到MANUSCRIPT

## 阶段二：实体类修复检查

- [x] Video实体类与数据库videos表字段完全匹配
- [x] Video实体类已添加@TableName("videos")注解
- [x] Comment实体类已移除video相关字段
- [x] Comment实体类只保留manuscriptId字段
- [x] UserInteraction实体类已添加@TableName("user_interactions")注解
- [x] FavoriteFolder实体类使用collectCount字段

## 阶段三：服务层重构检查

- [x] 互动服务：所有TARGET_TYPE_VIDEO改为TARGET_TYPE_MANUSCRIPT
- [x] 互动服务：所有setVideoCount改为setCollectCount
- [x] 评论服务：移除所有targetType相关逻辑
- [x] 评论服务：只使用manuscript_id关联

## 阶段四：编译检查

- [x] mybilibili-common模块编译成功
- [x] mybilibili-interaction模块编译成功
- [x] mybilibili-comment模块编译成功
