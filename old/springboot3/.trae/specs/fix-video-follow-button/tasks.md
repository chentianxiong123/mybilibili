# Tasks

- [x] Task 1: 修复 VideoView.vue 中 handleFollow 函数的API调用逻辑
  - [x] SubTask 1.1: 将 `userApi.follow(videoInfo.value.uploader.id, true)` 改为 `userApi.follow(videoInfo.value.uploader.id, !isFollowing.value)`
  - [x] SubTask 1.2: 调整响应处理逻辑，根据API返回结果更新状态

- [x] Task 2: 修复后端获取稿件详情时返回正确的关注状态
  - [x] SubTask 2.1: 修改 ManuscriptController.getManuscriptById，添加 HttpServletRequest 参数获取当前用户ID
  - [x] SubTask 2.2: 修改 ManuscriptService.getManuscriptWithVideos，添加 currentUserId 参数
  - [x] SubTask 2.3: 修改 UserClient，添加 checkFollowStatus 方法
  - [x] SubTask 2.4: FollowController 已有检查关注状态接口 `/check/{userId}`
  - [x] SubTask 2.5: 修改 ManuscriptServiceImpl，使用正确的逻辑查询关注状态

# Task Dependencies
- [Task 2] 依赖 [Task 1] 完成
