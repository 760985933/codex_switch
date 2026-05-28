<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import type { UsageStats, UsageStatsResponse } from '../types'
import UsageStatsCard from '../components/UsageStatsCard.vue'
import TokenStatsPanel from '../components/TokenStatsPanel.vue'

const { t } = useI18n()
const store = useAppStore()
const loading = ref(false)
const activeTab = ref<'today' | 'week' | 'month' | 'year'>('today')

type TabKey = 'today' | 'week' | 'month' | 'year'

const tabKeyMap: Record<TabKey, keyof UsageStatsResponse> = {
  today: 'today',
  week: 'thisWeek',
  month: 'thisMonth',
  year: 'thisYear',
}

const tabs = computed<TabKey[]>(() => ['today', 'week', 'month', 'year'])

const tabLabels = computed(() => ({
  today: t('monitoring.today'),
  week: t('monitoring.week'),
  month: t('monitoring.month'),
  year: t('monitoring.year'),
}))

const currentStats = computed<UsageStats[]>(() => {
  if (!store.usageStats) return []
  const key = tabKeyMap[activeTab.value]
  return (store.usageStats[key] as UsageStats[]) ?? []
})

const totalRequests = computed(() => {
  return currentStats.value.reduce((sum, s) => sum + s.requestCount, 0)
})

const totalTokens = computed(() => {
  return currentStats.value.reduce((sum, s) => sum + s.totalTokens, 0)
})

async function loadStats() {
  if (loading.value) return
  loading.value = true
  try {
    await store.getUsageStats()
  } catch (err) {
    console.error('Failed to load usage stats:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadStats()
})
</script>

<template>
  <div class="monitoring-page">
    <div class="page-header">
      <h2>{{ t('monitoring.title') }}</h2>
      <n-button size="small" tertiary :loading="loading" @click="loadStats">
        <template #icon>
          <span>↻</span>
        </template>
      </n-button>
    </div>

    <n-tabs
      v-model:value="activeTab"
      type="line"
      animated
      class="time-tabs"
    >
      <n-tab v-for="tab in tabs" :key="tab" :name="tab" :tab="tabLabels[tab]" />
    </n-tabs>

    <div v-if="loading" class="loading-state">
      <n-spin size="medium" />
    </div>

    <div v-else-if="currentStats.length === 0" class="empty-state">
      <div class="empty-icon">📊</div>
      <p>{{ t('monitoring.noData') }}</p>
      <p class="empty-hint">
        启动代理并发送请求后，用量数据将自动记录
      </p>
    </div>

    <div v-if="currentStats.length > 0" class="stats-summary">
      <n-card class="summary-card" size="small" :bordered="true">
        <div class="summary-grid">
          <div class="summary-item">
            <span class="summary-value">{{ totalRequests }}</span>
            <span class="summary-label">{{ t('monitoring.requestCount') }}</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ totalTokens.toLocaleString() }}</span>
            <span class="summary-label">{{ t('monitoring.totalTokens') }}</span>
          </div>
        </div>
      </n-card>
    </div>

    <TokenStatsPanel
      v-if="currentStats.length > 0"
      :stats="currentStats"
      :total-tokens="totalTokens"
    />

    <div v-if="currentStats.length > 0" class="stats-grid">
      <UsageStatsCard
        v-for="s in currentStats"
        :key="s.provider"
        :stats="s"
      />
    </div>
  </div>
</template>

<style scoped>
.monitoring-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 8px 0;
}
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--text);
}
.time-tabs {
  margin-bottom: 20px;
}
.loading-state {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--muted);
}
.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}
.empty-state p {
  margin: 0 0 8px;
  font-size: 15px;
}
.empty-hint {
  font-size: 13px;
  opacity: 0.7;
}
.stats-summary {
  margin-bottom: 20px;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.summary-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text);
}
.summary-label {
  font-size: 12px;
  color: var(--muted);
}
.stats-grid {
  display: grid;
  gap: 16px;
}
</style>
