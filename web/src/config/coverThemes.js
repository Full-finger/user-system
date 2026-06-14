/**
 * 封面主题预设（纯 CSS 渐变，复用 design-tokens.css 色彩 DNA）
 *
 * 注意：新增/修改主题时需同步更新后端 internal/service/cover_theme.go 的 ValidCoverThemes 白名单。
 * key 为空串 "" 表示"无封面"（默认）。
 */

/**
 * @typedef {Object} CoverTheme
 * @property {string} key  - 主题 key，空串表示无封面
 * @property {string} name - 中文名称
 * @property {string} css  - background 渐变值（亮色模式）
 */

/** @type {CoverTheme[]} */
export const coverThemes = [
  { key: '', name: '无封面', css: 'none' },
  {
    key: 'sunset',
    name: '落日',
    css: 'linear-gradient(135deg, var(--accent), var(--peach))',
  },
  {
    key: 'lavender',
    name: '薰衣草',
    css: 'linear-gradient(135deg, var(--lavender), var(--accent))',
  },
  {
    key: 'ocean',
    name: '海洋',
    css: 'linear-gradient(135deg, var(--sky), var(--lavender))',
  },
  {
    key: 'forest',
    name: '森林',
    css: 'linear-gradient(135deg, var(--mint), var(--sky))',
  },
  {
    key: 'amber',
    name: '琥珀',
    css: 'linear-gradient(135deg, var(--amber), var(--peach))',
  },
  {
    key: 'blossom',
    name: '花漾',
    css: 'linear-gradient(135deg, var(--accent), var(--lavender))',
  },
  {
    key: 'mint',
    name: '薄荷',
    css: 'linear-gradient(135deg, var(--mint), var(--lavender))',
  },
  {
    key: 'dawn',
    name: '晨曦',
    css: 'linear-gradient(135deg, var(--peach), var(--amber) 50%, var(--mint))',
  },
]

/** key -> css 的快速映射 */
const coverThemeMap = coverThemes.reduce((acc, t) => {
  acc[t.key] = t.css
  return acc
}, {})

/**
 * 根据主题 key 获取渐变 CSS。
 * @param {string} key - 主题 key，空串或未知返回 'none'（无封面）
 * @returns {string}
 */
export function getCoverCSS(key) {
  return coverThemeMap[key] || 'none'
}