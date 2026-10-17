<template>
  <div class="fans-wrap">
    <!-- 顶部 tab：粉丝列表 / 粉丝勋章 / 骑士团 -->
    <div class="link-top-wrap">
      <div class="link-top-container">
        <span class="link-top-item active">粉丝列表</span>
        <span class="link-top-item">粉丝勋章</span>
        <span class="link-top-item">骑士团</span>
      </div>
    </div>

    <div class="manage-wrap content">
      <!-- 头部：我的粉丝数 + 筛选下拉 -->
      <div class="fans-header clearfix">
        <span class="fans-num left">我的粉丝数 {{ totalFans }}</span>
        <div class="dropdown-wrp right">
          <div class="dropdown" v-click-outside="closeDropdown">
            <a class="dropdown-btn" @click.stop="dropdownOpen = !dropdownOpen">
              <span class="text ellipsis">{{ filterLabel }}</span>
              <span class="iconfont icon-xiajiantou" :class="{ open: dropdownOpen }"></span>
            </a>
            <ul v-show="dropdownOpen" class="dropdown-menu">
              <li @click="setFilter('all')">
                <span :class="{ selected: fansFilter === 'all' }">全部粉丝</span>
              </li>
              <li @click="setFilter('mutual')">
                <span :class="{ selected: fansFilter === 'mutual' }">互相关注</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 粉丝列表 -->
      <div class="fans-body" v-loading="fansLoading">
        <template v-if="fans.length">
          <div v-for="fan in fans" :key="fan.id" class="card-wrp clearfix">
            <!-- 头像 -->
            <a class="u-face left" @click="goSpace(fan.id)">
              <img :src="fan.avatar || '/assets/placeholder-avatar.svg'" :alt="fan.nickname" />
            </a>

            <!-- 昵称 -->
            <div class="info left">
              <div class="u-name">
                <a @click="goSpace(fan.id)">{{ fan.nickname || '未知用户' }}</a>
              </div>
            </div>

            <!-- 右侧操作 -->
            <div class="meta-view right">
              <div class="follow">
                <a
                  class="bili-btn"
                  :class="{ 'is-followed': fan.isFollowing }"
                  @click="handleFollowToggle(fan)"
                >
                  {{ fan.isFollowing ? '已关注' : '关注' }}
                </a>
              </div>
              <div class="option">
                <div class="dropdown" v-click-outside="() => closeOption(fan.id)">
                  <a class="act-btn" @click.stop="toggleOption(fan.id)">
                    <i class="more-icon"></i>
                  </a>
                  <div v-show="openOptionId === fan.id" class="dropdown-menu option-menu">
                    <ul>
                      <li class="menu-item" @click="goWhisper(fan.id)">私信</li>
                      <li class="menu-item" @click="removeFan(fan)">移除粉丝</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <el-empty
          v-else-if="!fansLoading"
          :description="fansFilter === 'mutual' ? '暂无互相关注' : '暂无粉丝'"
        />
      </div>

      <!-- 分页 -->
      <div class="fans-footer">
        <BccPagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="totalFans"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, Directive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi, followApi } from '@/api/creator'
import BccPagination from './BccPagination.vue'

const props = defineProps({
  active: {
    type: Boolean,
    default: false
  }
})

const router = useRouter()

// 状态
const fansLoading = ref(false)
const fansFilter = ref<'all' | 'mutual'>('all')
const fans = ref<any[]>([])
const totalFans = ref(0)
const currentPage = ref(1)
const pageSize = 20

const dropdownOpen = ref(false)
const openOptionId = ref<number | null>(null)

const filterLabel = computed(() =>
  fansFilter.value === 'mutual' ? '互相关注' : '全部粉丝'
)

let initialized = false

// 点击外部关闭下拉
const vClickOutside = {
  beforeMount(el: any, binding: any) {
    el._clickOutside = (e: Event) => {
      if (!el.contains(e.target)) binding.value()
    }
    document.addEventListener('click', el._clickOutside)
  },
  unmounted(el: any) {
    document.removeEventListener('click', el._clickOutside)
  }
} as Directive

