# 视频处理服务移植计划

## 一、现状分析

### 1.1 老项目功能（已验证）
老项目 `mybilibili-admin` 中包含完整的视频处理服务：

1. **VideoProcessServiceImpl** - 视频处理主服务
   - `transcodeVideo()` - 视频转码（使用FFmpeg）
   - `extractAudio()` - 音频提取
   - `generateSubtitle()` - 字幕生成（使用Whisper）
   - `aiSummary()` - AI摘要生成
   - `processAll()` - 全流程处理

2. **AdminSubtitleServiceImpl** - 字幕管理服务
   - `importSystemSubtitle()` - 字幕存入MongoDB
   - 使用 `SubtitleRepository` 操作MongoDB

3. **相关工具类**
   - `FFmpegUtils` - FFmpeg工具类
   - `SubtitleTextUtils` - 字幕文本处理工具

### 1.2 微服务项目现状
微服务项目 `mybilibili-ai` 中的 `VideoProcessConsumer` 消费MQ消息，但处理逻辑不完整：
- 转码、音频提取、字幕生成、AI摘要各有Service
- 但字幕生成后没有存入MongoDB
- FFmpegUtils等工具类可能缺失

## 二、移植任务

### 2.1 移植FFmpegUtils工具类
**来源**：`d:\files\mybilibili\mybilibili-java\mybilibili-admin\src\main\java\com\mybilibili\admin\utils\FFmpegUtils.java`

**目标**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\utils\FFmpegUtils.java`

### 2.2 移植SubtitleTextUtils工具类
**来源**：`d:\files\mybilibili\mybilibili-java\mybilibili-admin\src\main\java\com\mybilibili\admin\utils\SubtitleTextUtils.java`

**目标**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\utils\SubtitleTextUtils.java`

### 2.3 移植UploadFilePathUtils工具类
**来源**：`d:\files\mybilibili\mybilibili-java\mybilibili-admin\src\main\java\com\mybilibili\admin\utils\UploadFilePathUtils.java`

**目标**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-common\src\main\java\com\mybilibili\common\utils\UploadFilePathUtils.java`

### 2.4 移植Subtitle实体类和MongoDB配置
**来源**：老项目的 `Subtitle` 实体类
**目标**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-common\src\main\java\com\mybilibili\common\entity\Subtitle.java`

### 2.5 移植SubtitleRepository
**来源**：`d:\files\mybilibili\mybilibili-java\mybilibili-admin\src\main\java\com\mybilibili\admin\repository\SubtitleRepository.java`

**目标**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\repository\SubtitleRepository.java`

### 2.6 完善视频处理服务
**需要修改**：`VideoTranscodeServiceImpl`, `AudioExtractServiceImpl`, `AiSubtitleServiceImpl`, `AiSummaryServiceImpl`

主要改动：
- 字幕生成后自动存入MongoDB
- 完善各处理阶段的错误处理和状态更新

### 2.7 添加MongoDB配置
在 `mybilibili-ai` 的 application.yml 中添加MongoDB连接配置

## 三、字幕存入MongoDB的逻辑

根据老项目 `AdminSubtitleServiceImpl` 的 `importSystemSubtitle` 方法：

```java
// 1. 读取SRT字幕文件
String srtContent = new String(Files.readAllBytes(Paths.get(subtitlePath)), StandardCharsets.UTF_8);

// 2. 解析SRT为字幕项列表
List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);

// 3. 创建Subtitle对象
Subtitle subtitle = new Subtitle();
subtitle.setVideoId(videoId);
subtitle.setLanguage(language);
subtitle.setFormat("srt");
subtitle.setContent(items);
subtitle.setIsDefault(true);
subtitle.setSource("system"); // 来源：系统生成
subtitle.setStatus(3); // 系统生成状态

// 4. 保存到MongoDB
subtitleRepository.save(subtitle);
```

## 四、详细任务清单

| 序号 | 任务 | 优先级 | 状态 |
|------|------|--------|------|
| 1 | 移植FFmpegUtils工具类 | 高 | 待完成 |
| 2 | 移植SubtitleTextUtils工具类 | 高 | 待完成 |
| 3 | 移植UploadFilePathUtils工具类 | 高 | 待完成 |
| 4 | 移植Subtitle实体类 | 高 | 待完成 |
| 5 | 创建SubtitleRepository | 高 | 待完成 |
| 6 | 添加MongoDB配置 | 高 | 待完成 |
| 7 | 完善AiSubtitleServiceImpl - 字幕存入MongoDB | 高 | 待完成 |
| 8 | 完善其他处理服务 | 中 | 待完成 |

## 五、依赖检查

### 5.1 需要引入的Maven依赖
- spring-boot-starter-data-mongodb
- mybilibili-common（已存在）

### 5.2 Whisper配置
```yaml
ai:
  whisper:
    cli-path: whisper
    model-path: base
    language: zh
    threads: 4
```
