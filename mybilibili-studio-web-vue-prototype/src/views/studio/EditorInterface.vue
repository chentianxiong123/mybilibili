<template>
  <div class="editor-root">
    <Toolbar />
    <div class="editor-grid" :style="gridStyle">
      <aside class="col col-left" :style="{ width: ui.mediaWidth + 'px' }">
        <AssetsPanel />
      </aside>
      <section class="col col-center">
        <Preview />
      </section>
      <aside class="col col-right" :style="{ width: ui.inspectorWidth + 'px' }">
        <InspectorPanel />
      </aside>
    </div>
    <div class="editor-bottom" :style="{ height: ui.timelineHeight + 'px' }">
      <Timeline />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUIStore } from '@/stores/ui'
import Toolbar from './components/Toolbar.vue'
import AssetsPanel from './components/AssetsPanel.vue'
import Preview from './components/Preview.vue'
import InspectorPanel from './components/InspectorPanel.vue'
import Timeline from './components/Timeline.vue'

const ui = useUIStore()
const gridStyle = computed(() => ({
  gridTemplateRows: `1fr ${ui.timelineHeight}px`,
  gridTemplateColumns: `${ui.mediaWidth}px 1fr ${ui.inspectorWidth}px`
}))
</script>

<style scoped>
.editor-root {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #0f1115;
  color: #e6e8eb;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  overflow: hidden;
}
.editor-grid {
  flex: 1;
  display: grid;
  gap: 6px;
  padding: 6px;
  min-height: 0;
}
.col {
  background: #15181f;
  border: 1px solid #232733;
  border-radius: 6px;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.editor-bottom {
  border-top: 1px solid #232733;
  background: #15181f;
  padding: 0 6px 6px 6px;
}
</style>
