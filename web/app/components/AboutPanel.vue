<script setup lang="ts">
// 关于：项目名片 + 主要功能 + 技术栈 + 相关链接（GitHub 仓库等，点击新标签页打开）。
// 内容与 README 口径保持一致，功能有增减时同步改这里即可。
const REPO_URL = 'https://github.com/z91300/cookbook'
const IMAGE = 'ghcr.io/z91300/cookbook'

// 项目名片上的能力标签
const aboutBadges = [
  'AGPL-3.0 开源',
  'Go 1.26 + GoFrame v2',
  'Nuxt 4',
  'PostgreSQL / SQLite',
  'Docker 单镜像',
]

// 主要功能：标题 + 一句话说明
const aboutFeatures = [
  { title: '菜谱管理', desc: '卡片墙浏览，关键词 / 标签 / 用餐时段多维筛选；食材与图文步骤结构化编辑，均支持拖拽排序。' },
  { title: '多收藏夹', desc: '一道菜可同时收进多个收藏夹，卡片上的爱心实时显示收藏次数。' },
  { title: '家庭成员', desc: '记录成员角色与口味、忌口、过敏备注，作为「按家人口味编排菜单」的依据。' },
  { title: '每周菜单编排', desc: '周历 × 早 / 午 / 晚 / 加餐四餐段，一格可排多道菜，已排条目拖拽即可换日期或餐段。' },
  { title: '标签与设置', desc: '标签完全自维护、可拖拽排序；通用设置 key-value 存储，内置 AI 生图提示词模板与变量说明。' },
  { title: '图片与附件', desc: '上传即转 WebP、按内容去重，文件本体存库（备份 = 全库备份），列表小图走 360 缩略图。' },
  { title: '用户与权限', desc: '首个注册用户即管理员；access 2 小时 + refresh 30 天双令牌登录态，失效自动续期。' },
]

// 技术栈：层 + 选型
const aboutStack = [
  { label: '后端', value: 'Go 1.26 + GoFrame v2（OpenAPI / Swagger 即接口文档）' },
  { label: '前端', value: 'Nuxt 4 + Vue 3 + Tailwind CSS 4' },
  { label: '数据库', value: 'PostgreSQL 或 SQLite，附件二进制直接入库' },
  { label: '部署', value: `Docker 单镜像，同容器内跑 Nuxt SSR（3000）与 Go API（8000），镜像 ${IMAGE}` },
]

// 相关链接：全部新标签页打开
const aboutLinks = [
  { label: 'GitHub 仓库', url: REPO_URL, desc: '源码、更新日志与完整使用说明' },
  { label: '问题反馈', url: `${REPO_URL}/issues`, desc: '遇到问题或有新想法，欢迎提 Issue 交流' },
  { label: '版本发布', url: `${REPO_URL}/releases`, desc: '查看每个版本的更新内容' },
  { label: '开源许可', url: `${REPO_URL}/blob/main/LICENSE`, desc: 'AGPL-3.0，可自由自托管使用' },
]
</script>

