<script setup lang="ts">
import { computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import { useUiStore } from '../stores/ui'

const emit = defineEmits<{
  copy: [value: string]
  health: []
}>()

const props = defineProps<{
  listenAddress: string
  loading: boolean
}>()

const store = useAppStore()
const ui = useUiStore()
const message = useMessage()
const { t } = useI18n()

const codexBaseURL = computed(() => {
  if (!props.listenAddress) return ''
  return props.listenAddress.replace(/\/+$/, '') + '/v1'
})

async function handleRestoreCodex() {
  try {
    const path = await store.restoreCodexConfigToml()
    message.success(t('settings.toast.restored', { path }))
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  }
}
</script>

<template>
  <div class="guide-card">
    <div class="guide-header">
      <div>
        <h3>{{ t('guide.title') }}</h3>
      </div>
      <n-button
        tertiary
        type="primary"
        :disabled="!codexBaseURL"
        @click="emit('copy', codexBaseURL)"
      >
        {{ t('guide.actions.copyBaseUrl') }}
      </n-button>
    </div>

    <div class="steps">
      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 1</span>
          <span class="step-title">{{ t('guide.step.one.title') }}</span>
        </div>
        <div class="step-body">
          <div class="mono">{{ props.listenAddress || t('guide.step.one.notRunning') }}</div>
          <div class="hint">{{ t('guide.step.one.hint') }}</div>
        </div>
      </div>

      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 2</span>
          <span class="step-title">{{ t('guide.step.two.title') }}</span>
        </div>
        <div class="step-body">
          <div class="actions">
            <n-button secondary @click="ui.showSettings = true">{{ t('guide.actions.openPreferences') }}</n-button>
            <n-button tertiary @click="handleRestoreCodex">{{ t('guide.actions.restoreDefault') }}</n-button>
          </div>
          <div class="kv">
            <span>{{ t('guide.step.two.baseUrl') }}</span>
            <strong class="mono">{{ codexBaseURL || t('guide.step.two.baseUrlAuto') }}</strong>
          </div>
          <div class="kv">
            <span>{{ t('guide.step.two.apiKey') }}</span>
            <strong class="mono">{{ t('guide.step.two.apiKeyNone') }}</strong>
          </div>
          <div class="hint">{{ t('guide.step.two.hint') }}</div>
        </div>
      </div>

      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 3</span>
          <span class="step-title">{{ t('guide.step.three.title') }}</span>
        </div>
        <div class="step-body">
          <div class="actions">
            <n-button secondary :loading="props.loading" @click="emit('health')">{{ t('guide.step.three.healthCheck') }}</n-button>
          </div>
          <div class="hint">{{ t('guide.step.three.hint') }}</div>
          <div class="cmd">
            <div class="cmd-label">{{ t('guide.step.three.quickVerify') }}</div>
            <div class="mono">curl {{ props.listenAddress || 'http://127.0.0.1:11434' }}/health</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.guide-card {
  display: grid;
  gap: 14px;
  padding: 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.guide-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.guide-header h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text);
}

.steps {
  display: grid;
  gap: 10px;
}

.step {
  border-radius: 18px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.82);
  padding: 12px 12px 14px;
  display: grid;
  gap: 10px;
}

.step-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.step-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(22, 119, 255, 0.12);
  color: rgba(22, 119, 255, 0.92);
}

.step-title {
  font-size: 13px;
  font-weight: 600;
  color: rgba(11, 18, 32, 0.92);
}

.step-body {
  display: grid;
  gap: 8px;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  word-break: break-all;
  color: rgba(11, 18, 32, 0.9);
}

.hint {
  font-size: 12px;
  line-height: 1.6;
  color: var(--muted);
}

.kv {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: baseline;
  font-size: 12px;
  color: var(--muted);
}

.kv strong {
  color: rgba(11, 18, 32, 0.9);
  font-weight: 600;
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

.actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

@media (max-width: 920px) {
  .guide-header {
    flex-direction: column;
  }
}
</style>
