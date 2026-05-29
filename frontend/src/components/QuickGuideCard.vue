<script setup lang="ts">
import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { useUiStore } from '../stores/ui'
import type { ProxyStatusPayload, HealthCheckResult } from '../types'

const emit = defineEmits<{
  copy: [value: string]
  health: []
  stop: []
  refresh: []
}>()

const props = defineProps<{
  listenAddress: string
  loading: boolean
  status: ProxyStatusPayload
  health: HealthCheckResult | null
}>()

const store = useAppStore()
const ui = useUiStore()
const router = useRouter()
const message = useMessage()
const { t } = useI18n()

const statusLabel = computed(() => {
  switch (props.status.status) {
    case 'running': return t('app.status.running')
    case 'starting': return t('app.status.starting')
    case 'error': return t('app.status.error')
    default: return t('app.status.stopped')
  }
})

const healthSummary = computed(() => {
  if (!props.health) return null
  const failed = props.health.checks.filter((item) => !item.ok)
  if (props.health.ok) return { tone: 'success' as const, text: t('console.health.ok') }
  return { tone: 'warning' as const, text: t('console.health.failed', { count: failed.length }) }
})

const failedChecks = computed(() => {
  if (!props.health) return []
  return props.health.checks.filter((item) => !item.ok)
})

const codexBaseURL = computed(() => {
  if (!props.listenAddress) return ''
  return props.listenAddress.replace(/\/+$/, '') + '/v1'
})

const currentProfile = computed(() => store.currentProfile)

// ── Login actions ──
const activeLoginAction = ref<'plugin' | 'noaccount' | null>(null)
const loginProfileId = ref<string | null>(null)

async function handlePluginLogin() {
  const id = store.config.currentProfileId
  const profile = store.config.profiles[id]
  if (!profile?.apiKey) {
    message.warning(t('guide.monitor.noKey'))
    return
  }
  loginProfileId.value = id
  activeLoginAction.value = 'plugin'
  try {
    if (!store.isRunning) await store.startProxy()
    const path = await store.pluginUnlockLogin()
    const hintPath = await store.getCodexConfigPath()
    message.success(t('app.toast.codexTomlWritten', { path: path || hintPath }))
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    loginProfileId.value = null
    activeLoginAction.value = null
  }
}

async function handleNoAccountLogin() {
  const id = store.config.currentProfileId
  const profile = store.config.profiles[id]
  if (!profile?.apiKey) {
    message.warning(t('guide.monitor.noKey'))
    return
  }
  loginProfileId.value = id
  activeLoginAction.value = 'noaccount'
  try {
    if (!store.isRunning) await store.startProxy()
    const path = await store.writeCodexConfigTomlProfiles()
    const hintPath = await store.getCodexConfigPath()
    message.success(t('app.toast.codexTomlWritten', { path: path || hintPath }))
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    loginProfileId.value = null
    activeLoginAction.value = null
  }
}
</script>

<template>
  <div class="dashboard">
    <!-- Active Profile Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('dashboard.currentProfile') }}</span>
        <n-button text size="small" type="primary" @click="router.push('/models')">
          {{ t('dashboard.manageModels') }}
        </n-button>
      </div>
      <div v-if="currentProfile" class="profile-summary">
        <div class="profile-summary-main">
          <span class="profile-name">{{ currentProfile.name }}</span>
          <span class="profile-badge">{{ currentProfile.provider }}</span>
        </div>
        <div class="profile-summary-meta">
          <span class="meta-item"><span class="meta-label">Model:</span> {{ currentProfile.defaultModel }}</span>
          <span class="meta-item"><span class="meta-label">API:</span> {{ currentProfile.baseURL }}</span>
        </div>
        <div class="profile-actions" v-if="currentProfile.apiKey">
          <n-button
            size="small"
            type="primary"
            :title="t('guide.actions.pluginUnlockLoginTooltip')"
            :loading="loginProfileId === store.config.currentProfileId && activeLoginAction === 'plugin'"
            @click="handlePluginLogin"
          >
            {{ t('guide.actions.pluginUnlockLogin') }}
          </n-button>
          <n-button
            size="small"
            secondary
            type="primary"
            :loading="loginProfileId === store.config.currentProfileId && activeLoginAction === 'noaccount'"
            @click="handleNoAccountLogin"
          >
            {{ t('guide.actions.noAccountLogin') }}
          </n-button>
        </div>
      </div>
      <div v-else class="no-profile">
        <p>{{ t('dashboard.noProfile') }}</p>
        <n-button size="small" type="primary" @click="router.push('/models')">
          {{ t('dashboard.goToModels') }}
        </n-button>
      </div>
    </div>

    <!-- Proxy Status Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('dashboard.proxyStatus') }}</span>
        <n-button text size="small" type="primary" @click="router.push('/proxy')">
          {{ t('dashboard.proxySettings') }}
        </n-button>
      </div>
      <div class="status-section">
        <div class="s-status">
          <span class="s-dot" :data-status="status.status" />
          <span>{{ statusLabel }}</span>
        </div>

        <div class="s-meta">
          <span class="s-meta-item">
            <span class="s-meta-label">{{ t('console.meta.listenAddress') }}:</span>
            <strong>{{ status.listenAddress || t('console.meta.notRunning') }}</strong>
          </span>
          <span class="s-meta-item">
            <span class="s-meta-label">{{ t('console.meta.requestCount') }}:</span>
            <strong>{{ status.requestCount }}</strong>
          </span>
          <span v-if="status.lastError" class="s-meta-item" data-tone="error">
            <span class="s-meta-label">{{ t('console.meta.lastError') }}:</span>
            <strong>{{ status.lastError }}</strong>
          </span>
        </div>

        <div v-if="healthSummary" class="s-health" :data-tone="healthSummary.tone">
          <span class="h-dot" />
          <span>{{ healthSummary.text }}</span>
        </div>
        <div v-if="failedChecks.length" class="s-fails">
          <div v-for="item in failedChecks" :key="item.name" class="s-fail">
            <strong>{{ item.name }}</strong>
            <p>{{ item.message }}</p>
          </div>
        </div>

        <div class="actions">
          <n-button type="primary" :loading="loading" @click="emit('health')">{{ t('guide.step.three.healthCheck') }}</n-button>
          <n-button secondary :loading="loading" @click="emit('refresh')">{{ t('console.actions.refresh') }}</n-button>
        </div>

        <div class="cmd">
          <div class="cmd-label">{{ t('guide.step.three.quickVerify') }}</div>
          <div class="mono">浏览器访问 {{ props.listenAddress || 'http://127.0.0.1:11434' }}/health</div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('dashboard.quickActions') }}</span>
      </div>
      <div class="actions">
        <n-button tertiary @click="ui.showSettings = true">{{ t('guide.actions.preferences') }}</n-button>
        <n-button
          tertiary
          type="primary"
          :disabled="!codexBaseURL"
          @click="emit('copy', codexBaseURL)"
        >
          {{ t('guide.actions.copyBaseUrl') }}
        </n-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  display: grid;
  gap: 16px;
  max-width: 720px;
}

