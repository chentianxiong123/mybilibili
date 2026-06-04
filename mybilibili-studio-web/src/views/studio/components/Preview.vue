<template>
  <div class="prev-root">
    <div class="prev-head">
      <span>预览</span>
      <span class="dim">{{ project.project.width }}×{{ project.project.height }} · WebAV</span>
    </div>
    <div class="prev-stage">
      <div class="prev-wrap" :style="wrapStyle">
        <canvas ref="canvasEl" class="prev-canvas" />
        <div v-if="engineError" class="prev-error">
          <el-alert type="error" :title="webavSupported ? '画布错误' : '浏览器不支持 WebCodecs'" :description="engineError" :closable="false" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useEngineStore } from '@/stores/engine'
import { initCanvas, canvasInstance, webavError } from '@/adapters/webav'

const project = useProjectStore()
const engine = useEngineStore()
const canvasEl = ref(null)
const engineError = ref(null)

const wrapStyle = computed(() => ({
  aspectRatio: `${project.project.width} / ${project.project.height}`
}))

onMounted(async () => {
  try {
    await initCanvas(canvasEl.value, project.project.width, project.project.height)
    engineError.value = null
  } catch (e) {
    engineError.value = e.message
  }
})

onBeforeUnmount(() => {
  canvasInstance.value?.destroy?.()
  canvasInstance.value = null
})

// 监听 project 宽高变化:重新初始化画布
watch(() => [project.project.width, project.project.height], async ([w, h]) => {
  if (!canvasEl.value) return
  try {
    canvasInstance.value?.destroy?.()
    await initCanvas(canvasEl.value, w, h)
  } catch (e) { engineError.value = e.message }
})
</script>

<style scoped>
.prev-root { display: flex; flex-direction: column; height: 100%; min-height: 0; }
.prev-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 6px 10px; font-size: 12px; color: #b8bcc7;
  border-bottom: 1px solid #232733;
}
.dim { color: #8a8f9b; }
.prev-stage {
  flex: 1; display: flex; align-items: center; justify-content: center;
  background: #0a0c10; padding: 16px; min-height: 0;
}
.prev-wrap {
  max-width: 100%; max-height: 100%; background: #000;
  border-radius: 4px; overflow: hidden; position: relative;
}
.prev-canvas { width: 100%; height: 100%; display: block; }
.prev-error { position: absolute; top: 8px; left: 8px; right: 8px; }
</style>
