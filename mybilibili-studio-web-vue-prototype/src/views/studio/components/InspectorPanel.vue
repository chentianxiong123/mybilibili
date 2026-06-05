<template>
  <div class="ins-root">
    <div class="ins-head">属性</div>
    <div class="ins-body">
      <el-empty v-if="!clip" description="选中时间轴上的片段" :image-size="60" />
      <el-form v-else label-position="top" size="small">
        <el-form-item label="名称">
          <el-input :model-value="clip.label" @change="patch({ label: clip.label })" />
        </el-form-item>
        <el-form-item label="轨道">
          <el-tag size="small">{{ clip.trackId }} · {{ clip.trackKind }}</el-tag>
        </el-form-item>
        <el-form-item label="开始时间 (秒)">
          <el-input-number :model-value="Number(clip.startTime.toFixed(2))" :min="0" :step="0.1" :precision="2"
            @change="(v) => patch({ startTime: v })" />
        </el-form-item>
        <el-form-item label="时长 (秒)">
          <el-input-number :model-value="Number(clip.duration.toFixed(2))" :min="0.1" :step="0.1" :precision="2"
            @change="(v) => patch({ duration: v })" />
        </el-form-item>
        <el-form-item v-if="clip.trackKind === 'video'" label="不透明度">
          <el-slider :model-value="clip.opacity ?? 1" :min="0" :max="1" :step="0.05"
            @change="(v) => patch({ opacity: v })" />
        </el-form-item>
        <el-form-item v-if="clip.trackKind === 'audio'" label="音量">
          <el-slider :model-value="clip.volume ?? 1" :min="0" :max="1" :step="0.05"
            @change="(v) => patch({ volume: v })" />
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useProjectStore } from '@/stores/project'
import { syncClipToSprite } from '@/adapters/webav'

const project = useProjectStore()
const clip = computed(() => project.selectedClip)

function patch(p) {
  project.updateClip(clip.value.id, p)
  syncClipToSprite(clip.value.id, p)
}
</script>

<style scoped>
.ins-root { display: flex; flex-direction: column; height: 100%; }
.ins-head {
  padding: 8px 12px; font-size: 13px; color: #b8bcc7;
  border-bottom: 1px solid #232733;
}
.ins-body { flex: 1; overflow: auto; padding: 12px; }
</style>
