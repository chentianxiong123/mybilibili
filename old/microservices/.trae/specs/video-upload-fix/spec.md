# 视频上传参数修复 Spec

## Why
前端上传视频时，参数名格式与后端Controller期望的格式不匹配，导致视频文件无法被后端接收，稿件上传成功但视频列表为空。

## 问题分析
**前端发送的参数**：
- `video0`, `video1`... (视频文件)
- `videoTitle0`, `videoTitle1`... (标题)
- `videoOrder0`, `videoOrder1`... (排序)
- `videoDuration0`, `videoDuration1`... (时长)

**后端Controller期望的参数**：
- `videos[0].file`, `videos[1].file`... (视频文件)
- `videos[0].title`, `videos[1].title`... (标题)
- `videos[0].videoOrder`, `videos[1].videoOrder`... (排序)
- `videos[0].durationSeconds`, `videos[1].durationSeconds`... (时长)

## What Changes
- 修改前端API文件，使参数格式与后端Controller匹配
- 修改后端Controller，简化参数接收逻辑

## Impact
- Affected code: 
  - `mybilibili-web/src/api/manuscript.js` (前端API)
  - `mybilibili-video/src/main/java/com/mybilibili/video/controller/ManuscriptController.java` (后端Controller)

## ADDED Requirements
### Requirement: 视频上传参数匹配
系统 SHALL 确保前端发送的视频参数格式与后端Controller期望的格式一致。

#### Scenario: 上传包含视频的稿件
- **WHEN** 用户上传稿件并添加视频文件
- **THEN** 后端Controller应能正确接收视频文件
- **AND** 视频记录应被正确保存到数据库
- **AND** 视频文件应被正确保存到服务器目录
