<script setup lang="ts">
// 首页：标签栏 + 搜索框 + 封面 3:2 卡片网格（无限加载）
import { apis } from '~/api'
import type { Cookbook_internal_model_recipe_list_item, Cookbook_internal_model_tag_item } from '~/api/components'

const searchInput = ref('')
const keywords = ref('')
const activeTagId = ref<number | null>(null) // 单选标签，null=全部
const activeMeal = ref<number | null>(null) // 单选用餐时间，null=全部
const sentinel = ref<Element | null>(null)

// 编辑/新建/删除入口的可见性：未登录只能看（canEdit=false），删除仅管理员（后端口径一致）
const { isAdmin, canEdit } = useAuth()

// 用餐时间单选项（位掩码值与后端 model 常量对齐）
const mealOptions = [
  { label: '早餐', value: 1 },
  { label: '午餐', value: 2 },
  { label: '晚餐', value: 4 },
  { label: '加餐', value: 8 },
]
// 炊具固定选项（前端写死，编辑/新建弹窗多选）
const toolOptions = ['炒锅', '高压锅', '电饭煲', '平底锅', '砂锅', '汤锅', '奶锅']
// 难度单选项（radio 按钮组用；不再提供未填，编辑时 0 按中等处理）
const difficultyOptions = [
  { label: '简单', value: 1 },
  { label: '中等', value: 2 },
  { label: '较难', value: 3 },
]

const { data: tags } = await useAsyncData('tag-list', () => apis.tag.getList(), { default: () => ({ list: [] }), server: false })
const tagList = computed<Cookbook_internal_model_tag_item[]>(() => tags.value?.list ?? [])

const { items, total, status, error, reset } = useInfiniteList<Cookbook_internal_model_recipe_list_item>(
  (page) =>
    apis.recipe.getList({
      params: {
        page,
        pageSize: 12,
        keywords: keywords.value || undefined,
        tagIds: activeTagId.value ? [activeTagId.value] : undefined,
        meal: activeMeal.value ?? undefined,
      },
    }),
  { target: sentinel },
)

// 搜索防抖：输入停顿 300ms 后触发筛选并重新从第 1 页加载
let debounceTimer: ReturnType<typeof setTimeout> | undefined
watch(searchInput, (val) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    const trimmed = val.trim()
    if (trimmed === keywords.value) return // 值未变化（如仅首尾空格变化）不重复请求
    keywords.value = trimmed
    reset()
  }, 300)
})
watch(activeTagId, () => reset())
watch(activeMeal, () => reset())


function selectTag(id: number) {
  activeTagId.value = activeTagId.value === id ? null : id
}

function selectMeal(value: number) {
  activeMeal.value = activeMeal.value === value ? null : value
}

function clearFilters() {
  keywords.value = ''
  searchInput.value = ''
  activeTagId.value = null
  activeMeal.value = null
}

const difficultyText: Record<number, string> = { 1: '简单', 2: '中等', 3: '较难' }

// 位掩码 → 用餐时间文案；0（未设置）视为全部时段
function mealText(mask: number): string {
  if (!mask) return '全部时段'
  return mealOptions.filter(o => (mask & o.value) !== 0).map(o => o.label).join('·')
}

// 编辑弹窗内：切换分类标签选中态
function toggleEditTag(id: number) {
  if (!editForm.value) return
  editForm.value.tagIds = editForm.value.tagIds.includes(id)
    ? editForm.value.tagIds.filter(t => t !== id)
    : [...editForm.value.tagIds, id]
}

// 编辑弹窗内：勾选用餐时段（位或叠加），勾掉则清位
function toggleEditMeal(value: number) {
  if (!editForm.value) return
  editForm.value.mealMask = (editForm.value.mealMask & value) !== 0
    ? editForm.value.mealMask & ~value
    : editForm.value.mealMask | value
}

// 编辑弹窗内：切换工具选中态
function toggleEditTool(tool: string) {
  if (!editForm.value) return
  editForm.value.tools = editForm.value.tools.includes(tool)
    ? editForm.value.tools.filter(t => t !== tool)
    : [...editForm.value.tools, tool]
}

// ============ 右键菜单 ============
const ctxMenu = ref<{ x: number, y: number, item: Cookbook_internal_model_recipe_list_item | null }>({ x: 0, y: 0, item: null })

function openCtxMenu(e: MouseEvent, item: Cookbook_internal_model_recipe_list_item) {
  e.preventDefault()
  ctxMenu.value = { x: e.clientX, y: e.clientY, item }
}

function closeCtxMenu() {
  ctxMenu.value.item = null
}

// ============ 复制生图提示词（后端按设置模板渲染后复制） ============
const copyPromptBusyId = ref<number | null>(null)
const copyPromptTip = ref('')

async function copyGenPrompt(item: Cookbook_internal_model_recipe_list_item) {
  closeCtxMenu()
  copyPromptBusyId.value = item.id
  try {
    const res = await apis.setting.render({ pathParams: { key: 'image_gen_prompt', recipeId: item.id } })
    await navigator.clipboard.writeText(res.rendered ?? '')
    copyPromptTip.value = '生图提示词已复制'
  }
  catch (e) {
    copyPromptTip.value = e instanceof Error ? e.message : '复制失败'
  }
  copyPromptBusyId.value = null
  setTimeout(() => (copyPromptTip.value = ''), 2500)
}

// ============ 详情弹窗（点击卡片） ============
const detailOpen = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref<import('~/api/components').Cookbook_api_recipe_v1_get_one_res | null>(null)
// 封面大图预览（详情封面点击打开）
const coverViewerOpen = ref(false)

async function openDetail(item: Cookbook_internal_model_recipe_list_item) {
  detailError.value = ''
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await apis.recipe.getOne({ pathParams: { id: item.id } })
  }
  catch (e) {
    detailError.value = e instanceof Error ? e.message : '加载失败'
  }
  detailLoading.value = false
}

function closeDetail() {
  detailOpen.value = false
  detail.value = null
  detailError.value = ''
  coverViewerOpen.value = false
}

// ============ 封面大图预览（点击详情弹窗封面） ============
function openCoverViewer() {
  coverViewerOpen.value = true
}

function closeCoverViewer() {
  coverViewerOpen.value = false
}

