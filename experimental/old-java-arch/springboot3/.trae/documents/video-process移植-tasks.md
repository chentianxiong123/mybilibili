# 视频处理服务移植任务清单

## 阶段1：工具类移植

- [ ] Task 1: 移植FFmpegUtils工具类
  - 从老项目复制 `FFmpegUtils.java`
  - 检查并修正包路径
  - 添加必要的import

- [ ] Task 2: 移植SubtitleTextUtils工具类
  - 从老项目复制 `SubtitleTextUtils.java`
  - 检查并修正包路径

- [ ] Task 3: 移植UploadFilePathUtils工具类
  - 从老项目复制 `UploadFilePathUtils.java`
  - 检查路径方法是否与微服务一致

## 阶段2：MongoDB相关

- [ ] Task 4: 移植Subtitle实体类
  - 从老项目复制 `Subtitle.java` 到 `mybilibili-common`
  - 确保 `@Document`, `@Id` 等注解正确

- [ ] Task 5: 创建SubtitleRepository
  - 在 `mybilibili-ai` 中创建MongoDB Repository接口
  - 实现 `findByVideoId`, `findByVideoIdAndLanguage` 等方法

- [ ] Task 6: 添加MongoDB配置
  - 在 `mybilibili-ai/application.yml` 添加MongoDB连接配置

## 阶段3：服务完善

- [ ] Task 7: 完善AiSubtitleServiceImpl
  - 字幕生成后调用 `importSystemSubtitle` 存入MongoDB
  - 确保状态正确更新

- [ ] Task 8: 检查其他处理服务
  - 检查 `VideoTranscodeServiceImpl`
  - 检查 `AudioExtractServiceImpl`
  - 检查 `AiSummaryServiceImpl`

## 阶段4：测试验证

- [ ] Task 9: 编译验证
  - 编译 `mybilibili-common`
  - 编译 `mybilibili-ai`

- [ ] Task 10: 功能测试
  - 上传视频触发审核
  - 审核通过后检查MQ消息发送
  - 检查字幕是否存入MongoDB

## 任务依赖关系
- Task 4, 5, 6 依赖于 Task 1, 2, 3
- Task 7 依赖于 Task 4, 5, 6
- Task 8 依赖于 Task 1, 2, 3
- Task 9 依赖于所有移植任务完成
- Task 10 依赖于 Task 9
