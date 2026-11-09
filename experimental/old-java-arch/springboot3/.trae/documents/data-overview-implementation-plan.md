# 数据概览页面实现计划

## 一、需求分析

数据概览页面 (`/create-center/data`) 需要展示创作者的核心数据指标，包括：
1. 核心数据概览（播放量、粉丝、点赞、收藏、投币、评论、弹幕、分享等）
2. 数据趋势图表（近7天/30天/90天的播放量趋势）
3. 账号诊断（雷达图展示各项指标对比）
4. 稿件分析（稿件表现排行）
5. 粉丝分析（粉丝增长趋势）

## 二、现有资源

### 前端
- `CreateCenterView.vue` 已有数据概览的基础UI框架
- 已有 `CreatorStatsVO` 和 `FansStatsVO` 数据结构
- 已有 `creator.js` API文件

### 后端
- `ManuscriptController` 已有稿件统计接口
- 数据库表：
  - `manuscripts` - 包含播放量、点赞、投币、收藏、分享、评论、弹幕等字段
  - `user_interactions` - 用户互动记录
  - `users` - 用户信息

### 数据库字段
```
manuscripts表:
- view_count (播放量)
- like_count (点赞数)
- coin_count (投币数)
- collect_count (收藏数)
- share_count (分享数)
- comment_count (评论数)
- danmaku_count (弹幕数)
- status (状态)
- upload_time (上传时间)
```

## 三、实现步骤

### 步骤1: 创建后端数据统计API

#### 1.1 创建 CreatorStatsController
位置: `mybilibili-cloud/mybilibili-video/src/main/java/com/mybilibili/video/controller/CreatorStatsController.java`

提供以下接口：
- `GET /creator/stats/overview` - 获取创作者核心数据概览
- `GET /creator/stats/trend` - 获取数据趋势（按时间范围）
- `GET /creator/stats/ranking` - 获取稿件排行

#### 1.2 创建 CreatorStatsService
位置: `mybilibili-cloud/mybilibili-video/src/main/java/com/mybilibili/video/service/CreatorStatsService.java`

实现以下方法：
- `getCreatorOverview(Integer userId)` - 汇总所有核心数据
- `getPlayTrend(Integer userId, String startDate, String endDate)` - 播放量趋势
- `getManuscriptRanking(Integer userId, Integer limit)` - 稿件排行

#### 1.3 数据统计SQL实现
需要聚合查询：
```sql
-- 总播放量、点赞、投币、收藏、分享、评论、弹幕
SELECT 
  SUM(view_count) as totalViews,
  SUM(like_count) as totalLikes,
  SUM(coin_count) as totalCoins,
  SUM(collect_count) as totalCollections,
  SUM(share_count) as totalShares,
  SUM(comment_count) as totalComments,
  SUM(danmaku_count) as totalDanmaku,
  COUNT(*) as totalManuscripts
FROM manuscripts 
WHERE user_id = ? AND status = 3
```

### 步骤2: 配置网关路由

在 `mybilibili-gateway` 的 `application.yml` 添加路由配置：
```yaml
- id: creator-stats
  uri: lb://mybilibili-video
  predicates:
    - Path=/api/creator/stats/**
  filters:
    - StripPrefix=1
```

### 步骤3: 前端API接口

在 `mybilibili-web/src/api/creator.js` 添加：
```javascript
export const statsApi = {
  // 获取数据概览
  getOverview: () => api.get('/creator/stats/overview'),
  
  // 获取趋势数据
  getTrend: (params) => api.get('/creator/stats/trend', { params }),
  
  // 获取稿件排行
  getRanking: (limit) => api.get('/creator/stats/ranking', { params: { limit } })
}
```

### 步骤4: 前端页面实现

#### 4.1 数据概览Tab
- 核心数据卡片（9个指标）
- 播放量趋势图（使用ECharts）
- 时间范围选择器

#### 4.2 账号诊断Tab
- 雷达图展示各项指标
- 与同类UP主对比

#### 4.3 稿件分析Tab
- 稿件表现排行
- 播放量TOP10
- 互动率TOP10

#### 4.4 粉丝分析Tab
- 粉丝增长趋势
- 粉丝画像

### 步骤5: 集成ECharts图表库

安装ECharts：
```bash
npm install echarts
```

创建图表组件：
- 折线图（趋势）
- 雷达图（诊断）
- 柱状图（排行）

## 四、详细实现清单

### 后端任务
1. [ ] 创建 `CreatorStatsController.java`
2. [ ] 创建 `CreatorStatsService.java` 接口
3. [ ] 创建 `CreatorStatsServiceImpl.java` 实现
4. [ ] 创建 `CreatorOverviewVO.java` 数据返回对象
5. [ ] 创建 `TrendDataVO.java` 趋势数据对象
6. [ ] 配置网关路由

### 前端任务
1. [ ] 安装ECharts依赖
2. [ ] 创建 `statsApi` API接口
3. [ ] 实现数据概览Tab
4. [ ] 实现账号诊断Tab
5. [ ] 实现稿件分析Tab
6. [ ] 实现粉丝分析Tab
7. [ ] 添加加载状态和错误处理

## 五、数据结构设计

### CreatorOverviewVO
```java
public class CreatorOverviewVO {
    // 核心指标
    private Integer totalViews;        // 总播放量
    private Integer totalLikes;        // 总点赞
    private Integer totalCoins;        // 总投币
    private Integer totalCollections;  // 总收藏
    private Integer totalShares;       // 总分享
    private Integer totalComments;     // 总评论
    private Integer totalDanmaku;      // 总弹幕
    private Integer totalFollowers;    // 粉丝数
    private Integer totalManuscripts;  // 稿件数
    
    // 增量数据（近7天）
    private Integer viewsIncrease;
    private Integer likesIncrease;
    private Integer followersIncrease;
    
    // 更新时间
    private String updateTime;
}
```

### TrendDataVO
```java
public class TrendDataVO {
    private List<String> dates;        // 日期列表
    private List<Integer> views;       // 播放量列表
    private List<Integer> likes;       // 点赞列表
    private List<Integer> followers;   // 粉丝列表
}
```

## 六、预估工作量

- 后端开发：约2-3小时
- 前端开发：约3-4小时
- 测试调试：约1-2小时

## 七、注意事项

1. 数据统计需要考虑性能，大量数据时使用缓存
2. 趋势数据需要按天聚合，可能需要定时任务预计算
3. 前端图表需要响应式适配
4. 需要处理数据为空的情况
