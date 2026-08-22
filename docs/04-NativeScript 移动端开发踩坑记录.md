# NativeScript-Vue 移动端开发踩坑记录

## 项目概况

- **项目路径**: `/mnt/shared/mybilibili/mybilibili-ns/`
- **技术栈**: NativeScript 9.0.7 + @nativescript/core ~9.0.0 + nativescript-vue ^3.0.2
- **模板**: `@nativescript/template-blank-vue`（官方 Vue 模板）
- **目标平台**: Android（x86_64 模拟器 / arm64-v8a 真机）
- **构建输出**: `platforms/android/app/build/outputs/apk/debug/app-debug.apk`

## 目录结构

```
app/
├── app.ts              # 入口：createApp(App).start()
├── app.css             # 全局样式
├── App.vue             # 根组件：GridLayout(Frame + BottomTabBar)
├── components/         # 共享组件
│   ├── BottomTabBar.vue
│   ├── VideoItem.vue
│   └── Home.vue        # 模板自带，已废弃
├── views/              # 页面（25 个）
│   ├── home/Index.vue          # 首页（分区 TabBar + 轮播图 + 视频列表）
│   ├── video/Detail.vue        # 视频详情（WebView 播放器 + 原生 UI）
│   ├── live/Room.vue           # 直播间（WebView 播放器 + 聊天区）
│   ├── search/Search.vue       # 搜索
│   ├── space/Space.vue         # 我的
│   ├── dynamic/Index.vue       # 动态
│   ├── message/Message.vue     # 消息列表
│   ├── message/Chat.vue        # 聊天详情
│   ├── creator/Center.vue      # 创作中心
│   ├── channel/Channel.vue     # 频道分区
│   ├── ranking/Ranking.vue     # 排行榜
│   ├── mall/Index.vue          # 会员购（占位）
│   ├── Login.vue               # 登录
│   └── NotFound.vue            # 404
├── api/                # API 层（15 个模块，基于 fetch）
│   ├── client.ts       # HTTP 客户端，baseURL: /api/v1
│   ├── index.ts        # 首页 API
│   ├── video.ts / live.ts / search.ts / user.ts / ...
│   └── ...
├── utils/              # 工具函数
│   ├── storage.ts      # ApplicationSettings 封装（替代 localStorage）
│   ├── format.ts       # 格式化函数
│   └── theme.ts        # 主题常量
└── assets/
    └── player.html     # WebView 视频播放器（ArtPlayer + hls.js + 弹幕插件）
```

## 环境搭建

### 系统要求

| 组件 | 版本 | 说明 |
|---|---|---|
| Node.js | v22+ | 必须 |
| npm | 全局安装 `nativescript` CLI |
| Java | JDK 21 | 编译 Android |
| Android SDK | target 35, build-tools 35.0.0 | 必须 |
| Gradle | 9.3.1 | SDK 自带 |

### 安装 NativeScript CLI

```bash
npm install -g nativescript
```

注意：CLI 包名是 `nativescript` 而不是 `@nativescript/cli`。

### 创建项目

```bash
ns create myapp --template @nativescript/template-blank-vue
```

### 生成 Webpack 配置

```bash
cd myapp
npx nativescript-webpack init
```

这会在项目根目录生成 `webpack.config.js`。

## 踩坑记录

### 坑 1：环境检查不通过 — emulator 不存在

**现象**：
```
✖ WARNING: The Android SDK is not installed or is not configured properly.
Your environment is not configured properly and you will not be able to execute local builds.
```

**原因**：`ns build` 会检查 `$ANDROID_HOME/emulator/emulator` 是否存在，并执行 `emulator -help` 验证。如果 SDK 没有安装 emulator，检查会失败并阻止构建。

**解决**：创建一个虚拟的 emulator 可执行文件：
```bash
mkdir -p $ANDROID_HOME/emulator
cat > $ANDROID_HOME/emulator/emulator << 'EOF'
#!/bin/bash
echo "usage: emulator -help"
EOF
chmod +x $ANDROID_HOME/emulator/emulator
```

### 坑 2：Gradle 依赖下载被墙

**现象**：构建时卡在 `Could not resolve com.android.tools.lint:lint-gradle` 并超时。

**原因**：Gradle 的 `google()` 仓库指向 `maven.google.com`，在墙内无法访问。

**解决**：创建全局 Gradle 初始化脚本，不走修改项目文件的方式：

`~/.gradle/init.gradle.kts`：
```kotlin
settingsEvaluated {
    pluginManagement {
        repositories {
            maven(url = "https://maven.aliyun.com/repository/gradle-plugin")
            maven(url = "https://maven.aliyun.com/repository/google")
            maven(url = "https://maven.aliyun.com/repository/central")
        }
    }
}

allprojects {
    buildscript {
        repositories {
            maven(url = "https://maven.aliyun.com/repository/google")
            maven(url = "https://maven.aliyun.com/repository/public")
            maven(url = "https://maven.aliyun.com/repository/gradle-plugin")
        }
    }
    repositories {
        maven(url = "https://maven.aliyun.com/repository/google")
        maven(url = "https://maven.aliyun.com/repository/public")
        maven(url = "https://maven.aliyun.com/repository/gradle-plugin")
    }
}
```

**原理**：`init.gradle.kts` 是 Gradle 的全局初始化脚本，自动应用到所有项目。项目里的 `google()` 和 `mavenCentral()` 在运行时会被 Aliyun 镜像拦截，不需要改项目文件。

### 坑 3：`@nativescript/android` 运行时未安装

**现象**：`ns build` 报错 `Cannot find a compatible Android SDK for compilation`，且列出不存在的 SDK 版本范围。

**原因**：NativeScript Android 运行时 `@nativescript/android` 未在 `node_modules` 中，或者版本不兼容。

