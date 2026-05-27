<script setup lang="ts">
import { computed, h, onMounted, ref, resolveComponent } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import {
  CountLegacySessions,
  GetCodexSessionContent,
  ListCodexSessionBackups,
  ListCodexSessions,
  MigrateCodexProviders,
  RestoreCodexSessions,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const sessions = ref<CodexSession[]>([])
const legacyCount = ref(0)
const migrating = ref(false)

const selectedSession = ref<SessionDetail | null>(null)
const detailLoading = ref(false)
const showDetail = ref(false)

const backups = ref<string[]>([])
const restoringBackup = ref(false)

interface CodexSession {
  id: string
  title: string
  model: string
  modelProvider: string
  messageCount: number
  createdAt: string
  isArchived: boolean
}

interface SessionMessage {
  role: string
  content: string
  timestamp: string
}

interface SessionDetail {
  session: CodexSession
  messages: SessionMessage[]
}

function formatBackupName(path: string): string {
  const parts = path.replace(/\\/g, '/').split('/')
  const name = parts[parts.length - 1] || path
  return name.replace(/^sessions_backup_/, '').replace(/\.tar$/, '')
}

function formatTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const columns = computed(() => {
  const NTag = resolveComponent('NTag')
  return [
    {
      title: t('sessions.table.title'),
      key: 'title',
      ellipsis: { tooltip: true },
      minWidth: 180,
      render(row: CodexSession) {
        return row.title || row.id.slice(0, 12) + '…'
      },
    },
    {
      title: t('sessions.table.model'),
      key: 'model',
      width: 160,
      ellipsis: { tooltip: true },
    },
    {
      title: t('sessions.table.messages'),
      key: 'messageCount',
      width: 100,
      align: 'center' as const,
    },
    {
      title: t('sessions.table.time'),
      key: 'createdAt',
      width: 180,
      render(row: CodexSession) {
        return formatTime(row.createdAt)
      },
    },
    {
      title: t('sessions.table.status'),
      key: 'isArchived',
      width: 100,
      align: 'center' as const,
      render(row: CodexSession) {
        return row.isArchived
          ? h(NTag, { size: 'small', type: 'warning' }, () => t('sessions.status.archived'))
          : h(NTag, { size: 'small', type: 'success' }, () => t('sessions.status.active'))
      },
    },
  ]
})

async function loadSessions() {
  loading.value = true
  try {
    sessions.value = (await ListCodexSessions()) ?? []
    legacyCount.value = await CountLegacySessions()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    loading.value = false
  }
}

async function viewSession(row: CodexSession) {
  detailLoading.value = true
  showDetail.value = true
  try {
    selectedSession.value = await GetCodexSessionContent(row.id)
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
    showDetail.value = false
  } finally {
    detailLoading.value = false
  }
}

function closeDetail() {
  showDetail.value = false
  selectedSession.value = null
}

async function confirmMigrate() {
  dialog.warning({
    title: t('sessions.migration.confirmTitle'),
    content: t('sessions.migration.confirmContent'),
    positiveText: t('sessions.migration.button'),
    negativeText: '取消',
    onPositiveClick: async () => {
      await doMigrate()
    },
  })
}

async function doMigrate() {
  migrating.value = true
  try {
    const result = await MigrateCodexProviders('Local', 'openai')
    if (result.error) {
      message.error(result.error)
      return
    }
    message.success(t('sessions.migration.success', { count: result.migratedCount, path: result.backupPath }))
    await loadSessions()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    migrating.value = false
  }
}

async function loadBackups() {
  try {
    backups.value = (await ListCodexSessionBackups()) ?? []
  } catch {
    backups.value = []
  }
}

async function confirmRestore(backupPath: string) {
  dialog.warning({
    title: t('sessions.backup.restoreConfirm'),
    content: t('sessions.backup.restoreConfirmContent'),
    positiveText: t('sessions.backup.restoreButton'),
    negativeText: '取消',
    onPositiveClick: async () => {
      await doRestore(backupPath)
    },
  })
}

async function doRestore(backupPath: string) {
  restoringBackup.value = true
  try {
    const result = await RestoreCodexSessions(backupPath)
    if (result.error) {
      message.error(result.error)
      return
    }
    message.success(t('sessions.backup.restoreSuccess', { count: result.migratedCount }))
    await loadSessions()
    await loadBackups()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    restoringBackup.value = false
  }
}

onMounted(() => {
  loadSessions()
  loadBackups()
})
</script>

<template>
  <div class="sessions-page">
    <div class="page-head">
      <div>
        <h2>{{ t('sessions.title') }}</h2>
        <p>{{ t('sessions.desc') }}</p>
      </div>
      <n-space>
        <n-button secondary :loading="loading" @click="loadSessions">刷新</n-button>
      </n-space>
    </div>

    <!-- Migration banner -->
    <div v-if="legacyCount > 0" class="migration-banner">
      <div class="migration-content">
        <span class="migration-icon">⚠️</span>
        <span>{{ t('sessions.migration.banner', { count: legacyCount }) }}</span>
      </div>
      <n-button type="warning" :loading="migrating" @click="confirmMigrate">
        {{ t('sessions.migration.button') }}
      </n-button>
    </div>

    <!-- Backup management -->
    <div v-if="backups.length > 0" class="backup-section">
      <div class="backup-header">
        <span class="backup-title">{{ t('sessions.backup.title') }}</span>
      </div>
      <div class="backup-list">
        <div v-for="(bp, idx) in backups" :key="idx" class="backup-item">
          <span class="backup-name">{{ formatBackupName(bp) }}</span>
          <n-button
            size="tiny"
            secondary
            type="warning"
            :loading="restoringBackup"
            @click="confirmRestore(bp)"
          >
            {{ t('sessions.backup.restoreButton') }}
          </n-button>
        </div>
      </div>
    </div>

    <!-- Session list -->
    <div class="session-list-wrap">
      <n-data-table
        v-if="sessions.length > 0"
        :columns="columns"
        :data="sessions"
        :loading="loading"
        :row-props="(row: CodexSession) => ({ style: 'cursor: pointer;', onClick: () => viewSession(row) })"
        :bordered="false"
        :single-line="false"
        size="small"
      />
      <div v-else-if="!loading" class="empty-state">
        <n-empty :description="t('sessions.empty')" />
      </div>
    </div>

    <!-- Session detail drawer -->
    <n-drawer :show="showDetail" :width="620" placement="right" @mask-click="closeDetail" @esc="closeDetail">
      <n-drawer-content :title="t('sessions.detail.title')" closable @close="closeDetail">
        <div v-if="detailLoading" class="detail-loading">
          <n-spin :description="t('sessions.detail.loading')" />
        </div>
        <div v-else-if="selectedSession" class="detail-content">
          <div class="session-meta">
            <n-space vertical :size="4">
              <n-text depth="3">ID: {{ selectedSession.session.id }}</n-text>
              <n-text depth="3">{{ t('sessions.table.model') }}: {{ selectedSession.session.model }}</n-text>
              <n-text depth="3">{{ t('sessions.table.time') }}: {{ formatTime(selectedSession.session.createdAt) }}</n-text>
            </n-space>
          </div>
          <div class="messages-list">
            <div
              v-for="(msg, idx) in selectedSession.messages"
              :key="idx"
              class="message-row"
              :data-role="msg.role"
            >
              <div class="message-role">
                <span class="message-role-label">{{
                  msg.role === 'user' ? t('sessions.detail.roleUser') : t('sessions.detail.roleAssistant')
                }}</span>
                <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
              </div>
              <div class="message-content">{{ msg.content }}</div>
            </div>
            <div v-if="selectedSession.messages && selectedSession.messages.length === 0" class="no-messages">
              <n-text depth="3">{{ t('sessions.detail.noContent') }}</n-text>
            </div>
          </div>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<style scoped>
.sessions-page {
  display: grid;
  gap: 14px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 18px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
  box-shadow: 0 10px 30px rgba(14, 30, 68, 0.08);
}

.page-head h2 {
  margin: 0 0 6px;
  font-size: 18px;
  color: var(--text);
}

.page-head p {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.migration-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 18px;
  border-radius: 18px;
  border: 1px solid rgba(216, 150, 20, 0.3);
  background: rgba(216, 150, 20, 0.06);
}

.migration-content {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text);
}

