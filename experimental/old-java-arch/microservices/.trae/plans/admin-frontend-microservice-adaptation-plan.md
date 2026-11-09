# 管理员前端适配微服务后端方案

## 一、现状分析

### 1.1 管理员前端现有接口

管理员前端目前使用的接口路径：
- **稿件管理**: `/manuscript/*` (审核、发布、删除等)
- **用户管理**: `/users/*` (列表、详情、状态更新)
- **视频管理**: `/videos/*` (列表、详情、删除)
- **评论管理**: `/comments/*` (列表、详情、删除、状态更新)
- **分类管理**: `/categories/*` (CRUD操作)
- **统计数据**: `/statistics/*` (各种统计接口)
- **内容审核**: `/admin/content-review/*` (内容审核操作)
- **管理员**: `/admin/*` (登录、注册、列表、角色管理)

### 1.2 微服务后端现有接口

微服务后端已有的管理员相关接口：
- **管理员管理**: `/admin/*` (AdminUserController)
  - `/admin/login` - 管理员登录 ✅
  - `/admin/register` - 管理员注册 ✅
  - `/admin/list` - 管理员列表 ✅
  - `/admin/{id}` - 管理员详情 ✅
  - `/admin/{id}/roles` - 获取/设置管理员角色 ✅

- **用户管理**: `/user/admin/*` (UserController)
  - `/user/admin/list` - 用户列表 ✅
  - `/user/admin/{id}` - 用户详情 ✅
  - `/user/admin/{id}/status` - 更新用户状态 ✅
  - `/user/admin/{id}/password` - 重置用户密码 ✅

- **视频管理**: `/video/admin/*` (VideoAdminController)
  - `/video/admin/list` - 视频列表 ✅
  - `/video/admin/{id}` - 视频详情 ✅
  - `/video/admin/{id}` - 删除视频 ✅
  - `/video/admin/batch` - 批量删除视频 ✅

- **评论管理**: `/comment/admin/*` (CommentAdminController)
  - `/comment/admin/list` - 评论列表 ✅
  - `/comment/admin/{id}` - 评论详情 ✅
  - `/comment/admin/{id}` - 删除评论 ✅
  - `/comment/admin/{id}/status` - 更新评论状态 ✅

- **分类管理**: `/category/*` (CategoryController)
  - `/category` - 获取全部分区 ✅
  - `/category/{id}` - 获取单个分区 ✅
  - ❌ 缺少：添加、更新、删除分类的管理接口

- **创作者统计**: `/creator/stats/*` (CreatorStatsController)
  - `/creator/stats/overview` - 数据概览 ✅
  - `/creator/stats/trend` - 数据趋势 ✅
  - `/creator/stats/ranking` - 稿件排行 ✅
  - `/creator/stats/latest-comments` - 最新评论 ✅
  - `/creator/stats/fans-ranking` - 粉丝排行 ✅

## 二、接口映射方案

### 2.1 需要修改的前端接口路径

| 前端原接口 | 微服务新接口 | 状态 | 说明 |
|-----------|------------|------|------|
| `/manuscript/*` | 需要新建 | ⚠️ 缺失 | 稿件管理接口需要新建 |
| `/users` | `/user/admin/list` | ✅ 已有 | 用户列表 |
| `/users/{id}` | `/user/admin/{id}` | ✅ 已有 | 用户详情 |
| `/users/{id}/status` | `/user/admin/{id}/status` | ✅ 已有 | 更新用户状态 |
| `/users/{id}/password` | `/user/admin/{id}/password` | ✅ 已有 | 重置密码 |
| `/videos` | `/video/admin/list` | ✅ 已有 | 视频列表 |
| `/videos/{id}` | `/video/admin/{id}` | ✅ 已有 | 视频详情 |
| `/comments` | `/comment/admin/list` | ✅ 已有 | 评论列表 |
| `/comments/{id}` | `/comment/admin/{id}` | ✅ 已有 | 评论详情 |
| `/comments/{id}/status` | `/comment/admin/{id}/status` | ✅ 已有 | 更新评论状态 |
| `/categories` | `/category` | ✅ 已有 | 分类列表 |
| `/categories/{id}` | `/category/{id}` | ✅ 已有 | 分类详情 |
| `/categories` (POST) | 需要新建 | ⚠️ 缺失 | 添加分类 |
| `/categories/{id}` (PUT) | 需要新建 | ⚠️ 缺失 | 更新分类 |
| `/categories/{id}` (DELETE) | 需要新建 | ⚠️ 缺失 | 删除分类 |
| `/statistics/*` | `/creator/stats/*` | ✅ 已有 | 统计数据 |
| `/admin/content-review/*` | 需要新建 | ⚠️ 缺失 | 内容审核 |
| `/admin/*` | `/admin/*` | ✅ 已有 | 管理员管理 |

