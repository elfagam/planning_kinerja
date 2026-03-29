import { createRouter, createWebHistory } from 'vue-router'
import KontrolPagu from '../views/KontrolPagu.vue'

const router = createRouter({
  history: createWebHistory('/spa/'),
  routes: [
    {
      path: '/pagu-control',
      name: 'PaguControl',
      component: KontrolPagu
    },
    {
      path: '/',
      redirect: '/pagu-control'
    }
  ]
})

export default router
