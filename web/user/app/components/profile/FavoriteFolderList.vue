<script setup lang="ts">
import { MoreFilled } from '@element-plus/icons-vue'
import { ref } from 'vue'
import VPopover from '@/components/site/popover/VPopover.vue'

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
</script>

<template>
  <div class="favlist-aside">
    <div class="vui_collapse">
      <!-- 我创建的收藏夹 分组 -->
      <div class="vui_collapse_item">
        <div class="vui_collapse_item_header" @click="emit('toggle-expand')">
          <p>{{ isOwnSpace ? '我创建的收藏夹' : 'TA创建的收藏夹' }}</p>
          <i class="iconfont icon-xiajiantou" :style="myCollectionsExpanded ? 'transform: rotate(180deg);' : ''"></i>
        </div>

        <div class="vui_collapse_item_content" v-show="myCollectionsExpanded">
          <div class="fav-collapse-wrap">
            <!-- 新建收藏夹（仅本人） -->
            <div class="fav-collapse-create" v-if="isOwnSpace" @click="emit('open-create')">
              <svg t="1711976556258" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="4395" width="20" height="20"><path d="M512 62c247.5 0 450 202.5 450 450s-202.5 450-450 450-450-202.5-450-450 202.5-450 450-450z m28.125 421.875V287c0-16.875-11.25-28.125-28.125-28.125s-28.125 11.25-28.125 28.125v196.875H287c-16.875 0-28.125 11.25-28.125 28.125s11.25 28.125 28.125 28.125h196.875V737c0 16.875 11.25 28.125 28.125 28.125s28.125-11.25 28.125-28.125V540.125H737c16.875 0 28.125-11.25 28.125-28.125s-11.25-28.125-28.125-28.125H540.125z" p-id="4396"></path></svg>
              <span class="text">新建收藏夹</span>
            </div>

            <div class="vui_sidebar">
              <div
                class="fav-sidebar-item"
                v-for="(item, index) in collections"
                :key="index"
                @click="emit('select-folder', item)"
              >
                <div class="vui_sidebar-item" :class="{'vui_sidebar-item--active': activeFolderId !== null && Number(activeFolderId) === Number(item.id)}">
                  <div class="vui_sidebar-item-title">
                    <svg class="folder-icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="20" height="20"><path d="M149.333333 170.666667h213.333334l85.333333 85.333333h426.666667a42.666667 42.666667 0 0 1 42.666666 42.666667v512a42.666667 42.666667 0 0 1-42.666666 42.666666H149.333333a42.666667 42.666667 0 0 1-42.666666-42.666666V213.333333a42.666667 42.666667 0 0 1 42.666666-42.666666z" fill="currentColor"></path></svg>
                    <span class="vui_ellipsis" :title="item.name">{{ item.name }}</span>
                  </div>
                  <span class="vui_sidebar-item-right">{{ item.count }}</span>
                  <VPopover v-if="isOwnSpace" class="more" popStyle="z-index: 1000;">
                    <template #reference>
                      <div class="tool-more">
                        <i class="iconfont icon-gengduo"></i>
                      </div>
                    </template>
                    <template #content>
                      <ul class="be-dropdown-menu">
                        <li class="be-dropdown-item" @click.stop="emit('open-edit', item)">编辑信息</li>
                        <li class="be-dropdown-item be-dropdown-item-delimiter" @click.stop="emit('delete-folder', item)">删除</li>
                      </ul>
                    </template>
                  </VPopover>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TA收藏的 分组（仅访客视图且有数据时显示） -->
    <div class="vui_collapse" v-if="myFavorites.length">
      <div class="vui_collapse_item">
        <div class="vui_collapse_item_header" @click="isFavnavOpen = !isFavnavOpen">
          <p>TA收藏的</p>
          <i class="iconfont icon-xiajiantou" :style="isFavnavOpen ? 'transform: rotate(180deg);' : ''"></i>
        </div>
        <div class="vui_collapse_item_content" v-show="isFavnavOpen">
          <div class="fav-collapse-wrap">
            <div class="vui_sidebar">
              <div
                class="fav-sidebar-item"
                v-for="(item, index) in myFavorites"
                :key="index"
                @click="emit('select-folder', item)"
              >
                <div class="vui_sidebar-item" :class="{'vui_sidebar-item--active': activeFolderId !== null && Number(activeFolderId) === Number(item.id)}">
                  <div class="vui_sidebar-item-title">
                    <svg class="folder-icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="20" height="20"><path d="M149.333333 170.666667h213.333334l85.333333 85.333333h426.666667a42.666667 42.666667 0 0 1 42.666666 42.666667v512a42.666667 42.666667 0 0 1-42.666666 42.666666H149.333333a42.666667 42.666667 0 0 1-42.666666-42.666666V213.333333a42.666667 42.666667 0 0 1 42.666666-42.666666z" fill="currentColor"></path></svg>
                    <span class="vui_ellipsis" :title="item.name">{{ item.name }}</span>
                  </div>
                  <span class="vui_sidebar-item-right">{{ item.count }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.favlist-aside {
  width: 269px;
  flex-shrink: 0;
  padding: 0 5px 20px 0;
  margin: 18px 16px 0 0;
  font-size: 14px;
  color: #18191c;
}

.vui_collapse_item_header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 48px;
  padding: 0 16px;
  line-height: 48px;
  font-size: 16px;
  font-weight: 500;
  color: #18191c;
  cursor: pointer;
}

