<script setup lang="ts">
// 登录 / 注册页
// 说明：库里还没有可登录的管理员时，第一个完成注册的用户自动成为管理员。
const route = useRoute()
const { isLoggedIn, isAdmin, displayName, login, register } = useAuth()

type Mode = 'login' | 'register'
const mode = ref<Mode>('login')

const username = ref('')
const password = ref('')
const nickname = ref('')
const busy = ref(false)
const error = ref('')

const modes: Array<{ key: Mode, label: string }> = [
  { key: 'login', label: '登录' },
  { key: 'register', label: '注册' },
]

// 已登录时直接给出状态与返回入口，避免重复登录
const tips: Record<Mode, string> = {
  login: '登录后即可在设置页管理标签与用户。',
  register: '首次使用？注册即可；系统里的第一个注册用户会自动成为管理员。',
}

async function submit() {
  if (busy.value) return
  error.value = ''
  const name = username.value.trim()
  if (!name || !password.value) {
    error.value = '请填写用户名和密码'
    return
  }
  if (mode.value === 'register' && password.value.length < 6) {
    error.value = '密码至少 6 位'
    return
  }
  busy.value = true
  try {
    if (mode.value === 'login') {
      await login(name, password.value)
    }
    else {
      await register(name, password.value, nickname.value.trim())
    }
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await navigateTo(redirect)
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '操作失败，请稍后重试'
  }
  busy.value = false
}

function switchMode(next: Mode) {
  mode.value = next
  error.value = ''
}
</script>

<template>
  <main class="page page--narrow">
    <div class="page-header">
      <h1 class="page-title">账号</h1>
      <NuxtLink to="/" class="text-btn text-btn--accent">← 返回首页</NuxtLink>
    </div>

    <!-- 已登录：状态卡片 -->
    <section v-if="isLoggedIn" class="card mt-6 p-6">
      <p class="text-sm text-zinc-600 dark:text-zinc-300">
        当前已登录：<span class="font-medium">{{ displayName }}</span>
        <span v-if="isAdmin" class="tag-badge tag-badge--green ml-2">管理员</span>
      </p>
      <p class="form-hint mt-2">如需切换账号，请先退出登录。</p>
      <div class="mt-4 flex gap-3">
        <NuxtLink to="/" class="btn btn--outline">返回首页</NuxtLink>
        <NuxtLink to="/settings" class="btn btn--primary">前往设置</NuxtLink>
      </div>
    </section>

    <!-- 未登录：登录 / 注册表单 -->
    <section v-else class="card mt-6 p-6">
      <div class="auth-tabs">
        <button
          v-for="m in modes"
          :key="m.key"
          type="button"
          class="auth-tab"
          :class="{ 'auth-tab--active': mode === m.key }"
          @click="switchMode(m.key)"
        >
          {{ m.label }}
        </button>
      </div>

      <p class="form-hint mt-3">{{ tips[mode] }}</p>

      <form class="mt-5 space-y-4" @submit.prevent="submit">
        <div>
          <label for="auth-username" class="form-label">用户名</label>
          <input
            id="auth-username"
            v-model="username"
            type="text"
            autocomplete="username"
            maxlength="20"
            placeholder="2-20 个字符"
            class="input w-full"
          >
        </div>

        <div v-if="mode === 'register'">
          <label for="auth-nickname" class="form-label">昵称（可选）</label>
          <input
            id="auth-nickname"
            v-model="nickname"
            type="text"
            maxlength="20"
            placeholder="留空则与用户名相同"
            class="input w-full"
          >
        </div>

        <div>
          <label for="auth-password" class="form-label">密码</label>
          <input
            id="auth-password"
            v-model="password"
            type="password"
            :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
            maxlength="32"
            placeholder="至少 6 位"
            class="input w-full"
          >
        </div>

        <p v-if="error" class="error-alert">{{ error }}</p>

        <button type="submit" class="btn btn--primary w-full" :disabled="busy">
          {{ busy ? '处理中…' : (mode === 'login' ? '登录' : '注册并登录') }}
        </button>
      </form>
    </section>
  </main>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.auth-tabs { @apply flex gap-6 border-b border-zinc-200 dark:border-zinc-800; }
.auth-tab { @apply -mb-px border-b-2 border-transparent px-1 pb-2 text-sm text-zinc-500 transition-colors hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200; }
.auth-tab--active { @apply border-green-600 font-medium text-green-700 hover:text-green-700 dark:border-green-500 dark:text-green-400 dark:hover:text-green-400; }
</style>
