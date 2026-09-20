<script setup lang="ts">
// 设置页：标签管理（管理员）/ 用户管理（管理员）/ 通用设置 / 关于
import { apis } from '~/api'

const { isAdmin, isLoggedIn, canEdit, user, fetchProfile } = useAuth()

interface SettingItem {
  key: string
  value: string
}

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const savedTip = ref('')
// 编辑态：key -> value（key 不可改，value 可编辑）
const entries = ref<SettingItem[]>([])

// 提示词模板可用变量说明（与后端 renderTemplate 支持的 {recipe.*} 一致）
const templateVars = [  { name: '{recipe.title}', desc: '菜谱标题' },
  { name: '{recipe.summary}', desc: '一句话简介' },
  { name: '{recipe.ingredients}', desc: '食材列表（逐行「名称 用量」）' },
  { name: '{recipe.steps}', desc: '步骤列表（逐行 ①②③ 编号）' },
  { name: '{recipe.tools}', desc: '工具列表（逐行）' },
  { name: '{recipe.tips}', desc: '小贴士（空时输出省略提示）' },
  { name: '{recipe.stepCount}', desc: '步骤条数，如 5' },
  { name: '{recipe.ingredientCount}', desc: '食材条数，如 11' },
  { name: '{recipe.infoLine}', desc: '信息条，如「约 100 分钟 · 3 人份 · 难度中等」' },
  { name: '{recipe.calories}', desc: '热量，如 160 kcal/份（未填为空）' },
  { name: '{recipe.difficulty}', desc: '难度：简单/中等/较难（未填为空）' },
  { name: '{recipe.servings}', desc: '份量，如 3 人份（未填为空）' },
  { name: '{recipe.cookMinutes}', desc: '耗时，如 90 分钟（未填为空）' },
  { name: '{recipe.mealMask}', desc: '用餐时段，如 早餐、午餐（未填为空）' },
]

