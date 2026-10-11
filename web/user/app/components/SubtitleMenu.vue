<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  anchorX: { type: Number, default: 0 },
  subtitles: { type: Array, default: () => [] },
  currentLanguage: { type: String, default: '' },
  enabled: { type: Boolean, default: true },
  settings: {
    type: Object,
    default: () => ({
      fontScale: 1,
      color: '#ffffff',
      stroke: 0,
      position: 'bottom-center',
      backgroundOpacity: 0.75,
      fitScale: true
    })
  }
})

const emit = defineEmits([
  'update:visible',
  'update:enabled',
  'select-language',
  'update-settings',
  'reset-settings'
])

// 'main' = 语言列表页, 'settings' = 字幕设置页（横向滑动切换）
const panel = ref<'main' | 'settings'>('main')
const menuRef = ref<HTMLElement | null>(null)

// 菜单水平定位：居中于字幕按钮（anchorX），并夹在播放器内
const menuLeftStyle = computed(() => {
  const w = panel.value === 'settings' ? 266 : 168
  const hostW = menuRef.value?.parentElement?.clientWidth || 0
  let left = props.anchorX - w / 2
  if (hostW > 0) {
    left = Math.max(8, Math.min(left, hostW - w - 8))
  }
  return { left: `${Math.max(0, left)}px`, width: `${w}px` }
})

// ===== 下拉状态 =====
const openSelect = ref<string | null>(null)

const closeSelect = (e: Event) => {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    openSelect.value = null
  }
}
watch(
  () => props.visible,
  (v) => {
    if (v) {
      panel.value = 'main'
      openSelect.value = null
      document.addEventListener('click', closeSelect, true)
    } else {
      document.removeEventListener('click', closeSelect, true)
      openSelect.value = null
    }
  }
)

const toggleSelect = (name: string) => {
  openSelect.value = openSelect.value === name ? null : name
}

// ===== 选项定义（对齐B站） =====
const FONT_SCALES = [
  { value: 0.6, label: '最小' },
  { value: 0.8, label: '较小' },
  { value: 1, label: '适中' },
  { value: 1.3, label: '较大' },
  { value: 1.6, label: '最大' }
]

const COLORS = [
  { value: '#ffffff', label: '白色' },
  { value: '#F44336', label: '红色' },
  { value: '#9C27B0', label: '紫色' },
  { value: '#673AB7', label: '深紫色' },
  { value: '#3F51B5', label: '靛青色' },
  { value: '#2196F3', label: '蓝色' },
  { value: '#03A9F4', label: '亮蓝色' }
]

const STROKES = [
  { value: 0, label: '无描边' },
  { value: 1, label: '重墨' },
  { value: 2, label: '描边' },
  { value: 3, label: '45°投影' }
]

const POSITIONS = [
  { value: 'bottom-left', label: '左下角' },
  { value: 'bottom-center', label: '底部居中' },
  { value: 'bottom-right', label: '右下角' },
  { value: 'top-left', label: '左上角' },
  { value: 'top-center', label: '顶部居中' },
  { value: 'top-right', label: '右上角' }
]

const s = computed(() => props.settings)

const currentFontLabel = computed(() => FONT_SCALES.find(i => i.value === s.value.fontScale)?.label || '适中')
const currentColorLabel = computed(() => COLORS.find(i => i.value === s.value.color)?.label || '白色')
const currentStrokeLabel = computed(() => STROKES.find(i => i.value === s.value.stroke)?.label || '无描边')
const currentPositionLabel = computed(() => POSITIONS.find(i => i.value === s.value.position)?.label || '底部居中')
const opacityPercent = computed(() => Math.round(s.value.backgroundOpacity * 100))

const hasSubtitles = computed(() => props.subtitles.length > 0)

const patch = (p: any) => emit('update-settings', p)

// ===== 字幕开关（B站的"关闭"项） =====
const selectLanguage = (language: string) => {
  if (!props.enabled) emit('update:enabled', true)
  emit('select-language', language)
}

const closeMenu = () => emit('update:visible', false)
</script>