### 2.2 需要新建的后端接口

#### 2.2.1 稿件管理接口 (高优先级)

需要在 `mybilibili-video` 服务中新建 `ManuscriptAdminController`：

```java
@RestController
@RequestMapping("/manuscript/admin")
public class ManuscriptAdminController {
    // GET /manuscript/admin/pending - 获取待审核稿件
    // GET /manuscript/admin/processing - 获取处理中稿件
    // GET /manuscript/admin/ready - 获取待上架稿件
    // GET /manuscript/admin/all - 获取所有稿件
    // GET /manuscript/admin/{id} - 获取稿件详情
    // POST /manuscript/admin/approve/{id} - 审核通过
    // POST /manuscript/admin/reject/{id} - 审核拒绝
    // POST /manuscript/admin/publish/{id} - 上架稿件
    // POST /manuscript/admin/unpublish/{id} - 下架稿件
    // GET /manuscript/admin/{id}/videos - 获取稿件视频列表
    // GET /manuscript/admin/statistics - 获取稿件统计
    // POST /manuscript/admin/retry/{id} - 重试处理
    // POST /manuscript/admin/transcode/{videoId} - 手动转码
    // POST /manuscript/admin/extract-audio/{videoId} - 提取音频
    // POST /manuscript/admin/generate-subtitle/{videoId} - 生成字幕
    // POST /manuscript/admin/ai-summary/{videoId} - AI总结
    // POST /manuscript/admin/process-all/{videoId} - 一键处理
    // POST /manuscript/admin/reset/{videoId} - 重置视频状态
    // GET /manuscript/admin/video-status/{videoId} - 获取视频处理状态
    // GET /manuscript/admin/video-source/{videoId} - 获取源视频URL
}
```

#### 2.2.2 分类管理接口 (中优先级)

需要在 `mybilibili-video` 服务中扩展 `CategoryController`：

```java
// 在 CategoryController 中添加管理员接口
@RestController
@RequestMapping("/category")
public class CategoryController {
    // 现有接口...
    
    // 管理员接口
    // POST /category/admin - 添加分类
    // PUT /category/admin/{id} - 更新分类
    // DELETE /category/admin/{id} - 删除分类
}
```

#### 2.2.3 内容审核接口 (中优先级)

需要在 `mybilibili-comment` 服务中新建 `ContentReviewController`：

```java
@RestController
@RequestMapping("/admin/content-review")
public class ContentReviewController {
    // GET /admin/content-review/pending - 获取待审核列表
    // GET /admin/content-review/all - 获取所有内容
    // PUT /admin/content-review/restore/{type}/{id} - 恢复内容
    // DELETE /admin/content-review/{type}/{id} - 删除内容
    // POST /admin/content-review/batch - 批量处理
}
```

#### 2.2.4 统计接口适配 (低优先级)

前端使用的 `/statistics/*` 接口可以映射到 `/creator/stats/*`，但需要注意：
- 创作者统计接口是基于用户ID的，管理员统计需要全局数据
- 需要新建管理员专用的统计接口

## 三、前端修改方案

### 3.1 修改 API 文件

修改 `mybilibili-admin-web/src/api/` 下的各个文件：

#### 3.1.1 `user.js` 修改

```javascript
// 获取用户列表
export const getUserList = (params) => {
  return request({
    url: '/user/admin/list',  // 修改路径
    method: 'get',
    params
  })
}

// 获取用户详情
export const getUserById = (id) => {
  return request({
    url: `/user/admin/${id}`,  // 修改路径
    method: 'get'
  })
}

// 更新用户状态
export const updateUserStatus = (id, status) => {
  return request({
    url: `/user/admin/${id}/status`,  // 修改路径
    method: 'put',
    params: { status }
  })
}

// 重置用户密码
export const resetPassword = (id, newPassword) => {
  return request({
    url: `/user/admin/${id}/password`,  // 修改路径
    method: 'put',
    params: { newPassword }
  })
}
```

#### 3.1.2 `video.js` 修改

```javascript
export const getVideoList = (params) => {
  return request({
    url: '/video/admin/list',  // 修改路径
    method: 'get',
    params
  })
}

export const getVideoById = (id) => {
  return request({
    url: `/video/admin/${id}`,  // 修改路径
    method: 'get'
  })
}

export const deleteVideo = (id) => {
  return request({
    url: `/video/admin/${id}`,  // 修改路径
    method: 'delete'
  })
}

export const deleteVideos = (ids) => {
  return request({
    url: '/video/admin/batch',  // 修改路径
    method: 'delete',
    data: ids
  })
}
```

#### 3.1.3 `comment.js` 修改

