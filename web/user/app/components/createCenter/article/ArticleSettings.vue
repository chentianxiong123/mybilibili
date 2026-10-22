<template>
  <div class="publish-settings">
    <div class="publish-settings-title">发布设置</div>
    <div class="form">
      <!-- 可见范围 -->
      <div class="form-item">
        <label class="form-item-label">可见范围</label>
        <div class="form-item-control">
          <label class="radio" @click="patch('visibility', 1)">
            <span class="radio-box" :class="{ checked: modelValue.visibility === 1 }"></span>
            <span class="radio-label">所有人可见</span>
          </label>
          <label class="radio" @click="patch('visibility', 0)">
            <span class="radio-box" :class="{ checked: modelValue.visibility === 0 }"></span>
            <span class="radio-label">仅自己可见 <span class="tip">（将不支持分享、商业推广）</span></span>
          </label>
        </div>
      </div>

      <!-- 自定义封面 -->
      <div class="form-item">
        <label class="form-item-label">自定义封面</label>
        <div class="form-item-control">
          <div class="form-item-center">
            <span class="switch" :class="{ on: modelValue.coverEnabled }" @click="patch('coverEnabled', !modelValue.coverEnabled)">
              <span class="switch-dot"></span>
            </span>
            <span class="tip">（未设置自定义封面则自动抓取正文开头部分文字作为封面）</span>
          </div>
          <div v-if="modelValue.coverEnabled" class="cover-box" @click="pickCover">
            <img v-if="modelValue.coverUrl" :src="modelValue.coverUrl" alt="" />
            <template v-else>
              <span class="cover-plus">+</span>
              <span class="cover-hint">点击上传封面</span>
            </template>
          </div>
        </div>
      </div>

      <!-- 评论开关 -->
      <div class="form-item">
        <label class="form-item-label">评论开关</label>
        <div class="form-item-control form-item-center">
          <span class="switch" :class="{ on: modelValue.commentEnabled }" @click="patch('commentEnabled', !modelValue.commentEnabled)">
            <span class="switch-dot"></span>
          </span>
          <span class="tip">（关闭后，用户无法在你的专栏发表评论）</span>
        </div>
      </div>

      <!-- 精选评论 -->
      <div class="form-item">
        <label class="form-item-label">精选评论</label>
        <div class="form-item-control form-item-center">
          <span class="switch" :class="{ on: modelValue.featuredComment }" @click="patch('featuredComment', !modelValue.featuredComment)">
            <span class="switch-dot"></span>
          </span>
          <span class="tip">（开启后，经过你筛选的评论才会向用户展示）</span>
        </div>
      </div>

      <!-- 创作声明 -->
      <div class="form-item">
        <label class="form-item-label">创作声明</label>
        <div class="form-item-control">
          <label class="checkbox" @click="patch('isOriginal', !modelValue.isOriginal)">
            <span class="checkbox-box" :class="{ checked: modelValue.isOriginal }"></span>
            <span class="checkbox-label">声明此文章为原创，未经授权禁止转载</span>
          </label>
          <label class="checkbox" @click="patch('aiDeclaration', !modelValue.aiDeclaration)">
            <span class="checkbox-box" :class="{ checked: modelValue.aiDeclaration }"></span>
            <span class="checkbox-label">
              <span class="decl-title">AI辅助创作声明</span>
              <span class="decl-tip">勾选后，图片将标识该内容使用人工智能合成技术</span>
            </span>
          </label>
        </div>
      </div>

      <!-- 话题 -->
      <div class="form-item">
        <label class="form-item-label">话题</label>
        <div class="form-item-control form-item-center">
          <button class="btn-topic" @click="addTopic">
            <svg viewBox="0 0 16 16" width="14" height="14"><path d="M8 3v10M3 8h10" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
            添加话题
          </button>
          <span v-if="modelValue.topics" class="topic-tag">{{ modelValue.topics }}</span>
        </div>
      </div>

      <!-- 文集 -->
      <div class="form-item">
        <label class="form-item-label">文集</label>
        <div class="form-item-control form-item-center">
          <button class="btn-topic" @click="addCollection">
            <svg viewBox="0 0 16 16" width="14" height="14"><path d="M8 3v10M3 8h10" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
            选择文集
          </button>
          <span v-if="modelValue.collectionName" class="topic-tag">{{ modelValue.collectionName }}</span>
        </div>
      </div>
    </div>

    <input ref="coverFileRef" type="file" accept="image/*" style="display:none" @change="onCoverPicked" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { articleApi } from '@/api/article'