<template>
  <div v-show="visible" ref="menuRef" class="bsub-menu" :class="`panel-${panel}`" :style="menuLeftStyle" @click.stop>
    <div class="bsub-move" :class="{ 'move-settings': panel === 'settings' }">
      <!-- 面板1: 语言列表 -->
      <div class="bsub-panel bsub-panel-main">
        <div class="bsub-header">
          <span class="bsub-title">字幕</span>
          <span class="bsub-close" :class="{ active: !enabled }" @click="() => { emit('update:enabled', !enabled); if (enabled) closeSelect({ stopPropagation(){}, target: menuRef } as any) }">
            {{ enabled ? '关闭' : '已关闭' }}
          </span>
        </div>

        <div class="bsub-language">
          <div v-if="!hasSubtitles" class="bsub-nolan">暂无字幕</div>
          <div v-else class="bsub-language-list">
            <div
              v-for="sub in subtitles"
              :key="sub.language"
              class="bsub-language-item"
              :class="{ active: sub.language === currentLanguage && enabled }"
              @click="selectLanguage(sub.language)"
            >
              <span class="bsub-language-item-text">{{ sub.languageName || sub.language_name || sub.language }}</span>
              <span v-show="sub.language === currentLanguage && enabled" class="bsub-cc-icon">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 12">
                  <rect width="13.5" height="11.5" x=".25" y=".25" stroke="#fff" stroke-opacity=".5" stroke-width=".5" rx="5.75" />
                  <path fill="#fff" d="M5.248 2.788h.76L8.256 8.5h-.712l-.608-1.6H4.312l-.608 1.6H3l2.248-5.712Zm-.728 3.56h2.208l-1.08-2.856h-.032L4.52 6.348Zm4.362-3.56h.648V8.5h-.648V2.788Z" />
                </svg>
              </span>
            </div>
          </div>
        </div>

        <div class="bsub-separator"></div>
        <div class="bsub-setting-entry" @click="panel = 'settings'">
          <span class="bsub-setting-entry-text">字幕设置</span>
          <span class="bsub-chevron">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 6 10"><path fill="#fff" d="M4.93.15c.231-.213.613-.197.865.035l.038.034a.556.556 0 0 1 0 .795L1.538 5l4.295 4.017a.557.557 0 0 1 0 .796l-.053.047c-.26.19-.63.19-.851-.014L.167 5.396a.557.557 0 0 1 0-.795L4.93.15Z" /></svg>
          </span>
        </div>
      </div>

      <!-- 面板2: 字幕设置 -->
      <div class="bsub-panel bsub-panel-settings">
        <div class="bsub-header">
          <span class="bsub-back" @click="panel = 'main'">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 6 10"><path fill="#fff" d="M1.07.15C.839-.063.457-.047.205.185l-.038.034a.556.556 0 0 0 0 .795L4.462 5 .167 9.017a.557.557 0 0 0 0 .796l.053.047c.26.19.63.19.851-.014L5.833 5.396a.557.557 0 0 0 0-.795L1.07.15Z" /></svg>
          </span>
          <span class="bsub-title">字幕设置</span>
          <span class="bsub-close" @click="closeMenu">关闭</span>
        </div>

        <div class="bsub-settings-body">
          <div class="bsub-item-row">
            <div class="bsub-field">
              <div class="bsub-field-title">字体大小</div>
              <div class="bsub-select" :class="{ open: openSelect === 'font' }" @click.stop="toggleSelect('font')">
                <span class="bsub-select-result">{{ currentFontLabel }}</span>
                <span class="bsub-select-arrow"></span>
                <ul class="bsub-select-list">
                  <li
                    v-for="opt in FONT_SCALES" :key="opt.value"
                    :class="{ active: opt.value === s.fontScale }"
                    @click="patch({ fontScale: opt.value })"
                  >{{ opt.label }}</li>
                </ul>
              </div>
            </div>
            <div class="bsub-field">
              <div class="bsub-field-title">字体颜色</div>
              <div class="bsub-select" :class="{ open: openSelect === 'color' }" @click.stop="toggleSelect('color')">
                <span class="bsub-select-result"><i class="bsub-color-dot" :style="{ background: s.color }"></i>{{ currentColorLabel }}</span>
                <span class="bsub-select-arrow"></span>
                <ul class="bsub-select-list">
                  <li
                    v-for="opt in COLORS" :key="opt.value"
                    :class="{ active: opt.value === s.color }"
                    @click="patch({ color: opt.value })"
                  ><i class="bsub-color-dot" :style="{ background: opt.value }"></i>{{ opt.label }}</li>
                </ul>
              </div>
            </div>
          </div>

          <div class="bsub-item-row">
            <div class="bsub-field">
              <div class="bsub-field-title">描边方式</div>
              <div class="bsub-select" :class="{ open: openSelect === 'stroke' }" @click.stop="toggleSelect('stroke')">
                <span class="bsub-select-result">{{ currentStrokeLabel }}</span>
                <span class="bsub-select-arrow"></span>
                <ul class="bsub-select-list">
                  <li
                    v-for="opt in STROKES" :key="opt.value"
                    :class="{ active: opt.value === s.stroke }"
                    @click="patch({ stroke: opt.value })"
                  >{{ opt.label }}</li>
                </ul>
              </div>
            </div>
            <div class="bsub-field">
              <div class="bsub-field-title">默认位置</div>
              <div class="bsub-select" :class="{ open: openSelect === 'position' }" @click.stop="toggleSelect('position')">
                <span class="bsub-select-result">{{ currentPositionLabel }}</span>
                <span class="bsub-select-arrow"></span>
                <ul class="bsub-select-list">
                  <li
                    v-for="opt in POSITIONS" :key="opt.value"
                    :class="{ active: opt.value === s.position }"
                    @click="patch({ position: opt.value })"
                  >{{ opt.label }}</li>
                </ul>
              </div>
            </div>
          </div>

          <div class="bsub-item-single">
            <div class="bsub-opacity-header">
              <span class="bsub-field-title">背景不透明度</span>
              <span class="bsub-opacity-percent">{{ opacityPercent }}%</span>
            </div>
            <div class="bsub-slider">
              <input
                type="range" min="0" max="100" :value="opacityPercent"
                @input="patch({ backgroundOpacity: Number(($event.target as HTMLInputElement).value) / 100 })"
              />
            </div>
          </div>

          <div class="bsub-item-single">
            <div class="bsub-field-title">其它设置</div>
            <div class="bsub-checkbox-row" @click="patch({ fitScale: !s.fitScale })">
              <span class="bsub-checkbox" :class="{ checked: s.fitScale }">
                <svg v-show="s.fitScale" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="#00AEEC" d="m6.2 11.4-3-3 1.2-1.2 1.8 1.8 5-5L12.4 5z" /></svg>
              </span>
              <span class="bsub-checkbox-name">等比缩放</span>
            </div>
          </div>
        </div>

        <div class="bsub-footer">
          <div class="bsub-separator"></div>
          <div class="bsub-reset" @click="emit('reset-settings')">恢复默认设置</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bsub-menu {
  position: absolute;
  bottom: 52px;
  z-index: 120;
  overflow: hidden;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.82);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  color: #fff;
  font-size: 12px;
  line-height: 1.4;
  transition: width 0.28s cubic-bezier(0.25, 0.1, 0.25, 1);
}