const KEY_IMAGE_GEN_PROMPT = 'image_gen_prompt'
const IMAGE_GEN_PROMPT_HINT = [
  '帮我生成一张 {recipe.title} 的菜谱卡片图片，最终必须输出一张完整的 PNG 格式单张图片。',
  '**提示词目标：** 生成一张严格 9:16 纵向比例的简约信息图表式菜谱卡片；单张图像一次性直出（不要分镜、不要多图、不要拼图、不要前后对比、不要附带文字说明页），输出 PNG 格式。',
  '**0. 画布与高度分配（硬约束，最高优先级）**',
  '* 单图直出：仅输出这一张 9:16 卡片，不允许输出多张、分镜或对比图。',
  '* 画布：纵向 9:16，按总高 1920px、总宽 1080px 构图，整图高度记为 100%。',
  '* 上部实拍图区：固定 32% 高度（约 614px），不得压缩、不得放大、不得留白；照片满幅铺满 1080px 宽，盘中菜品主体完整不裁切、不偏心、不歪斜。',
  '* 下部信息区：占剩余 68% 高度（约 1306px）；其中正文排版区不得少于 60%（约 1152px），余下约 8%（约 154px）作为底部留白与边框内边距。',
  '* 全图必须 100% 铺满，底部不许出现大片空白，内容不许溢出或裁切；元素间距按内容多少均匀自适应拉伸，禁止靠留白补齐。',
  '**1. 版式与视觉风格**',
  '* 背景：整张卡片使用米白色（#FAF7F0 附近）带细腻纸张纹理的底色。',
  '* 上下切分：实拍图与纸张区之间是一条清晰的横向硬切分，直接接边，不加阴影、不加圆角、不留缝隙。',
  '* 边框：信息区四周一圈极细的浅灰绿色（#8A9A7B 附近）矩形边框，四角点缀简约植物叶片/藤蔓线稿装饰。',
  '* 字体：标题用优雅的中文衬线字体（宋体/思源宋体类），正文用清晰细衬线；标题下方配一条极简波浪线装饰。',
  '* 插画：全部图示为黑色细线稿 + 单色淡彩填充的手绘风插画，线条粗细统一、笔触干净，不要扁平色块图标、不要写实插画。',
  '* 配色：米白底 + 黑灰线稿 + 菜品主色调的少量点睛色，整体低饱和、干净不杂乱。',
  '**2. 上部实拍图区（真实度要求最高，优先保证）**',
  '* 必须是真实美食摄影照片：单反/微单实拍质感，绝不是插画、3D 渲染、CG 或 AI 涂抹感。',
  '* 菜品：{recipe.title} 成品特写——完整呈现这道菜的关键质感（炖菜汤汁浓稠挂勺带自然油润高光、炒菜油亮入味、汤品清澈或浓醇按实际情况），盛在浅口米白色粗陶汤盘里。',
  '* 器物：哑光米白色粗陶浅口汤盘，放在浅色原木桌面上。',
  '* 摄影参数：等效 50-85mm 焦距，f/2.8-f/4，菜品主体全清晰、背景柔和虚化；自然光从左后上方 45° 侧逆光射入，桌面有柔和自然的投影。',
  '* 色彩：自然真实、低饱和耐看；严禁过饱和、荧光色、HDR 过曝、塑料蜡质感。',
  '* 画面内不得出现任何文字、水印、logo、手指、人手、多余餐具或无关食材。',
  '**3. 下部信息区（自上而下固定顺序，不得增删模块）**',
  '* 标题区：居中大字显示菜名「{recipe.title}」；下方一行小字简介「{recipe.summary}」；再下方一条波浪线装饰。',
  '* 原料区：标题「原料」，下方按图标在上、文字在下的方式等宽排列，每行最多 6 项，按原料总条数（本道 {recipe.ingredientCount} 项）自动换行，项间距均匀；每项写「名称 + 用量」，用量可换行。',
  '* 原料内容（逐字照抄，不得增删改写）：',
  '{recipe.ingredients}',
  '* 做法区：标题「做法」，按步骤实际条数依次排列（本道 {recipe.stepCount} 步，不得增删）；每条 = 左侧黑色圆圈数字序号（①②③…）+ 中间手绘线稿插画（画该步骤的厨具或动作）+ 右侧两行以内的动作文字；条目之间用极细虚线分隔。',
  '* 做法内容（逐字照抄，不得增删改写、不得删减或自行概括）：',
  '{recipe.steps}',
  '* 小贴士区：底部一个细边框圆角矩形，标题「小贴士」，内含 1-2 行提示文字；若小贴士内容为空，则整块省略，其高度并入做法区间距，整图仍保持 100% 铺满。',
  '{recipe.tips}',
  '* 信息条：{recipe.infoLine}。',
  '**4. 一致性锁定（每次生成必须完全一致，只允许内容变）**',
  '* 固定不变：画布比例、32%/68% 高度分配、米白纸张底色、上下硬切分、细边框与四角叶片装饰、字体体系与标题波浪线、线稿插画风格与线条粗细、圆圈数字样式、虚线分隔样式、小贴士框样式、整体低饱和配色。',
  '* 允许变化：菜名、简介、原料内容与条数、步骤内容与条数及对应插画题材、小贴士文字。',
  '* 版式必须稳定：不得改变分区顺序、不得新增模块、不得移动标题位置。',
  '* 文字完整性：做法与原料文字必须逐字照抄、完整展示，不得缩写、合并、省略任何步骤或任一字句；步骤数严格等于所给条数，不得自行删减。',
  '**5. 菜谱内容（严格照抄所给文字，不得增删改写、不得删减或自行概括）**',
  '* 原料：',
  '{recipe.ingredients}',
  '* 做法：',
  '{recipe.steps}',
  '* 小贴士：{recipe.tips}',
  '* 信息条：{recipe.infoLine}',
  '**6. 负向约束 (Negative Prompt)**',
  '* 禁止：画面杂乱、深色背景、背景过暗、文字重叠或错位、错别字、英文、乱码、过于写实的烹饪过程图（过程只做线稿）、分区比例失调、上部图片被压缩拉伸、下部内容溢出裁切、底部大片空白、多余模块、水印与 logo、文字删减/步骤合并/自行概括、多图输出/分镜/拼图/前后对比图、非 PNG 格式。',
].join('\n')

function hintText(key: string): string {
  return key === KEY_IMAGE_GEN_PROMPT ? '模板支持 {recipe.*} 变量（见下方说明），复制到菜谱右键菜单时会替换为对应菜谱内容' : ''
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await apis.setting.getList()
    entries.value = (res.settings ?? []).map(s => ({ key: s.key, value: s.value ?? '' }))
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  }
  loading.value = false
}

async function save() {
  saving.value = true
  error.value = ''
  savedTip.value = ''
  try {
    await apis.setting.upsert({ body: { items: entries.value.map(e => ({ key: e.key, value: e.value })) } })
    savedTip.value = '已保存'
    setTimeout(() => (savedTip.value = ''), 2000)
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '保存失败'
  }
  saving.value = false
}

function fillExample(item: SettingItem) {
  item.value = IMAGE_GEN_PROMPT_HINT
}

// ============ 标签管理：列表（含菜谱数）/ 新增 / 重命名 / 删除 ============
interface TagRow {
  id: number
  name: string
  sort: number
  recipeCount: number
}

// 挂载即拉取，初始置 true 避免首帧闪现「暂无标签」
const tagsLoading = ref(true)
const tagRows = ref<TagRow[]>([])
const tagError = ref('')
const tagSavedTip = ref('')

// 新增
const newTagName = ref('')
const addingTag = ref(false)

// 重命名（行内编辑）
const renamingId = ref<number | null>(null)
const renameValue = ref('')
const renamingBusy = ref(false)

// 删除确认弹窗
const deleteTagTarget = ref<TagRow | null>(null)
const deletingTag = ref(false)
const deleteTagError = ref('')

