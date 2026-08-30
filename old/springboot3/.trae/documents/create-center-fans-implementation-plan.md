# 创建中心粉丝页面实现计划

## 项目概述
实现 `http://localhost:5173/create-center/fans` 页面的粉丝列表和互关粉丝列表功能，包括后端API调用、前端展示和交互功能。

## 任务分解和优先级

### [ ] 任务1: 实现后端粉丝列表API
- **Priority**: P0
- **Depends On**: None
- **Description**:
  - 在 `mybilibili-user` 服务中实现获取粉丝列表的API
  - 支持分页和筛选（全部粉丝/互关粉丝）
  - 实现关注/取消关注粉丝的API
- **Success Criteria**:
  - API能够返回正确的粉丝列表数据
  - 支持分页和筛选功能
  - 关注/取消关注功能正常工作
- **Test Requirements**:
  - `programmatic` TR-1.1: API返回状态码200
  - `programmatic` TR-1.2: API返回数据格式正确
  - `programmatic` TR-1.3: 关注/取消关注API调用成功

### [ ] 任务2: 实现前端粉丝列表数据获取
- **Priority**: P0
- **Depends On**: 任务1
- **Description**:
  - 在 `CreateCenterView.vue` 中实现获取粉丝列表的函数
  - 实现分页加载功能
  - 实现筛选（全部粉丝/互关粉丝）功能
- **Success Criteria**:
  - 页面能够正确加载粉丝列表
  - 分页功能正常工作
  - 筛选功能正常工作
- **Test Requirements**:
  - `programmatic` TR-2.1: 页面加载时能够成功调用API
  - `programmatic` TR-2.2: 分页按钮能够正确触发API调用
  - `human-judgement` TR-2.3: 粉丝列表显示正确，包含用户头像、用户名和互关标签

### [ ] 任务3: 实现关注/取消关注功能
- **Priority**: P1
- **Depends On**: 任务1
- **Description**:
  - 实现点击关注/取消关注按钮的逻辑
  - 处理关注状态的更新
  - 添加加载状态和错误处理
- **Success Criteria**:
  - 点击关注按钮后，按钮状态变为"已关注"
  - 点击已关注按钮后，按钮状态变为"关注"
  - 操作过程中有加载状态提示
- **Test Requirements**:
  - `programmatic` TR-3.1: 关注/取消关注API调用成功
  - `human-judgement` TR-3.2: 按钮状态能够正确更新
  - `human-judgement` TR-3.3: 操作过程中有加载状态提示

### [ ] 任务4: 实现互关粉丝筛选功能
- **Priority**: P1
- **Depends On**: 任务1, 任务2
- **Description**:
  - 实现点击"互关粉丝"筛选的逻辑
  - 确保筛选后只显示互关的粉丝
- **Success Criteria**:
  - 点击"互关粉丝"选项后，列表只显示互关的粉丝
  - 点击"全部粉丝"选项后，列表显示所有粉丝
- **Test Requirements**:
  - `programmatic` TR-4.1: 筛选API调用成功
  - `human-judgement` TR-4.2: 筛选结果正确显示

### [ ] 任务5: 优化用户体验
- **Priority**: P2
- **Depends On**: 任务2, 任务3, 任务4
- **Description**:
  - 添加加载动画
  - 添加错误提示
  - 优化列表滚动体验
  - 添加空状态提示
- **Success Criteria**:
  - 页面加载时有加载动画
  - 错误时显示错误提示
  - 列表滚动流畅
  - 无粉丝时显示空状态提示
- **Test Requirements**:
  - `human-judgement` TR-5.1: 加载动画显示正常
  - `human-judgement` TR-5.2: 错误提示显示正常
  - `human-judgement` TR-5.3: 列表滚动流畅
  - `human-judgement` TR-5.4: 空状态提示显示正常

## 技术实现细节

### 后端实现
1. **粉丝列表API**:
   - 路径: `/user/{userId}/followers`
   - 支持参数: `page`, `size`, `mutual` (是否只返回互关粉丝)
   - 返回数据: 粉丝列表，包含用户ID、用户名、头像、是否互关

2. **关注/取消关注API**:
   - 关注: `POST /follow/{userId}`
   - 取消关注: `DELETE /follow/{userId}`

### 前端实现
1. **数据获取**:
   - 使用 `userApi.getFollowerList` 方法获取粉丝列表
   - 实现分页逻辑，默认每页10条
   - 实现筛选逻辑，根据 `fansFilter` 值调用不同的API

2. **UI交互**:
   - 使用 Element Plus 组件库的 `el-avatar`, `el-button`, `el-tag` 等组件
   - 实现加载状态使用 `v-loading` 指令
   - 实现错误提示使用 `el-message`

3. **状态管理**:
   - 使用 `ref` 管理粉丝列表、加载状态、当前页码等
   - 使用 `reactive` 管理筛选条件

## 预期效果
- 粉丝管理页面能够正确显示粉丝列表
- 支持分页加载更多粉丝
- 支持筛选互关粉丝
- 支持关注/取消关注粉丝
- 页面交互流畅，有良好的用户体验

## 风险评估
- **API调用失败**：需要添加错误处理，确保页面不会崩溃
- **数据加载缓慢**：需要添加加载动画，提升用户体验
- **分页逻辑复杂**：需要确保分页逻辑正确，避免重复加载数据
- **筛选逻辑错误**：需要确保筛选逻辑正确，避免显示错误的数据
