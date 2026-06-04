import { createRouter, createWebHashHistory } from 'vue-router'
import EditView from '@/views/studio/EditView.vue'
import WorkflowView from '@/views/studio/workflow/WorkflowView.vue'

const routes = [
  { path: '/', redirect: '/edit' },
  { path: '/edit', name: 'edit', component: EditView, meta: { title: '剪辑工作区' } },
  { path: '/workflow', name: 'workflow', component: WorkflowView, meta: { title: 'AI 工作流' } }
]

export default createRouter({
  history: createWebHashHistory(),
  routes
})
