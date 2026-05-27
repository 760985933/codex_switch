<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import type { Profile } from '../types'
import KeyValueEditor from './KeyValueEditor.vue'
import { maskSecret } from '../utils/format'

const store = useAppStore()
const emit = defineEmits<{
  save: []
  copy: [value: string]
}>()

const formProfile = ref<Profile>({ ...store.currentProfile ?? {} as Profile })
const showAdvanced = ref(false)
const { t } = useI18n()
let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
const autoSaveScheduled = ref(false)

// Sync form when switching profiles
watch(
  () => store.config.currentProfileId,
  () => syncForm(),
)
watch(
  () => store.config.profiles,
  () => syncForm(),
  { deep: true },
)

function syncForm() {
  const p = store.currentProfile
  if (p) {
    formProfile.value = {
      ...p,
      mappings: { ...p.mappings },
      headers: { ...p.headers },
    }
  }
}
syncForm()

// Auto-save on form changes (debounced)
watch(
  formProfile,
  () => {
    if (autoSaveTimer) clearTimeout(autoSaveTimer)
    autoSaveScheduled.value = true
    autoSaveTimer = setTimeout(() => {
      submitSave()
    }, 600)
  },
  { deep: true },
)

const isRunning = computed(() => store.isRunning)
const maskedApiKey = computed(() => maskSecret(formProfile.value.apiKey))
const apiKeyHint = computed(() =>
  t('config.fields.apiKeyHint', {
    masked: maskedApiKey.value || t('config.fields.apiKeyMissing'),
  }),
)

async function submitSave() {
  if (autoSaveTimer) {
    clearTimeout(autoSaveTimer)
    autoSaveTimer = null
  }
  autoSaveScheduled.value = false
  const profiles = { ...store.config.profiles }
  profiles[store.config.currentProfileId] = { ...formProfile.value }
  const updated = {
    ...store.config,
    profiles,
  }
  await store.saveConfig(updated)
  emit('save')
}
</script>

<template>
  <div class="config-panel">
    <div class="panel-head">
      <div>
        <h3>{{ t('config.title') }}</h3>
        <p v-if="store.currentProfile">
          {{ t('config.editing') }}: <strong>{{ store.currentProfile.name }}</strong>
        </p>
        <p v-else>{{ t('config.noProfile') }}</p>
      </div>
    </div>

    <n-form label-placement="top" :model="formProfile" size="small">
      <div class="form-grid">
        <n-form-item :label="t('config.fields.profileName')">
          <n-input v-model:value="formProfile.name" size="small" />
        </n-form-item>
        <n-form-item :label="t('config.fields.defaultModel')">
          <n-input v-model:value="formProfile.defaultModel" placeholder="deepseek-v4-flash" size="small" />
        </n-form-item>
        <n-form-item label="API Base URL" class="span-2">
          <n-input v-model:value="formProfile.baseURL" placeholder="https://api.deepseek.com/v1" size="small" />
        </n-form-item>
        <n-form-item label="API Key" class="span-2">
          <n-input
            v-model:value="formProfile.apiKey"
            type="password"
            show-password-on="click"
            placeholder="sk-..."
            size="small"
          />
          <div class="field-hint">{{ apiKeyHint }}</div>
        </n-form-item>
        <n-form-item :label="t('config.fields.listenHost')">
          <n-input v-model:value="store.config.listenHost" placeholder="127.0.0.1" size="small" />
        </n-form-item>
        <n-form-item :label="t('config.fields.listenPort')">
          <n-input-number v-model:value="store.config.listenPort" :min="1" :max="65535" size="small" />
        </n-form-item>
        <n-form-item :label="t('config.fields.requestTimeout')">
          <n-input-number v-model:value="formProfile.requestTimeoutMs" :min="1000" :step="1000" size="small" />
        </n-form-item>
        <n-form-item :label="t('config.fields.maxRetries')">
          <n-input-number v-model:value="formProfile.maxRetries" :min="0" :max="5" size="small" />
        </n-form-item>
      </div>
    </n-form>

    <div class="action-bar">
      <n-button type="primary" :loading="store.isBusy" @click="submitSave">
        {{ autoSaveScheduled ? t('config.actions.saving') : t('config.actions.save') }}
      </n-button>
      <n-button
        tertiary
        size="tiny"
        :disabled="!store.status.listenAddress"
        @click="emit('copy', store.status.listenAddress)"
      >
        {{ t('config.actions.copyLocal') }}
      </n-button>
    </div>

    <n-collapse-transition :show="showAdvanced">
      <div class="advanced-panel">
        <KeyValueEditor
          v-model:model-value="formProfile.mappings"
          :title="t('config.advanced.modelMapping.title')"
          :description="t('config.advanced.modelMapping.desc')"
          :key-placeholder="t('config.advanced.modelMapping.keyPlaceholder')"
          :value-placeholder="t('config.advanced.modelMapping.valuePlaceholder')"
          size="small"
        />
        <KeyValueEditor
          v-model:model-value="formProfile.headers"
          :title="t('config.advanced.headers.title')"
          :description="t('config.advanced.headers.desc')"
          :key-placeholder="t('config.advanced.headers.keyPlaceholder')"
          :value-placeholder="t('config.advanced.headers.valuePlaceholder')"
          size="small"
        />
      </div>
    </n-collapse-transition>

    <n-button text type="primary" @click="showAdvanced = !showAdvanced">
      {{ showAdvanced ? t('config.actions.collapseAdvanced') : t('config.actions.expandAdvanced') }}
    </n-button>
  </div>
</template>

<style scoped>
.config-panel {
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.panel-head h3 {
  margin: 0 0 2px;
  font-size: 15px;
  color: var(--text);
}

.panel-head p {
  margin: 0;
  font-size: 11px;
  line-height: 1.5;
  color: var(--muted);
}

.panel-head strong {
  color: var(--text);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 12px;
}

.span-2 {
  grid-column: span 2;
}

.field-hint {
  margin-top: 4px;
  font-size: 11px;
  color: var(--muted);
}

.action-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.advanced-panel {
  display: grid;
  gap: 14px;
  padding-top: 4px;
}

@media (max-width: 920px) {
  .panel-head,
  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .span-2 {
    grid-column: span 1;
  }
}
</style>
