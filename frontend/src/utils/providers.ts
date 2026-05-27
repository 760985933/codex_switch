export interface ProviderPreset {
  id: string
  label: string
  defaultBaseURL: string
  defaultModel: string
  placeholderApiKey: string
  placeholderModel: string
}

export const PROVIDER_PRESETS: ProviderPreset[] = [
  {
    id: 'deepseek',
    label: 'DeepSeek',
    defaultBaseURL: 'https://api.deepseek.com/v1',
    defaultModel: 'deepseek-v4-flash',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'deepseek-v4-flash',
  },
  {
    id: 'alibaba',
    label: '阿里通义千问',
    defaultBaseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    defaultModel: 'qwen3.6-plus',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'qwen3.6-plus',
  },
  {
    id: 'xiaomi',
    label: '小米 MiMo',
    defaultBaseURL: 'https://api.xiaomimimo.com/v1',
    defaultModel: 'mimo-v2.5-pro',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'mimo-v2.5-pro',
  },
  {
    id: 'zhipu',
    label: '智谱 GLM',
    defaultBaseURL: 'https://open.bigmodel.cn/api/paas/v4',
    defaultModel: 'glm-4.7-flash',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'glm-4.7-flash',
  },
  {
    id: 'custom',
    label: '自定义',
    defaultBaseURL: '',
    defaultModel: '',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'gpt-4o',
  },
]

export function getProviderPreset(id: string): ProviderPreset | undefined {
  return PROVIDER_PRESETS.find((p) => p.id === id)
}

export function getDefaultProviderPreset(): ProviderPreset {
  return PROVIDER_PRESETS[0]
}
