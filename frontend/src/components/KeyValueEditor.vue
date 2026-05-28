<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

interface PairItem {
  key: string
  value: string
}

const props = defineProps<{
  title: string
  description: string
  modelValue: Record<string, string>
  keyPlaceholder: string
  valuePlaceholder: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

const { t } = useI18n()

// Local editing state — avoids focus loss from computed rebuilding on every keystroke
const localItems = ref<PairItem[]>([])
// Tracks the last JSON we emitted to avoid the emit → parent → watch → reset loop
const lastEmitted = ref('')

watch(
  () => props.modelValue,
  (val) => {
    const incoming = JSON.stringify(val ?? {})
    if (incoming !== lastEmitted.value) {
      localItems.value = Object.entries(val ?? {}).map(([k, v]) => ({ key: k, value: v }))
    }
  },
  { deep: true, immediate: true },
)

// Append an empty row for the "add new" affordance
const displayItems = computed<PairItem[]>(() => [...localItems.value, { key: '', value: '' }])

function syncToParent() {
  const normalized: Record<string, string> = {}
  for (const item of localItems.value) {
    const k = item.key.trim()
    const v = item.value.trim()
    if (k && v) normalized[k] = v
  }
  lastEmitted.value = JSON.stringify(normalized)
  emit('update:modelValue', normalized)
}

function updateEntry(index: number, field: keyof PairItem, value: string) {
  const next = localItems.value.map((item) => ({ ...item }))
  // If index is past the end, we're filling in the "add new" row
  if (index >= next.length) {
    const item: PairItem = { key: '', value: '' }
    item[field] = value
    next.push(item)
  } else {
    next[index] = { ...next[index], [field]: value }
  }
  localItems.value = next
  syncToParent()
}

function removeEntry(key: string) {
  localItems.value = localItems.value.filter((e) => e.key !== key)
  syncToParent()
}
</script>

<template>
  <div class="kv-editor">
    <div class="section-head">
      <div>
        <h4>{{ title }}</h4>
        <p>{{ description }}</p>
      </div>
    </div>

    <div class="kv-list">
      <div
        v-for="(item, index) in displayItems"
        :key="index"
        class="kv-row"
      >
        <n-input
          :value="item.key"
          :placeholder="keyPlaceholder"
          @update:value="(value: string) => updateEntry(index, 'key', value)"
        />
        <n-input
          :value="item.value"
          :placeholder="valuePlaceholder"
          @update:value="(value: string) => updateEntry(index, 'value', value)"
        />
        <n-button
          tertiary
          type="error"
          :disabled="!item.key"
          @click="removeEntry(item.key)"
        >
          {{ t('common.delete') }}
        </n-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.kv-editor {
  display: grid;
  gap: 12px;
}

.section-head h4 {
  margin: 0 0 4px;
  font-size: 14px;
  color: var(--text);
}

.section-head p {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.kv-list {
  display: grid;
  gap: 10px;
}

.kv-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 72px;
  gap: 10px;
}

@media (max-width: 920px) {
  .kv-row {
    grid-template-columns: 1fr;
  }
}
</style>
