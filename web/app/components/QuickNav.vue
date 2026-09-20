<script setup lang="ts">
// 右下角快捷导航：悬浮球 + 点击展开菜单（speed-dial）。
// 收起时只占一个圆钮，避免常驻胶囊按钮遮挡移动端内容。
// 未登录整体隐藏：这几个入口都需要登录后才能用（只读浏览不打扰）。
const { isLoggedIn } = useAuth()

const open = ref(false)

const links = [
  { label: '收藏', to: '/favorites', icon: 'star' },
  { label: '成员', to: '/members', icon: 'users' },
  { label: '食谱编排', to: '/scheduling', icon: 'calendar' },
  { label: '设置', to: '/settings', icon: 'cog' },
] as const

// 退出登录时收起菜单，避免残留的全屏遮罩挡住页面点击
watch(isLoggedIn, (logged) => {
  if (!logged) open.value = false
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') open.value = false
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

// 移动端返回手势 / 浏览器返回键：收起菜单而不是退出页面。
// 菜单项点击不做 UI 关闭（避免关闭补偿回退与路由跳转竞态），靠路由切换卸载组件清理。
useModalBackClose(() => open.value && isLoggedIn.value, () => {
  open.value = false
})
</script>

<template>
  <Teleport v-if="isLoggedIn" to="body">
    <!-- 展开期间铺满全屏接管点击：点空白处收起 -->
    <div v-if="open" class="fixed inset-0 z-40" @click="open = false" />

    <nav aria-label="快捷导航" class="quick-nav">
      <!-- 展开项：自下而上错落弹出（显隐/位移动画态用内联原子类驱动） -->
      <NuxtLink
        v-for="(link, i) in links"
        :key="link.to"
        :to="link.to"
        class="quick-nav__item"
        :class="open ? 'translate-y-0 scale-100 opacity-100' : 'pointer-events-none translate-y-2 scale-90 opacity-0'"
        :style="{ transitionDelay: open ? `${i * 40}ms` : '0ms' }"
        :aria-hidden="!open"
        :tabindex="open ? undefined : -1"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="size-4.5 shrink-0"
          aria-hidden="true"
        >
          <path v-if="link.icon === 'star'" d="M11.48 3.5a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
          <path v-else-if="link.icon === 'users'" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
          <path v-else-if="link.icon === 'calendar'" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
          <template v-else>
            <path d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
            <path d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
          </template>
        </svg>
        {{ link.label }}
      </NuxtLink>

      <!-- 悬浮球：网格图标 ↔ 关闭图标交叉淡切（动画态用内联原子类驱动） -->
      <button
        type="button"
        class="quick-nav__toggle"
        :aria-expanded="open"
        aria-label="快捷导航"
        @click="open = !open"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="col-start-1 row-start-1 size-6 transition-all duration-200"
          :class="open ? 'scale-50 opacity-0' : 'scale-100 opacity-100'"
          aria-hidden="true"
        >
          <path d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" />
        </svg>
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="col-start-1 row-start-1 size-6 transition-all duration-200"
          :class="open ? 'scale-100 opacity-100' : 'scale-50 opacity-0'"
          aria-hidden="true"
        >
          <path d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </nav>
  </Teleport>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.quick-nav {
    @apply fixed bottom-[calc(1.25rem+env(safe-area-inset-bottom))] right-[calc(1.25rem+env(safe-area-inset-right))] z-40 flex flex-col items-end gap-2.5;
  }

.quick-nav__item {
    @apply flex items-center gap-2 rounded-full border border-zinc-200 bg-white py-2 pl-3.5 pr-4 text-sm font-medium text-zinc-700 shadow-lg shadow-zinc-900/10 transition-all duration-200 hover:border-green-500 hover:text-green-600 active:scale-95 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-300 dark:shadow-black/40 dark:hover:border-green-500 dark:hover:text-green-400;
  }

.quick-nav__toggle {
    @apply grid size-14 place-items-center rounded-full bg-green-600 text-white shadow-lg shadow-green-600/40 transition-all duration-200 hover:bg-green-700 active:scale-90;
  }
</style>