// 获取粉丝列表
const fetchFans = async () => {
  fansLoading.value = true
  try {
    const res = await creatorApi.getMyFans({
      page: currentPage.value,
      size: pageSize,
      mutual: fansFilter.value === 'mutual'
    })
    if (res.code === 200) {
      fans.value = res.data.list || []
      totalFans.value = res.data.total || 0
      await enrichFollowingStatus()
    }
  } catch (error) {
    console.error('获取粉丝列表失败:', error)
    ElMessage.error('获取粉丝列表失败')
  } finally {
    fansLoading.value = false
  }
}

// 给每个粉丝补 isFollowing
const enrichFollowingStatus = async () => {
  if (!fans.value.length) return
  try {
    const res = await creatorApi.getMyFollowing()
    if (res.code === 200) {
      const followingIds = new Set((res.data || []).map((u: any) => u.id))
      fans.value = fans.value.map(f => ({
        ...f,
        isFollowing: followingIds.has(f.id)
      }))
    }
  } catch {
    // 静默
  }
}

// 切换筛选
const setFilter = (filter: 'all' | 'mutual') => {
  dropdownOpen.value = false
  if (fansFilter.value === filter) return
  fansFilter.value = filter
  currentPage.value = 1
  fetchFans()
}

const closeDropdown = () => {
  dropdownOpen.value = false
}

const toggleOption = (id: number) => {
  openOptionId.value = openOptionId.value === id ? null : id
}

const closeOption = (id: number) => {
  if (openOptionId.value === id) openOptionId.value = null
}

// 翻页
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchFans()
}

// 关注/取关
const handleFollowToggle = async (fan: any) => {
  try {
    if (fan.isFollowing) {
      await followApi.unfollow(fan.id)
      fan.isFollowing = false
      ElMessage.success('已取消关注')
    } else {
      await followApi.follow(fan.id)
      fan.isFollowing = true
      ElMessage.success('关注成功')
    }
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败，请重试')
  }
}

// 移除粉丝
const removeFan = async (fan: any) => {
  openOptionId.value = null
  try {
    await ElMessageBox.confirm(`确定要移除粉丝「${fan.nickname}」吗？`, '移除粉丝', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    // 取关即可移除粉丝关系
    await followApi.unfollow(fan.id)
    ElMessage.success('已移除')
    fetchFans()
  } catch (e) {
    if (e !== 'cancel') console.error('移除失败:', e)
  }
}

// 跳转私信
const goWhisper = (uid: number) => {
  openOptionId.value = null
  router.push(`/message/whisper?uid=${uid}`)
}

// 跳转空间
const goSpace = (uid: number) => {
  router.push(`/space/${uid}`)
}

// 初始化
const init = () => {
  if (initialized) return
  initialized = true
  fetchFans()
}

watch(
  () => props.active,
  (val) => {
    if (val) init()
  },
  { immediate: true }
)

onMounted(() => {
  if (!props.active) init()
})
</script>

<style scoped>
/* ===== 顶部 tab ===== */
.link-top-wrap {
  padding: 0;
}

.link-top-container {
  display: flex;
  border-bottom: 1px solid #e7e7e7;
}

.link-top-item {
  font-size: 16px;
  color: #505050;
  font-weight: 400;
  padding: 26px 0 0;
  margin-right: 36px;
  cursor: pointer;
  border-bottom: 3px solid transparent;
  margin-bottom: -1px;
  transition: color .2s;
}

.link-top-item:last-child {
  margin-right: 0;
}

.link-top-item.active {
  color: #00a1d6;
  font-weight: 700;
  border-bottom-color: #00a1d6;
}

.link-top-item:hover {
  color: #00a1d6;
}

/* ===== 主体 ===== */
.manage-wrap {
  padding: 0 0 40px;
}

/* 头部 */
.fans-header {
  padding: 20px 0 0;
}

.fans-num {
  font-size: 18px;
  font-weight: 700;
  color: #222;
  line-height: 30px;
}

.fans-header .left {
  float: left;
}

.fans-header .right {
  float: right;
}

/* 筛选下拉 */
.dropdown-wrp {
  position: relative;
}

.dropdown {
  position: relative;
  display: inline-block;
  border: 1px solid #e5e9ef;
  border-radius: 4px;
}

.dropdown-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #99a2aa;
  cursor: pointer;
  padding: 5px 10px;
}

