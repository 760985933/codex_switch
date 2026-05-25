<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import type { AppConfig } from '../types'
import { useAppStore } from '../stores/app'

const props = defineProps<{
  modelValue: boolean
  config: AppConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  save: [value: AppConfig]
  export: []
  codexCopy: []
  codexWrite: []
}>()

const localConfig = ref<AppConfig>({ ...props.config })
const store = useAppStore()
const message = useMessage()
const dialog = useDialog()

const codexPath = ref('')
const codexRaw = ref('')
const codexBusy = ref(false)
const codexBackups = ref<string[]>([])
const selectedBackup = ref<string>('')

const backupOptions = computed(() => {
  return codexBackups.value.map((p) => ({
    label: p.split('/').slice(-1)[0] || p,
    value: p,
  }))
})

const needsWireApiFix = computed(() => {
  if (!codexRaw.value) return false
  const value = codexRaw.value
  const providerBlock = /\[\s*model_providers\.local-bridge\s*\][\s\S]*?(\n\[|$)/.exec(value)
  if (!providerBlock) return false
  return /wire_api\s*=\s*"chat"/.test(providerBlock[0])
})

watch(
  () => props.config,
  (value) => {
    localConfig.value = {
      ...value,
      mappings: { ...value.mappings },
      headers: { ...value.headers },
    }
  },
  { deep: true, immediate: true },
)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      void loadCodexRaw()
    }
  },
)

function submit() {
  emit('save', localConfig.value)
}

async function loadCodexRaw() {
  codexBusy.value = true
  try {
    codexPath.value = await store.getCodexConfigPath()
    codexRaw.value = await store.readCodexConfigToml()
    await refreshCodexBackups()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    codexBusy.value = false
  }
}

async function refreshCodexBackups() {
  codexBackups.value = await store.listCodexConfigBackups()
  if (selectedBackup.value && !codexBackups.value.includes(selectedBackup.value)) {
    selectedBackup.value = ''
  }
}

async function generateCodexRaw() {
  codexBusy.value = true
  try {
    codexRaw.value = await store.generateCodexConfigToml()
    message.success('已生成 TOML（可直接保存）')
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    codexBusy.value = false
  }
}

async function saveCodexRaw() {
  codexBusy.value = true
  try {
    const path = await store.writeCodexConfigTomlRaw(codexRaw.value)
    message.success(`已保存: ${path || codexPath.value}`)
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    codexBusy.value = false
  }
}

async function mergeWriteCodex() {
  codexBusy.value = true
  try {
    const path = await store.writeCodexConfigToml()
    message.success(`已合并写入: ${path || codexPath.value}`)
    await loadCodexRaw()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    codexBusy.value = false
  }
}

async function restoreCodex() {
  dialog.warning({
    title: '恢复 Codex 配置',
    content: '将使用最新备份覆盖恢复；若无备份则尝试移除 local-bridge 配置。',
    positiveText: '恢复',
    negativeText: '取消',
    onPositiveClick: async () => {
      codexBusy.value = true
      try {
        const path = await store.restoreCodexConfigToml()
        message.success(`已恢复: ${path || codexPath.value}`)
        await loadCodexRaw()
      } catch (error) {
        message.error(error instanceof Error ? error.message : String(error))
      } finally {
        codexBusy.value = false
      }
    },
  })
}

async function restoreSelectedBackup() {
  if (!selectedBackup.value) {
    message.warning('请选择一个备份')
    return
  }
  codexBusy.value = true
  try {
    const path = await store.restoreCodexConfigTomlFromBackup(selectedBackup.value)
    message.success(`已恢复: ${path || codexPath.value}`)
    await loadCodexRaw()
  } catch (error) {
    message.error(error instanceof Error ? error.message : String(error))
  } finally {
    codexBusy.value = false
  }
}

