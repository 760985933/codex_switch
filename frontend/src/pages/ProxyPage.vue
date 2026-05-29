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
      <n-button type="primary" size="small" :loading="store.isBusy" @click="handleSave">
        {{ t('proxy.save') }}
      </n-button>
    </div>

    <n-card :title="t('proxy.section.network')" size="small">
      <n-form label-placement="top" size="small">
        <div class="form-grid">
          <n-form-item :label="t('proxy.listenHost')">
            <n-input v-model:value="localConfig.listenHost" placeholder="127.0.0.1" size="small" />
          </n-form-item>
          <n-form-item :label="t('proxy.listenPort')">
            <n-input-number v-model:value="localConfig.listenPort" :min="1" :max="65535" size="small" />
          </n-form-item>
        </div>
        <n-alert type="info" :bordered="false" class="proxy-address-hint">
          <template #header>{{ t('proxy.proxyAddress') }}</template>
          <code>{{ proxyAddress }}</code>
        </n-alert>
      </n-form>
    </n-card>

    <n-card :title="t('proxy.section.transport')" size="small">
      <n-form label-placement="top" size="small">
        <div class="form-grid">
          <n-form-item :label="t('proxy.requestTimeout')">
            <n-input-number v-model:value="localConfig.requestTimeoutMs" :min="1000" :step="1000" size="small" />
          </n-form-item>
          <n-form-item :label="t('proxy.maxRetries')">
            <n-input-number v-model:value="localConfig.maxRetries" :min="0" :max="5" size="small" />
          </n-form-item>
        </div>
      </n-form>
    </n-card>

    <n-card :title="t('proxy.section.headers')" size="small">
      <KeyValueEditor
        v-model:model-value="localConfig.headers"
        :title="t('proxy.customHeaders')"
        :description="t('proxy.customHeadersDesc')"
        :key-placeholder="t('proxy.headerKeyPlaceholder')"
        :value-placeholder="t('proxy.headerValuePlaceholder')"
        size="small"
      />
    </n-card>

    <n-card :title="t('proxy.section.behavior')" size="small">
      <n-space vertical size="small">
        <n-form-item :label="t('proxy.autoStart')" label-placement="left">
          <n-switch v-model:value="localConfig.enableAutoStart" />
        </n-form-item>
        <n-form-item :label="t('proxy.minimizeToTray')" label-placement="left">
          <n-switch v-model:value="localConfig.minimizeToTray" />
        </n-form-item>
        <n-form-item :label="t('proxy.compactMode')" label-placement="left">
          <n-switch v-model:value="localConfig.compactMode" />
        </n-form-item>
        <n-form-item :label="t('proxy.logRetentionDays')" label-placement="left">
          <n-input-number v-model:value="localConfig.logRetentionDays" :min="1" :max="30" size="small" />
        </n-form-item>
      </n-space>
    </n-card>
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 12px;
}

.proxy-address-hint {
  margin-top: 8px;
}

.proxy-address-hint code {
  font-size: 13px;
  font-weight: 600;
  user-select: all;
}
</style>
