# 数据库字段修正计划

## 概述
根据数据库结构的变更，需要修正所有服务代码以符合最新的数据库结构。

## 数据库变更摘要
1. **users表**：
   - `video_count` → `manuscript_count`
   - 删除 `point_count` 字段

2. **user_dynamics表**：
   - 删除 `ref_video_id` 字段

## 需要修改的文件

### 1. mybilibili-common 模块

#### 1.1 User.java (实体类)
**文件路径**: `mybilibili-common/src/main/java/com/mybilibili/common/entity/User.java`

**修改内容**:
- 将 `videoCount` 字段重命名为 `manuscriptCount`
- 删除 `pointCount` 字段

**修改前**:
```java
private Integer videoCount;
private Integer pointCount;
```

**修改后**:
```java
private Integer manuscriptCount;
```

#### 1.2 UserDynamic.java (实体类)
**文件路径**: `mybilibili-common/src/main/java/com/mybilibili/common/entity/UserDynamic.java`

**修改内容**:
- 删除 `refVideoId` 字段

**修改前**:
```java
private Integer refVideoId;
private Integer refManuscriptId;
```

**修改后**:
```java
private Integer refManuscriptId;
```

#### 1.3 DynamicVO.java (VO类)
**文件路径**: `mybilibili-common/src/main/java/com/mybilibili/common/vo/DynamicVO.java`

**修改内容**:
- 删除 `refVideoId` 字段
- 删除 `refVideo` 字段（因为不再引用视频）

**修改前**:
```java
private Integer refVideoId;
private Integer refManuscriptId;
private VideoRefVO refVideo;
```

**修改后**:
```java
private Integer refManuscriptId;
```

### 2. mybilibili-search 模块

#### 2.1 ManuscriptDocument.java (搜索文档)
**文件路径**: `mybilibili-search/src/main/java/com/mybilibili/search/document/ManuscriptDocument.java`

**修改内容**:
- `videoCount` 字段保留，但含义为稿件下的视频数量（分P数），无需修改

### 3. mybilibili-common 模块（搜索相关）

#### 3.1 ManuscriptDocument.java (搜索文档)
**文件路径**: `mybilibili-common/src/main/java/com/mybilibili/common/document/ManuscriptDocument.java`

**修改内容**:
- `videoCount` 字段保留，但含义为稿件下的视频数量（分P数），无需修改

#### 3.2 VideoSearchVO.java (搜索结果VO)
**文件路径**: `mybilibili-common/src/main/java/com/mybilibili/common/vo/VideoSearchVO.java`

**修改内容**:
- `videoCount` 字段保留，但含义为稿件下的视频数量（分P数），无需修改

## 修改优先级

### P0 - 必须修改（影响数据库映射）
1. User.java - 实体类字段与数据库不匹配
2. UserDynamic.java - 实体类字段与数据库不匹配

### P1 - 建议修改（保持代码一致性）
1. DynamicVO.java - 移除废弃字段引用

### P2 - 可选修改（不影响功能）
1. ManuscriptDocument.java - 字段含义正确，无需修改
2. VideoSearchVO.java - 字段含义正确，无需修改

## 验证步骤

1. 编译所有模块，确保无编译错误
2. 运行单元测试，确保功能正常
3. 启动各服务，验证数据库操作正常
4. 检查Swagger文档，确认接口正常

## 注意事项

1. 修改实体类字段名时，MyBatis-Plus会自动映射到数据库字段（驼峰转下划线）
2. 删除字段后，确保相关业务逻辑不再使用这些字段
3. VO类的修改不影响数据库映射，但需要确保前端不再使用这些字段
