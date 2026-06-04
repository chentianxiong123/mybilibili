<template>
  <div class="wf-root">
    <header class="wf-head">
      <span class="logo">mybilibili Studio · AI 工作流</span>
      <router-link to="/edit" custom v-slot="{ navigate }">
        <el-button size="small" @click="navigate">返回剪辑</el-button>
      </router-link>
    </header>
    <div class="wf-grid">
      <aside class="col col-left">
        <div class="pane-h">输入</div>
        <div class="pane-b">
          <el-alert
            v-if="!project.selectedClip"
            type="info" :closable="false"
            title="未选中片段"
            description="在 /edit 选中时间轴上的片段,这里会显示。"
          />
          <div v-else>
            <div class="hint">当前选中</div>
            <el-tag>{{ project.selectedClip.label }}</el-tag>
            <div class="hint" style="margin-top:12px">待处理任务</div>
            <el-tag
              v-for="j in jobs" :key="j.id" type="warning" style="margin:2px">
              {{ j.intent }} · {{ j.clipId.slice(0,6) }}
            </el-tag>
          </div>
          <el-divider />
          <div class="hint">添加节点</div>
          <el-button-group>
            <el-button @click="addNode('script')">脚本</el-button>
            <el-button @click="addNode('tts')">TTS</el-button>
            <el-button @click="addNode('subtitle')">字幕</el-button>
            <el-button @click="addNode('image')">配图</el-button>
          </el-button-group>
        </div>
      </aside>
      <section class="col col-center">
        <div class="pane-h">
          <span>AI 工作流</span>
          <div>
            <el-button size="small" type="primary" @click="run">运行</el-button>
            <el-button size="small" @click="clear">清空</el-button>
          </div>
        </div>
        <div class="canvas-wrap">
          <VueFlow
            v-model:nodes="nodes"
            v-model:edges="edges"
            :default-edge-options="{ animated: true }"
            fit-view-on-init
          >
            <Background pattern-color="#2a2f3a" :gap="16" />
            <MiniMap />
            <Controls />
            <template #node-default="{ data }">
              <div class="wf-node">
                <div class="wf-node-title">{{ data.label }}</div>
                <div class="wf-node-status">{{ data.status || 'idle' }}</div>
              </div>
            </template>
          </VueFlow>
        </div>
      </section>
      <aside class="col col-right">
        <div class="pane-h">输出 / 日志</div>
        <div class="pane-b">
          <div class="hint">最近 10 条</div>
          <ul class="log">
            <li v-for="(l, i) in recentLog" :key="i">{{ formatTime(l.ts) }} · {{ l.msg }}</li>
          </ul>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, shallowRef } from 'vue'
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'
import { useProjectStore } from '@/stores/project'

const project = useProjectStore()

// 跨工作区信号:Edit 选中 clip 时,自动记录到 jobs
const jobs = ref([])
watch(() => project.selectedClipId, (id) => {
  if (!id) return
  const c = project.selectedClip
  if (!c) return
  jobs.value.push({
    id: 'j_' + Date.now(),
    clipId: c.id, intent: 'enhance', ts: Date.now()
  })
})

const nodes = ref([])
const edges = ref([])
const runLog = ref([])

const recentLog = computed(() => runLog.value.slice(-10).reverse())

function formatTime(ts) {
  return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false })
}

const nodeMap = {
  script:   { type: 'script',   label: '脚本生成' },
  tts:      { type: 'tts',      label: 'TTS 配音' },
  subtitle: { type: 'subtitle', label: '字幕生成' },
  image:    { type: 'image',    label: '配图生成' }
}

function addNode(type) {
  nodes.value.push({
    id: 'n_' + Date.now() + '_' + Math.random().toString(36).slice(2, 6),
    position: { x: 80, y: 80 + nodes.value.length * 90 },
    data: { ...nodeMap[type], status: 'idle' }
  })
}

function clear() { nodes.value = []; edges.value = [] }

function run() {
  runLog.value.push({ ts: Date.now(), msg: `运行 ${nodes.value.length} 个节点` })
  for (const n of nodes.value) n.data.status = 'running'
  setTimeout(() => {
    for (const n of nodes.value) n.data.status = 'done'
    // 回填字幕到 Edit 时间轴
    project.tracks.find(t => t.id === 't1').clips.push({
      id: 'c_' + Date.now(),
      mediaId: null,
      label: 'AI 工作流字幕',
      startTime: 0, duration: 4,
      opacity: 1, volume: 1
    })
    runLog.value.push({ ts: Date.now(), msg: '已回填字幕到时间轴 /t1' })
  }, 800)
}
</script>

<style scoped>
.wf-root { display: flex; flex-direction: column; height: 100vh; background: #0f1115; color: #e6e8eb; }
.wf-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 12px; background: #15181f; border-bottom: 1px solid #232733;
}
.logo { font-weight: 700; }
.wf-grid { flex: 1; display: grid; grid-template-columns: 280px 1fr 320px; gap: 6px; padding: 6px; min-height: 0; }
.col { background: #15181f; border: 1px solid #232733; border-radius: 6px; min-height: 0; display: flex; flex-direction: column; overflow: hidden; }
.pane-h {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 12px; font-size: 13px; color: #b8bcc7; border-bottom: 1px solid #232733;
}
.pane-b { padding: 12px; overflow: auto; flex: 1; }
.canvas-wrap { flex: 1; min-height: 0; }
.hint { font-size: 12px; color: #8a8f9b; margin: 6px 0; }
.log { padding: 0; margin: 0; list-style: none; font-size: 12px; }
.log li { padding: 4px 0; color: #b8bcc7; }
.wf-node { background: #1c2030; border: 1px solid #2a3145; border-radius: 6px; padding: 8px 10px; color: #e6e8eb; min-width: 120px; }
.wf-node-title { font-size: 13px; font-weight: 600; }
.wf-node-status { font-size: 11px; color: #8a8f9b; margin-top: 2px; }
</style>
