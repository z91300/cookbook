export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',

  devtools: { enabled: true },

  // 显式绑定 IPv4 回环。默认 host 为 localhost，本机 DNS 把 localhost 优先解析为 ::1，
  // 导致 dev server 只监听 [::1]:3000，http://127.0.0.1:3000 访问被拒。
  devServer: {
    host: '127.0.0.1',
  },

  modules: ['@nuxt/ui'],

  // 站点为纯浅色设计（扁平/小圆角/无阴影），不跟随系统暗色
  ui: {
    colorMode: false,
  },

  css: ['~/assets/css/main.css'],

  telemetry: false,

  fonts: {
    providers: {
      google: false,
      googleicons: false,
      bunny: false,
      fontshare: false,
    },
  },

  runtimeConfig: {
    public: {
      apiBase: '/api',
    },
  },

  nitro: {
    // 生产环境 /api 反代（与下方 devProxy 行为一致：去掉 /api 前缀转发到后端根路径路由）。
    // 合并镜像容器内前端与后端同机，目标固定为容器回环 http://127.0.0.1:8000（本文件默认值），
    // 无需构建期注入；保留 NITRO_API_UPSTREAM 覆盖入口以备未来拆容器。
    routeRules: {
      '/api/**': {
        proxy: `${process.env.NITRO_API_UPSTREAM ?? 'http://127.0.0.1:8000'}/**`,
      },
      // 附件内容路由（/attachments/{id}/content，后端 coverUrl 直出绝对路径）
      '/attachments/**': {
        proxy: `${process.env.NITRO_API_UPSTREAM ?? 'http://127.0.0.1:8000'}/attachments/**`,
      },
    },
    devProxy: {
      '/api/**': {
        target: 'http://127.0.0.1:8000',
        changeOrigin: true,
        prependPath: false,
      },
      '/attachments/**': {
        target: 'http://127.0.0.1:8000',
        changeOrigin: true,
        prependPath: false,
      },
    },
  },
})
