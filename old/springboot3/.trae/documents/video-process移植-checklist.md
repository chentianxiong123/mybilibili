# 视频处理服务移植检查清单

## 工具类移植

- [ ] FFmpegUtils.java 已复制到微服务项目
- [ ] SubtitleTextUtils.java 已复制到微服务项目
- [ ] UploadFilePathUtils.java 已复制到微服务项目
- [ ] 所有import语句正确

## MongoDB相关

- [ ] Subtitle.java 实体类存在于 mybilibili-common
- [ ] SubtitleRepository.java 存在于 mybilibili-ai
- [ ] MongoDB连接配置已添加
- [ ] Spring Data MongoDB依赖已添加

## 服务完善

- [ ] AiSubtitleServiceImpl 字幕生成后存入MongoDB
- [ ] VideoTranscodeServiceImpl 转码逻辑完整
- [ ] AudioExtractServiceImpl 音频提取逻辑完整
- [ ] AiSummaryServiceImpl AI摘要逻辑完整

## 编译验证

- [ ] mybilibili-common 编译通过
- [ ] mybilibili-ai 编译通过
- [ ] 无缺少类或方法的错误

## 功能验证

- [ ] 审核通过后MQ消息正确发送
- [ ] 视频处理流程正确执行
- [ ] 字幕生成后存入MongoDB
- [ ] 视频状态正确更新
