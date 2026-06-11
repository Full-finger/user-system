/**
 * 相对时间格式化（刚刚 / N分钟前 / N小时前 / N天前 / 日期）
 */
export function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const diff = Math.floor((Date.now() - d) / 1000)
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  if (diff < 604800) return Math.floor(diff / 86400) + ' 天前'
  return d.toLocaleDateString('zh-CN')
}

/**
 * 数量简写（1000 → 1.0K, 10000 → 1.0W）
 */
export function formatCount(n) {
  if (!n) return '0'
  if (n >= 10000) return (n / 10000).toFixed(1) + 'W'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}