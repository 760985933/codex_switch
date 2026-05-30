<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import { GetUsageBalance } from '../../wailsjs/go/main/App'
import type { Profile, UsageBalance } from '../types'

const props = withDefaults(defineProps<{
  profiles: Profile[]
  currentProfileId: string
  loading: boolean
  showDelete?: boolean
  sortable?: boolean
}>(), {
  showDelete: true,
  sortable: false,
})

const emit = defineEmits<{
  edit: [id: string]
  delete: [id: string]
  select: [id: string]
  reorder: [ids: string[]]
}>()

const store = useAppStore()
const message = useMessage()
const { t } = useI18n()

const usageData = reactive<Record<string, UsageBalance | null>>({})
const usageLoadingMap = reactive<Record<string, boolean>>({})

async function fetchUsage(id: string) {
  const profile = props.profiles.find(p => p.id === id)
  if (!profile?.apiKey) return
  usageLoadingMap[id] = true
  try {
    usageData[id] = await GetUsageBalance(id)
  } catch (err) {
    usageData[id] = { availableBalance: '', totalBalance: '', currency: '', isDepleted: false, error: String(err) }
  } finally {
    usageLoadingMap[id] = false
  }
}

watch(() => props.profiles, (profiles) => {
  profiles.forEach(p => {
    if (p.apiKey && !usageData[p.id]) fetchUsage(p.id)
  })
  // Reset local display list when prop changes from outside
  if (!isDragging.value) {
    displayProfiles.value = [...profiles]
  }
}, { immediate: true })

function handleEdit(id: string) {
  emit('edit', id)
}

function friendlyError(raw: string): string {
  if (!raw) return ''
  const lower = raw.toLowerCase()
  if (lower.includes('timeout') || lower.includes('deadline') || lower.includes('exceeded')) {
    return t('guide.usage.timeout')
  }
  if (lower.includes('connection refused') || lower.includes('no such host') || lower.includes('dns') || lower.includes('unreachable')) {
    return t('guide.usage.networkError')
  }
  if (lower.includes('unauthorized') || lower.includes('401') || lower.includes('invalid api key') || lower.includes('authentication') || lower.includes('auth')) {
    return t('guide.usage.authError')
  }
  if (raw.length > 40) {
    return t('guide.usage.queryFailed')
  }
  return raw
}

// ── Delete confirmation ──
const deleteConfirmId = ref<string | null>(null)
const deleteTargetName = computed(() => {
  if (!deleteConfirmId.value) return ''
  return props.profiles.find(p => p.id === deleteConfirmId.value)?.name ?? ''
})

function handleDelete(id: string) {
  if (store.profileList.length < 2) {
    message.warning(t('profile.cannotDeleteLast'))
    return
  }
  deleteConfirmId.value = id
}

function confirmDelete() {
  if (deleteConfirmId.value) {
    emit('delete', deleteConfirmId.value)
  }
  deleteConfirmId.value = null
}

function cancelDelete() {
  deleteConfirmId.value = null
}

// ── Drag-and-drop with live reordering ──
const isDragging = ref(false)
const dragIndex = ref<number | null>(null)
const displayProfiles = ref<Profile[]>([...props.profiles])

function onDragStart(e: DragEvent, index: number) {
  if (!props.sortable) return
  isDragging.value = true
  dragIndex.value = index
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(index))
  }
}

function onDragOver(e: DragEvent, targetIndex: number) {
  if (!props.sortable || dragIndex.value === null) return
  e.preventDefault()
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = 'move'
  }
  if (dragIndex.value === targetIndex) return

  // Live reorder: move the dragged item to the target position
  const items = [...displayProfiles.value]
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const midY = rect.top + rect.height / 2
  const insertAt = e.clientY < midY ? targetIndex : targetIndex + 1

  // Remove from old position (adjust for the fact that the source index may shift)
  const sourceIndex = items.findIndex(p => p.id === displayProfiles.value[dragIndex.value].id)
  const [moved] = items.splice(sourceIndex, 1)
  const adjustedInsert = sourceIndex < insertAt ? insertAt - 1 : insertAt
  items.splice(adjustedInsert, 0, moved)

  displayProfiles.value = items
  dragIndex.value = adjustedInsert
}

