import { createRouter, createWebHashHistory } from 'vue-router'
import OverviewPage from '../pages/OverviewPage.vue'
import LogsPage from '../pages/LogsPage.vue'

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/overview',
    },
    {
      path: '/overview',
      name: 'overview',
      component: OverviewPage,
    },
    {
      path: '/logs',
      name: 'logs',
      component: LogsPage,
    },
  ],
})
