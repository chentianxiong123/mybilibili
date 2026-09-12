<template>
  <aside class="sidebar">
    <!-- 独立的投稿按钮 -->
    <el-button type="primary" class="upload-btn-large" @click="goToUpload">
      <el-icon><Upload /></el-icon>
      <span>投稿</span>
    </el-button>

    <!-- 侧边导航菜单 -->
    <el-menu
      ref="menuRef"
      default-active="home"
      class="sidebar-menu"
      :unique-opened="false"
      @select="handleMenuSelect"
    >
      <el-menu-item index="home">
        <el-icon><House /></el-icon>
        <span>首页</span>
      </el-menu-item>
      <el-sub-menu index="content">
        <template #title>
          <el-icon><Document /></el-icon>
          <span>内容管理</span>
        </template>
        <el-menu-item index="content-articles">
          <el-icon><Menu /></el-icon>
          <span>稿件管理</span>
        </el-menu-item>
        <el-menu-item index="drafts">
          <el-icon><FolderOpened /></el-icon>
          <span>草稿箱</span>
        </el-menu-item>

      </el-sub-menu>
      <el-menu-item index="data">
        <el-icon><DataAnalysis /></el-icon>
        <span>数据中心</span>
      </el-menu-item>
      <el-menu-item index="fans">
        <el-icon><UserFilled /></el-icon>
        <span>粉丝管理</span>
      </el-menu-item>
    </el-menu>
  </aside>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Upload,
  House,
  Document,
  DataAnalysis,
  UserFilled,
  Menu
} from '@element-plus/icons-vue'

const props = defineProps({
  activeKey: {
    type: String,
    default: 'home'
  }
})

const router = useRouter()

// 侧边栏菜单引用
const menuRef = ref(null)

// 菜单选择事件处理
const handleMenuSelect = (index, indexPath) => {
  // 根据索引导航到对应的路由
  const routeMap = {
    'home': '/create-center/home',
    'upload': '/create-center/upload',
    'content': '/create-center/content',
    'content-articles': '/create-center/content-articles',
    'drafts': '/create-center/drafts',

    'data': '/create-center/data',
    'fans': '/create-center/fans',
  }

  if (routeMap[index]) {
    router.push(routeMap[index])
  }

  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 跳转到投稿页面（修改为在创作中心内部显示）
const goToUpload = () => {
  router.push('/create-center/upload')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 监听当前激活菜单变化，同步菜单高亮
watch(
  () => props.activeKey,
  (newVal) => {
    if (menuRef.value) {
      menuRef.value.activeIndex = newVal
    }
  },
  { immediate: true }
)
</script>

<style scoped>
/* 侧边栏样式 */
.sidebar {
  width: 200px;
  background-color: #fff;
  border-right: 1px solid #e0e0e0;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.upload-btn-large {
  width: 100%;
  height: 40px;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background-color: #fb7299;
  border: none;
  color: #fff;
}

.upload-btn-large:hover {
  background-color: #f75982;
  color: #fff;
}

.sidebar-menu {
  border-right: none;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  font-size: 15px;
}

.sidebar-menu .el-menu-item.is-active {
  color: #1890ff;
  background-color: #ecf5ff;
}
</style>