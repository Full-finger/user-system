/**
 * Site Configuration — 站点外观与行为配置
 *
 * 修改此文件后刷新页面即可生效，无需重新构建前端。
 * 将图片文件放到 public/ 目录下即可通过相对路径引用，例如:
 *   siteLogo: '/my-logo.png'
 */
window.__SITE_CONFIG__ = {
  /* ---- 基础信息 ---- */

  // 站点名称
  siteName: 'DevMoe',

  // 站点副标题 / 描述
  siteDescription: '开发者社区',

  /* ---- Logo & 图标 ---- */

  // 站点 Logo 图片 URL。支持相对路径（public/ 目录下的文件）或绝对 URL。
  // 设为 null 则使用默认的 PhSparkle 图标。
  // 示例: '/logo.png', '/avatar.jpg', 'https://example.com/logo.svg'
  siteLogo: null,

  // 网站 Favicon URL。设为 null 则使用默认的 /favicon.svg
  siteFavicon: null,

  /* ---- 界面文案 ---- */

  // 搜索框占位提示文字
  searchPlaceholder: '搜索帖子、用户、标签...',

  // 右侧栏公告内容（仅展示，不影响功能）
  announcement: '欢迎使用 DevMoe 社区！',
}
