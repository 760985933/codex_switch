<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import ProxySettingsPanel from '../components/ProxySettingsPanel.vue'

const store = useAppStore()
const { t } = useI18n()
</script>

<template>
  <div class="proxy-page">
    <div class="page-header">
      <div>
        <h2>{{ t('proxy.title') }}</h2>
        <p class="page-desc">{{ t('proxy.description') }}</p>
      </div>
    </div>

    <ProxySettingsPanel :config="store.config" />
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
</style>
