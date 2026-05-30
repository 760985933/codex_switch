import { createRouter, createWebHashHistory } from 'vue-router'
import ProxyPage from '../pages/ProxyPage.vue'
import LogsPage from '../pages/LogsPage.vue'
import SessionsPage from '../pages/SessionsPage.vue'
import ContactPage from '../pages/ContactPage.vue'
import MonitoringPage from '../pages/MonitoringPage.vue'
import ModelsPage from '../pages/ModelsPage.vue'
import SettingsPage from '../pages/SettingsPage.vue'

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/proxy',
    },
    {
      path: '/proxy',
      name: 'proxy',
      component: ProxyPage,
    },
    {
      path: '/models',
      name: 'models',
      component: ModelsPage,
    },
    {
      path: '/logs',
      name: 'logs',
      component: LogsPage,
    },
    {
      path: '/sessions',
      name: 'sessions',
      component: SessionsPage,
    },
    {
      path: '/monitoring',
      name: 'monitoring',
      component: MonitoringPage,
    },
    {
      path: '/contact',
      name: 'contact',
      component: ContactPage,
    },
    {
      path: '/settings',
      name: 'settings',
      component: SettingsPage,
    },
  ],
})
