# Tasks

## Task 1: 修改稿件管理页面审核通过交互
修改ManuscriptsView.vue，将"审核通过"按钮改为弹出选项对话框。
- [x] SubTask 1.1: 修改handleApprove方法，使用ElMessageBox.confirm改为自定义对话框
- [x] SubTask 1.2: 添加"仅审核通过"和"审核通过并处理视频"两个选项
- [x] SubTask 1.3: 调用新的后端接口（审核通过并处理视频）
- [x] SubTask 1.4: 将"转码状态"列改为"处理状态"列

## Task 2: 后端新增审核通过并处理视频接口
在ManuscriptAdminController中新增接口，支持审核通过时可选自动处理视频。
- [x] SubTask 2.1: 创建新接口 POST /api/manuscript/admin/{id}/approve-with-process
- [x] SubTask 2.2: 接收参数 autoProcess: boolean
- [x] SubTask 2.3: 如果autoProcess=true，获取稿件下所有视频并创建全流程任务队列
- [x] SubTask 2.4: 返回处理结果

## Task 3: 实现全流程自动执行服务
创建或修改服务，支持自动按顺序执行视频的所有处理步骤。
- [x] SubTask 3.1: 创建全流程任务队列管理器
- [x] SubTask 3.2: 实现步骤完成后自动触发下一步的逻辑
- [x] SubTask 3.3: 确保单线程顺序执行（并发数为1）
- [x] SubTask 3.4: 处理步骤失败的情况（停止后续步骤，继续下一个视频）

## Task 4: 优化SSE状态推送
确保SSE正确推送状态更新，前端实时更新按钮颜色。
- [x] SubTask 4.1: 确认SSE消息格式包含必要信息（videoId, stage, progress, status, nextStage）
- [x] SubTask 4.2: 前端VideoProcessView.vue正确处理SSE消息
- [x] SubTask 4.3: 优化getStepClass函数，确保颜色映射正确
- [x] SubTask 4.4: 测试按钮颜色实时更新

## Task 5: 前端API更新
更新前端API调用，支持新的接口。
- [x] SubTask 5.1: 在manuscript.js中添加approveWithProcess接口
- [x] SubTask 5.2: 更新ManuscriptsView.vue中的API调用

## Task 6: 集成测试
测试完整流程，确保功能正常。
- [x] SubTask 6.1: 测试"仅审核通过"功能
- [x] SubTask 6.2: 测试"审核通过并处理视频"功能
- [x] SubTask 6.3: 测试SSE按钮颜色实时更新
- [x] SubTask 6.4: 测试单线程顺序执行
- [x] SubTask 6.5: 测试步骤失败处理

# Task Dependencies
- [Task 2] depends on [Task 3]
- [Task 1] depends on [Task 2, Task 5]
- [Task 6] depends on [Task 1, Task 2, Task 3, Task 4, Task 5]