.bsub-move {
  position: relative;
  display: block;
  width: 168px;
  height: 148px;
  transition: transform 0.28s cubic-bezier(0.25, 0.1, 0.25, 1), height 0.28s cubic-bezier(0.25, 0.1, 0.25, 1);
  transform: translateX(0);
}
.bsub-move.move-settings {
  transform: translateX(-168px);
  height: 296px;
}

.bsub-panel {
  box-sizing: border-box;
  position: absolute;
  top: 0;
  display: flex;
  flex-direction: column;
}
.bsub-panel-main {
  left: 0;
  width: 168px;
  height: 148px;
  padding-bottom: 4px;
}
.bsub-panel-settings {
  left: 168px;
  width: 266px;
  height: 296px;
}

/* 头部 */
.bsub-header {
  display: flex;
  align-items: center;
  padding: 10px 12px 6px;
  gap: 8px;
}
.bsub-title {
  flex: 1;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
}
.bsub-close {
  cursor: pointer;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
}
.bsub-close:hover,
.bsub-close.active {
  color: #00AEEC;
}
.bsub-back {
  cursor: pointer;
  display: flex;
  width: 14px;
}
.bsub-back svg { width: 6px; height: 10px; }

/* 语言列表 */
.bsub-language {
  padding: 2px 8px 4px;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.bsub-nolan {
  padding: 12px;
  text-align: center;
  color: rgba(255, 255, 255, 0.5);
}
.bsub-language-item {
  display: flex;
  align-items: center;
  padding: 7px 12px;
  border-radius: 6px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.9);
}
.bsub-language-item:hover {
  background: rgba(255, 255, 255, 0.1);
}
.bsub-language-item.active {
  color: #00AEEC;
  background: rgba(0, 174, 236, 0.12);
}
.bsub-language-item-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bsub-cc-icon {
  display: flex;
  margin-left: 6px;
}
.bsub-cc-icon svg { width: 14px; height: 12px; }