.vui_collapse_item_header p {
  margin: 0;
}

.vui_collapse_item_header .iconfont {
  font-size: 14px;
  color: #9499a0;
  transition: transform .2s ease;
}

.vui_collapse_item_content {
  padding-bottom: 16px;
}

.fav-collapse-create {
  display: flex;
  align-items: center;
  height: 48px;
  padding: 14px 16px;
  border-radius: 6px;
  font-size: 14px;
  color: #18191c;
  cursor: pointer;
  transition: background-color .2s ease;
}

.fav-collapse-create:hover {
  background-color: #f6f7f8;
}

.fav-collapse-create .icon {
  fill: #00aeec;
  margin-right: 8px;
  flex-shrink: 0;
}

.fav-collapse-create .text {
  line-height: 21px;
}

.vui_sidebar {
  display: block;
}

.fav-sidebar-item {
  position: relative;
}

.vui_sidebar-item {
  display: flex;
  align-items: center;
  height: 48px;
  padding: 14px 16px;
  border-radius: 6px;
  font-size: 14px;
  line-height: 21px;
  color: #18191c;
  cursor: pointer;
  transition: background-color .2s ease;
}

.vui_sidebar-item:hover {
  background-color: #f6f7f8;
}

.vui_sidebar-item--active {
  background-color: #00aeec;
  color: #fff;
}

.vui_sidebar-item--active:hover {
  background-color: #00aeec;
}

.vui_sidebar-item-title {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  margin-right: 8px;
}

.vui_sidebar-item-title .folder-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  margin-right: 8px;
  color: currentColor;
}

.vui_ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 21px;
}

.vui_sidebar-item-right {
  flex-shrink: 0;
  font-size: 12px;
  line-height: 18px;
  color: #9499a0;
}

.vui_sidebar-item--active .vui_sidebar-item-right {
  color: #fff;
}

.vui_sidebar-item .more {
  display: none !important;
}

.fav-sidebar-item:hover .more {
  display: inline-block !important;
}

.tool-more {
  cursor: pointer;
  line-height: 21px;
}

.icon-gengduo {
  color: #9499a0;
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
  color: #18191c;
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: background-color .2s ease;
  cursor: pointer;
}

.be-dropdown-item:hover {
  color: #00aeec;
  background-color: #f0f0f0;
}

.be-dropdown-item.be-dropdown-item-delimiter {
  border-bottom: 1px solid #e3e5e7;
}
</style>
