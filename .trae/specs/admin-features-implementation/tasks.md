# Tasks

## Task 1: 实现索引管理后端接口
在 mybilibili-search 模块中创建 IndexAdminController，提供索引管理功能。
- [x] SubTask 1.1: 创建 IndexAdminController.java，包含 /admin/index/status、/admin/index/bulk、/admin/index/rebuild 接口
- [x] SubTask 1.2: 在 VideoSearchService 中添加索引管理相关方法
- [x] SubTask 1.3: 更新网关路由，确保 /api/search/admin/** 正确转发

## Task 2: 完善稿件管理服务实现
检查并完善 ManuscriptServiceImpl 中的管理功能实现。
- [x] SubTask 2.1: 验证 getPendingManuscripts() 方法实现
- [x] SubTask 2.2: 验证 getProcessingManuscripts() 方法实现
- [x] SubTask 2.3: 验证 getReadyManuscripts() 方法实现
- [x] SubTask 2.4: 验证 getAllManuscripts() 方法实现
- [x] SubTask 2.5: 验证 approveManuscript() 方法实现
- [x] SubTask 2.6: 验证 rejectManuscript() 方法实现
- [x] SubTask 2.7: 验证 publishManuscript() 方法实现
- [x] SubTask 2.8: 验证 unpublishManuscript() 方法实现
- [x] SubTask 2.9: 验证 getManuscriptStatistics() 方法实现

## Task 3: 完善视频处理服务实现
检查并完善 ManuscriptServiceImpl 中的视频处理功能实现。
- [x] SubTask 3.1: 验证 manualTranscode() 方法实现
- [x] SubTask 3.2: 验证 manualExtractAudio() 方法实现
- [x] SubTask 3.3: 验证 manualGenerateSubtitle() 方法实现
- [x] SubTask 3.4: 验证 manualAiSummary() 方法实现
- [x] SubTask 3.5: 验证 manualProcessAll() 方法实现
- [x] SubTask 3.6: 验证 resetVideoStatus() 方法实现
- [x] SubTask 3.7: 验证 getVideoProcessStatus() 方法实现
- [x] SubTask 3.8: 验证 getVideoSourceUrl() 方法实现

## Task 4: 完善内容审核服务实现
检查并完善 CommentServiceImpl 中的内容审核功能实现。
- [x] SubTask 4.1: 验证 getPendingList() 方法实现
- [x] SubTask 4.2: 验证 getAllContent() 方法实现
- [x] SubTask 4.3: 验证 restoreContent() 方法实现
- [x] SubTask 4.4: 验证 deleteContent() 方法实现
- [x] SubTask 4.5: 验证 batchProcess() 方法实现

## Task 5: 集成测试与验证
编译并启动所有服务，验证前后端联调。
- [x] SubTask 5.1: 编译所有模块
- [x] SubTask 5.2: 启动 Nacos、Gateway、Video、Comment、Search 服务
- [x] SubTask 5.3: 测试稿件管理页面功能
- [x] SubTask 5.4: 测试视频处理页面功能
- [x] SubTask 5.5: 测试索引管理页面功能
- [x] SubTask 5.6: 测试内容审核页面功能

# Task Dependencies
- [Task 5] depends on [Task 1, Task 2, Task 3, Task 4]
- [Task 2] and [Task 3] can run in parallel
- [Task 4] can run in parallel with [Task 2, Task 3]
