<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import QuickGuideCard from '../components/QuickGuideCard.vue'
import CodexLoginActions from '../components/CodexLoginActions.vue'
import { useProxyEvents } from '../composables/useProxyEvents'
import { useAppStore } from '../stores/app'
import { useCodexStore } from '../stores/codex'
import { useUiStore } from '../stores/ui'

const store = useAppStore()
const codexStore = useCodexStore()
const ui = useUiStore()
const message = useMessage()
const { t } = useI18n()
const busy = ref(false)

async function wrapAction<T>(
  task: () => Promise<T>,
  successMessage?: string,
  options?: { timeoutMs?: number; onTimeout?: () => Promise<void> },
) {
  busy.value = true
  try {
    const timeoutMs = options?.timeoutMs ?? 5000
    const timeoutSeconds = Math.max(1, Math.round(timeoutMs / 1000))
    const timeoutError = new Error(t('app.errors.timeoutStopped', { seconds: timeoutSeconds }))
    timeoutError.name = 'TimeoutError'
    const timeoutPromise = new Promise<never>((_, reject) => {
      window.setTimeout(() => reject(timeoutError), timeoutMs)
    })

    const result = await Promise.race([task(), timeoutPromise])
    await store.refreshLogs()
    if (successMessage) {
      message.success(successMessage)
    }
    return result
  } catch (error) {
    if (error instanceof Error && error.name === 'TimeoutError') {
      if (options?.onTimeout) {
        await options.onTimeout()
      }
      await store.refreshLogs()
      message.error(error.message)
      return null as T
    }

    message.error(error instanceof Error ? error.message : String(error))
    throw error
  } finally {
    busy.value = false
  }
}

async function handleStop() {
  await wrapAction(async () => store.stopProxy(), t('overview.toast.proxyStopped'))
}

async function handleHealth() {
  const result = await wrapAction(async () => store.runHealthCheck())
  if (result) {
    message[result.ok ? 'success' : 'warning'](result.ok ? t('overview.health.ok') : t('overview.health.bad'))
  }
}

async function copyText(value: string) {
  await ClipboardSetText(value)
  message.success(t('overview.toast.clipboardCopied'))
}

// ── Codex Desktop tab state ──
const activeTab = ref('dashboard')
const proxyAddress = ref('')
const codexConfigPath = ref('')
const codexConfigContent = ref('')

function updateProxyAddress() {
  const host = store.config.listenHost || '127.0.0.1'
  const port = store.config.listenPort || 17419
  proxyAddress.value = `http://${host}:${port}/v1`
}

async function loadCodexConfig() {
  try {
    codexConfigPath.value = await codexStore.getCodexConfigPath()
    codexConfigContent.value = await codexStore.readCodexConfigToml()
  } catch {
    // file may not exist yet
  }
}

async function handleWriteCodexConfig() {
  try {
    const path = await codexStore.writeCodexConfigTomlRaw(codexConfigContent.value)
    message.success(t('app.toast.codexTomlWritten', { path }))
    await loadCodexConfig()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  }
}

async function handleCopyProxyAddress() {
  await ClipboardSetText(proxyAddress.value)
  message.success(t('overview.toast.clipboardCopied'))
}

async function handleCopyEnvVar() {
  await ClipboardSetText(`export ANTHROPIC_BASE_URL="${proxyAddress.value}"`)
  message.success(t('overview.toast.clipboardCopied'))
}

function handleTabChange(tab: string) {
  updateProxyAddress()
  if (tab === 'codex') loadCodexConfig()
}

useProxyEvents({
  onStatus(payload) {
    store.applyStatus(payload)
  },
  onLog(entry) {
    store.pushLog(entry)
  },
})

onMounted(async () => {
  updateProxyAddress()
  if (!store.lastLoadedAt) {
    await wrapAction(async () => store.initialize())
  }
})
</script>

