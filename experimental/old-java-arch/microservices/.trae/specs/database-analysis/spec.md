# 数据库结构分析与优化 - 产品需求文档

## Overview
- **Summary**: 分析当前MySQL数据库结构，识别存在的问题并提出优化方案，确保数据库设计符合业务逻辑和性能要求。
- **Purpose**: 解决数据库中稿件(manuscript)与视频(video)概念混淆、冗余字段、约束缺失等问题，提高数据库设计的合理性和性能。
- **Target Users**: 开发团队、数据库管理员

## Goals
- 解决稿件与视频概念混淆问题
- 清理冗余字段和表结构
- 完善数据库约束和索引
- 统一命名规范
- 优化数据库性能

## Non-Goals (Out of Scope)
- 不涉及数据迁移策略
- 不修改业务逻辑代码
- 不调整数据库服务器配置

## Background & Context
当前数据库结构存在以下问题：
- 概念混淆：favorite_videos表实际存储manuscript_id
- 冗余字段：videos表中存在与manuscripts表重复的字段
- 约束缺失：缺少外键约束，数据完整性无法保证
- 命名不一致：表名和字段名命名风格不统一
- 索引优化：部分查询可能存在性能问题

## Functional Requirements
- **FR-1**: 清理favorite_videos表，重命名为favorite_manuscripts
- **FR-2**: 清理comments表中冗余的target_type和target_id字段
- **FR-3**: 清理videos表中与manuscripts表重复的字段
- **FR-4**: 完善数据库外键约束
- **FR-5**: 统一时间字段命名规范
- **FR-6**: 优化数据库索引结构

## Non-Functional Requirements
- **NFR-1**: 保持数据库向后兼容性
- **NFR-2**: 确保修改后数据库性能不下降
- **NFR-3**: 维持数据完整性和一致性

## Constraints
- **Technical**: MySQL 8.0+
- **Business**: 最小化对现有业务的影响
- **Dependencies**: 与现有代码的兼容性

## Assumptions
- 所有修改都在测试环境中验证后再应用到生产环境
- 数据迁移脚本将在后续阶段实现

## Acceptance Criteria

### AC-1: 概念清晰化
- **Given**: 数据库结构分析完成
- **When**: 执行表重命名和结构调整
- **Then**: favorite_manuscripts表正确存储稿件收藏关系
- **Verification**: `programmatic`

### AC-2: 冗余字段清理
- **Given**: 冗余字段分析完成
- **When**: 删除重复字段
- **Then**: 数据库结构更加简洁，无冗余字段
- **Verification**: `programmatic`

### AC-3: 约束完善
- **Given**: 约束分析完成
- **When**: 添加外键约束
- **Then**: 数据完整性得到保证
- **Verification**: `programmatic`

### AC-4: 命名规范统一
- **Given**: 命名分析完成
- **When**: 统一命名规范
- **Then**: 数据库结构更加规范，易于维护
- **Verification**: `human-judgment`

### AC-5: 性能优化
- **Given**: 索引分析完成
- **When**: 优化索引结构
- **Then**: 数据库查询性能得到提升
- **Verification**: `programmatic`

## Open Questions
- [ ] 是否需要保留历史数据迁移策略？
- [ ] 索引优化的具体方案需要进一步分析