.migration-icon {
  font-size: 18px;
}

.backup-section {
  border-radius: 18px;
  border: 1px solid var(--border);
  background: var(--surface);
  overflow: hidden;
}

.backup-header {
  padding: 12px 16px 0;
}

.backup-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.backup-list {
  display: grid;
}

.backup-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border);
}

.backup-item:last-child {
  border-bottom: none;
}

.backup-name {
  font-size: 13px;
  color: var(--text);
  font-family: monospace;
}

.session-list-wrap {
  min-height: 200px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  border-radius: 22px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.detail-loading {
  display: flex;
  justify-content: center;
  padding: 60px 0;
}

.session-meta {
  padding: 12px 0 16px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 16px;
}

.messages-list {
  display: grid;
  gap: 14px;
  max-height: calc(100vh - 260px);
  overflow-y: auto;
  padding: 4px 0;
}

.message-row {
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.message-row[data-role='assistant'] {
  background: rgba(22, 119, 255, 0.04);
  border-color: rgba(22, 119, 255, 0.12);
}

.message-role {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--accent);
  margin-bottom: 6px;
}

.message-time {
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
  opacity: 0.55;
  font-size: 11px;
}

.message-row[data-role='user'] .message-role {
  color: var(--muted);
}

.message-content {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 400px;
  overflow-y: auto;
}

.no-messages {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

@media (max-width: 920px) {
  .page-head {
    flex-direction: column;
    align-items: stretch;
  }

  .migration-banner {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
