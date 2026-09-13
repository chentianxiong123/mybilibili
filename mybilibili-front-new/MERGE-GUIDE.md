# old → new 合并文档

## 一、项目概况

| | old（mybilibili-front） | new（mybilibili-front-new） |
|---|---|---|
| 技术栈 | Nuxt 4 + Vue 3 + Element Plus | Nuxt 4 + Vue 3 + Element Plus |
| UI 风格 | bilibili 原版风格 | teriteri 风格 |
| 状态管理 | Pinia（3 个 store） | Pinia + Vuex shim（1 个 teriteri store） |
| SSR | 多页面 SSR + 部分 ssr:false | 全站 ssr:false |
| 组件扫描 | 自动扫描 components | components:false 手动 import |
| 数据层 | api/client.ts 封装 + composables | teriteri-src/network/request.js + teriteri store |
| 样式 | @mybilibili/ui + 全局 CSS 变量 | teriteri assets（CSS 变量） |

---

## 二、功能对比

### new 独有（teriteri 迁移，old 没有）
- 首页冬季动画头图 + 频道栏 + 轮播
- 弹幕播放器 + 弹幕发送
- teriteri 风格评论系统（点赞/点踩/表情/回复/楼中楼）
- teriteri 风格搜索（视频 tab / 用户 tab / 排序）
- teriteri 风格空间（视频/动态/粉丝/收藏/设置）
- teriteri 风格消息（悄悄话/系统/回复/点赞/提醒/配置）
- teriteri 风格创作中心（上传/稿件/评论/弹幕/数据/审核）
- teriteri 风格账户设置（头像/信息/安全）
- 404 页（[...slug].vue catch-all）
- Pinia store + Vuex shim
- plugins/teriteri.client.ts（全局注册 $get/$post/$axios/$message + EP 图标）
- server/api/[...path].ts（反向代理层，PATH_MAP）
- server/routes/uploads/[...path].ts（MinIO 静态代理）
- server/middleware/auth.global.ts（JWT→X-User-Id 注入）

### old 独有（new 缺少）
- 登录弹窗（密码/邮箱验证码/图形验证码 + Token 刷新）
- 注册弹窗 + 兴趣选择弹窗
- 直播页（live/）
- 动态页（dynamic/）
- 收藏夹管理（collection/）
- 观看历史（history）
- 注册页（register）
- 个人中心（personal-center/）
- 个人主页（profile/）
- @mybilibili/ui 组件库
- layouts/（home/none/simple）
- composables/（18 个 hook）
- components/（54 个业务组件）
- api/client.ts（axios 封装 + 拦截器）

---

## 三、合并操作

### 1. 复制 old 独有目录进 new

```
# 复制 old 的 pages（覆盖前检查冲突）
cp -r mybilibili-front/apps/web/app/pages/live      mybilibili-front-new/apps/web/app/pages/
cp -r mybilibili-front/apps/web/app/pages/dynamic   mybilibili-front-new/apps/web/app/pages/
cp -r mybilibili-front/apps/web/app/pages/collection mybilibili-front-new/apps/web/app/pages/
cp mybilibili-front/apps/web/app/pages/history.vue  mybilibili-front-new/apps/web/app/pages/
cp mybilibili-front/apps/web/app/pages/register.vue mybilibili-front-new/apps/web/app/pages/
cp -r mybilibili-front/apps/web/app/pages/personal-center mybilibili-front-new/apps/web/app/pages/
cp -r mybilibili-front/apps/web/app/pages/profile   mybilibili-front-new/apps/web/app/pages/
cp mybilibili-front/apps/web/app/pages/404.vue      mybilibili-front-new/apps/web/app/pages/
cp mybilibili-front/apps/web/app/pages/user          mybilibili-front-new/apps/web/app/pages/user
```

**冲突文件（new 已有同名文件，需要决定保留哪个）：**
- `pages/search.vue` vs `pages/search/`（teriteri 搜索目录）→ 保留 new，old 搜索功能由 teriteri 替代
- `pages/create-center/[...slug].vue` vs `pages/platform/`（teriteri 创作中心）→ 保留 new，old 创作中心由 teriteri 替代

### 2. 复制 old 独有基础设施

