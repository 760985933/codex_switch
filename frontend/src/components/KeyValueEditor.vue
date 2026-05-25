<script setup lang="ts">
import { computed } from 'vue'

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

const items = computed<PairItem[]>(() => {
  const source = Object.entries(props.modelValue ?? {})
  return [...source.map(([key, value]) => ({ key, value })), { key: '', value: '' }]
})

function updateEntry(index: number, field: keyof PairItem, value: string) {
  const nextItems = items.value.map((item) => ({ ...item }))
  nextItems[index][field] = value

  const normalized = nextItems.reduce<Record<string, string>>((acc, item) => {
    const key = item.key.trim()
    const currentValue = item.value.trim()
    if (key && currentValue) {
      acc[key] = currentValue
    }
    return acc
  }, {})

  emit('update:modelValue', normalized)
}

function removeEntry(key: string) {
  const normalized = { ...(props.modelValue ?? {}) }
  delete normalized[key]
  emit('update:modelValue', normalized)
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
        v-for="(item, index) in items"
        :key="`${item.key}-${index}`"
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
          删除
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
