<template>
  <UApp>
    <div class="app-shell">
      <header class="app-header">
        <div class="app-header__inner">
          <NuxtLink to="/" class="app-logo">
            <span class="app-logo__badge">C</span>
            <span class="app-logo__title">Cookbook · 项目模板</span>
          </NuxtLink>
          <ThemeToggle />
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

.app-logo {
    @apply flex min-w-0 items-center gap-2.5;
  }

.app-logo__badge {
    @apply flex size-8 shrink-0 items-center justify-center rounded-md bg-green-600 text-sm font-bold text-white;
  }

.app-logo__title {
    @apply truncate text-sm font-semibold tracking-wide;
  }
</style>
