# 粉丝增长图表首次加载不显示问题修复计划

## 问题分析

### 根本原因
1. `activeDataTab` 默认值是 `'overview'`
2. 粉丝分析区域使用 `v-if="activeDataTab === 'fans'"`，所以默认情况下粉丝图表的DOM元素不存在
3. 当用户切换到粉丝分析tab时，`watch(activeDataTab)` 触发，但存在时序问题：
   - `v-if` 切换后DOM需要时间渲染
   - `nextTick` 只等待一个tick，不够等待DOM完全渲染
   - 导致 `initFansTrendChart()` 执行时DOM元素还不存在

### 为什么选择天数后能显示
- 选择天数时，`watch(fansTimeRange)` 触发
- 此时用户已经在fans tab，DOM已经渲染完成
- 所以图表能正常初始化

## 修复方案

### 方案：使用 setTimeout 确保 DOM 渲染完成

修改 `watch(activeDataTab)` 中 fans 分支的逻辑，使用 `setTimeout` 替代 `nextTick`，确保DOM有足够时间渲染。

### 具体修改

**文件**: `d:\files\mybilibili-next\mybilibili-web\src\views\CreateCenterView.vue`

**修改位置**: 第2494-2514行

**修改内容**:
```javascript
// 监听数据中心标签页变化
watch(activeDataTab, async (newVal) => {
  if (newVal === 'overview') {
    await Promise.all([
      loadDataOverview(),
      loadTrendData(),
      loadManuscriptRanking('views', 5)
    ])
    nextTick(() => {
      initPlayTrendChart()
      initManuscriptRankingChart()
    })
  } else if (newVal === 'article') {
    await loadManuscriptRanking(rankingSortBy.value)
  } else if (newVal === 'fans') {
    await loadTrendData()
    // 使用 setTimeout 确保 v-if 切换后 DOM 完全渲染
    setTimeout(() => {
      console.log('延迟初始化粉丝图表')
      initFansTrendChart()
    }, 100)
  }
  updateCharts()
})
```

### 同样修改 onMounted

如果用户直接访问粉丝分析页面（通过URL参数），也需要在 `onMounted` 中做同样的处理：

**修改位置**: 第2536-2559行

**修改内容**:
```javascript
onMounted(async () => {
  console.log('onMounted 执行, activeDataTab:', activeDataTab.value)
  window.addEventListener('resize', handleResize)

  if (activeDataTab.value === 'overview') {
    console.log('加载 overview 数据')
    await Promise.all([
      loadDataOverview(),
      loadTrendData(),
      loadManuscriptRanking('views', 5)
    ])
    nextTick(() => {
      initPlayTrendChart()
      initManuscriptRankingChart()
    })
  } else if (activeDataTab.value === 'fans') {
    console.log('加载 fans 数据')
    await loadTrendData()
    // 使用 setTimeout 确保 v-if 切换后 DOM 完全渲染
    setTimeout(() => {
      console.log('延迟初始化粉丝图表')
      initFansTrendChart()
    }, 100)
  }
})
```

## 执行步骤

1. 修改 `watch(activeDataTab)` 中 fans 分支，使用 `setTimeout` 替代 `nextTick`
2. 修改 `onMounted` 中 fans 分支，使用 `setTimeout` 替代 `nextTick`
3. 测试验证：刷新页面后直接切换到粉丝分析，图表应该立即显示
