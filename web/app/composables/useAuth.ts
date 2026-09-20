// 登录态（双令牌）：
//   · access  短效 2 小时，存 localStorage['cookbook_token']，随请求头 Authorization 带上
//   · refresh 长效 30 天，存 localStorage['cookbook_refresh_token']，access 过期时自动换新
// 令牌读写与自动刷新都在 app/api/request.ts（含 4401 单飞刷新 + 重试）。
//
// 权限极简：只有「管理员 / 普通用户」两种身份。
//   未登录：只读（页面上的新建/编辑入口一律隐藏，后端也会拒绝写操作）
//   普通用户：除「标签管理 / 删除菜谱 / 用户管理」外与过去一致（含新建、编辑菜谱）
//   管理员：可管理标签、删除菜谱、管理用户
import { apis } from '~/api'
// token 读写工具来自 request.ts（worma 生成时不会被覆盖）；
// 不要从 ~/api 聚合入口导入这些——index.ts 是生成文件，只重导出 ApiError/request/ApiRequestConfig。
import { clearAuthTokens, getAuthToken, setAuthExpiredHandler, setAuthTokens } from '~/api/request'
import type { Cookbook_internal_model_user_profile } from '~/api/components'

export function useAuth() {
  const user = useState<Cookbook_internal_model_user_profile | null>('auth-user', () => null)
  // ready：令牌校验过一轮（避免首帧把「登录中」错判成「未登录」）
  const ready = useState<boolean>('auth-ready', () => false)
  const busy = useState<boolean>('auth-busy', () => false)

  const isLoggedIn = computed(() => user.value !== null)
  const isAdmin = computed(() => user.value?.isAdmin === true)
  /** canEdit：能否写（新建/编辑/收藏/编排等），未登录一律只读 */
  const canEdit = computed(() => user.value !== null)
  const displayName = computed(() => user.value?.nickname || user.value?.username || '')

  // 刷新令牌也失效时（被禁用、会话被删、refresh 过期）由 request.ts 回调，清掉页面用户态
  if (import.meta.client) {
    setAuthExpiredHandler(() => {
      user.value = null
    })
  }

  /** 用当前令牌拉取用户信息；access 过期会自动刷新，刷新也失败则视为未登录 */
  async function fetchProfile() {
    if (!getAuthToken()) {
      user.value = null
      ready.value = true
      return
    }
    try {
      const res = await apis.user.getProfile()
      user.value = { id: res.id, username: res.username, nickname: res.nickname, isAdmin: res.isAdmin }
    }
    catch {
      clearAuthTokens()
      user.value = null
    }
    ready.value = true
  }

  /** 登录（后端返回 access + refresh 双令牌） */
  async function login(username: string, password: string) {
    busy.value = true
    try {
      const res = await apis.user.login({ body: { username, password } })
      setAuthTokens(res.token ?? '', res.refreshToken ?? '')
      user.value = res.user ?? null
      ready.value = true
    }
    finally {
      busy.value = false
    }
  }

  /** 注册（库中还没有可用管理员时，首个注册者自动成为管理员） */
  async function register(username: string, password: string, nickname: string) {
    busy.value = true
    try {
      const res = await apis.user.register({ body: { username, password, nickname } })
      setAuthTokens(res.token ?? '', res.refreshToken ?? '')
      user.value = res.user ?? null
      ready.value = true
    }
    finally {
      busy.value = false
    }
  }

  /** 退出登录：先通知后端删除会话（双令牌同时失效），本地无论成败都清干净 */
  async function logout() {
    busy.value = true
    try {
      await apis.user.logout()
    }
    catch {
      // 会话可能已过期，忽略
    }
    finally {
      clearAuthTokens()
      user.value = null
      busy.value = false
    }
  }

  return { user, ready, busy, isLoggedIn, isAdmin, canEdit, displayName, fetchProfile, login, register, logout }
}
