<script setup lang="ts">
import { MoreFilled } from '@element-plus/icons-vue'
import { ref } from 'vue'
import VPopover from '@/components/teriteri/popover/VPopover.vue'

const props = defineProps({
  collections: {
    type: Array,
    required: true
  },
  myFavorites: {
    type: Array,
    required: true
  },
  activeCategory: {
    type: String,
    required: true
  },
  activeFolderId: {
    type: [String, Number],
    default: null
  },
  myCollectionsExpanded: {
    type: Boolean,
    required: true
  },
  isOwnSpace: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'select-folder',
  'toggle-expand',
  'open-create',
  'open-edit',
  'delete-folder'
])

const isFavnavOpen = ref(true)

const placeholderCover = '/assets/teriteri/img/bilibili/video-placeholder.png'
</script>

<template>
  <div class="fav-sidenav">
    <div class="nav-container">
      <!-- 我的创建分组标题 -->
      <div class="favlist-title" @click="emit('toggle-expand')">
        <p>{{ isOwnSpace ? '我的创建' : 'TA的创建' }}</p>
        <i class="iconfont icon-xiajiantou" :style="myCollectionsExpanded ? 'transform: rotate(180deg);' : ''"></i>
      </div>

      <div class="fav-list-container" v-show="myCollectionsExpanded">
        <!-- 新建收藏夹（仅本人） -->
        <div class="nav-title nav-add" v-if="isOwnSpace" @click="emit('open-create')">
          <svg t="1711976556258" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4395" width="18" height="18"><path d="M512 62c247.5 0 450 202.5 450 450s-202.5 450-450 450-450-202.5-450-450 202.5-450 450-450z m28.125 421.875V287c0-16.875-11.25-28.125-28.125-28.125s-28.125 11.25-28.125 28.125v196.875H287c-16.875 0-28.125 11.25-28.125 28.125s11.25 28.125 28.125 28.125h196.875V737c0 16.875 11.25 28.125 28.125 28.125s28.125-11.25 28.125-28.125V540.125H737c16.875 0 28.125-11.25 28.125-28.125s-11.25-28.125-28.125-28.125H540.125z" p-id="4396"></path></svg>
          <div class="text">新建收藏夹</div>
        </div>

        <ul class="fav-list">
          <li
            class="fav-item"
            :class="{'cur': activeFolderId !== null && Number(activeFolderId) === Number(item.id), 'owner': isOwnSpace}"
            v-for="(item, index) in collections"
            :key="index"
            @click="emit('select-folder', item)"
          >
            <img v-if="item.cover" loading="lazy" decoding="async" :src="item.cover" alt="" class="folder-cover">
            <span v-else class="folder-cover placeholder"><i class="iconfont icon-wenjianjia"></i></span>
            <a class="text" :title="item.name">{{ item.name }}</a>
            <span class="num">{{ item.count }}</span>
            <VPopover class="more" popStyle="z-index: 1000;">
              <template #reference>
                <div class="tool-more">
                  <i class="iconfont icon-gengduo"></i>
                </div>
              </template>
              <template #content>
                <ul class="be-dropdown-menu">
                  <li class="be-dropdown-item" @click="emit('open-edit', item)">编辑信息</li>
                  <li class="be-dropdown-item be-dropdown-item-delimiter" @click="emit('delete-folder', item)">删除</li>
                </ul>
              </template>
            </VPopover>
          </li>
        </ul>
      </div>
    </div>

    <!-- TA的收藏分组 -->
    <div class="nav-container" v-if="myFavorites.length">
      <div class="favlist-title" @click="isFavnavOpen = !isFavnavOpen">
        <p>TA的收藏</p>
        <i class="iconfont icon-xiajiantou" :style="isFavnavOpen ? 'transform: rotate(180deg);' : ''"></i>
      </div>
      <div class="fav-list-container" v-show="isFavnavOpen">
        <ul class="fav-list">
          <li
            class="fav-item"
            :class="{'cur': activeFolderId !== null && Number(activeFolderId) === Number(item.id)}"
            v-for="(item, index) in myFavorites"
            :key="index"
            @click="emit('select-folder', item)"
          >
            <span class="folder-cover placeholder"><i class="iconfont icon-wenjianjia"></i></span>
            <a class="text" :title="item.name">{{ item.name }}</a>
            <span class="num">{{ item.count }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fav-sidenav {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid #e5e9ef;
  font-size: 14px;
  color: #222;
}

.nav-container {
  border-bottom: 1px solid #e5e9ef;
}

.favlist-title {
  padding-right: 11px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  height: 54px;
  padding-left: 19px;
  line-height: 54px;
  font-size: 14px;
  color: #99a2aa;
}

.favlist-title p {
  margin: 0;
}

.fav-list-container {
  position: relative;
  max-height: 420px;
  margin-bottom: 10px;
  overflow: visible;
}

.nav-title {
  position: relative;
  height: 44px;
  padding-left: 20px;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.nav-add .icon {
  fill: #cbd0d8;
  margin-right: 10px;
}

.nav-add .icon:hover {
  fill: #97a2ab;
}

.nav-title .text {
  display: block;
  line-height: 24px;
  color: #222;
}

.fav-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.fav-item {
  position: relative;
  padding-left: 20px;
  padding-right: 8px;
  transition: background-color .3s ease;
  white-space: nowrap;
  font-size: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  height: 44px;
  cursor: pointer;
}

.fav-item:hover {
  background-color: #f4f5f7;
}

.fav-item.cur {
  background-color: #00a1d6;
}

.fav-item.cur .icon-gengduo, .fav-item.cur svg, .fav-item.cur .num, .fav-item.cur .text {
  color: #fff;
  fill: #fff;
}

.folder-cover {
  width: 32px;
  height: 32px;
  border-radius: 3px;
  object-fit: cover;
  margin-right: 10px;
  flex-shrink: 0;
}

.folder-cover.placeholder {
  background: #e5e9ef;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #99a2aa;
  font-size: 16px;
}

.fav-item .text {
  display: inline-block;
  line-height: 44px;
  flex: 1;
  margin-right: 5px;
  font-size: 14px;
  vertical-align: middle;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  min-width: 0;
}

.fav-item .num {
  display: inline-block;
  width: 32px;
  font-size: 12px;
  color: #99a2aa;
  vertical-align: middle;
  text-align: center;
  flex-shrink: 0;
}

.owner:hover .num {
  display: none;
}

.fav-item .more {
  display: none !important;
}

.owner:hover .more {
  display: inline-block !important;
  margin-left: 8px;
}

.tool-more {
  cursor: pointer;
}

.icon-gengduo {
  color: #999;
}

.be-dropdown-menu {
  padding: 6px 0;
  list-style: none;
  margin: 0;
}

.be-dropdown-item {
  height: 40px;
  padding: 0 20px;
  line-height: 40px;
  text-align: center;
  font-size: 14px;
  color: #222;
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: background-color .2s ease;
  cursor: pointer;
}

.be-dropdown-item:hover {
  color: #00a1d6;
  background-color: #f0f0f0;
}

.be-dropdown-item.be-dropdown-item-delimiter {
  border-bottom: 1px solid #e5e9ef;
}
</style>