async function deleteSelectedBackup() {
  if (!selectedBackup.value) {
    message.warning('请选择一个备份')
    return
  }
  dialog.warning({
    title: '删除备份',
    content: `确认删除备份：${selectedBackup.value.split('/').slice(-1)[0]}？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      codexBusy.value = true
      try {
        await store.deleteCodexConfigBackup(selectedBackup.value)
        selectedBackup.value = ''
        await refreshCodexBackups()
        message.success('已删除备份')
      } catch (error) {
        message.error(error instanceof Error ? error.message : String(error))
      } finally {
        codexBusy.value = false
      }
    },
  })
}

async function clearAllBackups() {
  dialog.warning({
    title: '清理备份',
    content: '将删除所有备份文件（不影响当前 config.toml）。',
    positiveText: '清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      codexBusy.value = true
      try {
        const removed = await store.clearCodexConfigBackups()
        selectedBackup.value = ''
        await refreshCodexBackups()
        message.success(`已清理 ${removed} 份备份`)
      } catch (error) {
        message.error(error instanceof Error ? error.message : String(error))
      } finally {
        codexBusy.value = false
      }
    },
  })
}
</script>

<template>
  <n-drawer
    :show="modelValue"
    placement="right"
    :width="420"
    @update:show="(value: boolean) => emit('update:modelValue', value)"
  >
    <n-drawer-content title="偏好设置" closable>
      <div class="drawer-body">
        <n-card size="small" embedded>
          <n-space vertical size="large">
            <n-switch v-model:value="localConfig.enableAutoStart">
              <template #checked>自动启动桥接</template>
              <template #unchecked>自动启动桥接</template>
            </n-switch>
            <n-switch v-model:value="localConfig.minimizeToTray">
              <template #checked>关闭时隐藏窗口</template>
              <template #unchecked>关闭时隐藏窗口</template>
            </n-switch>
            <n-switch v-model:value="localConfig.compactMode">
              <template #checked>紧凑布局</template>
              <template #unchecked>紧凑布局</template>
            </n-switch>
          </n-space>
        </n-card>

        <n-form label-placement="top">
          <n-form-item label="日志保留天数">
            <n-input-number v-model:value="localConfig.logRetentionDays" :min="1" :max="30" />
          </n-form-item>
        </n-form>

        <n-space>
          <n-button type="primary" @click="submit">保存设置</n-button>
          <n-button secondary @click="emit('export')">导出配置</n-button>
        </n-space>

        <n-card size="small" embedded>
          <n-space vertical size="small">
            <div>
              <n-text style="font-weight: 600">Codex config.toml</n-text>
              <n-text depth="3" style="display: block; margin-top: 6px; line-height: 1.6">
                用于让 Codex 走本地桥接。支持直接编辑、保存，并自动生成历史备份用于回滚。
              </n-text>
            </div>
            <n-space>
              <n-button secondary @click="emit('codexCopy')">复制 TOML</n-button>
              <n-button type="primary" @click="emit('codexWrite')">写入文件</n-button>
            </n-space>
            <n-form label-placement="top">
              <n-form-item label="文件路径">
                <n-input :value="codexPath" readonly />
              </n-form-item>
              <n-form-item label="内容（可直接编辑）">
                <n-input
                  v-model:value="codexRaw"
                  type="textarea"
                  :autosize="{ minRows: 10, maxRows: 22 }"
                  :disabled="codexBusy"
                />
              </n-form-item>
            </n-form>

            <div v-if="needsWireApiFix" class="warning-text">
              检测到 local-bridge 的 wire_api 仍为 "chat"。点击“合并写入”可自动修复为 "responses" 并保留其它配置项。
            </div>

            <n-form label-placement="top">
              <n-form-item label="历史备份">
                <n-select
                  v-model:value="selectedBackup"
                  :options="backupOptions"
                  placeholder="选择一个备份用于恢复"
                  :disabled="codexBusy"
                  filterable
                />
              </n-form-item>
            </n-form>

            <n-space>
              <n-button tertiary :loading="codexBusy" @click="loadCodexRaw">读取文件</n-button>
              <n-button tertiary :loading="codexBusy" @click="generateCodexRaw">生成模板</n-button>
              <n-button secondary :loading="codexBusy" @click="saveCodexRaw">保存覆盖</n-button>
              <n-button type="primary" :loading="codexBusy" @click="mergeWriteCodex">合并写入</n-button>
              <n-button tertiary :loading="codexBusy" @click="refreshCodexBackups">刷新备份</n-button>
              <n-button secondary :loading="codexBusy" @click="restoreSelectedBackup">恢复所选</n-button>
              <n-button tertiary :loading="codexBusy" @click="deleteSelectedBackup">删除所选</n-button>
              <n-button tertiary :loading="codexBusy" @click="clearAllBackups">清理备份</n-button>
              <n-button tertiary :loading="codexBusy" @click="restoreCodex">恢复最新</n-button>
            </n-space>
          </n-space>
        </n-card>
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.drawer-body {
  display: grid;
  gap: 20px;
}

.warning-text {
  font-size: 12px;
  line-height: 1.6;
  color: var(--warning);
}
</style>
