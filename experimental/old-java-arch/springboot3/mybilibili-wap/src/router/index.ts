import { createRouter, createWebHistory } from 'vue-router'
import mobileRoutes from './mobile'

const router = createRouter({
  history: createWebHistory(),
  routes: mobileRoutes
})

export default router