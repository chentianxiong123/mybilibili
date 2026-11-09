# Checklist

- [x] 前端关注按钮API调用逻辑修复
- [x] 后端 ManuscriptController 接收当前用户ID
- [x] 后端 ManuscriptService 传递用户ID参数
- [x] 后端 UserClient 添加 checkFollowStatus 方法
- [x] 后端 FollowController 已有检查关注状态接口 `/check/{userId}`
- [x] 后端 ManuscriptServiceImpl 正确查询关注状态
- [x] 端到端测试：已登录用户访问稿件详情，关注状态正确显示
- [x] 端到端测试：未登录用户访问稿件详情，关注状态为 false
- [x] 端到端测试：点击关注后刷新页面，状态保持为已关注
- [x] 端到端测试：点击取消关注后刷新页面，状态保持为未关注
