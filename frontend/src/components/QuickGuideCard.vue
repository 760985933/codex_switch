<script setup lang="ts">
import { computed } from 'vue'
import { useMessage } from 'naive-ui'
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

const codexBaseURL = computed(() => {
  if (!props.listenAddress) return ''
  return props.listenAddress.replace(/\/+$/, '') + '/v1'
})

async function handleRestoreCodex() {
  try {
    const path = await store.restoreCodexConfigToml()
    message.success(`已恢复 Codex 配置: ${path}`)
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  }
}
</script>

<template>
  <div class="guide-card">
    <div class="guide-header">
      <div>
        <h3>接入指引</h3>
      </div>
      <n-button
        tertiary
        type="primary"
        :disabled="!codexBaseURL"
        @click="emit('copy', codexBaseURL)"
      >
        复制 Base URL
      </n-button>
    </div>

    <div class="steps">
      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 1</span>
          <span class="step-title">启动桥接服务</span>
        </div>
        <div class="step-body">
          <div class="mono">{{ props.listenAddress || '未启动（先在左侧点击“启动桥接”）' }}</div>
          <div class="hint">启动成功后会在本机监听一个地址，Codex 会通过它访问 /v1 接口。</div>
        </div>
      </div>

      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 2</span>
          <span class="step-title">配置 Codex 指向本地</span>
        </div>
        <div class="step-body">
          <div class="actions">
            <n-button secondary @click="ui.showSettings = true">打开偏好设置</n-button>
            <n-button tertiary @click="handleRestoreCodex">恢复默认</n-button>
          </div>
          <div class="kv">
            <span>Base URL</span>
            <strong class="mono">{{ codexBaseURL || '启动后自动生成' }}</strong>
          </div>
          <div class="kv">
            <span>API Key</span>
            <strong class="mono">无需配置</strong>
          </div>
          <div class="hint">推荐用“偏好设置 → Codex config.toml → 写入文件”，自动合并并保留原有 MCP/approvals 配置。</div>
        </div>
      </div>

      <div class="step">
        <div class="step-head">
          <span class="step-badge">Step 3</span>
          <span class="step-title">验证与排障</span>
        </div>
        <div class="step-body">
          <div class="actions">
            <n-button secondary :loading="props.loading" @click="emit('health')">健康检查</n-button>
          </div>
          <div class="hint">点击右侧“健康检查”，并观察控制台最近日志；异常时优先确认端口占用、Key、上游可达性。</div>
          <div class="cmd">
            <div class="cmd-label">快速验证</div>
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
