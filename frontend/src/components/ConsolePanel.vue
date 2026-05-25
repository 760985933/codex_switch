<script setup lang="ts">
import { computed } from 'vue'
import type { BridgeStatusPayload, HealthCheckResult } from '../types'

const props = defineProps<{
  status: BridgeStatusPayload
  health: HealthCheckResult | null
  loading: boolean
}>()

const emit = defineEmits<{
  health: []
  refresh: []
}>()

const statusLabel = computed(() => {
  switch (props.status.status) {
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

const healthSummary = computed(() => {
  if (!props.health) return null
  const failed = props.health.checks.filter((item) => !item.ok)
  if (props.health.ok) return { tone: 'success', text: '健康检查通过' as const }
  return { tone: 'warning', text: `健康检查失败：${failed.length} 项异常` as const }
})
</script>

<template>
  <div class="console-panel">
    <div class="panel-head">
      <div class="title">
        <span class="status-icon" :data-status="status.status" />
        <div>
          <h3>控制台</h3>
          <p>{{ statusLabel }}</p>
        </div>
      </div>

      <n-space>
        <n-button secondary :loading="loading" @click="emit('refresh')">刷新</n-button>
        <n-button type="primary" :loading="loading" @click="emit('health')">健康检查</n-button>
      </n-space>
    </div>

    <div class="meta-grid">
      <div class="meta-item">
        <span>监听地址</span>
        <strong>{{ status.listenAddress || '未启动' }}</strong>
      </div>
      <div class="meta-item">
        <span>请求次数</span>
        <strong>{{ status.requestCount }}</strong>
      </div>
      <div v-if="status.lastError" class="meta-item span-2" data-tone="error">
        <span>最近错误</span>
        <strong>{{ status.lastError }}</strong>
      </div>
    </div>

    <div v-if="healthSummary" class="health-row" :data-tone="healthSummary.tone">
      <span class="health-dot" />
      <span class="health-text">{{ healthSummary.text }}</span>
    </div>
  </div>
</template>

<style scoped>
.console-panel {
  display: grid;
  gap: 14px;
  padding: 18px;
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

.title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.title h3 {
  margin: 0 0 4px;
  font-size: 16px;
  color: var(--text);
}

.title p {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.status-icon {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: rgba(11, 18, 32, 0.26);
  box-shadow: 0 0 0 4px rgba(11, 18, 32, 0.06);
}

.status-icon[data-status='running'] {
  background: var(--accent-2);
  box-shadow: 0 0 0 4px rgba(19, 194, 194, 0.16);
}

.status-icon[data-status='starting'] {
  background: var(--warning);
  box-shadow: 0 0 0 4px rgba(216, 150, 20, 0.16);
}

.status-icon[data-status='error'] {
  background: var(--danger);
  box-shadow: 0 0 0 4px rgba(212, 56, 13, 0.16);
}

.health-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.82);
}

.health-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--muted);
}

.health-row[data-tone='success'] .health-dot {
  background: var(--accent-2);
}

.health-row[data-tone='warning'] .health-dot {
  background: var(--warning);
}

.health-text {
  font-size: 13px;
  color: rgba(11, 18, 32, 0.86);
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 12px;
}

.span-2 {
  grid-column: span 2;
}

.meta-item {
  display: grid;
  gap: 6px;
  padding: 10px 12px;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.82);
}

.meta-item span {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.meta-item strong {
  font-size: 13px;
  font-weight: 600;
  color: rgba(11, 18, 32, 0.9);
  word-break: break-all;
}

.meta-item[data-tone='error'] {
  border-color: rgba(212, 56, 13, 0.22);
}

.meta-item[data-tone='error'] strong {
  color: rgba(212, 56, 13, 0.92);
}

@media (max-width: 920px) {
  .panel-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