<template>
  <div>
    <!-- 项目名片 -->
    <div class="about-card">
      <span class="about-card__badge">C</span>
      <div class="min-w-0">
        <h2 class="about-card__title">Cookbook · 家庭菜谱管理</h2>
        <p class="about-card__desc">
          自托管的个人 / 家庭菜谱中心：菜谱管理、多收藏夹、家庭成员口味、每周菜单编排。
          数据全部留在自己的数据库里，不依赖任何第三方服务。
        </p>
        <div class="about-card__badges">
          <span v-for="b in aboutBadges" :key="b" class="tag-badge">{{ b }}</span>
        </div>
      </div>
    </div>

    <!-- 主要功能 -->
    <h3 class="about-heading">主要功能</h3>
    <ul class="about-features">
      <li v-for="f in aboutFeatures" :key="f.title" class="about-feature">
        <span class="about-feature__dot" aria-hidden="true" />
        <p class="about-feature__text">
          <span class="about-feature__title">{{ f.title }}</span>
          <span class="about-feature__desc">{{ f.desc }}</span>
        </p>
      </li>
    </ul>

    <!-- 技术栈 -->
    <h3 class="about-heading">技术栈</h3>
    <dl class="about-stack">
      <div v-for="s in aboutStack" :key="s.label" class="about-stack__row">
        <dt class="about-stack__label">{{ s.label }}</dt>
        <dd class="about-stack__value">{{ s.value }}</dd>
      </div>
    </dl>

    <!-- 相关链接 -->
    <h3 class="about-heading">相关链接</h3>
    <ul class="about-links">
      <li v-for="l in aboutLinks" :key="l.url">
        <a
          class="about-link group"
          :href="l.url"
          target="_blank"
          rel="noopener noreferrer"
        >
          <span class="about-link__head">
            <span class="about-link__label">{{ l.label }}</span>
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="about-link__icon"
              aria-hidden="true"
            >
              <path d="M13.5 6H5.25A2.25 2.25 0 0 0 3 8.25v10.5A2.25 2.25 0 0 0 5.25 21h10.5A2.25 2.25 0 0 0 18 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
            </svg>
          </span>
          <span class="about-link__url">{{ l.url }}</span>
          <span class="about-link__desc">{{ l.desc }}</span>
        </a>
      </li>
    </ul>

    <p class="about-footnote">
      如果这个项目对你有帮助，欢迎到
      <a class="text-btn text-btn--accent" :href="REPO_URL" target="_blank" rel="noopener noreferrer">GitHub</a>
      点一个 Star，或提 Issue / PR 一起完善。
    </p>
  </div>
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 项目名片 ---- */
.about-card { @apply flex items-start gap-3 rounded-lg border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900; }
.about-card__badge { @apply flex size-10 shrink-0 items-center justify-center rounded-lg bg-green-600 text-lg font-bold text-white; }
.about-card__title { @apply text-base font-semibold text-zinc-900 dark:text-zinc-100; }
.about-card__desc { @apply mt-1 text-sm leading-relaxed text-zinc-500 dark:text-zinc-400; }
.about-card__badges { @apply mt-2 flex flex-wrap gap-1.5; }

/* ---- 分节标题 ---- */
.about-heading { @apply mt-8 text-sm font-semibold text-zinc-800 dark:text-zinc-200; }

/* ---- 主要功能 ---- */
.about-features { @apply mt-3 grid grid-cols-1 gap-x-6 gap-y-2 sm:grid-cols-2; }
.about-feature { @apply flex items-start gap-2; }
.about-feature__dot { @apply mt-1.5 size-1.5 shrink-0 rounded-full bg-green-600 dark:bg-green-500; }
.about-feature__text { @apply min-w-0 text-sm leading-relaxed text-zinc-500 dark:text-zinc-400; }
.about-feature__title { @apply font-medium text-zinc-800 dark:text-zinc-200; }
.about-feature__title::after { content: ' —— '; }
.about-feature__desc { @apply text-zinc-500 dark:text-zinc-400; }

/* ---- 技术栈 ---- */
.about-stack { @apply mt-3 divide-y divide-zinc-100 rounded-lg border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900; }
.about-stack__row { @apply flex flex-col gap-0.5 px-3 py-2 sm:flex-row sm:items-baseline sm:gap-4; }
.about-stack__label { @apply w-16 shrink-0 text-xs text-zinc-400 dark:text-zinc-500; }
.about-stack__value { @apply min-w-0 break-words text-sm text-zinc-700 dark:text-zinc-300; }

/* ---- 相关链接 ---- */
.about-links { @apply mt-3 divide-y divide-zinc-100 rounded-lg border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900; }
.about-link { @apply block px-3 py-2.5 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800/60; }
.about-link__head { @apply flex items-center justify-between gap-2; }
.about-link__label { @apply text-sm font-medium text-zinc-800 group-hover:text-green-700 dark:text-zinc-200 dark:group-hover:text-green-400; }
.about-link__icon { @apply size-3.5 shrink-0 text-zinc-300 transition-colors group-hover:text-green-600 dark:text-zinc-600 dark:group-hover:text-green-400; }
.about-link__url { @apply mt-0.5 block break-all font-mono text-xs text-zinc-400 dark:text-zinc-500; }
.about-link__desc { @apply mt-0.5 block text-xs text-zinc-400 dark:text-zinc-500; }

.about-footnote { @apply mt-8 text-xs leading-relaxed text-zinc-400 dark:text-zinc-500; }
</style>
