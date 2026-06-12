/**
 * 渲染帖子内容：HTML 转义 + mention 高亮 + 换行
 * @param {string} content - 原始帖子内容
 * @param {Array} mentions - 后端返回的 mentions 数组
 * @returns {string} 安全的 HTML 字符串
 */
export function renderContent(content, mentions) {
  if (!content) return ''
  const mentionMap = new Map()
  if (mentions?.length) {
    for (const m of mentions) {
      mentionMap.set(m.username.toLowerCase(), { username: m.username, nickname: m.nickname || m.username })
    }
  }
  return content
    .replace(/&/g, '\x26amp;')
    .replace(/</g, '\x26lt;')
    .replace(/>/g, '\x26gt;')
    .replace(/@([a-zA-Z0-9_]{3,30})/g, (match, name) => {
      const info = mentionMap.get(name.toLowerCase())
      if (info) {
        return `<a href="/users/${info.username}" class="mention-link" data-username="${info.username}">@${info.nickname}</a>`
      }
      return match
    })
    .replace(/\n/g, '<br>')
}