// 预览期间锁定页面滚动 + 侧滑/ESC 关闭（叠在详情弹窗之上，LIFO 先关预览）
useBodyScrollLock([() => coverViewerOpen.value])
useModalBackClose(() => coverViewerOpen.value, closeCoverViewer)

function onCoverViewerKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && coverViewerOpen.value) closeCoverViewer()
}
onMounted(() => window.addEventListener('keydown', onCoverViewerKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onCoverViewerKeydown))

// ============ 编辑弹窗 ============
interface EditForm {
  id: number
  title: string
  summary: string
  coverAttachmentId: number
  coverUrl: string
  ingredients: { name: string, amount: string, optional: number }[]
  tools: string[]
  steps: { title: string, content: string, media: number[] }[]
  mealMask: number
  difficulty: number // 难度 1=简单 2=中等 3=较难（表单默认中等，不再提供未填）
  cookMinutes: number // 耗时(分钟)，0=未填
  calories: number // 每份热量 kcal，0=未填
  tagIds: number[] // 分类（可多选）
}

const editOpen = ref(false)
const editSaving = ref(false)
const editLoading = ref(false)
const editError = ref('')
const editForm = ref<EditForm | null>(null)
const coverFile = ref<File | null>(null)
const coverPreview = ref('')
const stepUploading = ref<number[]>([]) // 正在上传图片的步骤下标
const dragFromIngredient = ref(-1) // 拖拽高亮：当前拖拽的食材下标
const dragFromStep = ref(-1) // 拖拽高亮：当前拖拽的步骤下标

// ============ 删除确认 ============
const deleteConfirm = ref<{ item: Cookbook_internal_model_recipe_list_item, deleting: boolean, error: string } | null>(null)

function askDelete(item: Cookbook_internal_model_recipe_list_item) {
  closeCtxMenu()
  deleteConfirm.value = { item, deleting: false, error: '' }
}

function cancelDelete() {
  deleteConfirm.value = null
}

async function confirmDelete() {
  if (!deleteConfirm.value) return
  deleteConfirm.value.deleting = true
  try {
    await apis.recipe.delete({ pathParams: { id: deleteConfirm.value.item.id } })
    const removedId = deleteConfirm.value.item.id
    deleteConfirm.value = null
    // 就地移除：不整页重拉，保持滚动位置与浏览状态
    const idx = items.value.findIndex(i => i.id === removedId)
    if (idx >= 0) items.value.splice(idx, 1)
    total.value = Math.max(0, total.value - 1)
  }
  catch (e) {
    deleteConfirm.value = { item: deleteConfirm.value.item, deleting: false, error: e instanceof Error ? e.message : '删除失败' }
  }
}

// 从详情弹窗进入编辑：复用已加载的详情数据，免二次请求
// 新建入口：复用编辑弹窗，空表单 + id=0
function openCreate() {
  editError.value = ''
  detailOpen.value = false
  editOpen.value = true
  editLoading.value = false
  editForm.value = {
    id: 0,
    title: '',
    summary: '',
    coverAttachmentId: 0,
    coverUrl: '',
    ingredients: [{ name: '', amount: '', optional: 0 }],
    tools: [],
    steps: [{ title: '', content: '', media: [] }],
    mealMask: 15, // 默认全时段
    difficulty: 2, // 默认中等
    cookMinutes: 0,
    calories: 0,
    tagIds: [],
  }
  coverFile.value = null
  coverPreview.value = ''
}

function openEditFromDetail() {
  const d = detail.value
  if (!d) return
  editError.value = ''
  detailOpen.value = false
  editOpen.value = true
  editLoading.value = false
  editForm.value = {
    id: d.id!,
    title: d.title ?? '',
    summary: d.summary ?? '',
    coverAttachmentId: d.coverAttachmentId ?? 0,
    coverUrl: d.coverUrl ?? '',
    ingredients: (d.ingredients ?? []).map(i => ({ ...i })),
    tools: [...(d.tools ?? [])],
    steps: (d.steps ?? []).map(s => ({ title: s.title ?? '', content: s.content ?? '', media: [...(s.media ?? [])] })),
    mealMask: (d.mealMask ?? 0) || 15, // 未填(0)按全时段处理
    difficulty: (d.difficulty ?? 0) || 2, // 未填按中等处理
    cookMinutes: d.cookMinutes ?? 0,
    calories: d.calories ?? 0,
    tagIds: (d.tags ?? []).map(t => t.id!),
  }
  coverPreview.value = d.coverUrl ?? ''
}

async function openEdit(item: Cookbook_internal_model_recipe_list_item) {
  closeCtxMenu()
  editError.value = ''
  editOpen.value = true
  editLoading.value = true
  editForm.value = null
  try {
    const detail = await apis.recipe.getOne({ pathParams: { id: item.id } })
    editForm.value = {
      id: detail.id!,
      title: detail.title ?? '',
      summary: detail.summary ?? '',
      coverAttachmentId: detail.coverAttachmentId ?? 0,
      coverUrl: detail.coverUrl ?? '',
      ingredients: (detail.ingredients ?? []).map(i => ({ ...i })),
      tools: [...(detail.tools ?? [])],
      steps: (detail.steps ?? []).map(s => ({ title: s.title ?? '', content: s.content ?? '', media: [...(s.media ?? [])] })),
      mealMask: (detail.mealMask ?? 0) || 15, // 未填(0)按全时段处理
      difficulty: (detail.difficulty ?? 0) || 2, // 未填按中等处理
      cookMinutes: detail.cookMinutes ?? 0,
      calories: detail.calories ?? 0,
      tagIds: (detail.tags ?? []).map(t => t.id!),
    }
    coverPreview.value = detail.coverUrl ?? ''
  }
  catch (e) {
    editError.value = e instanceof Error ? e.message : '加载失败'
  }
  editLoading.value = false
}

function closeEdit() {
  editOpen.value = false
  editForm.value = null
  coverFile.value = null
  coverPreview.value = ''
  stepUploading.value = []
  editError.value = ''
}

function addStep() {
  editForm.value?.steps.push({ title: '', content: '', media: [] })
}

function removeStep(idx: number) {
  editForm.value?.steps.splice(idx, 1)
}

