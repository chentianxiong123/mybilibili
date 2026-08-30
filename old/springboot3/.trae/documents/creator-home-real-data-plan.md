# 创作中心主页真实数据实现计划

## 一、需求分析

创作中心主页 (`/create-center/home`) 需要展示以下真实数据：
1. **统计数据卡片** - 粉丝总数、总播放量、总评论数、总弹幕数、总点赞数、总分享数、总收藏数、总投币数
2. **最新评论列表** - 展示最近收到的5条评论
3. **观看排行** - 按观看时长排序的粉丝排行
4. **互动排行** - 按互动指标排序的粉丝排行

## 二、现有资源

### 前端
- `CreateCenterView.vue` 已有主页UI框架
- `statsApi.getOverview()` 可复用于获取统计数据
- `creatorApi` 已有评论管理相关接口

### 后端
- `CreatorStatsController` 已有数据统计接口
- `CreatorCommentController` 已有评论管理接口
- 数据库表：`manuscripts`, `comments`, `user_interactions`, `users`

## 三、实现步骤

### 步骤1: 扩展后端API

#### 1.1 在 CreatorStatsController 添加新接口

**获取最新评论接口**
```java
GET /creator/stats/latest-comments?limit=5
```
返回最近收到的评论列表，包含用户信息、评论内容、时间等

**获取粉丝排行接口**
```java
GET /creator/stats/fans-ranking?type=view|interaction&limit=10
```
返回粉丝排行榜数据

#### 1.2 创建新的VO类

**LatestCommentVO** - 已存在，可复用
**FansRankingVO** - 新建，包含粉丝ID、用户名、头像、观看时长/互动指标等

### 步骤2: 扩展Service层

在 `CreatorStatsService` 添加方法：
- `getLatestComments(Integer userId, Integer limit)` - 获取最新评论
- `getFansRanking(Integer userId, String type, Integer limit)` - 获取粉丝排行

### 步骤3: 扩展Mapper层

在 `CreatorStatsMapper` 添加SQL查询：
```sql
-- 获取最新评论（联表查询用户信息和稿件信息）
SELECT c.id, c.content, c.created_at, u.username, u.avatar, m.title as manuscript_title
FROM comments c
JOIN users u ON c.user_id = u.id
JOIN manuscripts m ON c.manuscript_id = m.id
WHERE m.user_id = #{userId}
ORDER BY c.created_at DESC
LIMIT #{limit}

-- 获取观看排行（基于观看历史）
SELECT u.id, u.username, u.avatar, SUM(h.duration) as total_duration
FROM watch_history h
JOIN users u ON h.user_id = u.id
JOIN manuscripts m ON h.manuscript_id = m.id
WHERE m.user_id = #{userId}
GROUP BY u.id
ORDER BY total_duration DESC
LIMIT #{limit}

-- 获取互动排行（基于评论、点赞等）
SELECT u.id, u.username, u.avatar, COUNT(*) as interaction_count
FROM user_interactions i
JOIN users u ON i.user_id = u.id
JOIN manuscripts m ON i.target_id = m.id
WHERE m.user_id = #{userId} AND i.target_type = 'MANUSCRIPT'
GROUP BY u.id
ORDER BY interaction_count DESC
LIMIT #{limit}
```

### 步骤4: 前端API接口

在 `creator.js` 的 `statsApi` 添加：
```javascript
// 获取最新评论
getLatestComments: (limit = 5) => api.get(`/creator/stats/latest-comments?limit=${limit}`),

// 获取粉丝排行
getFansRanking: (type = 'view', limit = 10) => api.get(`/creator/stats/fans-ranking?type=${type}&limit=${limit}`)
```

### 步骤5: 前端页面实现

#### 5.1 修改数据加载逻辑
- 在 `onMounted` 和 `watch(currentActive)` 中调用数据加载函数
- 使用 `statsApi.getOverview()` 获取统计数据
- 使用 `statsApi.getLatestComments()` 获取最新评论
- 使用 `statsApi.getFansRanking()` 获取排行榜

#### 5.2 数据绑定
- 将 `statsData` 绑定到统计数据卡片
- 将 `latestComments` 绑定到评论列表
- 将 `viewRanking` 和 `interactionRanking` 绑定到排行榜

## 四、详细实现清单

### 后端任务
1. [ ] 创建 `FansRankingVO.java` 数据对象
2. [ ] 在 `CreatorStatsMapper` 添加查询方法
3. [ ] 在 `CreatorStatsService` 添加接口方法
4. [ ] 在 `CreatorStatsServiceImpl` 实现方法
5. [ ] 在 `CreatorStatsController` 添加API接口

### 前端任务
1. [ ] 在 `statsApi` 添加新接口
2. [ ] 修改主页数据加载逻辑
3. [ ] 绑定真实数据到UI组件

## 五、数据结构设计

### FansRankingVO
```java
public class FansRankingVO {
    private Integer id;           // 粉丝ID
    private String username;      // 用户名
    private String avatar;        // 头像
    private Long totalDuration;   // 总观看时长（秒）- 观看排行
    private Integer interactionCount; // 互动次数 - 互动排行
}
```

### LatestCommentVO (已存在)
```java
public class LatestCommentVO {
    private Integer id;
    private String username;
    private String avatar;
    private String content;
    private String manuscriptTitle;
    private String time;
}
```

## 六、注意事项

1. 观看历史表可能没有duration字段，需要根据实际情况调整SQL
2. 排行榜数据可能为空，需要处理空数据情况
3. 评论需要按时间倒序排列
4. 需要考虑分页和性能优化
