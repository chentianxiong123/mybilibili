# 稿件视频概念重构计划

## Why
当前微服务项目中存在严重的"稿件"和"视频"概念混乱问题，同时数据库表设计也存在混乱：
- 数据库表中有废弃字段未清理（如videos表的user_id/category_id标注为"已废弃"）
- comments表同时存在manuscript_id和target_id/target_type，功能重复
- favorite_folders表的video_count命名不准确（实际收藏的是稿件）
- 代码实体类字段与数据库表字段不匹配

## 数据库分析结果

### 表结构现状
| 表名 | 核心字段 | 问题 |
|------|----------|------|
| manuscripts | id, title, description, cover_url, user_id, **统计字段(view_count等)** | ✅ 正常，是内容核心 |
| videos | id, manuscript_id, title, play_url_hd/sd/ld, duration_seconds | user_id/category_id标注为"已废弃" |
| comments | id, manuscript_id, user_id, content, target_type, target_id | manuscript_id和target_id重复 |
| user_interactions | id, user_id, target_type, target_id, interaction_type | target_type支持VIDEO但应统一为MANUSCRIPT |
| dynamic_comments | id, dynamic_id, user_id, content | ✅ 正常，针对动态 |
| replies | id, comment_id, user_id, content | ✅ 正常，评论的回复 |
| favorite_folders | id, user_id, name, video_count | video_count应改为collect_count |

## What Changes

### 第一步：数据库表清理（不删除数据）

1. **videos表** - 清理废弃字段
   - 保留：id, manuscript_id, video_order, title, description, cover_url, play_url_hd, play_url_sd, play_url_ld, duration_seconds, upload_time, updated_at, process_progress, process_stage, has_subtitle, has_summary, process_status, process_error, source_video_url
   - 移除：user_id, category_id（通过manuscript关联即可，外键约束也要删除）

2. **comments表** - 简化结构
   - 移除：target_type, target_id（只保留manuscript_id）
   - 统一为针对稿件的评论

3. **favorite_folders表** - 字段重命名
   - video_count → collect_count

4. **user_interactions表** - 明确target_type
   - target_type只保留：MANUSCRIPT, DYNAMIC, COMMENT, REPLY, USER
   - 移除VIDEO类型（互动都针对稿件）

### 第二步：代码实体类修复

根据清理后的数据库表，重建所有实体类：

1. **Video.java** - 与videos表完全匹配
2. **Comment.java** - 只保留manuscript_id
3. **FavoriteFolder.java** - 使用collect_count
4. **UserInteraction.java** - 更新target_type枚举

### 第三步：服务层调整

1. 评论服务 - 只操作manuscript_id
2. 互动服务 - 只操作MANUSCRIPT类型
3. 弹幕服务 - 保持videoId（针对视频）

## Impact

### 受影响的表
- videos（清理废弃字段）
- comments（移除冗余字段）
- favorite_folders（重命名字段）

### 受影响的服务
- mybilibili-comment
- mybilibili-interaction
- mybilibili-common

## 实施原则
1. **不删除任何历史数据**
2. **废弃字段设为NULL或默认值**
3. **先分析再动手**
4. **每张表修改后验证**
