export interface ProviderPreset {
  id: string
  label: string
  defaultBaseURL: string
  defaultModel: string
  docsURL: string
  placeholderApiKey: string
  placeholderModel: string
}

export const PROVIDER_PRESETS: ProviderPreset[] = [
  {
    id: 'deepseek',
    label: 'DeepSeek',
    defaultBaseURL: 'https://api.deepseek.com/v1',
    defaultModel: 'deepseek-v4-flash',
    docsURL: 'https://api-docs.deepseek.com/',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'deepseek-v4-flash',
  },
  {
    id: 'alibaba',
    label: '阿里通义千问',
    defaultBaseURL: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    defaultModel: 'qwen3.6-plus',
    docsURL: 'https://help.aliyun.com/zh/model-studio/models',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'qwen3.6-plus',
  },
  {
    id: 'xiaomi',
    label: '小米 MiMo',
    defaultBaseURL: 'https://api.xiaomimimo.com/v1',
    defaultModel: 'mimo-v2.5-pro',
    docsURL: 'https://platform.xiaomimimo.com/#/docs/welcome',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'mimo-v2.5-pro',
  },
  {
    id: 'zhipu',
    label: '智谱 GLM',
    defaultBaseURL: 'https://open.bigmodel.cn/api/paas/v4',
    defaultModel: 'glm-4.7-flash',
    docsURL: 'https://docs.bigmodel.cn/',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'glm-4.7-flash',
  },
  {
    id: 'baidu',
    label: '百度千帆',
    defaultBaseURL: 'https://qianfan.baidubce.com/v2',
    defaultModel: 'ernie-5.1',
    docsURL: 'https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Fm2vrveyu',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'ernie-5.1',
  },
  {
    id: 'volcano',
    label: '火山引擎豆包',
    defaultBaseURL: 'https://ark.cn-beijing.volces.com/api/v3',
    defaultModel: 'doubao-seed-2-0-lite-260215',
    docsURL: 'https://www.volcengine.com/docs/82379/1330310',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'doubao-seed-2-0-lite-260215',
  },
  {
    id: 'tencent',
    label: '腾讯混元',
    defaultBaseURL: 'https://api.hunyuan.cloud.tencent.com/v1',
    defaultModel: 'hunyuan-2.0-thinking-20251109',
    docsURL: 'https://cloud.tencent.com/document/product/1729/104753',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'hunyuan-2.0-thinking-20251109',
  },
  {
    id: 'silicon',
    label: '硅基流动',
    defaultBaseURL: 'https://api.siliconflow.cn/v1',
    defaultModel: 'deepseek-ai/DeepSeek-V4-Flash',
    docsURL: 'https://docs.siliconflow.cn/',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'deepseek-ai/DeepSeek-V4-Flash',
  },
  {
    id: 'kimi',
    label: 'Kimi',
    defaultBaseURL: 'https://api.moonshot.cn/v1',
    defaultModel: 'kimi-k2.6',
    docsURL: 'https://platform.moonshot.cn/docs',
    placeholderApiKey: 'sk-...',
    placeholderModel: 'kimi-k2.6',
  },
  {
    id: 'custom',
    label: '自定义',
    defaultBaseURL: '',
    defaultModel: '',
    docsURL: '',
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