// 拖拽排序：松手后把整表顺序提交给后端（PUT /tags/sort），失败则按服务端顺序回滚
const dragTagId = ref<number | null>(null) // 正在拖的行
const dragOverTagId = ref<number | null>(null) // 悬停落点行
const savingOrder = ref(false)

function clearTagDrag() {
  dragTagId.value = null
  dragOverTagId.value = null
}

function onTagDragStart(e: DragEvent, tag: TagRow) {
  if (savingOrder.value || renamingId.value !== null) return
  dragTagId.value = tag.id
  tagError.value = ''
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    // Firefox 必须设置 data 才会真正开始拖拽
    e.dataTransfer.setData('text/plain', `tag:${tag.id}`)
  }
}

function onTagDragOver(e: DragEvent, tag: TagRow) {
  if (dragTagId.value === null || dragTagId.value === tag.id) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOverTagId.value = tag.id
}

/** 松手：把被拖的行插到落点行所在位置，再整体提交顺序 */
async function onTagDrop(target: TagRow) {
  const fromId = dragTagId.value
  clearTagDrag()
  if (fromId === null || fromId === target.id) return
  const from = tagRows.value.findIndex(t => t.id === fromId)
  const to = tagRows.value.findIndex(t => t.id === target.id)
  if (from < 0 || to < 0) return
  const next = [...tagRows.value]
  const [moved] = next.splice(from, 1)
  next.splice(to, 0, moved)
  tagRows.value = next
  await saveTagOrder()
}

/** 提交当前顺序；失败时重新拉取，回到服务端顺序 */
async function saveTagOrder() {
  if (savingOrder.value) return
  savingOrder.value = true
  tagError.value = ''
  try {
    await apis.tag.reorder({ body: { ids: tagRows.value.map(t => t.id) } })
    showTagTip('顺序已保存')
    await refreshHomeTags()
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '保存顺序失败'
    await loadTags()
  }
  savingOrder.value = false
}

function showTagTip(text: string) {
  tagSavedTip.value = text
  setTimeout(() => (tagSavedTip.value = ''), 2500)
}

async function loadTags() {
  tagsLoading.value = true
  tagError.value = ''
  try {
    const res = await apis.tag.getManageList()
    tagRows.value = (res.list ?? []).map(t => ({
      id: t.id!,
      name: t.name ?? '',
      sort: t.sort ?? 0,
      recipeCount: t.recipeCount ?? 0,
    }))
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '加载标签失败'
  }
  tagsLoading.value = false
}

// 变更后刷新首页标签栏缓存（useAsyncData key 与首页一致）
async function refreshHomeTags() {
  try {
    await refreshNuxtData('tag-list')
  }
  catch {
    // 首页缓存刷新失败不影响本页
  }
}

async function addTag() {
  const name = newTagName.value.trim()
  if (!name || addingTag.value) return
  addingTag.value = true
  tagError.value = ''
  try {
    await apis.tag.create({ body: { name } })
    newTagName.value = ''
    await loadTags()
    await refreshHomeTags()
    showTagTip('标签已添加')
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '添加失败'
  }
  addingTag.value = false
}

function startRename(tag: TagRow) {
  renamingId.value = tag.id
  renameValue.value = tag.name
  tagError.value = ''
}

function cancelRename() {
  renamingId.value = null
  renameValue.value = ''
}

async function confirmRename() {
  const id = renamingId.value
  const name = renameValue.value.trim()
  if (id === null || !name || renamingBusy.value) return
  renamingBusy.value = true
  tagError.value = ''
  try {
    await apis.tag.update({ pathParams: { id }, body: { name } })
    renamingId.value = null
    renameValue.value = ''
    await loadTags()
    await refreshHomeTags()
    showTagTip('标签已重命名')
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '重命名失败'
  }
  renamingBusy.value = false
}

function askDeleteTag(tag: TagRow) {
  deleteTagError.value = ''
  deleteTagTarget.value = tag
}

function cancelDeleteTag() {
  deleteTagTarget.value = null
}

async function confirmDeleteTag() {
  const tag = deleteTagTarget.value
  if (!tag || deletingTag.value) return
  deletingTag.value = true
  deleteTagError.value = ''
  try {
    const res = await apis.tag.delete({ pathParams: { id: tag.id } })
    deleteTagTarget.value = null
    await loadTags()
    await refreshHomeTags()
    const affected = res.affectedRecipes ?? 0
    showTagTip(affected > 0 ? `已删除标签，并从 ${affected} 道菜谱移除该标签` : '标签已删除')
  }
  catch (e) {
    deleteTagError.value = e instanceof Error ? e.message : '删除失败'
  }
  deletingTag.value = false
}

// ============ 用户管理（仅管理员）============
interface UserRow {
  id: number
  username: string
  nickname: string
  isAdmin: boolean
  status: number
  recipeCount: number
  lastLoginAt: number
  createdAt: number
}

