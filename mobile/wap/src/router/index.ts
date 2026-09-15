import { createRouter, createWebHistory } from 'vue-router'
import mobileRoutes from './mobile'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: mobileRoutes
})

export default router