**解决**：确保 `package.json` 中 `@nativescript/core` 版本与 Android 运行时匹配。如果缺失，手动安装：
```bash
npm install @nativescript/android --save
```

### 坑 4：Webpack 配置缺失

**现象**：
```
The bundler configuration file webpack.config.js does not exist.
```

**解决**：使用 `nativescript-webpack` 工具生成：
```bash
npx nativescript-webpack init
```

### 坑 5：TypeScript 版本兼容

**现象**：TypeScript 编译报错，如 `ApplicationSettings.getString(key, null)` 参数类型不匹配。

**原因**：`@nativescript/core` 的类型定义与 TypeScript 版本不兼容。

**解决**：检查并修正类型错误。例如 `getString(key, null)` → `getString(key)`（第二个参数不接受 null）。

### 坑 6：`App_Resources/Android/app.gradle` 配置错误

**现象**：构建时 Gradle 报错 build-tools 版本。

**原因**：`buildToolsVersion "35"` 缺少次要版本号，正确的格式是 `"35.0.0"`。

**解决**：
```groovy
android {
  compileSdkVersion 35
  buildToolsVersion "35.0.0"  // 不要写成 "35"
}
```

## 架构设计

### 导航方案

使用自定义底部导航栏（5 个 Tab），而不是 NativeScript 的 `<TabView>`，因为：
1. 中间需要异形 "+" 发布按钮（TabView 不支持）
2. 可以精确控制样式和交互

```
App.vue
└── GridLayout(rows="*, auto")
    ├── Frame (row=0)        # 内容区域，每个 Tab 用 $navigateTo 切换
    └── BottomTabBar (row=1) # 自定义底部栏
```

页面间导航：
- 普通页面跳转：`$navigateTo(PageComponent, { props })`
- 全屏页面（视频/直播/搜索/登录）：`$showModal(PageComponent, { fullscreen: true, props })`
- 返回：`$navigateBack()`

### 视频播放方案

混合架构：视频播放器用 WebView 嵌入，其余 UI 用原生 NativeScript 控件。

```
VideoDetail.vue
├── WebView (16:9)          → 加载 player.html（ArtPlayer + hls.js + 弹幕插件）
├── 原生 UI（UP主卡片、点赞/投币/收藏、评论列表、推荐视频）
└── ScrollView 包裹所有内容
```

**WebView 通信协议**：
```
NativeScript → WebView: webView.evaluateJavaScript("initVideo({...})")
WebView → NativeScript: window.location = 'native://action?param=value'（通过 loadStarted 拦截）
```

### API 层

使用 `fetch` 替代 axios（NativeScript 内置支持），`@nativescript/core/application-settings` 替代 localStorage。

```typescript
// client.ts
const BASE_URL = 'http://192.168.31.204:8080'
const api = {
  get: (url) => fetch(BASE_URL + url, { headers: {...} }),
  post: (url, data) => fetch(BASE_URL + url, { method: 'POST', body: JSON.stringify(data), headers: {...} }),
  // ...
}
```

## 构建 APK

### Debug 构建

```bash
ns build android --debug --bundle --for-device
```

### Release 构建（需要 keystore）

```bash
ns build android --release --bundle --env.uglify --key-store-* ...
```

### 安装到设备

```bash
adb install platforms/android/app/build/outputs/apk/debug/app-debug.apk
```

## 常见问题

### Q: WebView 加载本地 HTML 文件路径是什么？

A: 使用 `~/assets/player.html`。NativeScript 的 `~` 前缀指向 `app/` 目录，构建时会自动打包到 APK 的 assets 中。

### Q: 为什么不用 `vue-router`？

A: NativeScript-Vue 不使用 vue-router。导航通过 `$navigateTo` / `$navigateBack` / `$showModal` 实现，每个 Frame 维护独立的导航栈。

### Q: 图片资源路径怎么写？

A: 使用 `~` 前缀引用本地资源：`<Image src="~/assets/logo.png" />`。远程图片直接传 URL。

### Q: 样式和 Web CSS 有什么不同？

A: NativeScript CSS 是 CSS 子集：
- 不支持 flexbox（用 `FlexboxLayout` 组件替代）
- 不支持 `background` 缩写（用 `background-color`）
- 不支持 `display`、`position`、`float`
- 数值不需要 px 单位（直接写数字）
- 使用 `border-width`、`border-color`、`border-radius` 替代 `border` 缩写

## 维护提醒

### 依赖更新

```
ns update          # 检查并更新 NativeScript 版本
npm update         # 更新 npm 依赖
```

### 清理项目

```bash
ns clean           # 清理 hooks、platforms、node_modules
```

### 宿主机 IP 变化

如果宿主机 IP 变化，需要更新 `app/api/client.ts` 中的 `BASE_URL`。

### 插件不兼容

旧版 NativeScript 插件可能不支持 `@nativescript/core ~9.0.0`。安装前检查插件的 `nativescript-vue` 和 `@nativescript/core` 版本兼容性。

## 对比：Flutter vs NativeScript-Vue

| 维度 | Flutter | NativeScript-Vue |
|---|---|---|
| UI 控件 | 自绘引擎（Skia） | 原生 Android 控件 |
| 构建环境 | 需要下载引擎 blob（被墙） | 标准 Gradle + Maven |
| 镜像依赖 | `flutter-archive` 需要镜像 | 阿里云 Maven 镜像可用 |
| APK 体积 | ~32MB（arm64） | ~10MB（debug） |
| 开发语言 | Dart | TypeScript + Vue |
| 热重载 | 支持 | 不支持（需重新构建） |
| 稳定性 | 引擎版本耦合高 | 原生控件，系统级稳定 |