```
# 复制 old 的 utils（完全一致，无需操作）
# 复制 old 的 composables（new 缺少）
cp -r mybilibili-front/apps/web/app/composables/* mybilibili-front-new/apps/web/app/composables/

# 复制 old 的 layouts（new 缺少）
cp -r mybilibili-front/apps/web/app/layouts/* mybilibili-front-new/apps/web/app/layouts/

# 复制 old 的 components（new 只有 6 个，old 有 54 个）
cp -r mybilibili-front/apps/web/app/components/* mybilibili-front-new/apps/web/app/components/

# 复制 old 的 api（new 已有同名文件，跳过）
# 复制 old 的 middleware（new 已有同名文件，跳过）
```

### 3. 修改 nuxt.config

```typescript
// mybilibili-front-new/apps/web/nuxt.config.ts 需要修改：
// 1. 添加 old 的 @mybilibili/ui 引用
css: ['@/assets/teriteri/css/base.css', '@mybilibili/ui',
  'element-plus/es/components/message-box/style/css',
  'element-plus/es/components/notification/style/css',
  'element-plus/es/components/loading/style/css'],

// 2. 添加 old 的路由 SSR 控制
routeRules: {
  '/': { ssr: false },
  '/message': { redirect: '/message/private' },
  '/dynamic/**': { ssr: false },
  '/profile/**': { ssr: false },
  '/personal-center/**': { ssr: false },
  '/history': { ssr: false },
  '/avatar': { ssr: false },
  '/live/**': { ssr: false },
  '/create-center/**': { ssr: false },
  '/manuscript/**': { ssr: false },
  '/login': { ssr: false },
  '/collections': { ssr: false },
  '/collection/**': { ssr: false },
},

// 3. 添加 old 的路由重定向
routeRules: {
  '/message': { redirect: '/message/private' },
}
```

### 4. 修改 app.vue

old 的 `app.vue` 包含登录弹窗 + 兴趣弹窗 + Token 刷新 + 缩放适配。
new 的 `app.vue` 只有 loading 屏 + teriteri 初始化。

**方案：合并两者**

```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>

  <!-- loading 屏 -->
  <div class="loading-mark" :class="isMarkShow ? 'show' : 'hide'" :style="`display: ${markDisplay};`">
    <div class="loading-box">
      <img src="@/assets/teriteri/img/loading.gif" alt="">
    </div>
  </div>

  <!-- 复制 old 的登录弹窗 HTML -->
  <ClientOnly>
  <el-dialog v-model="showLoginDialog" ...>
    <!-- old 的登录/注册表单 -->
  </el-dialog>
  <el-dialog v-model="showInterestDialog" ...>
    <!-- old 的兴趣选择 -->
  </el-dialog>
  </ClientOnly>
</template>

<script setup>
// 合并 new 的 teriteri 初始化逻辑
// + old 的登录弹窗/Token 刷新/缩放适配逻辑
</script>
```

### 5. 修改 components: false

new 的 nuxt.config 有 `components: false`，阻止了自动扫描。

复制进来的 components 需要手动 import，或者删除 `components: false`。

**方案：删除 `components: false`**（让 Nuxt 恢复自动扫描）

---

## 四、冲突文件清单

| 文件 | old | new | 合并策略 |
|---|---|---|---|
| `app.vue` | 登录弹窗 + 兴趣弹窗 + Token 刷新 | loading 屏 + teriteri 初始化 | 合并两者 |
| `nuxt.config.ts` | @mybilibili/ui + 路由守卫 | teriteri CSS + components:false | 合并两者 |
| `stores/user.ts` | '@/utils/auth' | '@/teriteri-src/utils/auth' | 需决定用哪个 auth |
| `pages/search.vue` | bilibili 搜索 | - | 被 teriteri/search/ 替代 |
| `pages/create-center/` | bilibili 创作中心 | - | 被 teriteri/platform/ 替代 |
| `components/` | 54 个业务组件 | 6 个（已有同名） | 合并，解决同名冲突 |

---

## 五、执行顺序

1. 删除 `components: false`（恢复自动扫描）
2. 复制 old 的 `components/` 进 new（解决同名冲突）
3. 复制 old 的 `composables/` 进 new
4. 复制 old 的 `layouts/` 进 new
5. 复制 old 的 `pages/live`、`pages/dynamic`、`pages/collection`、`pages/history`、`pages/register`、`pages/personal-center`、`pages/profile`、`pages/user`、`pages/404.vue` 进 new
6. 修改 `nuxt.config`（添加 old 的路由规则）
7. 修改 `app.vue`（合并登录弹窗）
8. 测试所有功能