function onDragLeave() {
  // Only triggered on the parent container — reset if drag leaves the whole list
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  if (!props.sortable || dragIndex.value === null) return

  const ids = displayProfiles.value.map(p => p.id)
  isDragging.value = false
  dragIndex.value = null
  emit('reorder', ids)
}

function onDragEnd() {
  if (isDragging.value) {
    // Drag was cancelled (not dropped on valid target) — revert
    displayProfiles.value = [...props.profiles]
  }
  isDragging.value = false
  dragIndex.value = null
}

// Keep displayProfiles in sync when props change externally (e.g. after save)
watch(() => props.profiles, (profiles) => {
  if (!isDragging.value) {
    displayProfiles.value = [...profiles]
  }
}, { deep: true })
</script>

<template>
  <TransitionGroup
    name="profile-move"
    tag="div"
    class="profile-list"
    @dragleave="onDragLeave"
  >
    <div
      v-for="(profile, index) in displayProfiles"
      :key="profile.id"
      class="profile-item"
      :class="{
        active: profile.id === currentProfileId,
        dragging: isDragging && dragIndex === index,
      }"
      @click="emit('select', profile.id)"
      @dragover="onDragOver($event, index)"
      @drop="onDrop($event)"
    >
      <!-- Drag handle -->
      <div
        v-if="sortable"
        class="drag-handle"
        :title="t('profile.dragToReorder')"
        draggable="true"
        @click.stop
        @dragstart="onDragStart($event, index)"
        @dragend="onDragEnd"
      >
        <svg width="12" height="16" viewBox="0 0 12 16" fill="currentColor">
          <circle cx="3" cy="3" r="1.2" />
          <circle cx="9" cy="3" r="1.2" />
          <circle cx="3" cy="8" r="1.2" />
          <circle cx="9" cy="8" r="1.2" />
          <circle cx="3" cy="13" r="1.2" />
          <circle cx="9" cy="13" r="1.2" />
        </svg>
      </div>

      <div class="profile-item-main">
        <div class="profile-item-info">
          <div class="profile-item-name-row">
            <span v-if="profile.name" class="profile-item-label">{{ t('config.fields.profileName') }}:</span>
            <span class="profile-item-name">{{ profile.name }}</span>
            <span class="profile-badge">{{ profile.provider }}</span>
          </div>
          <span v-if="profile.baseURL" class="profile-item-meta">
            <span class="profile-item-label">API:</span> {{ profile.baseURL }}
          </span>
          <span v-if="profile.defaultModel" class="profile-item-meta">
            <span class="profile-item-label">{{ t('config.fields.defaultModel') }}:</span> {{ profile.defaultModel }}
          </span>
        </div>
        <div class="profile-item-right">
          <div v-if="usageData[profile.id]" class="profile-item-usage" @click.stop>
            <template v-if="usageData[profile.id]?.error">
              <span class="usage-error-icon">!</span>
              <span class="usage-error" :title="usageData[profile.id]?.error">{{ friendlyError(usageData[profile.id]?.error || '') }}</span>
            </template>
            <template v-else>
              <span>{{ t('guide.usage.available') }}: {{ usageData[profile.id]?.availableBalance }} {{ usageData[profile.id]?.currency }}</span>
              <span class="usage-sep">/</span>
              <span>{{ t('guide.usage.total') }}: {{ usageData[profile.id]?.totalBalance }} {{ usageData[profile.id]?.currency }}</span>
              <span v-if="usageData[profile.id]?.isDepleted" class="usage-depleted">{{ t('guide.usage.depleted') }}</span>
            </template>
            <n-button text size="tiny" :loading="usageLoadingMap[profile.id]" @click="fetchUsage(profile.id)">
              <template #icon>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
              </template>
            </n-button>
          </div>
          <div class="profile-item-actions" @click.stop>
            <slot name="actions" :profile="profile" />
            <n-button size="small" tertiary @click="handleEdit(profile.id)">
              {{ t('models.editModel') }}
            </n-button>
            <slot name="actions-after" :profile="profile" />
            <n-button v-if="showDelete" size="small" tertiary type="error" @click="handleDelete(profile.id)">
              {{ t('models.deleteModel') }}
            </n-button>
          </div>
        </div>
      </div>
    </div>
  </TransitionGroup>

  <!-- Delete Confirmation Dialog -->
  <n-modal
    :show="deleteConfirmId !== null"
    preset="dialog"
    :title="t('profile.delete')"
    :content="t('profile.confirmDelete', { name: deleteTargetName })"
    :positive-text="t('common.delete')"
    :negative-text="t('models.cancel')"
    @positive-click="confirmDelete"
    @negative-click="cancelDelete"
    @close="cancelDelete"
  />