.dropdown-btn:hover {
  color: #00a1d6;
}

.dropdown-btn .text {
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-btn .icon-arrowdown {
  font-size: 12px;
  transition: transform .2s;
}

.dropdown-btn .icon-arrowdown.open {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  min-width: 100px;
  background: #fff;
  border: 1px solid #ccd0d7;
  border-radius: 4px;
  box-shadow: 2px 2px 4px 0 rgba(0, 0, 0, .2);
  padding: 10px 0;
  z-index: 7;
  list-style: none;
  margin: 0;
}

.dropdown-menu li {
  padding: 6px 16px;
  font-size: 13px;
  color: #222;
  cursor: pointer;
  white-space: nowrap;
  transition: color .2s;
}

.dropdown-menu li:hover {
  color: #00a1d6;
}

.dropdown-menu li span.selected {
  color: #00a1d6;
}

/* ===== 粉丝列表 ===== */
.fans-body {
  min-height: 200px;
}

.card-wrp {
  display: flex;
  align-items: center;
  padding: 18px 0;
  border-bottom: 1px solid #e7e7e7;
}

.card-wrp:last-child {
  border-bottom: none;
}

.clearfix::after {
  content: '';
  display: table;
  clear: both;
}

/* 头像 */
.u-face {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  margin-right: 10px;
  cursor: pointer;
  display: block;
}

.u-face img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

/* 昵称 */
.info {
  flex: 1;
  min-width: 0;
}

.u-name {
  font-size: 14px;
  line-height: 18px;
}

.u-name a {
  font-size: 14px;
  color: #222;
  text-decoration: none;
  cursor: pointer;
}

.u-name a:hover {
  color: #00a1d6;
}

/* 右侧操作 */
.meta-view {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  height: 32px;
}

.follow {
  display: inline-block;
  margin-right: 12px;
}

.bili-btn {
  display: inline-block;
  height: 32px;
  padding: 0 10px;
  font-size: 12px;
  line-height: 32px;
  color: #505050;
  border: 1px solid #e5e9ef;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  transition: color .2s, border-color .2s;
  text-align: center;
  min-width: 68px;
}

.bili-btn:hover {
  color: #00a1d6;
  border-color: #00a1d6;
}

.bili-btn.is-followed {
  color: #99a2aa;
  border-color: #e5e9ef;
}

.bili-btn.is-followed:hover {
  color: #f56c6c;
  border-color: #f56c6c;
}

/* ··· 更多按钮 */
.option {
  display: inline-block;
}

.act-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 32px;
  cursor: pointer;
  border: 1px solid #e5e9ef;
  border-radius: 4px;
  transition: border-color .2s;
}

.act-btn:hover {
  border-color: #00a1d6;
}

.more-icon {
  display: inline-block;
  width: 16px;
  height: 4px;
  position: relative;
}

.more-icon::before,
.more-icon::after {
  content: '';
  position: absolute;
  top: 0;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #99a2aa;
}

.more-icon::before {
  left: 0;
}

.more-icon::after {
  right: 0;
}

.more-icon {
  background: radial-gradient(circle, #99a2aa 2px, transparent 2px) center / 4px 4px no-repeat;
}

.act-btn:hover .more-icon::before,
.act-btn:hover .more-icon::after {
  background: #00a1d6;
}

.act-btn:hover .more-icon {
  background: radial-gradient(circle, #00a1d6 2px, transparent 2px) center / 4px 4px no-repeat;
}

.option-menu {
  right: 0;
  min-width: 100px;
  padding: 10px 0;
}

.option-menu .menu-item {
  padding: 6px 16px;
  font-size: 13px;
  color: #222;
  cursor: pointer;
  white-space: nowrap;
}

.option-menu .menu-item:hover {
  color: #00a1d6;
}

/* ===== 分页 ===== */
.fans-footer {
  padding: 0 0 40px;
  margin-top: 20px;
  text-align: center;
}

.fans-footer :deep(.bcc-pagination-container) {
  justify-content: center;
  border-top: none;
  margin-top: 0;
  padding-top: 0;
}
</style>
