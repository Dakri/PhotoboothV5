import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import BuzzerView from '../views/BuzzerView.vue'
import CountdownView from '../views/CountdownView.vue'
import PreviewView from '../views/PreviewView.vue'
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
      path: '/buzzer',
      name: 'buzzer',
      component: BuzzerView
    },
    {
      path: '/countdown',
      name: 'countdown',
      component: CountdownView
    },
    {
      path: '/preview',
      name: 'preview',
      component: PreviewView
    },
    {
      path: '/gallery',
      name: 'gallery',
      component: GalleryView
    }
  ]
})

export default router
