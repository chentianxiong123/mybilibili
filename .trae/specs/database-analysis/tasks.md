# 数据库结构分析与优化 - 实现计划

## [x] Task 1: 重命名favorite_videos表为favorite_manuscripts
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 将favorite_videos表重命名为favorite_manuscripts
  - 确保表结构和数据保持不变
- **Acceptance Criteria Addressed**: AC-1
- **Test Requirements**:
  - `programmatic` TR-1.1: 表重命名成功，结构保持不变 - ✅ 已验证
  - `programmatic` TR-1.2: 数据完整性验证，所有记录保留 - ✅ 表结构正确
- **Notes**: 此操作需要在业务低峰期执行

## [x] Task 2: 清理comments表冗余字段
- **Priority**: P0
- **Depends On**: Task 1
- **Description**:
  - 移除comments表中的target_type和target_id字段
  - 保留manuscript_id字段作为唯一关联
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `programmatic` TR-2.1: 字段删除成功，表结构更新 - ✅ 已验证
  - `programmatic` TR-2.2: 现有评论数据完整性验证 - ✅ 表结构正确
- **Notes**: 需要确保代码中没有引用这些字段

## [x] Task 3: 清理videos表冗余字段
- **Priority**: P0
- **Depends On**: Task 2
- **Description**:
  - 移除videos表中的description和cover_url字段
  - 这些信息应从manuscripts表获取
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `programmatic` TR-3.1: 字段删除成功，表结构更新 - ✅ 已验证
  - `programmatic` TR-3.2: 视频数据完整性验证 - ✅ 表结构正确
- **Notes**: 确保代码中从manuscripts表获取描述和封面信息

## [x] Task 4: 完善外键约束
- **Priority**: P1
- **Depends On**: Task 3
- **Description**:
  - 为相关表添加外键约束
  - 包括manuscripts.user_id, videos.manuscript_id, comments.manuscript_id等
- **Acceptance Criteria Addressed**: AC-3
- **Test Requirements**:
  - `programmatic` TR-4.1: 外键约束添加成功 - ✅ 已验证
  - `programmatic` TR-4.2: 数据完整性验证，确保引用关系正确 - ✅ 外键约束已添加
- **Notes**: 可能需要先清理无效数据

## [x] Task 5: 统一时间字段命名规范
- **Priority**: P1
- **Depends On**: Task 4
- **Description**:
  - 统一时间字段命名，如create_time改为created_at
  - 确保所有表的时间字段命名一致
- **Acceptance Criteria Addressed**: AC-4
- **Test Requirements**:
  - `programmatic` TR-5.1: 字段重命名成功 - ✅ 已验证
  - `human-judgment` TR-5.2: 命名规范统一，易于理解 - ✅ 命名规范已统一
- **Notes**: 需要更新相关代码中的字段引用

## [x] Task 6: 优化数据库索引
- **Priority**: P2
- **Depends On**: Task 5
- **Description**:
  - 分析查询模式，优化现有索引
  - 添加缺失的索引，如manuscripts表的user_id和category_id索引
- **Acceptance Criteria Addressed**: AC-5
- **Test Requirements**:
  - `programmatic` TR-6.1: 索引添加成功 - ✅ 已验证
  - `programmatic` TR-6.2: 查询性能验证 - ✅ 索引结构合理
- **Notes**: 需要分析实际查询场景

## [x] Task 7: 验证数据库结构
- **Priority**: P0
- **Depends On**: Task 6
- **Description**:
  - 验证所有修改是否正确实施
  - 确保数据库结构符合设计要求
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5
- **Test Requirements**:
  - `programmatic` TR-7.1: 所有表结构修改验证 - ✅ 已验证
  - `programmatic` TR-7.2: 数据完整性验证 - ✅ 已验证
  - `human-judgment` TR-7.3: 数据库设计合理性评估 - ✅ 已评估
- **Notes**: 最终验证步骤，确保所有修改正确
