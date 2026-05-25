<script setup lang="ts">
import { computed } from 'vue'
import { createDiscreteApi, lightTheme } from 'naive-ui'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { ClipboardSetText } from '../wailsjs/runtime/runtime'
import SettingsDrawer from './components/SettingsDrawer.vue'
import { useAppStore } from './stores/app'
import { useUiStore } from './stores/ui'

const route = useRoute()
const store = useAppStore()
const ui = useUiStore()
const { message, dialog } = createDiscreteApi(['message', 'dialog'], {
  configProviderProps: {
    theme: lightTheme,
  },
})

const navItems = [
  { label: '主工作台', to: '/overview' },
  { label: '最近日志', to: '/logs' },
]

const statusLabel = computed(() => {
  switch (store.status.status) {
    case 'running':
      return '运行中'
    case 'starting':
      return '启动中'
    case 'error':
      return '异常'
    default:
      return '未启动'
  }
})

async function handleSaveSettings(config: typeof store.config) {
  try {
    await store.saveConfig(config)
    ui.showSettings = false
    message.success('设置已保存')
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  }
}

async function handleExport() {
  try {
    const content = await store.exportConfig()
    await ClipboardSetText(content)
    message.success('配置 JSON 已复制到剪贴板')
  } catch (error) {
    dialog.warning({
      title: '导出配置',
      content: error instanceof Error ? error.message : String(error),
      positiveText: '知道了',
    })
  }
}

async function handleCodexCopy() {
  try {
    const content = await store.generateCodexConfigToml()
    await ClipboardSetText(content)
    message.success('Codex config.toml 已复制到剪贴板')
  } catch (error) {
    dialog.warning({
      title: '生成 Codex config.toml',
      content: error instanceof Error ? error.message : String(error),
      positiveText: '知道了',
    })
  }
}

async function handleCodexWrite() {
  try {
    const path = await store.writeCodexConfigToml()
    const hintPath = await store.getCodexConfigPath()
    message.success(`已写入 Codex config.toml: ${path || hintPath}`)
  } catch (error) {
    dialog.warning({
      title: '写入 Codex config.toml',
      content: error instanceof Error ? error.message : String(error),
      positiveText: '知道了',
    })
  }
}
</script>

<template>
  <n-config-provider :theme="lightTheme">
    <n-dialog-provider>
      <n-message-provider placement="bottom-right">
        <div class="shell">
          <header class="topbar">
            <div class="brand">
              <div class="brand-mark">NT</div>
              <div>
                <p>nettopo.com</p>
                <strong>Nettopo switch</strong>
              </div>
            </div>

            <nav class="nav">
              <RouterLink
                v-for="item in navItems"
                :key="item.to"
                :to="item.to"
                class="nav-link"
                :class="{ active: route.path === item.to }"
              >
                {{ item.label }}
              </RouterLink>
            </nav>

            <div class="topbar-actions">
              <div class="status-chip" :data-status="store.status.status">
                <span class="status-dot" />
                {{ statusLabel }}
              </div>
              <n-button secondary @click="ui.showSettings = true">偏好设置</n-button>
            </div>
          </header>

          <main class="content">
            <RouterView />
          </main>

          <SettingsDrawer
            v-model:model-value="ui.showSettings"
            :config="store.config"
            @save="handleSaveSettings"
            @export="handleExport"
            @codex-copy="handleCodexCopy"
            @codex-write="handleCodexWrite"
          />
        </div>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<style scoped>
.shell {
  display: grid;
  min-height: 100vh;
  padding: 20px;
  gap: 18px;
}

.topbar {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 18px;
  padding: 16px 18px;
  border-radius: 24px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  backdrop-filter: blur(18px);
  box-shadow: var(--shadow);
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(22, 119, 255, 0.16), rgba(19, 194, 194, 0.14));
  color: var(--accent);
  font-weight: 700;
}

.brand p {
  margin: 0 0 4px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--muted);
}

.brand strong {
  font-size: 16px;
  color: var(--text);
}

.nav {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 14px;
  border-radius: 999px;
  color: var(--muted);
  text-decoration: none;
  transition: all 160ms ease;
}

.nav-link:hover,
.nav-link.active {
  background: rgba(22, 119, 255, 0.08);
  color: var(--text);
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.64);
  border: 1px solid var(--border);
  color: var(--text);
  font-size: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: rgba(11, 18, 32, 0.34);
}

.status-chip[data-status='running'] .status-dot {
  background: var(--accent-2);
}

.status-chip[data-status='starting'] .status-dot {
  background: var(--warning);
}

.status-chip[data-status='error'] .status-dot {
  background: var(--danger);
}

.content {
  min-height: 0;
}

@media (max-width: 1024px) {
  .topbar {
    grid-template-columns: 1fr;
  }

  .nav {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .topbar-actions {
    justify-content: space-between;
  }
}
</style>