// 拖拽排序：食材 / 步骤（HTML5 原生拖放）
// 把 from 位置的元素移动到 to 位置（to 可为数组长度，表示排到末尾）
function moveItem<T>(list: T[] | undefined, from: number, to: number) {
  if (!list || from === to || from < 0 || from >= list.length) return
  const t = Math.min(Math.max(to, 0), list.length - 1)
  const [item] = list.splice(from, 1)
  list.splice(t, 0, item)
}

function onDragStart(ev: DragEvent, idx: number, type: 'ingredient' | 'step') {
  if (!ev.dataTransfer) return
  ev.dataTransfer.effectAllowed = 'move'
  ev.dataTransfer.setData('text/plain', `${type}:${idx}`)
  if (type === 'ingredient') dragFromIngredient.value = idx
  else dragFromStep.value = idx
}

function onDragOver(ev: DragEvent, type: 'ingredient' | 'step') {
  ev.preventDefault()
  if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'move'
}

function onDrop(ev: DragEvent, to: number, type: 'ingredient' | 'step') {
  ev.preventDefault()
  const data = ev.dataTransfer?.getData('text/plain')
  if (!data) return
  const [draggedType, fromStr] = data.split(':')
  const from = Number(fromStr)
  if (draggedType !== type || Number.isNaN(from)) return
  if (type === 'ingredient') moveItem(editForm.value?.ingredients, from, to)
  else moveItem(editForm.value?.steps, from, to)
}

// 步骤图片：上传 / 移除 / 访问 URL（附件内容经 /attachments/{id}/content 直出）
function attachmentUrl(id: number): string {
  return `/attachments/${id}/content`
}

async function uploadStepImage(file: File): Promise<number | null> {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await apis.attachment.upload({ body: fd as never })
    return res?.attachmentId ?? null
  }
  catch (e) {
    editError.value = e instanceof Error ? `步骤图片上传失败：${e.message}` : '步骤图片上传失败'
    return null
  }
}

async function onStepImagesChange(e: Event, idx: number) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = '' // 重置以便重复选同名文件
  if (!files.length) return
  const step = editForm.value?.steps[idx]
  if (!step) return
  stepUploading.value = [...stepUploading.value, idx]
  for (const file of files) {
    const id = await uploadStepImage(file)
    if (id) step.media.push(id)
  }
  stepUploading.value = stepUploading.value.filter(i => i !== idx)
}

function removeStepImage(idx: number, mediaId: number) {
  const step = editForm.value?.steps[idx]
  if (!step) return
  step.media = step.media.filter(id => id !== mediaId)
}

function addIngredient() {
  editForm.value?.ingredients.push({ name: '', amount: '', optional: 0 })
}

function removeIngredient(idx: number) {
  editForm.value?.ingredients.splice(idx, 1)
}

function onCoverChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  coverFile.value = file
  coverPreview.value = URL.createObjectURL(file)
}

async function uploadCover(): Promise<number | null> {
  if (!coverFile.value) return null
  const fd = new FormData()
  fd.append('file', coverFile.value)
  try {
    const res = await apis.attachment.upload({ body: fd as never })
    return res?.attachmentId ?? null
  }
  catch (e) {
    editError.value = e instanceof Error ? `封面上传失败：${e.message}` : '封面上传失败'
    return null
  }
}

async function saveEdit() {
  if (!editForm.value) return
  editSaving.value = true
  editError.value = ''
  try {
    let coverId = editForm.value.coverAttachmentId
    if (coverFile.value) {
      const uploaded = await uploadCover()
      if (uploaded) coverId = uploaded
    }
    const body = {
      title: editForm.value.title,
      summary: editForm.value.summary,
      coverAttachmentId: coverId,
      ingredients: editForm.value.ingredients,
      tools: editForm.value.tools,
      tagIds: editForm.value.tagIds,
      steps: editForm.value.steps,
      mealMask: editForm.value.mealMask,
      difficulty: editForm.value.difficulty,
      cookMinutes: editForm.value.cookMinutes,
      calories: editForm.value.calories,
    }
    if (editForm.value.id > 0) {
      await apis.recipe.update({ pathParams: { id: editForm.value.id }, body })
    }
    else {
      await apis.recipe.create({ body })
    }
    editOpen.value = false
    reset() // 重新拉取列表反映变更
  }
  catch (e) {
    editError.value = e instanceof Error ? e.message : '保存失败'
  }
  editSaving.value = false
}

// ============ 收藏夹弹窗（卡片爱心与详情弹窗共用） ============
const favRecipeId = ref<number | null>(null)


function openFavModal(recipeId: number) {
  favRecipeId.value = recipeId
}

function closeFavModal() {
  favRecipeId.value = null
}

// ============ 弹窗滚动锁定 + 移动端返回手势 ============
// 弹窗打开期间锁定页面滚动，滚轮/键盘不再穿透到首页列表（多实例共享全局锁）
useBodyScrollLock([
  () => detailOpen.value,
  () => editOpen.value,
  () => deleteConfirm.value !== null,
  () => ctxMenu.value.item !== null,
  () => favRecipeId.value !== null,
])
// 移动端侧滑 / 浏览器返回：关闭最上层弹窗而非退出页面
useModalBackClose(() => detailOpen.value, closeDetail)
useModalBackClose(() => editOpen.value, closeEdit)
useModalBackClose(() => deleteConfirm.value !== null, cancelDelete)
useModalBackClose(() => ctxMenu.value.item !== null, closeCtxMenu)
</script>

