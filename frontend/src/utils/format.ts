export function formatUptime(seconds: number): string {
  if (!seconds || seconds < 0) {
    return '0 秒'
  }

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainSeconds = seconds % 60

  if (hours > 0) {
    return `${hours} 小时 ${minutes} 分`
  }

  if (minutes > 0) {
    return `${minutes} 分 ${remainSeconds} 秒`
  }

  return `${remainSeconds} 秒`
}

export function maskSecret(value: string): string {
  if (!value) {
    return ''
  }
  if (value.length <= 8) {
    return `${value.slice(0, 2)}****`
  }
  return `${value.slice(0, 4)}****${value.slice(-4)}`
}

export function formatTime(value: string): string {
  if (!value) {
    return '--'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`
}
