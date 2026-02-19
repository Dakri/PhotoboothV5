import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import ModeSelectView from '../views/ModeSelectView.vue'
import ClientView from '../views/ClientView.vue'
import GalleryView from '../views/GalleryView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: DashboardView
    },
    {
      path: '/modes',
      name: 'modes',
      component: ModeSelectView
    },
    {
      path: '/client',
      name: 'client',
      component: ClientView
    },
    {
      path: '/gallery',
      name: 'gallery',
      component: GalleryView
    },
    // Legacy routes redirect to new client system
    {
      path: '/buzzer',
      redirect: '/modes'
    },
    {
      path: '/countdown',
      redirect: '/modes'
    },
    {
      path: '/preview',
      redirect: '/modes'
    }
  ]
})

export default router
