# 视频上传参数修复任务列表

## Task 1: 修复前端API参数格式
- [x] Task 1.1: 修改manuscript.js中的uploadManuscript函数
  - [x] 1.1.1: 将`video${index}`改为`videos[${index}].file`
  - [x] 1.1.2: 将`videoTitle${index}`改为`videos[${index}].title`
  - [x] 1.1.3: 将`videoOrder${index}`改为`videos[${index}].videoOrder`
  - [x] 1.1.4: 将`videoDuration${index}`改为`videos[${index}].durationSeconds`

## Task 2: 验证修复效果
- [ ] Task 2.1: 重新上传稿件测试
  - [ ] 2.1.1: 上传包含视频的稿件
  - [ ] 2.1.2: 检查后端日志是否显示视频处理流程
  - [ ] 2.1.3: 检查数据库videos表是否有新记录
  - [ ] 2.1.4: 检查服务器目录是否有视频文件

## Task Dependencies
- Task 2 依赖 Task 1 完成
