<script setup lang="ts">
import { computed, watch } from 'vue'
import { lightTheme, dateDeDE, dateEnUS, dateEsAR, dateFrFR, dateJaJP, dateKoKR, dateZhCN, deDE, enUS, esAR, frFR, jaJP, koKR, zhCN } from 'naive-ui'
import { RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { WindowMinimise, Quit } from '../wailsjs/runtime/runtime'
import SettingsDrawer from './components/SettingsDrawer.vue'
import SidebarNav from './components/SidebarNav.vue'
import { useAppStore } from './stores/app'
import { useUiStore } from './stores/ui'

const store = useAppStore()
const ui = useUiStore()
const { t, locale } = useI18n()

function handleMinimise() { WindowMinimise() }
function handleClose() { Quit() }

watch(() => ui.locale, (value) => { locale.value = value }, { immediate: true })

const naiveLocale = computed(() => {
  switch (ui.locale) {
    case 'zh-CN': return zhCN
    case 'ja-JP': return jaJP
    case 'ko-KR': return koKR
    case 'fr-FR': return frFR
    case 'de-DE': return deDE
    case 'es-AR': return esAR
    default: return enUS
  }
})

const naiveDateLocale = computed(() => {
  switch (ui.locale) {
    case 'zh-CN': return dateZhCN
    case 'ja-JP': return dateJaJP
    case 'ko-KR': return dateKoKR
    case 'fr-FR': return dateFrFR
    case 'de-DE': return dateDeDE
    case 'es-AR': return dateEsAR
    default: return dateEnUS
  }
})
</script>

<template>
  <n-config-provider :theme="lightTheme" :locale="naiveLocale" :date-locale="naiveDateLocale">
    <n-dialog-provider>
      <n-message-provider placement="bottom-right">
        <div class="shell">
          <div class="titlebar">
            <div class="titlebar-drag" />
            <div class="titlebar-actions">
              <n-button tertiary size="small" @click="handleMinimise">—</n-button>
              <n-button tertiary type="error" size="small" @click="handleClose">×</n-button>
            </div>
          </div>
          <div class="shell-body">
            <SidebarNav @show-help="ui.showHelp = true" />
            <main class="content">
              <RouterView />
            </main>
          </div>

          <SettingsDrawer
            v-model:model-value="ui.showSettings"
            :config="store.config"
          />

          <n-modal v-model:show="ui.showHelp" preset="card" :title="'💡 ' + t('app.help.title')" style="max-width: 600px" :bordered="false" closable>
            <div class="help-content">
              <div class="help-section">
                <h4>{{ t('app.help.usage.title') }}</h4>
                <ol class="help-steps">
                  <li>{{ t('app.help.usage.step1') }}</li>
                  <li>{{ t('app.help.usage.step2') }}</li>
                  <li>{{ t('app.help.usage.step3') }}</li>
                  <li>{{ t('app.help.usage.step4') }}</li>
                  <li>{{ t('app.help.usage.step5') }}</li>
                </ol>
              </div>
              <div class="help-section">
                <h4>{{ t('app.help.backup.title') }}</h4>
                <p>{{ t('app.help.backup.desc') }}</p>
                <ol class="help-steps">
                  <li>{{ t('app.help.backup.step1') }}</li>
                  <li>{{ t('app.help.backup.step2') }}</li>
                </ol>
                <p class="help-note">{{ t('app.help.backup.note') }}</p>
              </div>
            </div>
          </n-modal>
        </div>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<style scoped>
.shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.titlebar {
  display: flex;
  align-items: center;
  height: 32px;
  min-height: 32px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border);
  -webkit-app-region: drag;
  user-select: none;
}

.titlebar-drag {
  flex: 1;
}

.titlebar-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-right: 10px;
  -webkit-app-region: no-drag;
}

.shell-body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.content {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding: 24px;
}
</style>

<style>
.help-content {
  display: grid;
  gap: 20px;
}

.help-section h4 {
  margin: 0 0 10px;
  font-size: 15px;
  color: var(--text);
}

.help-steps {
  margin: 0;
  padding-left: 20px;
  line-height: 2;
  font-size: 13px;
  color: var(--text);
}

.help-note {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--accent);
  opacity: 0.85;
}
</style>