</template>

<style scoped>
/* ── TransitionGroup move animation ── */
.profile-move-move {
  transition: transform 0.25s ease;
}
.profile-move-enter-active,
.profile-move-leave-active {
  transition: all 0.25s ease;
}
.profile-move-enter-from,
.profile-move-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
.profile-move-leave-active {
  position: absolute;
}

.profile-list {
  display: grid;
  gap: 8px;
  position: relative;
}

.profile-item {
  display: flex;
  align-items: stretch;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, box-shadow 0.15s, opacity 0.15s, box-shadow 0.2s;
}

.profile-item:hover {
  background: rgba(255, 255, 255, 0.8);
  border-color: rgba(22, 119, 255, 0.18);
  box-shadow: 0 2px 8px rgba(14, 30, 68, 0.06);
}

.profile-item.active {
  border-color: rgba(22, 119, 255, 0.35);
  background: rgba(22, 119, 255, 0.06);
}

.profile-item.dragging {
  opacity: 0.45;
  box-shadow: 0 4px 16px rgba(14, 30, 68, 0.12);
}

.drag-handle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  min-width: 20px;
  margin-right: 8px;
  color: rgba(11, 18, 32, 0.3);
  cursor: grab;
  border-radius: 4px;
  transition: color 0.15s, background 0.15s;
  align-self: center;
}

.drag-handle:hover {
  color: rgba(11, 18, 32, 0.6);
  background: rgba(11, 18, 32, 0.06);
}

.drag-handle:active {
  cursor: grabbing;
}

.profile-item-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.profile-item-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
  flex-shrink: 0;
}

.profile-item-info {
  display: grid;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.profile-item-name-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: wrap;
}

.profile-item-label {
  font-size: 11px;
  color: rgba(11, 18, 32, 0.55);
  white-space: nowrap;
}

.profile-item-name {
  font-size: 13px;
  font-weight: 600;
  color: rgba(11, 18, 32, 0.9);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.profile-badge {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  background: rgba(22, 119, 255, 0.12);
  color: rgba(22, 119, 255, 0.92);
}

.profile-item-meta {
  font-size: 11px;
  color: rgba(11, 18, 32, 0.6);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  word-break: break-all;
}

.profile-item-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
  align-items: center;
  flex-wrap: wrap;
}

.actions-sep {
  color: var(--border);
  font-size: 12px;
  margin: 0 2px;
}

.profile-item-usage {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: rgba(19, 160, 90, 0.88);
  font-weight: 500;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.usage-sep {
  color: var(--border);
}

.usage-depleted {
  color: rgba(212, 56, 13, 0.92);
  font-weight: 600;
}

.usage-error {
  color: rgba(212, 56, 13, 0.92);
  font-size: 11px;
  word-break: break-word;
  line-height: 1.4;
}

.usage-error-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: rgba(212, 56, 13, 0.92);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  flex-shrink: 0;
}
</style>