export interface ArticleSettings {
  visibility: number
  coverEnabled: boolean
  coverUrl: string
  commentEnabled: boolean
  featuredComment: boolean
  isOriginal: boolean
  aiDeclaration: boolean
  topics: string
  collectionName: string
}

const props = defineProps<{ modelValue: ArticleSettings }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: ArticleSettings): void }>()

const coverFileRef = ref<HTMLInputElement | null>(null)

function patch<K extends keyof ArticleSettings>(key: K, value: ArticleSettings[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function addTopic() {
  const t = window.prompt('请输入话题（可含 # 号）', props.modelValue.topics)
  if (t === null) return
  patch('topics', t.trim())
}
function addCollection() {
  const t = window.prompt('请输入文集名称', props.modelValue.collectionName)
  if (t === null) return
  patch('collectionName', t.trim())
}
function pickCover() {
  coverFileRef.value?.click()
}
async function onCoverPicked(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    const res: any = await articleApi.uploadImage(file)
    const url = res?.data?.url || res?.url
    if (url) patch('coverUrl', url)
    else ElMessage.error('封面上传失败')
  } catch {
    ElMessage.error('封面上传失败')
  }
}
</script>

<style scoped>
.publish-settings {
  background: #fff;
  border-radius: 2px;
  padding: 16px 20px;
}
.publish-settings-title {
  font-size: 16px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 12px;
}
.form-item { padding: 10px 0; border-bottom: 1px solid #f6f7f8; }
.form-item:last-child { border-bottom: none; }
.form-item-label {
  display: block;
  font-size: 13px;
  color: #61666d;
  margin-bottom: 8px;
}
.form-item-control { display: block; }
.form-item-center { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }

.radio { display: inline-flex; align-items: center; margin-right: 20px; cursor: pointer; }
.radio-box {
  width: 14px; height: 14px; border-radius: 50%; border: 1px solid #c9ccd0; margin-right: 6px; position: relative; box-sizing: border-box;
}
.radio-box.checked { border-color: #00aeec; }
.radio-box.checked::after {
  content: ''; position: absolute; inset: 3px; border-radius: 50%; background: #00aeec;
}
.radio-label { font-size: 13px; color: #18191c; }

.checkbox { display: flex; align-items: flex-start; margin-bottom: 8px; cursor: pointer; }
.checkbox-box {
  width: 14px; height: 14px; border-radius: 3px; border: 1px solid #c9ccd0; margin-right: 8px; margin-top: 2px; position: relative; box-sizing: border-box; flex: 0 0 auto;
}
.checkbox-box.checked { border-color: #00aeec; background: #00aeec; }
.checkbox-box.checked::after {
  content: ''; position: absolute; left: 4px; top: 1px; width: 4px; height: 8px; border: solid #fff; border-width: 0 1.6px 1.6px 0; transform: rotate(45deg);
}
.checkbox-label { font-size: 13px; color: #18191c; }
.decl-title { display: block; }
.decl-tip { display: block; font-size: 12px; color: #9499a0; margin-top: 2px; }

.tip { font-size: 12px; color: #9499a0; }

.switch {
  position: relative; width: 40px; height: 20px; border-radius: 10px; background: #c9ccd0; display: inline-block; transition: background .2s; cursor: pointer; flex: 0 0 auto;
}
.switch.on { background: #00aeec; }
.switch-dot {
  position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; border-radius: 50%; background: #fff; transition: left .2s;
}
.switch.on .switch-dot { left: 22px; }

.cover-box {
  margin-top: 10px; width: 160px; height: 100px; border: 1px dashed #d0d3d6; border-radius: 4px; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #9499a0; cursor: pointer; overflow: hidden;
}
.cover-box img { width: 100%; height: 100%; object-fit: cover; }
.cover-plus { font-size: 24px; line-height: 1; }
.cover-hint { font-size: 12px; margin-top: 4px; }

.btn-topic {
  display: inline-flex; align-items: center; gap: 4px; height: 28px; padding: 0 12px; border: 1px solid #e3e5e7; border-radius: 14px; background: #fff; color: #00aeec; font-size: 13px; cursor: pointer;
}
.btn-topic:hover { border-color: #00aeec; }
.topic-tag { font-size: 13px; color: #00aeec; }
</style>
