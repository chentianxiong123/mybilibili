# Flutter 构建环境踩坑记录

## 环境信息

| 项目 | 值 |
|---|---|
| 操作系统 | Linux x86_64 |
| Flutter 版本 | 3.44.9 stable（2026-08-05 发布，引擎 5a2a6a42cce...） |
| JDK | 21.0.11 |
| Android SDK | 编译目标 36，构建工具 36.0.0 |
| Gradle | 9.3.1 |
| 代理 | 192.168.31.82:7890（Clash），NO_PROXY 排除 *.cn |
| 数据目录 | `/mnt/shared/mybilibili/mybilibili-app-flutter` |

## 问题清单

### 1. Flutter 引擎 Maven 包找不到

**现象**：构建报错 `Could not find io.flutter:arm64_v8a_debug:1.0.0-5a2a6a42...`，搜了 Google Maven、Maven Central、华为云镜像都 404。

**原因**：Flutter 3.44.9 引擎版本太新（仅发布 2 周），国内镜像未同步。

**搜索路径**：
- `https://dl.google.com/dl/android/maven2/` — Google Maven 不放 flutter 引擎
- `https://repo.maven.apache.org/maven2/` — Maven Central 不放
- `https://download.flutter.io/` — 官方源，国内连不上（连接被墙/超时）
- `https://mirrors.huaweicloud.com/flutter/download.flutter.io/` — 华为云镜像，404（未同步新版）

**解决**：切到腾讯云镜像（已验证此版本存在，响应 0.1-1s）。

```bash
export FLUTTER_STORAGE_BASE_URL="https://mirrors.cloud.tencent.com/flutter"
```

**验证结果**：

| 镜像 | arm64_v8a_debug.pom | 速度 |
|---|---|---|
| 华为云 | 404 | 0.2s |
| 腾讯云 | 200 | 0.09s |
| 阿里云 | 404 | — |
| 清华 | 403 | — |
| 中科大/SJTU | 404 | — |

### 2. Dart Pub 包下载失败

**现象**：`flutter pub get` 报 `version solving failed` 或 `authorization failed`。

**原因**：`PUB_HOSTED_URL` 原指向清华镜像 `https://mirrors.tuna.tsinghua.edu.cn/dart-pub`，新版 `flutter_lints` 等包不存在或需要认证。

**解决**：切到官方 pub.dev（走代理，2.5s 响应）。

```bash
export PUB_HOSTED_URL="https://pub.dev"
```

### 3. media_kit mpv jar 下载超时

**现象**：Gradle 构建时下载 `https://github.com/media-kit/libmpv-android-video-build/releases/download/v1.1.7/default-arm64-v8a.jar` 等 4 个 jar 文件，超时或连接失败。

**原因**：GitHub 国内访问慢，30s 只能下 3MB（总需 5.7MB+）。

**解决**：手动下载全部 4 个 jar，校验 MD5 后放入构建缓存目录：

```bash
# 下载目录：/mnt/shared/.../build/media_kit_libs_android_video/v1.1.7/
# 4 个文件及 MD5（已校验通过）：
# default-arm64-v8a.jar  83df25b61193af8fa815e373143ac9af
# default-armeabi-v7a.jar  22e21526fefc0a2b8f17adbec9f57590
# default-x86_64.jar  6fa26bf0459b11f1c0b0dbc29e5b940d
# default-x86.jar  0d742b756dc9d1fcd84ea271d8b68f32
```

### 4. Kotlin/JVM 目标版本冲突

**现象**：`Inconsistent JVM Target Compatibility Between Java and Kotlin Tasks`，gradle 报 `compileDebugJavaWithJavac (11) and compileDebugKotlin (1.8)` 不匹配。

**原因**：`volume_controller` 和 `screen_brightness_android` 两个老插件用 Kotlin 1.7.21、设置 `jvmTarget = '1.8'`，但 JDK 21 默认 Java 目标 11，Kotlin 2.x 编译器要求两者一致。

**解决**：直接补丁 pub-cache 里的插件 build.gradle 文件：

