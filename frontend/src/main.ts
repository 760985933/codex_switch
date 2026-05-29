import { createApp } from 'vue'
import { createPinia } from 'pinia'
import {
  NButton,
  NCard,
  NCollapseTransition,
  NConfigProvider,
  NDataTable,
  NDialogProvider,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NMessageProvider,
  NScrollbar,
  NSpace,
  NSpin,
  NText,
  NSelect,
  NSwitch,
  NTabPane,
  NTabs,
  NTag,
  create,
} from 'naive-ui'
import App from './App.vue'
import { router } from './router'
import { i18n } from './i18n'
import './style.css'

const app = createApp(App)
const naive = create({
  components: [
    NButton,
    NCard,
    NCollapseTransition,
    NConfigProvider,
    NDataTable,
    NDialogProvider,
    NDrawer,
    NDrawerContent,
    NEmpty,
    NForm,
    NFormItem,
    NInput,
    NInputNumber,
    NModal,
    NMessageProvider,
    NScrollbar,
    NSpace,
    NSpin,
    NText,
    NSelect,
    NSwitch,
    NTabPane,
    NTabs,
    NTag,
  ],
})

app.use(createPinia())
app.use(router)
app.use(naive)
app.use(i18n)
app.mount('#app')
