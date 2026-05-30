import { reactive } from 'vue'

/**
 * 默认站点配置。
 * 用户可通过 public/site.config.js (window.__SITE_CONFIG__) 覆盖任意字段。
 */
const defaults = {
  siteName: 'DevMoe',
  siteDescription: '开发者社区',
  siteLogo: null,
  siteFavicon: null,
  searchPlaceholder: '搜索帖子、用户、标签...',
  announcement: '欢迎使用 DevMoe 社区！',
}

const userConfig =
  typeof window !== 'undefined' && window.__SITE_CONFIG__
    ? window.__SITE_CONFIG__
    : {}

/**
 * 响应式站点配置对象。
 * 所有组件应从此处读取，不要在组件中硬编码品牌信息。
 */
export const siteConfig = reactive({
  ...defaults,
  ...userConfig,
})

/**
 * 将 favicon <link> 的 href 更新为配置值。
 * 在应用初始化时调用一次即可。
 */
export function applyFavicon() {
  if (!siteConfig.siteFavicon) return
  const link =
    document.querySelector('link[rel="icon"]') ||
    document.querySelector('link[rel="shortcut icon"]')
  if (link) link.href = siteConfig.siteFavicon
}

/**
 * 将 document.title 更新为配置值。
 */
export function applyTitle() {
  document.title = `${siteConfig.siteName} — ${siteConfig.siteDescription}`
}
