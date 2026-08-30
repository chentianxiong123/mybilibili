# Checklist

## Task 1: 索引管理后端接口
- [x] IndexAdminController.java 文件已创建
- [x] /admin/index/status 接口实现并返回正确数据
- [x] /admin/index/bulk 接口实现并成功批量索引
- [x] /admin/index/rebuild 接口实现并成功重建索引
- [x] VideoSearchService 包含索引管理方法
- [x] 网关路由 /api/search/admin/** 正确配置

## Task 2: 稿件管理服务
- [x] getPendingManuscripts() 返回待审核稿件列表
- [x] getProcessingManuscripts() 返回处理中稿件列表
- [x] getReadyManuscripts() 返回待上架稿件列表
- [x] getAllManuscripts() 返回所有稿件列表
- [x] approveManuscript() 成功审核通过稿件
- [x] rejectManuscript() 成功审核拒绝稿件
- [x] publishManuscript() 成功上架稿件
- [x] unpublishManuscript() 成功下架稿件
- [x] getManuscriptStatistics() 返回正确统计数据

## Task 3: 视频处理服务
- [x] manualTranscode() 成功触发视频转码
- [x] manualExtractAudio() 成功触发音频提取
- [x] manualGenerateSubtitle() 成功触发字幕生成
- [x] manualAiSummary() 成功触发 AI 总结
- [x] manualProcessAll() 成功触发全流程处理
- [x] resetVideoStatus() 成功重置视频状态
- [x] getVideoProcessStatus() 返回正确处理状态
- [x] getVideoSourceUrl() 返回正确的视频源 URL

## Task 4: 内容审核服务
- [x] getPendingList() 返回待审核内容列表
- [x] getAllContent() 返回所有内容列表
- [x] restoreContent() 成功恢复内容
- [x] deleteContent() 成功删除内容
- [x] batchProcess() 成功批量处理内容

## Task 5: 集成测试
- [x] 所有模块编译成功
- [x] 所有服务启动成功
- [x] 稿件管理页面功能正常
- [x] 视频处理页面功能正常
- [x] 索引管理页面功能正常
- [x] 内容审核页面功能正常