const usersLoading = ref(false)
const userRows = ref<UserRow[]>([])
const userError = ref('')
const userSavedTip = ref('')
const userBusyId = ref<number | null>(null)

// 危险操作二次确认（禁用 / 取消管理员）
const userConfirm = ref<{ kind: 'disable' | 'revokeAdmin', row: UserRow } | null>(null)
const userConfirmBusy = ref(false)
const userConfirmError = ref('')

// 重置密码弹窗
const resetTarget = ref<UserRow | null>(null)
const resetPasswordValue = ref('')
const resetBusy = ref(false)
const resetError = ref('')

function showUserTip(text: string) {
  userSavedTip.value = text
  setTimeout(() => (userSavedTip.value = ''), 2500)
}

function formatDate(unix: number): string {
  if (!unix) return '—'
  const d = new Date(unix * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function loadUsers() {
  usersLoading.value = true
  userError.value = ''
  try {
    const res = await apis.user.getList({ params: { page: 1, pageSize: 100 } })
    userRows.value = (res.list ?? []).map(u => ({
      id: u.id!,
      username: u.username ?? '',
      nickname: u.nickname ?? '',
      isAdmin: u.isAdmin === true,
      status: u.status ?? 1,
      recipeCount: u.recipeCount ?? 0,
      lastLoginAt: u.lastLoginAt ?? 0,
      createdAt: u.createdAt ?? 0,
    }))
  }
  catch (e) {
    userError.value = e instanceof Error ? e.message : '加载用户失败'
  }
  usersLoading.value = false
}

// isSelf：当前登录账号在用户管理里不可被禁用/降权（后端亦有同样约束）
function isSelf(row: UserRow): boolean {
  return user.value?.id === row.id
}

async function patchUser(row: UserRow, payload: { nickname?: string, status?: number, isAdmin?: boolean }, tip: string) {
  userBusyId.value = row.id
  userError.value = ''
  try {
    await apis.user.update({ pathParams: { id: row.id }, body: payload })
    await loadUsers()
    if (isSelf(row)) await fetchProfile()
    showUserTip(tip)
  }
  catch (e) {
    userError.value = e instanceof Error ? e.message : '操作失败'
  }
  userBusyId.value = null
}

// 启用：直接生效；禁用走二次确认
function toggleStatus(row: UserRow) {
  if (row.status === 1) {
    userConfirmError.value = ''
    userConfirm.value = { kind: 'disable', row }
    return
  }
  patchUser(row, { status: 1 }, '已启用该账号')
}

// 取消管理员走二次确认；设为管理员直接生效
function toggleAdmin(row: UserRow) {
  if (row.isAdmin) {
    userConfirmError.value = ''
    userConfirm.value = { kind: 'revokeAdmin', row }
    return
  }
  patchUser(row, { isAdmin: true }, '已设为管理员')
}

function cancelUserConfirm() {
  userConfirm.value = null
  userConfirmError.value = ''
}

async function confirmUserAction() {
  const target = userConfirm.value
  if (!target || userConfirmBusy.value) return
  userConfirmBusy.value = true
  userConfirmError.value = ''
  const { kind, row } = target
  try {
    await apis.user.update({
      pathParams: { id: row.id },
      body: kind === 'disable' ? { status: 0 } : { isAdmin: false },
    })
    userConfirm.value = null
    await loadUsers()
    showUserTip(kind === 'disable' ? '已禁用该账号' : '已取消管理员')
  }
  catch (e) {
    userConfirmError.value = e instanceof Error ? e.message : '操作失败'
  }
  userConfirmBusy.value = false
}

// 昵称行内编辑
const renameUserId = ref<number | null>(null)
const userNicknameValue = ref('')
const userNicknameBusy = ref(false)

function startRenameUser(row: UserRow) {
  renameUserId.value = row.id
  userNicknameValue.value = row.nickname
  userError.value = ''
}

function cancelRenameUser() {
  renameUserId.value = null
  userNicknameValue.value = ''
}

async function confirmRenameUser(row: UserRow) {
  const name = userNicknameValue.value.trim()
  if (userNicknameBusy.value) return
  userNicknameBusy.value = true
  const ok = await patchUser(row, { nickname: name }, '昵称已更新')
  if (ok) renameUserId.value = null
  userNicknameBusy.value = false
}

function askResetPassword(row: UserRow) {
  resetError.value = ''
  resetPasswordValue.value = ''
  resetTarget.value = row
}

function cancelResetPassword() {
  resetTarget.value = null
  resetPasswordValue.value = ''
  resetError.value = ''
}

async function confirmResetPassword() {
  const row = resetTarget.value
  const password = resetPasswordValue.value
  if (!row || resetBusy.value) return
  if (password.length < 6) {
    resetError.value = '密码至少 6 位'
    return
  }
  resetBusy.value = true
  resetError.value = ''
  try {
    await apis.user.resetPassword({ pathParams: { id: row.id }, body: { password } })
    resetTarget.value = null
    resetPasswordValue.value = ''
    showUserTip(`已重置「${row.nickname || row.username}」的密码，其登录态已失效`)
  }
  catch (e) {
    resetError.value = e instanceof Error ? e.message : '重置失败'
  }
  resetBusy.value = false
}

// 弹窗期间锁定页面滚动 + 侧滑返回关闭
useBodyScrollLock([
  () => deleteTagTarget.value !== null,
  () => userConfirm.value !== null,
  () => resetTarget.value !== null,
])
useModalBackClose(() => resetTarget.value !== null, cancelResetPassword)
useModalBackClose(() => userConfirm.value !== null, cancelUserConfirm)
useModalBackClose(() => deleteTagTarget.value !== null, cancelDeleteTag)

// ============ 页签 ============
// 标签管理 / 用户管理仅管理员可见；通用设置、关于所有人可见
type SettingsTab = 'tags' | 'users' | 'general' | 'about'
const activeTab = ref<SettingsTab>('general')
const settingsTabs = computed<Array<{ key: SettingsTab, label: string }>>(() => {
  if (isAdmin.value) {
    return [
      { key: 'tags', label: '标签管理' },
      { key: 'users', label: '用户管理' },
      { key: 'general', label: '通用设置' },
      { key: 'about', label: '关于' },
    ]
  }
  return [
    { key: 'general', label: '通用设置' },
    { key: 'about', label: '关于' },
  ]
})

onMounted(() => {
  load()
  // 标签/用户管理仅管理员可用，非管理员不请求，避免无谓的 403
  if (isAdmin.value) {
    loadTags()
    loadUsers()
  }
  else {
    tagsLoading.value = false
  }
})

// 登录态异步就绪：管理员落在标签管理并补拉管理数据（含在本页完成登录的情形）
watch(isAdmin, (admin) => {
  activeTab.value = admin ? 'tags' : 'general'
  if (admin && !userRows.value.length && !usersLoading.value) {
    loadTags()
    loadUsers()
  }
}, { immediate: true })
</script>

<template>
  <main class="page page--narrow">
    <div class="page-header">
      <h1 class="page-title">设置</h1>
      <NuxtLink to="/" class="text-btn text-btn--accent">← 返回首页</NuxtLink>
    </div>

    <p v-if="error" class="error-alert mt-4">{{ error }}</p>

    <!-- 非管理员：说明可见范围（超管入口不展示，避免误操作） -->
    <p v-if="!isAdmin" class="form-hint mt-3">
      <template v-if="isLoggedIn">标签管理与用户管理仅管理员可见；通用设置为只读。</template>
      <template v-else>
        未登录只能查看，页面上的编辑入口已隐藏。
        <NuxtLink to="/login" class="text-btn text-btn--accent">登录</NuxtLink>
        后可修改设置，管理员还能在此管理标签与用户。
      </template>
    </p>

    <!-- 页签：标签管理 / 用户管理（管理员） + 通用设置 + 关于 -->
    <div class="settings-tabs">
      <button
        v-for="t in settingsTabs"
        :key="t.key"
        type="button"
        class="settings-tab"
        :class="{ 'settings-tab--active': activeTab === t.key }"
        @click="activeTab = t.key"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- 标签管理 -->
    <section v-show="activeTab === 'tags'" class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="section-title">标签管理</h2>
        <span v-if="tagSavedTip" class="saved-tip">{{ tagSavedTip }}</span>
      </div>
      <p class="form-hint mt-1">新增标签会排在末尾；拖动左侧 ⠿ 可调整顺序（松手即保存，首页标签栏同步）；删除标签会把它从使用中的菜谱上移除，菜谱本身保留。</p>

      <p v-if="tagError" class="error-alert mt-2">{{ tagError }}</p>

      <div v-if="tagsLoading" class="empty-note mt-3 text-sm">加载中…</div>
      <ul v-else class="settings-list">
        <li v-if="!tagRows.length" class="empty-note px-3 py-4 text-sm">暂无标签</li>
        <li
          v-for="tag in tagRows"
          :key="tag.id"
          class="settings-list__item"
          :class="{
            'settings-list__item--dragging': dragTagId === tag.id,
            'settings-list__item--drop': dragOverTagId === tag.id,
            'settings-list__item--busy': savingOrder,
          }"
          :draggable="!savingOrder && renamingId === null"
          @dragstart="onTagDragStart($event, tag)"
          @dragover="onTagDragOver($event, tag)"
          @drop.prevent="onTagDrop(tag)"
          @dragend="clearTagDrag"
        >
          <span v-if="renamingId !== tag.id" class="settings-list__handle" title="拖动调整顺序">⠿</span>
          <template v-if="renamingId === tag.id">
            <input
              v-model="renameValue"
              type="text"
              maxlength="20"
              class="input input--sm min-w-0 flex-1"
              @keyup.enter="confirmRename"
              @keyup.esc="cancelRename"
            >
            <button
              type="button"
              class="btn btn--primary btn--xs"
              :disabled="renamingBusy || !renameValue.trim()"
              @click="confirmRename"
            >{{ renamingBusy ? '保存中…' : '保存' }}</button>
            <button type="button" class="btn btn--outline btn--xs" @click="cancelRename">取消</button>
          </template>
          <template v-else>
            <span class="settings-list__name">{{ tag.name }}</span>
            <span class="settings-list__count">{{ tag.recipeCount }} 道菜谱</span>
            <button type="button" class="text-btn text-btn--edit text-btn--xs" @click="startRename(tag)">重命名</button>
            <button
              type="button"
              class="text-btn text-btn--delete text-btn--xs"
              @click="askDeleteTag(tag)"
            >删除</button>
          </template>
        </li>
      </ul>

      <div class="mt-3 flex gap-2">
        <input
          v-model="newTagName"
          type="text"
          maxlength="20"
          placeholder="新标签名（最长 20 字）"
          class="input w-56"
          @keyup.enter="addTag"
        >
        <button
          type="button"
          class="btn btn--primary"
          :disabled="addingTag || !newTagName.trim()"
          @click="addTag"
        >
          {{ addingTag ? '添加中…' : '添加标签' }}
        </button>
      </div>
    </section>

    <!-- 用户管理（仅管理员） -->
    <section v-if="isAdmin && activeTab === 'users'" class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="section-title">用户管理</h2>
        <span v-if="userSavedTip" class="saved-tip">{{ userSavedTip }}</span>
      </div>
      <p class="form-hint mt-1">
        管理员可管理标签、删除菜谱、管理用户；普通用户可正常新建与编辑菜谱。禁用账号会立即失效其登录态，系统至少保留一名可用管理员。
      </p>

      <p v-if="userError" class="error-alert mt-2">{{ userError }}</p>

      <div v-if="usersLoading" class="empty-note mt-3 text-sm">加载中…</div>
      <ul v-else class="settings-list">
        <li v-if="!userRows.length" class="empty-note px-3 py-4 text-sm">暂无用户</li>
        <li v-for="row in userRows" :key="row.id" class="user-row">
          <div class="user-row__main">
            <span class="user-row__avatar">{{ (row.nickname || row.username).slice(0, 1).toUpperCase() }}</span>
            <div class="min-w-0 flex-1">
              <template v-if="renameUserId === row.id">
                <div class="flex items-center gap-2">
                  <input
                    v-model="userNicknameValue"
                    type="text"
                    maxlength="20"
                    class="input input--sm min-w-0 flex-1"
                    @keyup.enter="confirmRenameUser(row)"
                    @keyup.esc="cancelRenameUser"
                  >
                  <button
                    type="button"
                    class="btn btn--primary btn--xs"
                    :disabled="userNicknameBusy"
                    @click="confirmRenameUser(row)"
                  >{{ userNicknameBusy ? '保存中…' : '保存' }}</button>
                  <button type="button" class="btn btn--outline btn--xs" @click="cancelRenameUser">取消</button>
                </div>
              </template>
              <template v-else>
                <div class="flex flex-wrap items-center gap-1.5">
                  <span class="user-row__name">{{ row.nickname || row.username }}</span>
                  <span v-if="isSelf(row)" class="tag-badge">当前账号</span>
                  <span v-if="row.isAdmin" class="tag-badge tag-badge--green">管理员</span>
                  <span v-if="row.status === 0" class="tag-badge tag-badge--amber">已禁用</span>
                </div>
                <p class="user-row__meta">
                  {{ row.username }} · {{ row.recipeCount }} 道菜谱 · 注册 {{ formatDate(row.createdAt) }} · 最近登录 {{ formatDate(row.lastLoginAt) }}
                </p>
              </template>
            </div>
          </div>

          <div class="user-row__actions">
            <button
              type="button"
              class="text-btn text-btn--edit text-btn--xs"
              :disabled="userBusyId === row.id"
              @click="startRenameUser(row)"
            >改昵称</button>
            <button
              type="button"
              class="text-btn text-btn--edit text-btn--xs"
              :disabled="userBusyId === row.id || isSelf(row)"
              @click="toggleAdmin(row)"
            >{{ row.isAdmin ? '取消管理员' : '设为管理员' }}</button>
            <button
              type="button"
              class="text-btn text-btn--edit text-btn--xs"
              :disabled="userBusyId === row.id || isSelf(row)"
              @click="toggleStatus(row)"
            >{{ row.status === 1 ? '禁用' : '启用' }}</button>
            <button
              type="button"
              class="text-btn text-btn--edit text-btn--xs"
              :disabled="resetBusy"
              @click="askResetPassword(row)"
            >重置密码</button>
          </div>
        </li>
      </ul>
    </section>

    <!-- 通用设置 -->
    <div v-show="activeTab === 'general'">
      <div v-if="loading" class="empty-note mt-8 text-sm">加载中…</div>

      <template v-else>
        <div v-if="!entries.length" class="settings-empty">
          暂无设置项
        </div>

        <div v-for="item in entries" :key="item.key" class="mt-6">
          <div class="flex items-center justify-between">
            <label :for="`setting-${item.key}`" class="section-title">{{ item.key }}</label>
            <div class="flex items-center gap-3">
              <span v-if="savedTip" class="saved-tip">{{ savedTip }}</span>
              <button
                v-if="item.key === KEY_IMAGE_GEN_PROMPT && canEdit"
                type="button"
                class="text-btn text-btn--edit text-btn--xs"
                @click="fillExample(item)"
              >
                填入推荐模板
              </button>
            </div>
          </div>
          <p v-if="hintText(item.key)" class="form-hint mt-1">{{ hintText(item.key) }}</p>
          <textarea
            :id="`setting-${item.key}`"
            v-model="item.value"
            rows="16"
            class="input mt-2 w-full p-3 font-mono leading-relaxed"
            :readonly="!canEdit"
            :placeholder="item.key === KEY_IMAGE_GEN_PROMPT ? '输入生图提示词模板，可用 {recipe.title} 等变量…' : '设置值'"
          />
        </div>

        <div v-if="entries.length && canEdit" class="mt-6 flex justify-end">
          <button
            type="button"
            class="btn btn--primary"
            :disabled="saving"
            @click="save"
          >
            {{ saving ? '保存中…' : '保存' }}
          </button>
        </div>

        <!-- 模板变量说明 -->
        <section class="settings-vars">
          <h2 class="section-title">提示词模板变量</h2>
          <p class="form-hint mt-1">模板中以 <code class="inline-code">{recipe.xxx}</code> 形式引用菜谱字段，复制提示词时自动替换为对应菜谱的内容。</p>
          <ul class="settings-vars__list">
            <li v-for="v in templateVars" :key="v.name" class="settings-vars__row">
              <code class="settings-vars__name">{{ v.name }}</code>
              <span class="settings-vars__desc">{{ v.desc }}</span>
            </li>
          </ul>
        </section>
      </template>
    </div>

    <!-- 关于：项目名片 / 主要功能 / 技术栈 / 相关链接（含 GitHub 地址） -->
    <AboutPanel v-show="activeTab === 'about'" class="mt-6" />
  </main>

  <!-- 删除标签确认弹窗 -->
  <Teleport to="body">
    <div
      v-if="deleteTagTarget"
      class="modal-overlay modal-overlay--confirm"
      @click.self="cancelDeleteTag"
    >
      <div class="modal-panel max-w-sm p-6">
        <h3 class="modal-title--sm">删除标签</h3>
        <p class="delete-tag-modal__text">
          确定要删除标签
          <span class="delete-tag-modal__highlight">「{{ deleteTagTarget.name }}」</span>？
        </p>
        <p
          v-if="deleteTagTarget.recipeCount > 0"
          class="warn-note mt-2"
        >
          该标签正被 {{ deleteTagTarget.recipeCount }} 道菜谱使用，删除后将从这些菜谱中移除该标签，菜谱本身不受影响。
        </p>
        <p v-if="deleteTagError" class="error-text mt-2">{{ deleteTagError }}</p>
        <div class="delete-tag-modal__actions">
          <button
            type="button"
            class="btn btn--outline"
            :disabled="deletingTag"
            @click="cancelDeleteTag"
          >
            取消
          </button>
          <button
            type="button"
            class="btn btn--danger"
            :disabled="deletingTag"
            @click="confirmDeleteTag"
          >
            {{ deletingTag ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 用户危险操作二次确认（禁用 / 取消管理员） -->
  <Teleport to="body">
    <div
      v-if="userConfirm"
      class="modal-overlay modal-overlay--confirm"
      @click.self="cancelUserConfirm"
    >
      <div class="modal-panel max-w-sm p-6">
        <h3 class="modal-title--sm">
          {{ userConfirm.kind === 'disable' ? '禁用账号' : '取消管理员' }}
        </h3>
        <p class="delete-tag-modal__text">
          确定要对
          <span class="delete-tag-modal__highlight">「{{ userConfirm.row.nickname || userConfirm.row.username }}」</span>
          {{ userConfirm.kind === 'disable' ? '执行禁用吗？' : '取消管理员权限吗？' }}
        </p>
        <p class="warn-note mt-2">
          {{ userConfirm.kind === 'disable'
            ? '禁用后该账号将无法登录，已有登录态立即失效；其创建的菜谱不受影响。'
            : '取消后该账号将无法再管理标签、删除菜谱与管理用户。' }}
        </p>
        <p v-if="userConfirmError" class="error-text mt-2">{{ userConfirmError }}</p>
        <div class="delete-tag-modal__actions">
          <button
            type="button"
            class="btn btn--outline"
            :disabled="userConfirmBusy"
            @click="cancelUserConfirm"
          >
            取消
          </button>
          <button
            type="button"
            class="btn btn--danger"
            :disabled="userConfirmBusy"
            @click="confirmUserAction"
          >
            {{ userConfirmBusy ? '处理中…' : '确定' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 重置密码弹窗 -->
  <Teleport to="body">
    <div
      v-if="resetTarget"
      class="modal-overlay modal-overlay--confirm"
      @click.self="cancelResetPassword"
    >
      <div class="modal-panel max-w-sm p-6">
        <h3 class="modal-title--sm">重置密码</h3>
        <p class="delete-tag-modal__text">
          为
          <span class="delete-tag-modal__highlight">「{{ resetTarget.nickname || resetTarget.username }}」</span>
          设置新密码，重置后其登录态全部失效，需用新密码重新登录。
        </p>
        <input
          v-model="resetPasswordValue"
          type="password"
          maxlength="32"
          placeholder="新密码（至少 6 位）"
          class="input mt-3 w-full"
          @keyup.enter="confirmResetPassword"
        >
        <p v-if="resetError" class="error-text mt-2">{{ resetError }}</p>
        <div class="delete-tag-modal__actions">
          <button
            type="button"
            class="btn btn--outline"
            :disabled="resetBusy"
            @click="cancelResetPassword"
          >
            取消
          </button>
          <button
            type="button"
            class="btn btn--primary"
            :disabled="resetBusy || resetPasswordValue.length < 6"
            @click="confirmResetPassword"
          >
            {{ resetBusy ? '重置中…' : '确认重置' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 页签 ---- */
.settings-tabs { @apply mt-6 flex gap-6 border-b border-zinc-200 dark:border-zinc-800; }
.settings-tab { @apply -mb-px border-b-2 border-transparent px-1 pb-2 text-sm text-zinc-500 transition-colors hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200; }
.settings-tab--active { @apply border-green-600 font-medium text-green-700 hover:text-green-700 dark:border-green-500 dark:text-green-400 dark:hover:text-green-400; }

/* ---- 标签管理 ---- */
.saved-tip { @apply text-xs text-green-600 dark:text-green-400; }
.settings-list { @apply mt-3 divide-y divide-zinc-100 rounded-md border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900; }
.settings-list__item { @apply flex items-center gap-3 px-3 py-2; }
.settings-list__name { @apply min-w-0 flex-1 truncate text-sm text-zinc-800 dark:text-zinc-200; }
.settings-list__count { @apply shrink-0 text-xs text-zinc-400 dark:text-zinc-500; }

/* ---- 标签拖拽排序 ---- */
.settings-list__handle { @apply shrink-0 cursor-grab select-none text-sm leading-none text-zinc-300 transition-colors hover:text-zinc-500 active:cursor-grabbing dark:text-zinc-600 dark:hover:text-zinc-400; }
.settings-list__item--dragging { @apply opacity-40; }
.settings-list__item--drop { @apply bg-green-50 ring-1 ring-inset ring-green-400 dark:bg-green-500/10 dark:ring-green-500; }
.settings-list__item--busy { @apply pointer-events-none opacity-70; }

/* ---- 用户管理 ---- */
.user-row { @apply flex flex-col gap-2 px-3 py-3 sm:flex-row sm:items-center sm:gap-4; }
.user-row__main { @apply flex min-w-0 flex-1 items-center gap-2.5; }
.user-row__avatar { @apply flex size-8 shrink-0 items-center justify-center rounded-full bg-zinc-200 text-sm font-semibold text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300; }
.user-row__name { @apply truncate text-sm font-medium text-zinc-800 dark:text-zinc-200; }
.user-row__meta { @apply mt-0.5 truncate text-xs text-zinc-400 dark:text-zinc-500; }
.user-row__actions { @apply flex shrink-0 flex-wrap items-center gap-3 sm:justify-end; }

/* ---- 通用设置 ---- */
.settings-empty { @apply mt-8 rounded-md border border-dashed border-zinc-300 p-8 text-center text-sm text-zinc-400 dark:border-zinc-700 dark:text-zinc-500; }
.settings-vars { @apply mt-10 rounded-md bg-zinc-50 p-4 dark:bg-zinc-900; }
.inline-code { @apply rounded bg-zinc-200 px-1 dark:bg-zinc-800; }
.settings-vars__list { @apply mt-3 grid grid-cols-1 gap-x-6 gap-y-1 text-xs sm:grid-cols-2; }
.settings-vars__row { @apply flex justify-between gap-2 border-b border-dashed border-zinc-200 py-1 dark:border-zinc-800; }
.settings-vars__name { @apply shrink-0 text-green-700 dark:text-green-400; }
.settings-vars__desc { @apply text-right text-zinc-500 dark:text-zinc-400; }

/* ---- 确认弹窗 ---- */
.delete-tag-modal__text { @apply mt-3 text-sm text-zinc-600 dark:text-zinc-400; }
.delete-tag-modal__highlight { @apply font-medium text-red-600 dark:text-red-400; }
.delete-tag-modal__actions { @apply mt-5 flex justify-end gap-3; }
</style>
