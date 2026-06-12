/**
 * 渲染帖子/评论内容：HTML 转义 + mention 高亮 + 外部链接处理 + 换行
 * @param {string} content - 原始内容
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
    .replace(/(https?:\/\/[^\s<>\x26]+)/g, (url) => {
      // 已经被 mention-link 包裹的 URL 不处理（理论上不会，但防御性编码）
      const displayUrl = url.length > 50 ? url.slice(0, 47) + '...' : url
      return `<a href="${url}" class="external-link" target="_blank" rel="nofollow noopener noreferrer">${displayUrl}</a>`
    })
    .replace(/\n/g, '<br>')
}

/**
 * 处理内容区域的点击事件：mention 跳转 + 外部链接确认。
 * 在组件的 @click 处理函数中调用。
 * @param {Event} e - 点击事件
 * @param {Object} router - Vue Router 实例
 */
export function handleRenderedContentClick(e, router) {
  // mention 链接
  const mentionLink = e.target.closest('.mention-link')
  if (mentionLink) {
    e.preventDefault()
    router.push({ name: 'UserProfile', params: { username: mentionLink.dataset.username } })
    return
  }

  // 外部链接跳转确认
  const externalLink = e.target.closest('.external-link')
  if (externalLink) {
    e.preventDefault()
    const url = externalLink.getAttribute('href')
    if (url && confirm(`即将离开本站，前往外部链接：\n\n${url}\n\n是否继续？`)) {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  }
}