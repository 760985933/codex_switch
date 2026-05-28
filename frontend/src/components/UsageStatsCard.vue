<script setup lang="ts">
import { computed } from 'vue'
import type { UsageStats } from '../types'
import { getProviderPreset } from '../utils/providers'

const props = defineProps<{
  stats: UsageStats
}>()

const providerInfo = computed(() => {
  const preset = getProviderPreset(props.stats.provider)
  return preset ?? { id: props.stats.provider, label: props.stats.provider, defaultBaseURL: '', defaultModel: '', docsURL: '' }
})

const successRate = computed(() => {
  const total = props.stats.successCount + props.stats.failureCount
  if (total === 0) return 100
  return Math.round((props.stats.successCount / total) * 100)
})

const tokenPercentage = computed(() => {
  const total = props.stats.totalTokens
  if (total === 0) return { prompt: 0, completion: 0 }
  return {
    prompt: Math.round((props.stats.promptTokens / total) * 100),
    completion: Math.round((props.stats.completionTokens / total) * 100),
  }
})

const isActive = computed(() => {
  return props.stats.requestCount > 0
})
</script>

<template>
  <n-card class="stats-card" :class="{ inactive: !isActive }" :bordered="true" size="small">
    <div class="stats-header">
      <div class="provider-label">{{ providerInfo.label }}</div>
      <n-tag v-if="stats.failureCount > 0" :type="successRate >= 90 ? 'success' : successRate >= 70 ? 'warning' : 'error'" size="small">
        {{ successRate }}%
      </n-tag>
      <n-tag v-else type="success" size="small">100%</n-tag>
    </div>

    <div class="stats-grid">
      <div class="stat-item">
        <span class="stat-value">{{ stats.requestCount }}</span>
        <span class="stat-label">{{ $t('monitoring.requestCount') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-value success">{{ stats.successCount }}</span>
        <span class="stat-label">{{ $t('monitoring.successCount') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-value failure">{{ stats.failureCount }}</span>
        <span class="stat-label">{{ $t('monitoring.failureCount') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-value">{{ stats.avgDurationMs.toFixed(0) }}<small>{{ $t('monitoring.ms') }}</small></span>
        <span class="stat-label">{{ $t('monitoring.avgDuration') }}</span>
      </div>
    </div>

    <div v-if="stats.totalTokens > 0" class="tokens-section">
      <div class="tokens-header">
        <span class="tokens-label">{{ $t('monitoring.tokens') }}</span>
        <span class="tokens-total">{{ stats.totalTokens.toLocaleString() }}</span>
      </div>
      <div class="token-bar">
        <div
          class="token-bar-prompt"
          :style="{ width: tokenPercentage.prompt + '%' }"
          :title="$t('monitoring.promptTokens') + ': ' + stats.promptTokens.toLocaleString()"
        />
        <div
          class="token-bar-completion"
          :style="{ width: tokenPercentage.completion + '%' }"
          :title="$t('monitoring.completionTokens') + ': ' + stats.completionTokens.toLocaleString()"
        />
      </div>
      <div class="token-details">
        <span class="token-detail">
          <span class="dot prompt" /> {{ $t('monitoring.promptTokens') }}: {{ stats.promptTokens.toLocaleString() }}
        </span>
        <span class="token-detail">
          <span class="dot completion" /> {{ $t('monitoring.completionTokens') }}: {{ stats.completionTokens.toLocaleString() }}
        </span>
      </div>
    </div>
  </n-card>
</template>

<style scoped>
.stats-card {
  border-radius: 12px;
  transition: opacity 0.2s;
}
.stats-card.inactive {
  opacity: 0.5;
}
.stats-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.provider-label {
  font-weight: 600;
  font-size: 15px;
  color: var(--text);
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}
.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text);
}
.stat-value.success {
  color: var(--accent-2);
}
.stat-value.failure {
  color: var(--danger);
}
.stat-value small {
  font-size: 12px;
  font-weight: 400;
  color: var(--muted);
  margin-left: 2px;
}
.stat-label {
  font-size: 11px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.tokens-section {
  border-top: 1px solid var(--border);
  padding-top: 12px;
}
.tokens-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.tokens-label {
  font-size: 12px;
  color: var(--muted);
}
.tokens-total {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}
.token-bar {
  display: flex;
  height: 6px;
  border-radius: 3px;
  overflow: hidden;
  background: var(--border);
  margin-bottom: 8px;
}
.token-bar-prompt {
  background: #1677ff;
  transition: width 0.3s;
}
.token-bar-completion {
  background: #13c2c2;
  transition: width 0.3s;
}
.token-details {
  display: flex;
  gap: 16px;
  font-size: 11px;
  color: var(--muted);
}
.token-detail {
  display: flex;
  align-items: center;
  gap: 4px;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.dot.prompt {
  background: #1677ff;
}
.dot.completion {
  background: #13c2c2;
}
</style>