<template>
  <main class="page">
    <div class="page-header">
      <h1 class="page-title">菜谱</h1>
      <!-- 未登录只读：不展示新建入口 -->
      <button v-if="canEdit" type="button" class="btn btn--primary gap-1.5" @click="openCreate">
        <PlusIcon class="size-4" />
        新建菜谱
      </button>
    </div>

    <!-- 筛选区：搜索 + 种类/时段（带分组标签，白底面板归组） -->
    <div class="card mt-4 p-4">
      <input
        v-model="searchInput"
        type="search"
        placeholder="搜索菜谱名称或简介…"
        class="input input--lg w-full"
      >

      <!-- 种类：单选标签，默认"全部" -->
      <div class="filter-group mt-4">
        <span class="filter-group__label">种类</span>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="chip chip--green"
            :class="{ 'chip--active': activeTagId === null }"
            @click="activeTagId = null"
          >
            全部
          </button>
          <button
            v-for="tag in tagList"
            :key="tag.id"
            type="button"
            class="chip chip--green"
            :class="{ 'chip--active': activeTagId === tag.id }"
            @click="selectTag(tag.id!)"
          >
            {{ tag.name }}
          </button>
        </div>
      </div>

      <!-- 时段：单选用餐时间，默认"全部时段" -->
      <div class="filter-group mt-3">
        <span class="filter-group__label">时段</span>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="chip chip--amber"
            :class="{ 'chip--active': activeMeal === null }"
            @click="activeMeal = null"
          >
            全部时段
          </button>
          <button
            v-for="opt in mealOptions"
            :key="opt.value"
            type="button"
            class="chip chip--amber"
            :class="{ 'chip--active': activeMeal === opt.value }"
            @click="selectMeal(opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- 筛选状态行 -->
    <p class="result-meta mt-4">
      共 {{ total }} 道菜谱
      <button
        v-if="keywords || activeTagId !== null || activeMeal !== null"
        type="button"
        class="text-btn text-btn--accent ml-2"
        @click="clearFilters"
      >
        清除筛选
      </button>
    </p>

    <!-- 卡片网格：正常横向封面（3:2） -->
    <div v-if="items.length" class="recipe-grid mt-3">
      <article
        v-for="item in items"
        :key="item.id"
        class="recipe-card"
        @click="openDetail(item)"
        @contextmenu="openCtxMenu($event, item)"
      >
        <div class="recipe-card__cover">
          <img
            v-if="item.coverUrl"
            :src="thumbUrl(item.coverUrl)"
            :alt="item.title"
            class="recipe-card__img"
            loading="lazy"
          >
          <div v-else class="recipe-card__cover-empty">暂无封面</div>
          <span v-if="item.difficulty" class="recipe-card__difficulty">{{ difficultyText[item.difficulty] }}</span>
          <button
            type="button"
            class="icon-btn recipe-card__fav"
            :class="item.favoriteCount ? 'recipe-card__fav--active' : 'recipe-card__fav--idle'"
            :title="item.favoriteCount ? `已收藏 ${item.favoriteCount} 个收藏夹` : '收藏'"
            :aria-label="item.favoriteCount ? `已收藏 ${item.favoriteCount} 个收藏夹` : '收藏'"
            @click.stop="openFavModal(item.id)"
          >
            <!-- 心形：未收藏描边、已收藏实心（fill 动态属性绑定，水合安全） -->
            <svg
              viewBox="0 0 24 24"
              :fill="item.favoriteCount ? 'currentColor' : 'none'"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="size-4"
              aria-hidden="true"
            >
              <path d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12Z" />
            </svg>
            <span v-if="item.favoriteCount" class="tabular-nums">{{ item.favoriteCount }}</span>
          </button>
        </div>
        <div class="recipe-card__body">
          <h2 class="recipe-card__title">{{ item.title }}</h2>
          <p class="recipe-card__summary">{{ item.summary }}</p>
          <div class="recipe-card__tags">
            <span v-for="tag in item.tags" :key="tag.id" class="tag-badge">{{ tag.name }}</span>
          </div>
        </div>
      </article>
    </div>

    <p v-else-if="status === 'done' && !error" class="empty-note mt-8 text-sm">没有找到匹配的菜谱</p>
    <p v-if="error" class="error-text mt-8 text-center">{{ error }}</p>

    <!-- 上拉加载哨兵 -->
    <div ref="sentinel" class="h-4" />
    <p v-if="status === 'loading'" class="empty-note py-4 text-sm">加载中…</p>
    <p v-else-if="status === 'done' && items.length" class="empty-note py-4 text-sm">已全部加载</p>
  </main>

  <!-- 食谱详情弹窗（点击卡片）：封面置顶 + 桌面端食材/步骤双栏 -->
  <Teleport to="body">
    <div v-if="detailOpen" class="modal-overlay" @click.self="closeDetail">
      <div class="detail-modal__panel">
        <!-- 悬浮关闭：固定在面板右上角，不随内容滚动（封面滚走后仍可关闭） -->
        <button
          v-if="!detailLoading"
          type="button"
          class="icon-btn detail-modal__close"
          aria-label="关闭"
          @click="closeDetail"
        >✕</button>

        <!-- 加载态 -->
        <div v-if="detailLoading" class="detail-modal__loading">
          <p class="empty-note text-sm">加载中…</p>
        </div>

        <!-- 错误态 -->
        <template v-else-if="detailError">
          <div class="detail-modal__loading">
            <p class="error-alert">{{ detailError }}</p>
          </div>
        </template>

        <!-- 详情：封面在滚动区内，随内容一起滚动 -->
        <template v-else-if="detail">
          <div class="detail-modal__body">
            <div v-if="detail.coverUrl" class="detail-modal__cover">
              <button
                type="button"
                class="detail-modal__cover-trigger"
                aria-label="查看大图"
                title="点击查看大图"
                @click="openCoverViewer"
              >
                <img
                  :src="detail.coverUrl"
                  :alt="detail.title ?? '封面'"
                  class="detail-modal__cover-img"
                >
              </button>
            </div>

            <div class="detail-modal__content">
              <h3 class="modal-title">{{ detail.title }}</h3>
              <p class="detail-modal__summary">{{ detail.summary }}</p>
            <div class="detail-modal__meta">
              <span v-if="detail.difficulty" class="tag-badge tag-badge--md">难度：{{ difficultyText[detail.difficulty] || '未填' }}</span>
              <span v-if="detail.cookMinutes" class="tag-badge tag-badge--md">耗时：{{ detail.cookMinutes }} 分钟</span>
              <span v-if="detail.calories" class="tag-badge tag-badge--md">热量：{{ detail.calories }} kcal/份</span>
              <span v-if="detail.servings" class="tag-badge tag-badge--md">{{ detail.servings }} 人份</span>
            </div>
            <div class="detail-modal__tags">
              <span v-if="detail.mealMask" class="tag-badge tag-badge--md tag-badge--amber">{{ mealText(detail.mealMask) }}</span>
              <span v-for="tag in detail.tags" :key="tag.id" class="tag-badge tag-badge--md tag-badge--green">{{ tag.name }}</span>
            </div>

            <!-- 单列上下：工具 → 食材 → 步骤 → 小贴士 -->
            <div class="detail-modal__sections">
              <!-- 工具 -->
              <div v-if="detail.tools?.length">
                <h4 class="section-title">工具</h4>
                <div class="detail-modal__tools">
                  <span v-for="(t, i) in detail.tools" :key="i" class="tag-badge tag-badge--md">{{ t }}</span>
                </div>
              </div>

              <!-- 食材（名称+用量两端对齐） -->
              <div v-if="detail.ingredients?.length">
                <h4 class="section-title">食材</h4>
                <ul class="detail-modal__ingredients">
                  <li v-for="(ing, i) in detail.ingredients" :key="i" class="detail-modal__ingredient">
                    <span class="min-w-0">{{ ing.name }}<span v-if="ing.optional" class="form-hint ml-1">（可选）</span></span>
                    <span class="detail-modal__amount">{{ ing.amount }}</span>
                  </li>
                </ul>
              </div>

              <!-- 步骤 -->
              <div v-if="detail.steps?.length">
                <h4 class="section-title">步骤</h4>
                <ol class="detail-modal__steps">
                  <li v-for="(step, i) in detail.steps" :key="i" class="detail-modal__step">
                    <span class="detail-modal__step-no">{{ i + 1 }}</span>
                    <div class="min-w-0 flex-1">
                      <p v-if="step.title" class="detail-modal__step-title">{{ step.title }}</p>
                      <p class="detail-modal__step-text" :class="step.title ? 'mt-1' : ''">{{ step.content }}</p>
                      <div v-if="step.media?.length" class="detail-modal__step-media">
                        <img
                          v-for="mediaId in step.media"
                          :key="mediaId"
                          :src="attachmentUrl(mediaId)"
                          :alt="`步骤图 ${mediaId}`"
                          class="detail-modal__step-img"
                          loading="lazy"
                        >
                      </div>
                    </div>
                  </li>
                </ol>
              </div>

              <!-- 注意事项 -->
              <div v-if="detail.tips" class="detail-modal__tips">
                <h4 class="detail-modal__tips-title">小贴士</h4>
                <p class="detail-modal__tips-text">{{ detail.tips }}</p>
              </div>
            </div>
            </div>
          </div>

          <!-- 操作：收藏（收藏夹弹窗）+ 编辑；未登录只读，改为登录引导 -->
          <div class="detail-modal__footer">
            <template v-if="canEdit">
              <button type="button" class="btn btn--outline" @click="openFavModal(detail.id)">
                收藏
              </button>
              <button type="button" class="btn btn--primary" @click="openEditFromDetail">
                编辑
              </button>
            </template>
            <NuxtLink v-else to="/login" class="text-btn text-btn--accent text-sm">
              登录后可编辑 / 收藏 →
            </NuxtLink>
          </div>
        </template>
      </div>
    </div>
  </Teleport>

  <!-- 封面大图预览：全屏灯箱，点击任意处 / ✕ / ESC / 侧滑关闭 -->
  <Teleport to="body">
    <div v-if="coverViewerOpen" class="modal-overlay image-viewer" @click="closeCoverViewer">
      <button type="button" class="image-viewer__close" aria-label="关闭预览" @click.stop="closeCoverViewer">✕</button>
      <img
        v-if="detail?.coverUrl"
        :src="detail.coverUrl"
        :alt="detail.title ?? '封面'"
        class="image-viewer__img"
      >
      <p v-if="detail?.title" class="image-viewer__caption">{{ detail.title }}</p>
    </div>
  </Teleport>

  <!-- 右键菜单：复制提示词（只读，人人可用）+ 编辑（登录后）+ 删除（仅管理员） -->
  <Teleport to="body">
    <div v-if="ctxMenu.item" class="fixed inset-0 z-50" @click="closeCtxMenu" @contextmenu.prevent="closeCtxMenu">
      <div class="ctx-menu" :style="{ left: `${ctxMenu.x}px`, top: `${ctxMenu.y}px` }">
        <button
          type="button"
          class="ctx-menu__item"
          :disabled="copyPromptBusyId !== null"
          @click.stop="copyGenPrompt(ctxMenu.item!)"
        >
          {{ copyPromptBusyId === ctxMenu.item!.id ? '渲染中…' : '复制生图提示词' }}
        </button>
        <button v-if="canEdit" type="button" class="ctx-menu__item" @click.stop="openEdit(ctxMenu.item!)">
          编辑
        </button>
        <button
          v-if="isAdmin"
          type="button"
          class="ctx-menu__item ctx-menu__item--danger"
          @click.stop="askDelete(ctxMenu.item!)"
        >
          删除
        </button>
      </div>
    </div>
  </Teleport>

  <!-- 删除确认弹窗 -->
  <Teleport to="body">
    <div v-if="deleteConfirm" class="modal-overlay modal-overlay--confirm" @click.self="cancelDelete">
      <div class="modal-panel max-w-sm p-6">
        <p class="confirm-modal__text">
          确定要删除食谱
          <span class="confirm-modal__highlight">「{{ deleteConfirm.item.title }}」</span>
        </p>
        <p v-if="deleteConfirm.error" class="error-text mt-2">{{ deleteConfirm.error }}</p>
        <div class="confirm-modal__actions">
          <button type="button" class="btn btn--outline" :disabled="deleteConfirm.deleting" @click="cancelDelete">
            取消
          </button>
          <button type="button" class="btn btn--danger" :disabled="deleteConfirm.deleting" @click="confirmDelete">
            {{ deleteConfirm.deleting ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 编辑弹窗 -->
  <Teleport to="body">
    <div v-if="editOpen" class="modal-overlay" @click.self="closeEdit">
      <div class="edit-modal__panel">
        <div class="edit-modal__header">
          <h3 class="modal-title">{{ (editForm?.id ?? 0) > 0 ? '编辑食谱' : '新建食谱' }}</h3>
          <button type="button" class="modal-close" @click="closeEdit">✕</button>
        </div>

        <p v-if="editLoading" class="empty-note py-10 text-sm">加载中…</p>
        <template v-else-if="editForm">
          <p v-if="editError" class="error-alert mt-3">{{ editError }}</p>

          <div class="edit-modal__columns">
            <!-- 左列：基础信息（桌面端独立滚动，不随右列联动） -->
            <div class="edit-modal__col--info">
              <!-- 封面图（左列首行） -->
              <div>
                <label class="form-label">封面图</label>
                <div class="edit-modal__cover-row">
                  <img v-if="coverPreview" :src="coverPreview" alt="封面预览" class="edit-modal__cover-preview">
                  <div v-else class="edit-modal__cover-empty">无封面</div>
                  <label class="btn btn--outline cursor-pointer">
                    选择文件
                    <input type="file" accept="image/*" class="hidden" @change="onCoverChange">
                  </label>
                </div>
              </div>

              <!-- 名称 + 难度（同一行） -->
              <div class="flex items-end gap-3">
                <div class="min-w-0 flex-1">
                  <label class="form-label">菜谱名称</label>
                  <input v-model="editForm.title" class="input w-full resize-y">
                </div>
                <div class="shrink-0">
                  <label class="form-label">难度</label>
                  <div class="edit-modal__radio" role="radiogroup" aria-label="难度">
                    <button
                      v-for="opt in difficultyOptions"
                      :key="opt.value"
                      type="button"
                      role="radio"
                      :aria-checked="editForm.difficulty === opt.value"
                      class="edit-modal__radio-btn"
                      :class="{ 'edit-modal__radio-btn--active': editForm.difficulty === opt.value }"
                      @click="editForm.difficulty = opt.value"
                    >
                      {{ opt.label }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- 菜谱描述（数据表 summary 字段） -->
              <div>
                <label class="form-label">菜谱描述</label>
                <textarea
                  v-model="editForm.summary"
                  rows="2"
                  maxlength="200"
                  placeholder="一句话介绍这道菜…"
                  class="input w-full resize-y"
                />
              </div>

              <!-- 分类（多选标签） -->
              <div>
                <label class="form-label">分类</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="tag in tagList"
                    :key="tag.id"
                    type="button"
                    class="chip chip--green"
                    :class="{ 'chip--active': editForm.tagIds.includes(tag.id!) }"
                    @click="toggleEditTag(tag.id!)"
                  >
                    {{ tag.name }}
                  </button>
                </div>
              </div>

              <!-- 用餐时段（多选） -->
              <div>
                <label class="form-label">适合时段（可多选）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="opt in mealOptions"
                    :key="opt.value"
                    type="button"
                    class="chip chip--amber"
                    :class="{ 'chip--active': (editForm.mealMask & opt.value) !== 0 }"
                    @click="toggleEditMeal(opt.value)"
                  >
                    {{ opt.label }}
                  </button>
                </div>
              </div>

              <!-- 炊具（多选，固定选项） -->
              <div>
                <label class="form-label">所需工具（可多选）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="tool in toolOptions"
                    :key="tool"
                    type="button"
                    class="chip chip--sky"
                    :class="{ 'chip--active': editForm.tools.includes(tool) }"
                    @click="toggleEditTool(tool)"
                  >
                    {{ tool }}
                  </button>
                </div>
              </div>

              <!-- 耗时 / 热量 -->
              <div class="edit-modal__group-row">
                <div>
                  <label class="form-label">耗时(分)</label>
                  <input v-model.number="editForm.cookMinutes" type="number" min="0" class="input w-full">
                </div>
                <div>
                  <label class="form-label">热量(kcal)</label>
                  <input v-model.number="editForm.calories" type="number" min="0" class="input w-full">
                </div>
              </div>
            </div>

            <!-- 右列：食材 + 步骤（桌面端独立滚动） -->
            <div class="edit-modal__col--steps">
              <!-- 食材（动态表单） -->
              <div>
                <div class="mb-1 flex items-center justify-between">
                  <label class="text-sm font-medium">食材</label>
                  <button
                    type="button"
                    class="text-btn text-btn--accent inline-flex items-center gap-1"
                    @click="addIngredient"
                  >
                    <PlusIcon class="size-3.5" />
                    添加食材
                  </button>
                </div>
                <div
                  v-for="(ing, idx) in editForm.ingredients"
                  :key="idx"
                  class="edit-modal__ingredient"
                  draggable="true"
                  :class="idx === dragFromIngredient ? 'opacity-50' : ''"
                  @dragstart="onDragStart($event, idx, 'ingredient')"
                  @dragover="onDragOver($event, 'ingredient')"
                  @drop="onDrop($event, idx, 'ingredient')"
                  @dragend="dragFromIngredient = -1"
                >
                  <span class="drag-handle" title="拖拽排序">⠿</span>
                  <input v-model="ing.name" type="text" placeholder="食材名" class="input input--sm w-40">
                  <input v-model="ing.amount" type="text" placeholder="用量（如 500g）" class="input input--sm w-32">
                  <button type="button" class="text-btn text-btn--danger" @click="removeIngredient(idx)">删除</button>
                </div>
              </div>

              <!-- 步骤（可多图） -->
              <div>
                <div class="mb-1 flex items-center justify-between">
                  <label class="text-sm font-medium">步骤</label>
                  <button
                    type="button"
                    class="text-btn text-btn--accent inline-flex items-center gap-1"
                    @click="addStep"
                  >
                    <PlusIcon class="size-3.5" />
                    添加步骤
                  </button>
                </div>

                <div
                  v-for="(step, idx) in editForm.steps"
                  :key="idx"
                  class="edit-modal__step"
                  draggable="true"
                  :class="idx === dragFromStep ? 'opacity-50' : ''"
                  @dragstart="onDragStart($event, idx, 'step')"
                  @dragover="onDragOver($event, 'step')"
                  @drop="onDrop($event, idx, 'step')"
                  @dragend="dragFromStep = -1"
                >
                  <div class="edit-modal__step-head">
                    <span class="edit-modal__step-label">
                      <span class="drag-handle" title="拖拽排序">⠿</span>
                      第 {{ idx + 1 }} 步
                    </span>
                    <button type="button" class="text-btn text-btn--danger text-btn--xs" @click="removeStep(idx)">删除步骤</button>
                  </div>

                  <!-- 步骤标题（可选） -->
                  <input
                    v-model="step.title"
                    type="text"
                    placeholder="步骤标题（可选），如 焯水去腥"
                    class="input input--sm mb-2 w-full"
                  >

                  <!-- 步骤内容 -->
                  <textarea
                    v-model="step.content"
                    rows="3"
                    placeholder="步骤内容"
                    class="input input--sm w-full resize-y"
                  />

                  <!-- 步骤图片：已有缩略图 + 添加 -->
                  <div class="edit-modal__media">
                    <div v-for="mediaId in step.media" :key="mediaId" class="edit-modal__thumb group">
                      <img
                        :src="attachmentUrl(mediaId)"
                        :alt="`步骤图 ${mediaId}`"
                        class="edit-modal__thumb-img"
                        loading="lazy"
                      >
                      <button
                        type="button"
                        class="icon-btn edit-modal__thumb-remove"
                        @click="removeStepImage(idx, mediaId)"
                      >×</button>
                    </div>
                    <label
                      class="edit-modal__media-add"
                      :class="stepUploading.includes(idx) ? 'pointer-events-none opacity-60' : ''"
                    >
                      <span v-if="stepUploading.includes(idx)" class="text-xs">上传中…</span>
                      <PlusIcon v-else class="size-5" />
                      <input
                        type="file"
                        accept="image/*"
                        multiple
                        class="hidden"
                        @change="onStepImagesChange($event, idx)"
                      >
                    </label>
                  </div>
                </div>

                <button type="button" class="edit-modal__add-step" @click="addStep">
                  <PlusIcon class="size-3.5" />
                  添加步骤
                </button>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="edit-modal__footer">
            <button type="button" class="btn btn--outline" :disabled="editSaving" @click="closeEdit">
              取消
            </button>
            <button type="button" class="btn btn--primary" :disabled="editSaving" @click="saveEdit">
              {{ editSaving ? '保存中…' : '保存' }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </Teleport>

  <!-- 右下角快捷导航：悬浮球 + 展开菜单 -->
  <QuickNav />

  <!-- 复制生图提示词结果提示 -->
  <Teleport to="body">
    <div v-if="copyPromptTip" class="toast">
      {{ copyPromptTip }}
    </div>
  </Teleport>

  <!-- 收藏夹弹窗（卡片爱心 / 详情弹窗收藏按钮共用） -->
  <FavoriteFolderModal :open="favRecipeId !== null" :recipe-id="favRecipeId" @close="closeFavModal" />
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 筛选区 ---- */
.filter-group { @apply flex items-start gap-3; }
.filter-group__label { @apply mt-1 w-8 shrink-0 text-sm font-medium text-zinc-500 dark:text-zinc-400; }
.result-meta { @apply text-sm text-zinc-500 dark:text-zinc-400; }

/* ---- 菜谱卡片 ---- */
.recipe-grid { @apply grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4; }
.recipe-card { @apply cursor-pointer overflow-hidden rounded-lg border border-zinc-200 bg-white transition-shadow hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900; }
.recipe-card__cover { @apply relative w-full overflow-hidden; aspect-ratio: 3 / 2; }
.recipe-card__img { @apply absolute inset-0 h-full w-full object-cover object-center; }
.recipe-card__cover-empty { @apply absolute inset-0 flex items-center justify-center bg-zinc-100 text-zinc-400 dark:bg-zinc-800 dark:text-zinc-500; }
.recipe-card__difficulty { @apply absolute left-2 top-2 rounded bg-black/50 px-1.5 py-0.5 text-xs text-white; }
.recipe-card__fav { @apply absolute right-2 top-2 z-10 flex items-center gap-1 rounded-full px-2.5 py-1.5 text-xs text-white shadow-sm backdrop-blur-md transition-all duration-200 hover:scale-105 active:scale-95; }
.recipe-card__fav--active { @apply bg-red-500/90 shadow-red-500/30 hover:bg-red-500; }
.recipe-card__fav--idle { @apply bg-black/40 hover:bg-black/60; }
.recipe-card__body { @apply p-3; }
.recipe-card__title { @apply truncate text-sm font-semibold; }
.recipe-card__summary { @apply mt-1 line-clamp-2 min-h-8 text-xs text-zinc-500 dark:text-zinc-400; }
.recipe-card__tags { @apply mt-2 flex flex-wrap gap-1; }

/* ---- 详情弹窗 ---- */
.detail-modal__panel { @apply relative flex h-[90vh] w-full max-w-3xl flex-col overflow-hidden rounded-lg bg-white shadow-lg dark:bg-zinc-900; }
.detail-modal__loading { @apply flex flex-1 items-center justify-center; }
.detail-modal__close { @apply absolute right-3 top-3 z-10 size-8 rounded-full bg-black/45 text-white backdrop-blur-sm hover:bg-black/65; }
.detail-modal__cover { @apply relative; }
.detail-modal__cover-trigger { @apply block w-full cursor-zoom-in transition-opacity hover:opacity-90; }
.detail-modal__cover-img { @apply h-52 w-full object-cover object-center sm:h-56; }

/* ---- 封面大图预览灯箱 ---- */
.image-viewer { @apply z-[80] cursor-zoom-out bg-black/85; }
.image-viewer__close { @apply absolute right-4 top-4 flex size-9 items-center justify-center rounded-full bg-white/10 text-lg text-white backdrop-blur-sm transition-colors hover:bg-white/20; }
.image-viewer__img { @apply max-h-[86vh] max-w-[92vw] object-contain; }
.image-viewer__caption { @apply absolute bottom-6 left-1/2 max-w-[80vw] -translate-x-1/2 truncate rounded bg-black/50 px-3 py-1.5 text-sm text-white; }
.detail-modal__body { @apply min-h-0 flex-1 overflow-y-auto; }
.detail-modal__content { @apply p-5; }
.detail-modal__summary { @apply mt-1.5 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400; }
.detail-modal__meta { @apply mt-3 flex flex-wrap gap-2; }
.detail-modal__tags { @apply mt-2.5 flex flex-wrap gap-2; }
.detail-modal__sections { @apply mt-5 space-y-5; }
.detail-modal__ingredients { @apply mt-2 text-sm text-zinc-700 dark:text-zinc-300; }
.detail-modal__ingredient { @apply flex items-baseline justify-between gap-2 border-b border-dashed border-zinc-200 py-1.5 dark:border-zinc-700; }
.detail-modal__amount { @apply shrink-0 text-zinc-500 dark:text-zinc-400; }
.detail-modal__tools { @apply mt-2 flex flex-wrap gap-2; }
.detail-modal__steps { @apply mt-2 space-y-4; }
.detail-modal__step { @apply flex gap-3; }
.detail-modal__step-no { @apply flex size-6 shrink-0 items-center justify-center rounded-full bg-green-600 text-xs font-semibold text-white; }
.detail-modal__step-title { @apply text-sm font-semibold text-zinc-800 dark:text-zinc-100; }
.detail-modal__step-text { @apply whitespace-pre-wrap text-sm leading-relaxed text-zinc-700 dark:text-zinc-300; }
.detail-modal__step-media { @apply mt-2 flex flex-wrap gap-2; }
.detail-modal__step-img { @apply size-24 rounded border border-zinc-200 object-cover dark:border-zinc-800; }
.detail-modal__tips { @apply rounded-md bg-amber-50 p-3 dark:bg-amber-500/10; }
.detail-modal__tips-title { @apply text-sm font-semibold text-amber-800 dark:text-amber-300; }
.detail-modal__tips-text { @apply mt-1 whitespace-pre-wrap text-sm text-amber-700 dark:text-amber-400; }
.detail-modal__footer { @apply flex justify-end gap-3 border-t border-zinc-200 px-5 py-4 dark:border-zinc-800; }

/* ---- 右键菜单 ---- */
.ctx-menu { @apply absolute min-w-32 rounded-md border border-zinc-200 bg-white py-1 shadow-lg dark:border-zinc-800 dark:bg-zinc-900; }
.ctx-menu__item { @apply block w-full px-4 py-2 text-left text-sm text-zinc-700 hover:bg-zinc-100 disabled:opacity-50 dark:text-zinc-300 dark:hover:bg-zinc-800; }
.ctx-menu__item--danger { @apply text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10; }

/* ---- 删除确认弹窗 ---- */
.confirm-modal__text { @apply mt-3 text-sm text-zinc-600 dark:text-zinc-400; }
.confirm-modal__highlight { @apply font-medium text-red-600 dark:text-red-400; }
.confirm-modal__actions { @apply mt-5 flex justify-end gap-3; }

/* ---- 编辑弹窗 ---- */
.edit-modal__panel { @apply max-h-[90vh] w-full max-w-5xl overflow-y-auto rounded-lg bg-white p-6 shadow-lg lg:flex lg:flex-col lg:overflow-hidden dark:bg-zinc-900; }
.edit-modal__header { @apply flex shrink-0 items-center justify-between; }
.edit-modal__columns { @apply mt-4 grid gap-6 lg:grid-cols-2 lg:min-h-0 lg:flex-1; }
.edit-modal__col--info { @apply space-y-4 lg:min-h-0 lg:overflow-y-auto; }
.edit-modal__col--steps { @apply space-y-5 lg:min-h-0 lg:overflow-y-auto; }
.edit-modal__cover-row { @apply flex items-center gap-4; }
.edit-modal__cover-preview { @apply h-24 w-44 rounded border border-zinc-200 object-cover dark:border-zinc-800; }
.edit-modal__cover-empty { @apply flex h-24 w-44 items-center justify-center rounded border border-dashed border-zinc-300 text-xs text-zinc-400 dark:border-zinc-700 dark:text-zinc-500; }
.edit-modal__radio { @apply flex overflow-hidden rounded-md border border-zinc-300 dark:border-zinc-700; }
.edit-modal__radio-btn { @apply border-r border-zinc-300 px-3 py-1.5 text-sm text-zinc-600 transition-colors last:border-r-0 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-400 dark:hover:bg-zinc-800; }
.edit-modal__radio-btn--active { @apply bg-green-600 text-white hover:bg-green-600 dark:text-white dark:hover:bg-green-600; }
.edit-modal__group-row { @apply grid grid-cols-3 gap-3; }
.edit-modal__ingredient { @apply mb-2 flex items-center gap-2; }
.drag-handle { @apply cursor-grab text-zinc-400 active:cursor-grabbing dark:text-zinc-600; }
.edit-modal__step { @apply mb-2 rounded-lg border border-zinc-200 p-3 dark:border-zinc-800; }
.edit-modal__step-head { @apply mb-2 flex items-center justify-between; }
.edit-modal__step-label { @apply flex items-center gap-1.5 text-sm font-medium text-zinc-600 dark:text-zinc-400; }
.edit-modal__media { @apply mt-2 flex flex-wrap items-center gap-2; }
.edit-modal__thumb { @apply relative size-16 overflow-hidden rounded border border-zinc-200 dark:border-zinc-800; }
.edit-modal__thumb-img { @apply size-full object-cover; }
.edit-modal__thumb-remove { @apply absolute right-0.5 top-0.5 size-5 rounded-full bg-black/60 text-xs text-white hover:bg-black/80; }
.edit-modal__media-add { @apply flex size-16 cursor-pointer items-center justify-center rounded border border-dashed border-zinc-300 text-xl text-zinc-400 hover:border-green-500 hover:text-green-600 dark:border-zinc-700 dark:text-zinc-500 dark:hover:border-green-500 dark:hover:text-green-400; }
.edit-modal__add-step { @apply flex w-full items-center justify-center gap-1 rounded-md border border-dashed border-zinc-300 py-2 text-sm text-zinc-500 hover:border-green-500 hover:text-green-600 dark:border-zinc-700 dark:text-zinc-400 dark:hover:border-green-500 dark:hover:text-green-400; }
.edit-modal__footer { @apply mt-6 flex shrink-0 justify-end gap-3; }

/* ---- 操作提示 ---- */
.toast { @apply fixed left-1/2 top-6 z-[80] -translate-x-1/2 rounded-md bg-zinc-900/90 px-4 py-2 text-sm text-white shadow-lg dark:bg-zinc-700/95; }
</style>
