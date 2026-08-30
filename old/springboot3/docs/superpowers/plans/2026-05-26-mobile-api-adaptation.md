# 移动端 API 对接项目接口计划

> **Goal:** 将所有 `src/mobile/api/*.ts` 文件从引用外部假 API (`112.74.99.5:3000`) 改为使用项目已有 API (`/api/*`)，并修正字段映射和响应码适配。

---

## 一、现有 API 盘点（可直接复用）

| API 模块 | 文件 | 可用于移动端 |
|----------|------|------------|
| 通用axios实例 | `api/index.js` | ✅ baseURL=/api，已处理token/响应拦截 |
| 轮播图 | `api/banner.js` | ✅ getHomeBanners() → `/banner-images/home` |
| 推荐视频 | `api/recommend.js` | ✅ getRecommendedVideos(), getHotVideos() |
| 视频详情 | `api/index.js` → videoApi | ✅ getVideoById, getVideoByManuscriptId |
| 分类/分区 | `api/index.js` → categoryApi | ✅ getCategoryList() |
| 搜索 | `api/search.js` | ✅ searchVideos, getSearchSuggestions, getHotSearch |
| 用户/UP主 | `api/index.js` → userApi | ✅ getUserById |
| 历史记录 | `api/history.js` 或 api/index.js | ✅ getHistoryList, clearHistory |
| 弹幕 | `api/index.js` → interactionApi | ✅ sendDanmaku, getDanmakus |
| 直播 | `api/live.js` | ✅ getRoom, getLiveList |

**缺失的 API（需要新建或扩展现有 API）：**
1. 首页分区列表 → 需要新建 `/category` 的子分区接口
2. 排行榜数据 → 需要新建 `ranking` 相关接口
3. 直播间弹幕 WebSocket → 需要复用现有 `/ws` 或新建
4. UP 主投稿视频列表 → 需要新建 `/user/{id}/videos` 接口

---

## 二、字段映射规则

react-bilibili 返回字段 vs 项目字段：

```
视频对象映射：
  aId          → id 或 manuscriptId
  title        → title (一致)
  pic          → coverUrl
  author       → username (用户API) 或 nickname
  mid          → userId
  play         → viewCount
  videoReview  → danmakuCount
  duration     → duration (一致，格式不同需转换)
  description  → description (一致)

评论对象映射：
  rpid         → id
  content      → content (一致)
  ctime        → createTime
  user.face    → avatarUrl
  user.name    → username

直播房间对象映射：
  roomId       → id
  title        → roomName
  cover        → coverUrl
  isLive       → status === 'live'
  onlineNum    → viewerCount
  playUrl      → streamUrl

UP主对象映射：
  mid          → id
  face         → avatarUrl
  name         → username 或 nickname
  follower     → followerCount
  following    → followingCount
  sign         → bio
  level        → level
```

响应码适配：
```
react-bilibili:  code === '1'   → 成功
项目API:         code === 200  → 成功
处理: 在每个 API 调用后统一做 .then(res => { if (res.code === 200) ... })
```

---

## 三、实施计划

### Task A: 重写移动端 API 适配层

**文件:** 全部重写 `src/mobile/api/*.ts`

- **index.ts** — 首页：调用 `banner.js` + `recommend.js`，映射字段
- **video.ts** — 视频详情：调用 `videoApi.getVideoByManuscriptId`
- **channel.ts** — 频道分区：调用 `categoryApi.getCategoryList`，需配合分区接口
- **ranking.ts** — 排行榜：新建 ranking.js API（复用 recommendApi.getHotVideos 按分类）
- **search.ts** — 搜索：直接调用 `searchApi` 的三个方法
- **up-user.ts** — UP主：调用 `userApi.getUserById`
- **live.ts** — 直播：调用 `liveApi.getRoom` + `liveApi.getLiveList`

### Task B: 统一响应码处理

在每个 API 文件中统一处理：
```ts
// 统一处理响应码
const handleRes = (res) => {
  if (res.code === 200) return res.data
  return null
}
```

### Task C: 字段映射适配器

在 `src/mobile/utils/adapter.ts` 创建统一映射：
```ts
// 视频字段映射
export function adaptVideo(raw): Video {
  return {
    aId: raw.id || raw.manuscriptId,
    title: raw.title,
    pic: raw.coverUrl || raw.cover,
    author: raw.username || raw.nickname || raw.author,
    mid: raw.userId,
    play: raw.viewCount || raw.play || 0,
    videoReview: raw.danmakuCount || raw.videoReview || 0,
    duration: formatDuration(raw.duration),
    description: raw.description,
    ...
  }
}

// 直播房间映射
export function adaptLive(raw): Live {
  return {
    roomId: raw.id,
    title: raw.roomName || raw.title,
    cover: raw.coverUrl || raw.cover,
    isLive: raw.status === 'live',
    onlineNum: raw.viewerCount || 0,
    playUrl: raw.streamUrl || raw.playUrl,
    ...
  }
}
```

### Task D: 修复字段名引用（逐个页面）

逐个检查 `src/mobile/views/` 下的 Vue 组件，确保：
1. template 中使用的字段名与 API 返回的字段名对应
2. v-for 遍历的视频对象属性存在
3. router-link 跳转参数名匹配（`/m/video/:aId` → `:aId` 实际传的是什么）

### Task E: 修复样式路径

检查所有 `src/mobile/assets/images/` 引用路径，确保图片资源存在。

### Task F: 验证首页 /m/index 可访问

启动 dev server，访问 `/m/index`，抓包看 API 请求是否成功，页面是否渲染。

---

## 四、执行顺序

1. **Task A** — 重写 API 适配层（核心）
2. **Task B** — 统一响应码处理
3. **Task C** — 字段映射适配器
4. **Task D** — 修复页面组件字段引用
5. **Task E** — 修复样式路径
6. **Task F** — 启动验证

---

## 五、预估工作量

| Task | 内容 | 复杂度 |
|------|------|--------|
| A | 重写 API 适配层（7个文件） | 中 |
| B | 响应码统一处理 | 低 |
| C | 字段映射适配器 | 中 |
| D | 逐页面字段修复（~15个.vue） | 高 |
| E | 样式路径修复 | 低 |
| F | 启动验证 + 修 bug | 中 |

**总计约 2-3 小时工作量。**