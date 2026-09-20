<script setup lang="ts">
// 全局外壳：顶栏（Logo + 主题切换 + 登录态/用户区）
const { isLoggedIn, isAdmin, displayName, fetchProfile, logout } = useAuth()

// 进站按本地令牌校验一次登录态（令牌失效会自动清除）
onMounted(fetchProfile)

const initial = computed(() => (displayName.value || '?').slice(0, 1).toUpperCase())

async function onLogout() {
  await logout()
  await navigateTo('/')
}
</script>

<template>
  <UApp>
    <div class="app-shell">
      <header class="app-header">
        <div class="app-header__inner">
          <NuxtLink to="/" class="app-logo">
            <span class="app-logo__badge">C</span>
            <span class="app-logo__title">Cookbook</span>
          </NuxtLink>

          <div class="app-header__right">
            <ThemeToggle />

            <!-- 已登录：昵称（管理员带标识）+ 退出 -->
            <div v-if="isLoggedIn" class="user-chip">
              <span class="user-chip__avatar">{{ initial }}</span>
              <span class="user-chip__name">{{ displayName }}</span>
              <span v-if="isAdmin" class="tag-badge tag-badge--green">管理员</span>
              <button type="button" class="text-btn text-btn--muted text-btn--xs" @click="onLogout">退出</button>
            </div>

            <!-- 未登录：登录 / 注册入口 -->
            <NuxtLink v-else to="/login" class="btn btn--outline btn--xs">登录</NuxtLink>
          </div>
        </div>
      </header>

      <NuxtPage />
    </div>
  </UApp>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.app-shell {
    @apply min-h-screen bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100;
  }

.app-header {
    @apply sticky top-0 z-40 border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900;
  }

.app-header__inner {
    @apply mx-auto flex h-14 max-w-6xl items-center justify-between gap-2 px-4;
  }

.app-header__right {
    @apply flex shrink-0 items-center gap-2 sm:gap-3;
  }

.app-logo {
    @apply flex min-w-0 items-center gap-2.5;
  }

.app-logo__badge {
    @apply flex size-8 shrink-0 items-center justify-center rounded-md bg-green-600 text-sm font-bold text-white;
  }

.app-logo__title {
    @apply truncate text-sm font-semibold tracking-wide;
  }

/* ---- 用户区 ---- */
.user-chip {
    @apply flex items-center gap-1.5 rounded-full border border-zinc-200 py-1 pl-1 pr-2 dark:border-zinc-700;
  }

.user-chip__avatar {
    @apply flex size-6 shrink-0 items-center justify-center rounded-full bg-green-600 text-xs font-bold text-white;
  }

.user-chip__name {
    @apply max-w-24 truncate text-xs font-medium text-zinc-700 dark:text-zinc-200;
  }
</style>
