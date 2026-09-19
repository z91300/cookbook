<template>
  <UApp>
    <div class="min-h-screen bg-zinc-50 text-zinc-900">
      <header class="sticky top-0 z-40 border-b border-zinc-200 bg-white">
        <div class="mx-auto flex h-14 max-w-6xl items-center px-4">
          <NuxtLink to="/" class="flex min-w-0 items-center gap-2.5">
            <span class="flex size-8 shrink-0 items-center justify-center rounded-md bg-green-600 text-sm font-bold text-white">C</span>
            <span class="truncate text-sm font-semibold tracking-wide">Cookbook · 项目模板</span>
          </NuxtLink>
        </div>
      </header>

      <NuxtPage />
    </div>
  </UApp>
</template>
