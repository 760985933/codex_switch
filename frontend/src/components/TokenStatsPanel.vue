<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getProviderPreset } from '../utils/providers'
import type { UsageStats } from '../types'

const { t } = useI18n()

const props = defineProps<{
  stats: UsageStats[]
  totalTokens: number
}>()

const sortedStats = computed(() => {
  return [...props.stats].sort((a, b) => b.totalTokens - a.totalTokens)
})

const maxTokens = computed(() => {
  if (sortedStats.value.length === 0) return 0
  return Math.max(...sortedStats.value.map((s) => s.totalTokens), 1)
})

function providerLabel(provider: string): string {
  const preset = getProviderPreset(provider)
  return preset?.label ?? provider
}

function tokenShare(stats: UsageStats): number {
  if (props.totalTokens === 0) return 0
  return Math.round((stats.totalTokens / props.totalTokens) * 100)
}

function avgTokensPerRequest(stats: UsageStats): number {
  if (stats.requestCount === 0) return 0
  return Math.round(stats.totalTokens / stats.requestCount)
}
</script>

<template>
  <n-card class="token-panel" :bordered="true" size="small">
    <template #header>
      <div class="panel-header">
        <span class="panel-title">{{ t('monitoring.tokenStats') }}</span>
      </div>
    </template>

    <div v-if="sortedStats.length === 0" class="panel-empty">
      {{ t('monitoring.noData') }}
    </div>

    <div v-else class="token-table">
      <div class="table-header">
        <span class="col-provider">{{ t('monitoring.requests') }}</span>
        <span class="col-tokens">{{ t('monitoring.totalTokens') }}</span>
        <span class="col-share">{{ t('monitoring.tokenShare') }}</span>
        <span class="col-avg">{{ t('monitoring.avgTokensPerRequest') }}</span>
        <span class="col-split">{{ t('monitoring.promptCompletionSplit') }}</span>
      </div>

      <div
        v-for="s in sortedStats"
        :key="s.provider"
        class="table-row"
      >
        <span class="col-provider">{{ providerLabel(s.provider) }}</span>

        <span class="col-tokens">
          <span class="token-bar-wrap">
            <span
              class="token-bar-fill"
              :style="{ width: (s.totalTokens / maxTokens * 100) + '%' }"
            />
          </span>
          <span class="token-value">{{ s.totalTokens.toLocaleString() }}</span>
        </span>

        <span class="col-share">{{ tokenShare(s) }}%</span>

        <span class="col-avg">{{ avgTokensPerRequest(s) }}</span>

        <span class="col-split">
          <span v-if="s.totalTokens > 0" class="split-bars">
            <span
              class="split-bar prompt"
              :style="{ width: (s.promptTokens / s.totalTokens * 100) + '%' }"
              :title="t('monitoring.promptTokens') + ': ' + s.promptTokens.toLocaleString()"
            />
            <span
              class="split-bar completion"
              :style="{ width: (s.completionTokens / s.totalTokens * 100) + '%' }"
              :title="t('monitoring.completionTokens') + ': ' + s.completionTokens.toLocaleString()"
            />
          </span>
          <span v-else class="split-none">–</span>
        </span>
      </div>
    </div>
  </n-card>
</template>

<style scoped>
.token-panel {
  border-radius: 12px;
  margin-bottom: 20px;
}
.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
}
.panel-title {
  font-weight: 600;
  font-size: 15px;
  color: var(--text);
}
.panel-empty {
  text-align: center;
  padding: 24px;
  color: var(--muted);
  font-size: 13px;
}
.token-table {
  display: flex;
  flex-direction: column;
}
.table-header {
  display: grid;
  grid-template-columns: 140px 1fr 70px 120px 160px;
  gap: 12px;
  padding: 0 0 8px;
  font-size: 11px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 1px solid var(--border);
}
.table-row {
  display: grid;
  grid-template-columns: 140px 1fr 70px 120px 160px;
  gap: 12px;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  color: var(--text);
}
.table-row:last-child {
  border-bottom: none;
}
.col-provider {
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.col-tokens {
  display: flex;
  align-items: center;
  gap: 10px;
}
.token-bar-wrap {
  flex: 1;
  height: 8px;
  border-radius: 4px;
  background: var(--border);
  overflow: hidden;
  max-width: 200px;
}
.token-bar-fill {
  display: block;
  height: 100%;
  border-radius: 4px;
  background: linear-gradient(90deg, #1677ff, #13c2c2);
  transition: width 0.3s;
}
.token-value {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  min-width: 60px;
  text-align: right;
}
.col-share,
.col-avg {
  font-variant-numeric: tabular-nums;
  text-align: right;
}
.col-split {
  display: flex;
  align-items: center;
}
.split-bars {
  display: flex;
  width: 100%;
  height: 6px;
  border-radius: 3px;
  overflow: hidden;
  background: var(--border);
}
.split-bar.prompt {
  background: #1677ff;
  transition: width 0.3s;
}
.split-bar.completion {
  background: #13c2c2;
  transition: width 0.3s;
}
.split-none {
  color: var(--muted);
}
</style>
