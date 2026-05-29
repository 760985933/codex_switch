<script setup lang="ts">
import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import KeyValueEditor from '../components/KeyValueEditor.vue'

const store = useAppStore()
const message = useMessage()
const { t } = useI18n()

const localConfig = ref({ ...store.config })

const proxyAddress = computed(() => {
  const host = localConfig.value.listenHost || '127.0.0.1'
  const port = localConfig.value.listenPort || 17419
  return `http://${host}:${port}/v1`
})

async function handleSave() {
  await store.saveConfig(localConfig.value)
  message.success(t('proxy.toast.saved'))
}
</script>

<template>
  <div class="proxy-page">
    <div class="page-header">
      <div>
        <h2>{{ t('proxy.title') }}</h2>
        <p class="page-desc">{{ t('proxy.description') }}</p>
      </div>
      <n-button type="primary" :loading="store.isBusy" @click="handleSave">
        {{ t('proxy.save') }}
      </n-button>
    </div>

    <!-- Network -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('proxy.section.network') }}</span>
      </div>
      <n-form label-placement="top" size="small">
        <div class="form-grid">
          <n-form-item :label="t('proxy.listenHost')">
            <n-input v-model:value="localConfig.listenHost" placeholder="127.0.0.1" />
          </n-form-item>
          <n-form-item :label="t('proxy.listenPort')">
            <n-input-number v-model:value="localConfig.listenPort" :min="1" :max="65535" style="width:100%" />
          </n-form-item>
        </div>
        <div class="address-hint">
          <div class="hint-label">{{ t('proxy.proxyAddress') }}</div>
          <code>{{ proxyAddress }}</code>
        </div>
      </n-form>
    </div>

    <!-- Request Behavior -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('proxy.section.transport') }}</span>
      </div>
      <n-form label-placement="top" size="small">
        <div class="form-grid">
          <n-form-item :label="t('proxy.requestTimeout')">
            <n-input-number v-model:value="localConfig.requestTimeoutMs" :min="1000" :step="1000" style="width:100%" />
          </n-form-item>
          <n-form-item :label="t('proxy.maxRetries')">
            <n-input-number v-model:value="localConfig.maxRetries" :min="0" :max="5" style="width:100%" />
          </n-form-item>
        </div>
      </n-form>
    </div>

    <!-- Custom Headers -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('proxy.section.headers') }}</span>
      </div>
      <KeyValueEditor
        v-model:model-value="localConfig.headers"
        :title="t('proxy.customHeaders')"
        :description="t('proxy.customHeadersDesc')"
        :key-placeholder="t('proxy.headerKeyPlaceholder')"
        :value-placeholder="t('proxy.headerValuePlaceholder')"
      />
    </div>

    <!-- Behavior -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">{{ t('proxy.section.behavior') }}</span>
      </div>
      <div class="behavior-grid">
        <div class="behavior-item">
          <span class="behavior-label">{{ t('proxy.autoStart') }}</span>
          <n-switch v-model:value="localConfig.enableAutoStart" />
        </div>
        <div class="behavior-item">
          <span class="behavior-label">{{ t('proxy.minimizeToTray') }}</span>
          <n-switch v-model:value="localConfig.minimizeToTray" />
        </div>
        <div class="behavior-item">
          <span class="behavior-label">{{ t('proxy.compactMode') }}</span>
          <n-switch v-model:value="localConfig.compactMode" />
        </div>
        <div class="behavior-item">
          <span class="behavior-label">{{ t('proxy.logRetentionDays') }}</span>
          <n-input-number v-model:value="localConfig.logRetentionDays" :min="1" :max="30" style="width:100px" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.proxy-page {
  display: grid;
  gap: 16px;
  max-width: 720px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.page-header h2 {
  margin: 0 0 4px;
  font-size: 18px;
  color: var(--text);
}

.page-desc {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
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

.address-hint code {
  font-size: 13px;
  font-weight: 600;
  user-select: all;
  color: rgba(11, 18, 32, 0.9);
}

.behavior-grid {
  display: grid;
  gap: 4px;
}

.behavior-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}

.behavior-label {
  font-size: 13px;
  color: rgba(11, 18, 32, 0.86);
}
</style>
