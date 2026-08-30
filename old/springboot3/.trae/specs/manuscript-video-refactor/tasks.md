# 稿件视频概念重构任务列表

## 阶段一：实体类修复（优先级最高）

- [ ] Task 1.1: 修复Video实体类
  - [ ] 1.1.1: 移除manuscript关联对象和manuscript相关字段
  - [ ] 1.1.2: 确保所有字段与数据库videos表完全匹配
  - [ ] 1.1.3: 添加@TableName("videos")注解

- [ ] Task 1.2: 修复Comment实体类
  - [ ] 1.2.1: 移除videoId字段，只保留manuscriptId
  - [ ] 1.2.2: 清理targetType相关逻辑

- [ ] Task 1.3: 修复UserInteraction实体类
  - [ ] 1.3.1: 确保@TableName("user_interactions")注解存在
  - [ ] 1.3.2: 确认targetType只支持MANUSCRIPT

- [ ] Task 1.4: 修复Danmaku实体类
  - [ ] 1.4.1: 移除manuscriptId字段，只保留videoId
  - [ ] 1.4.2: 确保与MongoDB danmakus集合字段匹配

## 阶段二：服务层重构

- [ ] Task 2.1: 重构评论服务
  - [ ] 2.1.1: 修改CommentServiceImpl，移除video相关逻辑
  - [ ] 2.1.2: 确保评论只关联稿件
  - [ ] 2.1.3: 更新CommentController注释说明

- [ ] Task 2.2: 重构互动服务
  - [ ] 2.2.1: 修改VideoInteractionServiceImpl
  - [ ] 2.2.2: 移除videoId相关逻辑，只操作manuscriptId
  - [ ] 2.2.3: 更新统计信息到manuscript表

- [ ] Task 2.3: 确认弹幕服务（无需修改）
  - [ ] 2.3.1: 验证弹幕服务已正确只使用videoId

## 阶段三：编译验证

- [ ] Task 3.1: 编译mybilibili-common模块
  - [ ] 3.1.1: 执行 mvn clean install
  - [ ] 3.1.2: 确认无编译错误

- [ ] Task 3.2: 编译各服务模块
  - [ ] 3.2.1: 编译mybilibili-comment
  - [ ] 3.2.2: 编译mybilibili-interaction
  - [ ] 3.2.3: 编译mybilibili-danmaku

## 阶段四：服务测试

- [ ] Task 4.1: 启动评论服务测试
  - [ ] 4.1.1: 启动服务
  - [ ] 4.1.2: 测试评论列表接口

- [ ] Task 4.2: 启动互动服务测试
  - [ ] 4.2.1: 启动服务
  - [ ] 4.2.2: 测试互动状态接口
  - [ ] 4.2.3: 测试点赞/投币/收藏接口

- [ ] Task 4.3: 启动弹幕服务测试
  - [ ] 4.3.1: 启动服务
  - [ ] 4.3.2: 测试弹幕列表接口

## 任务依赖关系
- 阶段一（1.1-1.4）完成后才能进行阶段二
- 阶段二完成后才能进行阶段三
- 阶段三完成后才能进行阶段四