.bsub-separator {
  height: 1px;
  background: rgba(255, 255, 255, 0.12);
  margin: 4px 0;
}

/* 字幕设置入口 */
.bsub-setting-entry {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
}
.bsub-setting-entry:hover { background: rgba(255, 255, 255, 0.08); }
.bsub-setting-entry-text { flex: 1; }
.bsub-chevron svg { width: 6px; height: 10px; opacity: 0.7; }

/* 设置面板 */
.bsub-settings-body {
  padding: 4px 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.bsub-item-row {
  display: flex;
  gap: 10px;
}
.bsub-field { flex: 1; min-width: 0; }
.bsub-field-title {
  color: rgba(255, 255, 255, 0.65);
  margin-bottom: 5px;
  font-size: 12px;
}

/* 下拉（仿 bui-select） */
.bsub-select {
  position: relative;
  display: flex;
  align-items: center;
  height: 28px;
  padding: 0 8px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.08);
  cursor: pointer;
  user-select: none;
}
.bsub-select:hover { background: rgba(255, 255, 255, 0.12); }
.bsub-select-result {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  white-space: nowrap;
}
.bsub-select-arrow {
  width: 0;
  height: 0;
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  border-top: 5px solid rgba(255, 255, 255, 0.6);
  transition: transform 0.15s;
}
.bsub-select.open .bsub-select-arrow { transform: rotate(180deg); }
.bsub-select-list {
  display: none;
  position: absolute;
  top: 32px;
  left: 0;
  right: 0;
  margin: 0;
  padding: 4px 0;
  list-style: none;
  border-radius: 6px;
  background: rgba(20, 20, 20, 0.96);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  z-index: 10;
  max-height: 160px;
  overflow-y: auto;
}
.bsub-select.open .bsub-select-list { display: block; }
.bsub-select-list li {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  cursor: pointer;
}
.bsub-select-list li:hover { background: rgba(255, 255, 255, 0.08); }
.bsub-select-list li.active { color: #00AEEC; }
.bsub-color-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.25);
  flex-shrink: 0;
}

/* 不透明度滑条 */
.bsub-opacity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.bsub-opacity-percent { color: rgba(255, 255, 255, 0.9); }
.bsub-slider input[type="range"] {
  width: 100%;
  height: 4px;
  appearance: none;
  -webkit-appearance: none;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
  outline: none;
  cursor: pointer;
}
.bsub-slider input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 0 2px rgba(0, 0, 0, 0.5);
}

/* 复选框 */
.bsub-checkbox-row {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  margin-top: 4px;
  user-select: none;
}
.bsub-checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.06);
}
.bsub-checkbox.checked {
  border-color: #00AEEC;
  background: rgba(0, 174, 236, 0.15);
}
.bsub-checkbox svg { width: 14px; height: 14px; }
.bsub-checkbox-name { color: rgba(255, 255, 255, 0.85); }

/* 底部恢复默认 */
.bsub-footer { padding-bottom: 6px; }
.bsub-reset {
  text-align: center;
  padding: 8px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.75);
  border-radius: 6px;
}
.bsub-reset:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

/* 滚动条 */
.bsub-language::-webkit-scrollbar,
.bsub-select-list::-webkit-scrollbar { width: 4px; }
.bsub-language::-webkit-scrollbar-thumb,
.bsub-select-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.25);
  border-radius: 2px;
}
</style>