```diff
# /home/a1/.pub-cache/hosted/pub.dev/volume_controller-2.0.8/android/build.gradle
- compileSdkVersion 31
+ compileSdkVersion 36
- kotlinOptions { jvmTarget = '1.8' }
+ kotlinOptions { jvmTarget = '17' }
+ compileOptions {
+     sourceCompatibility JavaVersion.VERSION_17
+     targetCompatibility JavaVersion.VERSION_17
+ }

# /home/a1/.pub-cache/hosted/pub.dev/screen_brightness_android-0.1.0+2/android/build.gradle
- compileSdkVersion 31
+ compileSdkVersion 36
- compileOptions { sourceCompatibility VERSION_1_8; targetCompatibility VERSION_1_8 }
+ compileOptions { sourceCompatibility VERSION_17; targetCompatibility VERSION_17 }
- kotlinOptions { jvmTarget = '1.8' }
+ kotlinOptions { jvmTarget = '17' }
```

### 5. 插件 compileSdk 太旧

**现象**：`22 issues were found when checking AAR metadata`，插件编译目标 31，依赖需要 36+。

**涉及的插件**：

| 插件 | 旧值 | 新值 |
|---|---|---|
| `media_kit_video` | 31 | 36 |
| `volume_controller` | 31 | 36 |
| `screen_brightness_android` | 31 | 36 |

**解决**：补丁每个插件的 `compileSdkVersion 31` → `compileSdkVersion 36`。

### 6. Release 版缺少 INTERNET 权限

**现象**：App 启动后报 `SocketException: Operation not permitted, errno = 1`，连接后端失败。

**原因**：Flutter 只在 debug 版自动加 `INTERNET` 权限，release 版 `AndroidManifest.xml` 必须显式声明。

**解决**：在主 manifest 加一行：

```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <uses-permission android:name="android.permission.INTERNET"/>
    <application ...>
```

### 7. API 地址写 localhost（容器内不通）

**现象**：App 连接 `localhost:8080`，容器内指向自己而非宿主机。

**原因**：App 硬编码了 `http://localhost:8080/api/v1`，在 Waydroid 容器内 localhost 是本机（Android 容器），不是宿主机后端。

**解决**：改两个文件：

| 文件 | 行 | 旧值 | 新值 |
|---|---|---|---|
| `lib/core/api/client.dart:7` | baseUrl | `http://localhost:8080` | `http://192.168.31.204:8080` |
| `lib/core/api/live_api.dart:42` | streamUrl | `http://localhost:8080` | `http://192.168.31.204:8080` |

### 8. Waydroid 平板样式

**现象**：Waydroid 显示为平板布局（1024x568，密度 180）。

**原因**：默认分辨率 1024x568，密度 180dpi，最小宽度 505dp，Android 呈现平板 UI。

**解决**：临时覆盖（重启后失效）：

```bash
adb shell wm size 1080x1920
adb shell wm density 420
```

持久化需要编辑 `/var/lib/waydroid/waydroid.cfg`，在 `[properties]` 段加：

```ini
[properties]
ro.sf.lcd_density = 420
```

## 最终环境配置

### 永久环境变量（`~/.bashrc`）

```bash
export FLUTTER_STORAGE_BASE_URL="https://mirrors.cloud.tencent.com/flutter"
export PUB_HOSTED_URL="https://pub.dev"
```

### 编译命令

```bash
# Debug 全架构（192MB，不适合发布）
flutter build apk --debug

# Release 按架构拆分（推荐）
flutter build apk --release --split-per-abi
```

### 产物位置

```bash
# Debug
build/app/outputs/flutter-apk/app-debug.apk                    # 192MB

# Release（按架构拆分）
build/app/outputs/flutter-apk/app-arm64-v8a-release.apk       # 32MB（现代手机）
build/app/outputs/flutter-apk/app-armeabi-v7a-release.apk     # 29MB（老设备，已弃用）
build/app/outputs/flutter-apk/app-x86_64-release.apk          # 37MB（模拟器/Waydroid）
```

## 后续维护注意事项

1. **`flutter pub upgrade` 后**：pub-cache 清空或重建，需重新打 `volume_controller`、`screen_brightness_android`、`media_kit_video` 三个插件的补丁
2. **Flutter SDK 升级后**：如果引擎版本更新，需先确认腾讯云镜像是否已有新版本——`curl -I "https://mirrors.cloud.tencent.com/flutter/download.flutter.io/io/flutter/arm64_v8a_debug/1.0.0-新哈希/...pom"`
3. **换网络环境后**：宿主机 IP 变化，需更新 `client.dart` 和 `live_api.dart` 中的 `192.168.31.204` 为新 IP
4. **`flutter clean` 后**：media_kit 的 mpv jar 缓存被清，需重新预下载或首次构建多等 3 分钟
5. **Waydroid 重启后**：分辨率/密度若未持久化，需重新 `adb shell wm size 1080x1920 && wm density 420`