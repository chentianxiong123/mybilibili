<template>
  <div class="bcc-pagination-container">
    <div class="bcc-pagination">
      <span
        v-if="props.currentPage > 1"
        class="bcc-pagination-item prev"
        @click="setPage(props.currentPage - 1)"
      >
        上一页
      </span>
      <span
        v-for="item in pageItems"
        :key="item.key"
        class="bcc-pagination-item"
        :class="{ selected: item.type === 'number' && item.page === currentPage }"
        @click="handleClick(item)"
      >
        {{ item.type === 'number' ? item.page : '⋯' }}
      </span>
      <span
        v-if="props.currentPage < totalPages"
        class="bcc-pagination-item next"
        @click="setPage(props.currentPage + 1)"
      >
        下一页
      </span>
      <span class="bcc-pagination-extra">
        共{{ totalPages }}页 / {{ total }}个， 跳转至
        <el-input-number
          v-model="jumpPage"
          :min="1"
          :max="totalPages"
          :controls="false"
          class="bcc-pagination-input"
          @change="handleJump"
        />
        页
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps({
  total: { type: Number, default: 0 },
  pageSize: { type: Number, default: 10 },
  currentPage: { type: Number, default: 1 }
})

const emit = defineEmits(['update:currentPage', 'current-change'])

const jumpPage = ref(props.currentPage)

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

// 生成页码项：1 ... 5 6 7 ... 9 （对齐 bcc-pagination）
const pageItems = computed(() => {
  const total = totalPages.value
  const current = props.currentPage
  const arr = []
  const add = (type, page, key) => {
    if (page < 1 || page > total) return
    arr.push({ type, page, key })
  }
  if (total <= 7) {
    for (let i = 1; i <= total; i++) add('number', i, `n-${i}`)
  } else {
    add('number', 1, 'n-1')
    if (current > 3) arr.push({ type: 'ellipsis', key: 'e-1' })
    for (let i = Math.max(2, current - 1); i <= Math.min(total - 1, current + 1); i++) {
      add('number', i, `n-${i}`)
    }
    if (current < total - 2) arr.push({ type: 'ellipsis', key: 'e-2' })
    add('number', total, `n-${total}`)
  }
  return arr
})

const handleClick = (item) => {
  if (item.type === 'number') {
    setPage(item.page)
  }
}

const setPage = (page) => {
  const p = Math.min(totalPages.value, Math.max(1, page))
  jumpPage.value = p
  emit('update:currentPage', p)
  emit('current-change', p)
}

const handleJump = (val) => {
  if (!val) return
  setPage(val)
}
</script>

<style scoped>
.bcc-pagination-container {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  height: 40px;
  margin-top: 20px;
  padding: 15px 0;
  border-top: 1px solid #e0e0e0;
}

.bcc-pagination {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
}

.bcc-pagination-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 40px;
  padding: 0 11px;
  margin-right: 8px;
  font-size: 14px;
  color: #505050;
  background: #fff;
  border: 1px solid #e7e7e7;
  border-radius: 2px;
  cursor: pointer;
  box-sizing: border-box;
  user-select: none;
  transition: border-color .2s, color .2s;
}

.bcc-pagination-item:hover {
  color: var(--brand_pink, #ff6699);
  border-color: var(--brand_pink, #ff6699);
}

.bcc-pagination-item.selected {
  color: #fff;
  background: var(--brand_pink, #ff6699);
  border-color: var(--brand_pink, #ff6699);
}

.bcc-pagination-item.selected:hover {
  color: #fff;
}

.bcc-pagination-item.next,
.bcc-pagination-item.prev {
  min-width: 76px;
}

.bcc-pagination-extra {
  margin-left: 32px;
  font-size: 12px;
  color: #505050;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.bcc-pagination-input :deep(.el-input__wrapper) {
  width: 48px;
  height: 28px;
  padding: 0 4px;
  font-size: 12px;
  color: #505050;
  background: #fff;
  box-shadow: 0 0 0 1px #e7e7e7 inset;
}

.bcc-pagination-input :deep(.el-input__inner) {
  text-align: center;
}
</style>