<template>
  <div class="proxy-page">
    <n-tabs
      v-model:value="activeTab"
      type="line"
      animated
      @update:value="handleTabChange"
    >
      <!-- Tab 1: Dashboard -->
      <n-tab-pane name="dashboard" :tab="t('overview.tab.dashboard')">
        <QuickGuideCard
          :listen-address="store.status.listenAddress"
          :status="store.status"
          :health="store.healthCheck"
          :loading="busy"
          @copy="copyText"
          @health="handleHealth"
          @stop="handleStop"
          @refresh="wrapAction(async () => { await store.refreshStatus(); await store.refreshLogs() })"
        />
      </n-tab-pane>

      <!-- Tab 2: Codex Desktop -->
      <n-tab-pane name="codex" :tab="t('overview.tab.codexDesktop')">
        <div class="tab-content">
          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.codex.overview') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.codex.overviewDesc') }}</p>
            <div class="address-hint">
              <div class="hint-label">{{ t('proxy.proxyAddress') }}</div>
              <div class="hint-row">
                <code>{{ proxyAddress }}</code>
                <n-button text size="small" type="primary" @click="handleCopyProxyAddress">
                  {{ t('guide.actions.copyBaseUrl') }}
                </n-button>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.codex.quickLogin') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.codex.quickLoginDesc') }}</p>
            <div class="login-actions">
              <CodexLoginActions
                v-for="p in store.profileList"
                :key="p.id"
                :profile-id="p.id"
              />
            </div>
            <div v-if="store.profileList.length === 0" class="empty-hint">
              {{ t('dashboard.noProfile') }}
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.codex.configToml') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.codex.configTomlDesc') }}</p>
            <n-form label-placement="top" size="small">
              <n-form-item :label="t('settings.codex.filePath')">
                <n-input :value="codexConfigPath" readonly />
              </n-form-item>
              <n-form-item :label="t('settings.codex.content')">
                <n-input
                  v-model:value="codexConfigContent"
                  type="textarea"
                  :autosize="{ minRows: 8, maxRows: 20 }"
                />
              </n-form-item>
            </n-form>
            <div class="action-row">
              <n-button type="primary" size="small" @click="handleWriteCodexConfig">
                {{ t('settings.codexActions.mergeWrite') }}
              </n-button>
              <n-button size="small" @click="loadCodexConfig">
                {{ t('settings.codexActions.readFile') }}
              </n-button>
            </div>
          </div>
        </div>
      </n-tab-pane>

      <!-- Tab 3: Claude Code -->
      <n-tab-pane name="claude" :tab="t('overview.tab.claudeCode')">
        <div class="tab-content">
          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.claude.overview') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.claude.overviewDesc') }}</p>
            <div class="address-hint">
              <div class="hint-label">{{ t('proxy.proxyAddress') }}</div>
              <div class="hint-row">
                <code>{{ proxyAddress }}</code>
                <n-button text size="small" type="primary" @click="handleCopyProxyAddress">
                  {{ t('guide.actions.copyBaseUrl') }}
                </n-button>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.claude.envConfig') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.claude.envConfigDesc') }}</p>
            <div class="code-block">
              <code>export ANTHROPIC_BASE_URL="{{ proxyAddress }}"</code>
              <n-button text size="small" type="primary" @click="handleCopyEnvVar">
                {{ t('logs.actions.copy') }}
              </n-button>
            </div>
            <p class="card-desc" style="margin-top: 16px">{{ t('overview.claude.apiKeyNote') }}</p>
            <div class="code-block">
              <code>export ANTHROPIC_API_KEY="your-api-key"</code>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <span class="card-title">{{ t('overview.claude.verify') }}</span>
            </div>
            <p class="card-desc">{{ t('overview.claude.verifyDesc') }}</p>
            <div class="code-block">
              <code>claude --version</code>
            </div>
            <p class="card-desc" style="margin-top: 12px">{{ t('overview.claude.runHint') }}</p>
            <div class="code-block">
              <code>claude</code>
            </div>
          </div>
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<style scoped>
.proxy-page {
  display: grid;
  gap: 16px;
  max-width: 780px;
}

.tab-content {
  display: grid;
  gap: 16px;
  padding-top: 8px;
}

.card {
  padding: 16px 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.card-header {
  margin-bottom: 12px;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

.card-desc {
  margin: 0 0 14px;
  font-size: 12px;
  color: var(--muted);
  line-height: 1.6;
}

.address-hint {
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px dashed rgba(22, 119, 255, 0.28);
  background: rgba(22, 119, 255, 0.06);
  display: grid;
  gap: 6px;
}

.hint-label {
  font-size: 12px;
  color: rgba(11, 18, 32, 0.72);
  font-weight: 600;
}

.hint-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.address-hint code {
  font-size: 13px;
  font-weight: 600;
  user-select: all;
  color: rgba(11, 18, 32, 0.9);
}

.login-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.empty-hint {
  font-size: 12px;
  color: var(--muted);
}

.action-row {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.code-block {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.04);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  word-break: break-all;
}

.code-block code {
  flex: 1;
  color: rgba(11, 18, 32, 0.9);
}
</style>