```javascript
// 获取评论列表
export const getCommentList = (params) => {
  return request({
    url: '/comment/admin/list',  // 修改路径
    method: 'get',
    params
  })
}

// 获取评论详情
export const getCommentById = (id) => {
  return request({
    url: `/comment/admin/${id}`,  // 修改路径
    method: 'get'
  })
}

// 删除评论
export const deleteComment = (id) => {
  return request({
    url: `/comment/admin/${id}`,  // 修改路径
    method: 'delete'
  })
}

// 更新评论状态
export const updateCommentStatus = (id, status) => {
  return request({
    url: `/comment/admin/${id}/status`,  // 修改路径
    method: 'put',
    params: { status }
  })
}
```

#### 3.1.4 `category.js` 修改

```javascript
// 获取分类列表
export const getCategoryList = (params) => {
  return request({
    url: '/category',  // 保持不变
    method: 'get',
    params
  })
}

// 获取分类详情
export const getCategoryById = (id) => {
  return request({
    url: `/category/${id}`,  // 保持不变
    method: 'get'
  })
}

// 添加分类 (需要后端新建接口)
export const addCategory = (data) => {
  return request({
    url: '/category/admin',  // 新路径
    method: 'post',
    params: data
  })
}

// 更新分类 (需要后端新建接口)
export const updateCategory = (id, data) => {
  return request({
    url: `/category/admin/${id}`,  // 新路径
    method: 'put',
    params: data
  })
}

// 删除分类 (需要后端新建接口)
export const deleteCategory = (id) => {
  return request({
    url: `/category/admin/${id}`,  // 新路径
    method: 'delete'
  })
}
```

### 3.2 网关配置

需要在 `mybilibili-gateway` 中配置路由规则，确保管理员请求正确路由到各个微服务：

```yaml
spring:
  cloud:
    gateway:
      routes:
        # 管理员服务路由
        - id: admin-service
          uri: lb://mybilibili-user
          predicates:
            - Path=/api/admin/**
          filters:
            - StripPrefix=1
        
        # 用户管理路由
        - id: user-admin
          uri: lb://mybilibili-user
          predicates:
            - Path=/api/user/admin/**
          filters:
            - StripPrefix=1
        
        # 视频管理路由
        - id: video-admin
          uri: lb://mybilibili-video
          predicates:
            - Path=/api/video/admin/**
          filters:
            - StripPrefix=1
        
        # 评论管理路由
        - id: comment-admin
          uri: lb://mybilibili-comment
          predicates:
            - Path=/api/comment/admin/**
          filters:
            - StripPrefix=1
        
        # 稿件管理路由
        - id: manuscript-admin
          uri: lb://mybilibili-video
          predicates:
            - Path=/api/manuscript/admin/**
          filters:
            - StripPrefix=1
        
        # 分类管理路由
        - id: category
          uri: lb://mybilibili-video
          predicates:
            - Path=/api/category/**
          filters:
            - StripPrefix=1
```

## 四、实施步骤

### 阶段一：基础适配 (优先级：高)

1. **修改前端 API 文件**
   - 修改 `user.js`、`video.js`、`comment.js` 的接口路径
   - 测试现有功能是否正常

2. **配置网关路由**
   - 添加管理员相关的路由规则
   - 确保请求正确路由到各个微服务

### 阶段二：稿件管理接口 (优先级：高)

1. **新建 ManuscriptAdminController**
   - 实现稿件审核、发布、删除等接口
   - 实现视频处理相关接口

2. **测试稿件管理功能**
   - 测试审核流程
   - 测试视频处理流程

### 阶段三：分类管理接口 (优先级：中)

1. **扩展 CategoryController**
   - 添加分类的增删改接口
   - 测试分类管理功能

### 阶段四：内容审核接口 (优先级：中)

1. **新建 ContentReviewController**
   - 实现内容审核相关接口
   - 测试内容审核功能

### 阶段五：统计接口优化 (优先级：低)

1. **新建管理员统计接口**
   - 实现全局数据统计
   - 适配前端统计页面

## 五、注意事项

1. **权限验证**
   - 确保所有管理员接口都有权限验证
   - 使用 JWT Token 验证管理员身份

2. **数据一致性**
   - 微服务之间的数据同步
   - 分布式事务处理

3. **错误处理**
   - 统一错误响应格式
   - 友好的错误提示

4. **性能优化**
   - 分页查询优化
   - 缓存策略

## 六、总结

通过以上方案，可以实现管理员前端与微服务后端的适配，主要工作包括：

1. ✅ **前端接口路径修改** - 修改 API 文件中的接口路径
2. ⚠️ **后端接口补充** - 新建稿件管理、分类管理、内容审核等接口
3. ✅ **网关路由配置** - 配置管理员请求的路由规则
4. ✅ **权限验证** - 确保管理员接口的安全性

这个方案最大程度地复用了现有的微服务接口，减少了重复开发工作，同时保证了系统的可维护性和扩展性。
