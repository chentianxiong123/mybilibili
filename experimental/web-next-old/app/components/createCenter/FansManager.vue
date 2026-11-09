<template>
  <div class="fans-manager">
    <div class="fans-container">
      <div class="fans-header">
        <h3>粉丝管理</h3>
      </div>
      <div class="fans-count-row">
        <p class="fans-count">全部粉丝 ({{ totalFans }})</p>
        <div class="fans-filter">
          <el-radio-group v-model="fansFilter" size="small">
            <el-radio-button value="all">全部粉丝</el-radio-button>
            <el-radio-button value="mutual">互关好友</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <FansList
        :users="fans"
        :loading="fansLoading"
        title=""
        @follow="handleFollowUser"
        @unfollow="handleUnfollowUser"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { creatorApi, followApi } from '@/api/creator'
import FansList from '@/components/FansList.vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

// 粉丝管理相关状态
const fansLoading = ref(false)
const totalFans = ref(0)
const fansCurrentPage = ref(1)
const fansPageSize = ref(10)
const fansFilter = ref('all')
const fans = ref([])
const fansStats = ref({
  totalFans: 0,
  newFansToday: 0,
  newFansThisWeek: 0,
  newFansThisMonth: 0
})

// 粉丝初始化标志
let fansInitialized = false

// 粉丝统计数据获取函数
const fetchFansStats = async () => {
  try {
    const res = await creatorApi.getMyFollowers()
    console.log('粉丝统计 - API响应:', res)
    if (res.code === 200) {
      totalFans.value = Array.isArray(res.data) ? res.data.length : 0
    }
  } catch (error) {
    console.error('获取粉丝统计失败:', error)
  }
}

// 粉丝列表获取函数
const fetchFansList = async () => {
  fansLoading.value = true
  try {
    let fansRes
    if (fansFilter.value === 'mutual') {
      fansRes = await creatorApi.getMyFollowing()
    } else {
      fansRes = await creatorApi.getMyFollowers()
    }
    
    console.log('粉丝列表 - API响应:', fansRes)
    if (fansRes.code === 200) {
      fans.value = fansRes.data || []
      totalFans.value = Array.isArray(fansRes.data) ? fansRes.data.length : 0
      
      const followingRes = await creatorApi.getMyFollowing()
      if (followingRes.code === 200) {
        const followingIds = new Set((followingRes.data || []).map(u => u.id))
        fans.value = fans.value.map(fan => ({
          ...fan,
          isFollowing: followingIds.has(fan.id)
        }))
        console.log('更新后的粉丝列表:', fans.value)
      }
    }
  } catch (error) {
    console.error('获取粉丝列表失败:', error)
    ElMessage.error('获取粉丝列表失败')
  } finally {
    fansLoading.value = false
  }
}

// 粉丝管理初始化函数
const initFansData = () => {
  if (!fansInitialized) {
    fansInitialized = true
    fetchFansStats()
    fetchFansList()
  }
}

const handleFollowUser = async (userId) => {
  try {
    await followApi.follow(userId)
    const fan = fans.value.find(f => f.id === userId)
    if (fan) fan.isFollowing = true
    ElMessage.success('关注成功')
  } catch (error) {
    console.error('关注失败:', error)
    ElMessage.error('操作失败，请重试')
  }
}

const handleUnfollowUser = async (userId) => {
  try {
    await followApi.unfollow(userId)
    const fan = fans.value.find(f => f.id === userId)
    if (fan) fan.isFollowing = false
    ElMessage.success('已取消关注')
  } catch (error) {
    console.error('取消关注失败:', error)
    ElMessage.error('操作失败，请重试')
  }
}

watch(fansFilter, () => {
  fansCurrentPage.value = 1
})

watch([fansCurrentPage, fansPageSize, fansFilter], () => {
  if (fansInitialized) {
    fetchFansList()
  }
})

// 监听路由变化，首次进入粉丝管理时初始化数据
watch(
  () => props.active,
  (newVal) => {
    if (newVal) {
      initFansData()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.fans-manager {
  width: 100%;
}

.fans-container {
  padding: 20px;
}

.fans-header {
  margin-bottom: 15px;
}

.fans-header h3 {
  font-size: 18px;
  font-weight: bold;
  margin: 0;
}

.fans-count-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.fans-count {
  font-size: 16px;
  color: #333;
  margin: 0;
}

.fans-filter {
  display: flex;
  align-items: center;
}
</style>