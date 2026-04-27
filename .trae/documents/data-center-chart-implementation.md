# 数据中心图表功能实现计划

## 目标
在 `/create-center/data` 页面引入真实的 ECharts 图表，替换当前的模拟图表，实现完整的数据可视化功能。

## 当前状态分析

### 前端
- 页面位置: `mybilibili-web/src/views/CreateCenterView.vue`
- 已安装 ECharts: `echarts@5.6.0`
- 当前状态: 使用 CSS 模拟的假图表（mock-chart），位于第 786-822 行
- 已有数据结构:
  - `dataOverview`: 核心数据概览
  - `trendData`: 趋势数据（dates, views, likes, comments）
  - `timeRange`: 时间范围选择器

### 后端
- 控制器: `CreatorStatsController.java`
- 当前接口:
  - `GET /creator/stats` - 获取创作者统计数据
  - `GET /creator/latest-comments` - 获取最新评论
  - `GET /creator/rankings` - 获取排行榜
- 缺失接口:
  - `GET /creator/stats/overview` - 数据概览（前端调用但后端未实现）
  - `GET /creator/stats/trend` - 趋势数据
  - `GET /creator/stats/ranking` - 稿件排行

## 实现步骤

### 第一阶段：后端接口开发

#### 1.1 添加趋势数据接口
- 文件: `CreatorStatsController.java`
- 新增接口: `GET /creator/stats/trend?days={days}`
- 返回数据结构:
```json
{
  "code": 200,
  "data": {
    "dates": ["2026-03-28", "2026-03-29", ...],
    "views": [120, 150, ...],
    "likes": [10, 15, ...],
    "comments": [5, 8, ...],
    "followers": [2, 3, ...]
  }
}
```

#### 1.2 添加数据概览接口
- 文件: `CreatorStatsController.java`
- 新增接口: `GET /creator/stats/overview`
- 返回数据结构:
```json
{
  "code": 200,
  "data": {
    "totalViews": 10000,
    "totalLikes": 500,
    "totalCoins": 100,
    "totalCollections": 200,
    "totalShares": 50,
    "totalComments": 300,
    "totalDanmaku": 400,
    "totalFollowers": 1000,
    "totalManuscripts": 20,
    "viewsIncrease": 100,
    "likesIncrease": 20,
    "updateTime": "2026-04-04"
  }
}
```

#### 1.3 添加稿件排行接口
- 文件: `CreatorStatsController.java`
- 新增接口: `GET /creator/stats/ranking?sortBy={sortBy}&limit={limit}`
- 返回数据结构:
```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "title": "视频标题",
        "coverUrl": "封面URL",
        "viewCount": 1000,
        "likeCount": 100,
        "commentCount": 50,
        "interactionRate": 15.5
      }
    ]
  }
}
```

#### 1.4 实现 Service 层
- 文件: `CreatorStatsService.java`
- 实现数据查询和聚合逻辑

### 第二阶段：前端图表开发

#### 2.1 创建 ECharts 图表组件
- 在数据中心页面引入 ECharts
- 创建播放量趋势折线图
- 支持时间范围切换（7天、30天、90天、1年）

#### 2.2 替换模拟图表
- 删除 mock-chart 相关代码（第 790-820 行）
- 添加真实的 ECharts 图表容器
- 实现图表初始化和更新逻辑

#### 2.3 添加更多图表
- 粉丝增长趋势图
- 互动数据分布图（饼图）
- 稿件表现对比图（柱状图）

### 第三阶段：数据联调

#### 3.1 前后端联调
- 确保接口数据格式正确
- 处理加载状态和错误状态
- 添加数据为空时的提示

#### 3.2 优化用户体验
- 添加图表加载动画
- 实现图表交互功能（tooltip、legend切换）
- 响应式布局适配

## 详细代码修改清单

### 后端修改

1. **CreatorStatsController.java** - 添加三个新接口
2. **CreatorStatsService.java** - 添加服务方法
3. **CreatorStatsMapper.java** (如需要) - 添加数据查询方法

### 前端修改

1. **CreateCenterView.vue**
   - 引入 ECharts
   - 添加图表容器 div
   - 实现图表初始化方法
   - 实现数据加载和图表更新
   - 删除 mock-chart 相关代码和样式

## 预期效果

1. 数据概览页面显示真实的播放量趋势图
2. 图表支持多数据系列（总播放量、粉丝播放量）
3. 时间范围切换时图表动态更新
4. 图表具有交互功能（hover显示详情、legend切换显示）
5. 数据加载时显示 loading 状态
