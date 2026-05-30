<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FetchChangelog } from '../../wailsjs/go/main/App'

const { t } = useI18n()

const content = ref('')
const loading = ref(true)
const fromCache = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await FetchChangelog()
    content.value = result.content
    fromCache.value = result.fromCache
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="changelog-page">
    <div class="page-head">
      <h2>{{ t('changelog.title') }}</h2>
      <span v-if="fromCache" class="cache-badge">{{ t('changelog.fromCache') }}</span>
    </div>

    <div class="card">
      <div v-if="loading" class="loading-state">
        <n-spin size="medium" />
        <span>{{ t('changelog.loading') }}</span>
      </div>

      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <n-button size="small" @click="load">{{ t('changelog.retry') }}</n-button>
      </div>

      <pre v-else class="changelog-content">{{ content }}</pre>
    </div>
  </div>
</template>

<style scoped>
.changelog-page {
  display: grid;
  gap: 14px;
}

.page-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.page-head h2 {
  margin: 0;
  font-size: 18px;
  color: var(--text);
}

.cache-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(255, 170, 0, 0.12);
  color: #b8860b;
}

.card {
  padding: 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
  min-height: 200px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: var(--muted);
  font-size: 13px;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 0;
  color: var(--danger);
  font-size: 13px;
}

.error-state p {
  margin: 0;
  text-align: center;
}

.changelog-content {
  margin: 0;
  font-family: inherit;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-wrap: break-word;
  color: var(--text);
}
</style>
