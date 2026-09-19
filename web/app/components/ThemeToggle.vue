<script setup lang="ts">
// 深浅色切换（页面右上角）：初始跟随系统，手动切换后由 colorMode 记忆
// （localStorage + cookie，SSR 首屏即正确着色，无闪烁）。
const colorMode = useColorMode()

// heroicons outline：sun / moon
const SUN_PATH = 'M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z'
const MOON_PATH = 'M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z'

const isDark = computed(() => colorMode.value === 'dark')

function toggle() {
  colorMode.preference = isDark.value ? 'light' : 'dark'
}
</script>

<template>
  <button
    type="button"
    class="icon-btn theme-toggle"
    :aria-label="isDark ? '切换到浅色模式' : '切换到深色模式'"
    :title="isDark ? '切换到浅色模式' : '切换到深色模式'"
    @click="toggle"
  >
    <!-- 图标随模式切换（深色显示太阳=点击回浅色；浅色显示月亮=点击去深色）。
         路径用动态 :d 而非 v-if 分支：SSR 首屏（未知系统偏好时回退浅色）与客户端
         分支可能不同，同标签 svg 分支在水合时不会被可靠替换，动态属性则会。 -->
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      class="size-5"
      aria-hidden="true"
    >
      <path :d="isDark ? SUN_PATH : MOON_PATH" />
    </svg>
  </button>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.theme-toggle {
    @apply size-9 shrink-0 text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-100;
  }
</style>
