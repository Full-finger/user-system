import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { applyFavicon, applyTitle, siteConfig } from './config/site'
import './styles/global.css'
import './styles/components.css'
import './styles/auth.css'

// 应用站点配置
applyFavicon()
applyTitle()

// 监听配置变化（site.config.js 热更场景）同步 title
import { watch } from 'vue'
watch(() => [siteConfig.siteName, siteConfig.siteDescription], applyTitle)

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
