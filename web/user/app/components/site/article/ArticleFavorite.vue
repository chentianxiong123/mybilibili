<template>
  <el-dialog
    :model-value="modelValue"
    title="收藏至"
    width="440px"
    append-to-body
    :close-on-click-modal="false"
    @update:model-value="close"
    @open="loadFolders"
  >
    <div class="af-body" v-loading="loading">
      <div class="af-new">
        <input
          v-model="newName"
          class="af-new__input"
          maxlength="20"
          placeholder="新建收藏夹"
          @keyup.enter="createFolder"
        />
        <button class="af-new__btn" :disabled="creating || !newName.trim()" @click="createFolder">新建</button>
      </div>

      <div class="af-list">
        <label v-for="f in folders" :key="f.id" class="af-item">
          <input type="checkbox" :checked="checked.has(f.id)" @change="toggle(f.id, $event)" />
          <span class="af-item__name">{{ f.name }}</span>
          <span class="af-item__count">{{ f.count }}个内容</span>
        </label>
        <div v-if="!loading && !folders.length" class="af-empty">还没有收藏夹，先新建一个吧</div>
      </div>
    </div>

    <template #footer>
      <button class="af-btn af-btn--plain" @click="close(false)">取消</button>
      <button class="af-btn af-btn--primary" :disabled="saving" @click="confirm">确定</button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { articleApi } from '@/api/article'

const props = defineProps<{
  modelValue: boolean
  articleId: number
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'changed', payload: { favorited: boolean; count: number }): void
}>()

const folders = ref<any[]>([])
const checked = ref<Set<number>>(new Set())
const origin = ref<Set<number>>(new Set())
const loading = ref(false)
const saving = ref(false)
const creating = ref(false)
const newName = ref('')

watch(
  () => props.modelValue,
  (v) => {
    if (v) {
      newName.value = ''
      loadFolders()
    }
  }
)

async function loadFolders() {
  if (!props.articleId) return
  loading.value = true
  try {
    const res: any = await articleApi.getFolders(props.articleId)
    if (res.code === 200) {
      folders.value = res.data || []
      const set = new Set<number>()
      folders.value.forEach((f) => {
        if (f.checked) set.add(f.id)
      })
      checked.value = new Set(set)
      origin.value = new Set(set)
    }
  } catch {
    folders.value = []
  } finally {
    loading.value = false
  }
}

function toggle(id: number, e: Event) {
  const on = (e.target as HTMLInputElement).checked
  const next = new Set(checked.value)
  if (on) next.add(id)
  else next.delete(id)
  checked.value = next
}

async function createFolder() {
  const name = newName.value.trim()
  if (!name) return
  creating.value = true
  try {
    const res: any = await articleApi.createFolder(name)
    if (res.code === 200) {
      newName.value = ''
      const id = res.data.id
      const next = new Set(checked.value)
      next.add(id)
      checked.value = next
      await loadFolders()
      const after = new Set(checked.value)
      after.add(id)
      checked.value = after
      ElMessage.success('新建成功')
    } else {
      ElMessage.error(res.message || '新建失败')
    }
  } catch {
    ElMessage.error('新建失败')
  } finally {
    creating.value = false
  }
}

async function confirm() {
  saving.value = true
  try {
    const adds = [...checked.value].filter((id) => !origin.value.has(id))
    const removes = [...origin.value].filter((id) => !checked.value.has(id))
    const res: any = await articleApi.collect(props.articleId, adds, removes)
    if (res.code === 200) {
      emit('changed', { favorited: res.data.favorited, count: res.data.count })
      ElMessage.success(res.data.favorited ? '已收藏' : '已取消收藏')
      emit('update:modelValue', false)
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch {
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

function close(v = false) {
  emit('update:modelValue', v)
}
</script>

<style scoped>
.af-body { min-height: 120px; }
.af-new { display: flex; gap: 8px; margin-bottom: 14px; }
.af-new__input {
  flex: 1; height: 32px; border: 1px solid #e3e5e7; border-radius: 4px; padding: 0 10px;
  font-size: 13px; outline: none; box-sizing: border-box;
}
.af-new__input:focus { border-color: #00aeec; }
.af-new__btn {
  height: 32px; padding: 0 16px; border: 1px solid #00aeec; border-radius: 4px;
  background: #fff; color: #00aeec; font-size: 13px; cursor: pointer;
}
.af-new__btn:disabled { border-color: #e3e5e7; color: #c9ccd0; cursor: not-allowed; }
.af-list { max-height: 300px; overflow-y: auto; }
.af-item {
  display: flex; align-items: center; gap: 10px; padding: 10px 4px; cursor: pointer;
  border-bottom: 1px solid #f6f7f8; font-size: 14px; color: #18191c;
}
.af-item:last-child { border-bottom: none; }
.af-item input { width: 16px; height: 16px; accent-color: #00aeec; }
.af-item__name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.af-item__count { font-size: 12px; color: #9499a0; }
.af-empty { text-align: center; color: #9499a0; font-size: 13px; padding: 30px 0; }
.af-btn {
  height: 34px; padding: 0 22px; border-radius: 4px; font-size: 14px; cursor: pointer; margin-left: 10px;
}
.af-btn--plain { border: 1px solid #e3e5e7; background: #fff; color: #61666d; }
.af-btn--primary { border: none; background: #00aeec; color: #fff; }
.af-btn--primary:disabled { background: #c9ccd0; cursor: not-allowed; }
</style>