.card {
  padding: 16px 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

/* Profile Summary */
.profile-summary {
  display: grid;
  gap: 8px;
}

.profile-summary-main {
  display: flex;
  align-items: center;
  gap: 8px;
}

.profile-name {
  font-size: 16px;
  font-weight: 700;
  color: rgba(11, 18, 32, 0.92);
}

.profile-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  background: rgba(22, 119, 255, 0.12);
  color: rgba(22, 119, 255, 0.92);
}

.profile-summary-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 16px;
  font-size: 12px;
}

.meta-item {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: rgba(11, 18, 32, 0.72);
}

.meta-label {
  font-family: inherit;
  color: var(--muted);
}

.profile-actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.no-profile {
  display: grid;
  gap: 8px;
}

.no-profile p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
}

/* Status */
.s-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: rgba(11, 18, 32, 0.86);
}

.s-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: rgba(11, 18, 32, 0.26);
  box-shadow: 0 0 0 4px rgba(11, 18, 32, 0.06);
  flex-shrink: 0;
}
.s-dot[data-status='running'] {
  background: var(--accent-2);
  box-shadow: 0 0 0 4px rgba(19, 194, 194, 0.16);
}
.s-dot[data-status='starting'] {
  background: var(--warning);
  box-shadow: 0 0 0 4px rgba(216, 150, 20, 0.16);
}
.s-dot[data-status='error'] {
  background: var(--danger);
  box-shadow: 0 0 0 4px rgba(212, 56, 13, 0.16);
}

.s-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  font-size: 12px;
}
.s-meta-item {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
}
.s-meta-label {
  color: var(--muted);
  font-size: 11px;
}
.s-meta-item strong {
  font-weight: 600;
  color: rgba(11, 18, 32, 0.9);
  word-break: break-all;
}
.s-meta-item[data-tone='error'] strong {
  color: rgba(212, 56, 13, 0.92);
}

.s-health {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.82);
  font-size: 13px;
  color: rgba(11, 18, 32, 0.86);
}
.h-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--muted);
  flex-shrink: 0;
}
.s-health[data-tone='success'] .h-dot { background: var(--accent-2); }
.s-health[data-tone='warning'] .h-dot { background: var(--warning); }

.s-fails {
  display: grid;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px solid rgba(216, 150, 20, 0.22);
  background: rgba(255, 255, 255, 0.82);
}
.s-fail {
  display: grid;
  gap: 4px;
}
.s-fail strong {
  font-size: 12px;
  color: rgba(11, 18, 32, 0.9);
}
.s-fail p {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: rgba(11, 18, 32, 0.72);
  word-break: break-word;
}

.actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.cmd {
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px dashed rgba(22, 119, 255, 0.28);
  background: rgba(22, 119, 255, 0.06);
  display: grid;
  gap: 6px;
}

.cmd-label {
  font-size: 12px;
  color: rgba(11, 18, 32, 0.72);
  font-weight: 600;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  word-break: break-all;
  color: rgba(11, 18, 32, 0.9);
}
</style>
