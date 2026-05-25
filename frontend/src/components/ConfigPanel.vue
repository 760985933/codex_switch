<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { AppConfig, BridgeStatusPayload } from '../types'
import KeyValueEditor from './KeyValueEditor.vue'
import { maskSecret } from '../utils/format'

const props = defineProps<{
  config: AppConfig
  status: BridgeStatusPayload
  loading: boolean
}>()

const emit = defineEmits<{
  save: [value: AppConfig]
  start: []
  stop: []
  restart: []
  copy: [value: string]
}>()

const formValue = ref<AppConfig>({ ...props.config })
const showAdvanced = ref(false)

watch(
  () => props.config,
  (value) => {
    formValue.value = {
      ...value,
      mappings: { ...value.mappings },
      headers: { ...value.headers },
    }
  },
  { immediate: true, deep: true },
)

const isRunning = computed(() => props.status.status === 'running')
const maskedApiKey = computed(() => maskSecret(formValue.value.apiKey))

function submitSave() {
  emit('save', formValue.value)
}
</script>

<template>
  <div class="config-panel">
    <div class="panel-head">
      <div>
        <h3>连接配置</h3>
        <p>主界面只保留高频字段，高级映射通过折叠区收纳。</p>
      </div>
      <n-space>
        <n-button type="primary" :loading="loading" @click="submitSave">保存配置</n-button>
      </n-space>
    </div>

    <n-form label-placement="top" :model="formValue">
      <div class="form-grid">
        <n-form-item label="DeepSeek Base URL">
          <n-input v-model:value="formValue.deepseekBaseURL" placeholder="https://api.deepseek.com/v1" />
        </n-form-item>
        <n-form-item label="默认模型">
          <n-input v-model:value="formValue.defaultModel" placeholder="deepseek-chat" />
        </n-form-item>
        <n-form-item label="API Key" class="span-2">
          <n-input
            v-model:value="formValue.apiKey"
            type="password"
            show-password-on="click"
            placeholder="sk-..."
          />
          <div class="field-hint">当前展示: {{ maskedApiKey || '未配置' }}</div>
        </n-form-item>
        <n-form-item label="监听地址">
          <n-input v-model:value="formValue.listenHost" placeholder="127.0.0.1" />
        </n-form-item>
        <n-form-item label="监听端口">
          <n-input-number v-model:value="formValue.listenPort" :min="1" :max="65535" />
        </n-form-item>
        <n-form-item label="请求超时 (ms)">
          <n-input-number v-model:value="formValue.requestTimeoutMs" :min="1000" :step="1000" />
        </n-form-item>
        <n-form-item label="最大重试次数">
          <n-input-number v-model:value="formValue.maxRetries" :min="0" :max="5" />
        </n-form-item>
      </div>
    </n-form>

    <div class="action-bar">
      <n-space>
        <n-button type="primary" :disabled="isRunning" :loading="loading" @click="emit('start')">
          启动桥接
        </n-button>
        <n-button secondary :disabled="!isRunning" :loading="loading" @click="emit('restart')">
          重启
        </n-button>
        <n-button tertiary type="error" :disabled="!isRunning" :loading="loading" @click="emit('stop')">
          停止
        </n-button>
      </n-space>
      <n-button
        tertiary
        :disabled="!status.listenAddress"
        @click="emit('copy', status.listenAddress)"
      >
        复制本地地址
      </n-button>
    </div>

    <n-collapse-transition :show="showAdvanced">
      <div class="advanced-panel">
        <KeyValueEditor
          v-model:model-value="formValue.mappings"
          title="模型映射"
          description="当 Codex 请求中的模型名与 DeepSeek 模型不一致时，在此做静态映射。"
          key-placeholder="如 gpt-4.1"
          value-placeholder="如 deepseek-chat"
        />
        <KeyValueEditor
          v-model:model-value="formValue.headers"
          title="附加请求头"
          description="只有在接入网关或代理平台时再填写，未设置时保持为空。"
          key-placeholder="如 X-Source"
          value-placeholder="如 codex-desktop"
        />
      </div>
    </n-collapse-transition>

    <n-button text type="primary" @click="showAdvanced = !showAdvanced">
      {{ showAdvanced ? '收起高级配置' : '展开高级配置' }}
    </n-button>
  </div>
</template>

<style scoped>
.config-panel {
  display: grid;
  gap: 18px;
  padding: 20px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 12px 34px rgba(14, 30, 68, 0.08);
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-head h3 {
  margin: 0 0 6px;
  font-size: 17px;
  color: var(--text);
}

.panel-head p {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--muted);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 16px;
}

.span-2 {
  grid-column: span 2;
}

.field-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--muted);
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.advanced-panel {
  display: grid;
  gap: 20px;
  padding-top: 8px;
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
