import { createRouter, createWebHistory } from 'vue-router'
import Devices from '../views/Devices.vue'
import Domains from '../views/Domains.vue'
import Auth from '../views/Auth.vue'
import Logs from '../views/Logs.vue'

const routes = [
  {
    path: '/',
    redirect: '/devices',
  },
  {
    path: '/devices',
    name: 'Devices',
    component: Devices,
  },
  {
    path: '/domains',
    name: 'Domains',
    component: Domains,
  },
  {
    path: '/auth',
    name: 'Auth',
    component: Auth,
  },
  {
    path: '/logs',
    name: 'Logs',
    component: Logs,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router

