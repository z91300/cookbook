<script setup lang="ts">
// 首页：标签栏 + 搜索框 + 封面 6:9 卡片网格（无限加载）
import { apis } from '~/api'
import type { Cookbook_internal_model_recipe_list_item, Cookbook_internal_model_tag_item } from '~/api/components'

const searchInput = ref('')
const keywords = ref('')
const activeTagId = ref<number | null>(null) // 单选标签，null=全部
const activeMeal = ref<number | null>(null) // 单选用餐时间，null=全部
const sentinel = ref<Element | null>(null)

// 用餐时间单选项（位掩码值与后端 model 常量对齐）
const mealOptions = [
  { label: '早餐', value: 1 },
  { label: '午餐', value: 2 },
  { label: '晚餐', value: 4 },
  { label: '加餐', value: 8 },
]
// 炊具固定选项（前端写死，编辑/新建弹窗多选）
const toolOptions = ['炒锅', '高压锅', '电饭煲', '平底锅', '砂锅', '汤锅', '奶锅']

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
}

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
  difficulty: number // 难度 0=未填 1=简单 2=中等 3=较难
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
    mealMask: 0,
    difficulty: 0,
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
    mealMask: d.mealMask ?? 0,
    difficulty: d.difficulty ?? 0,
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
      mealMask: detail.mealMask ?? 0,
      difficulty: detail.difficulty ?? 0,
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

// ============ 右下角常驻菜单：跳转收藏/成员/食谱编排 ============
function gotoFavorites() {
  navigateTo('/favorites')
}

function gotoMembers() {
  navigateTo('/members')
}

function gotoScheduling() {
  navigateTo('/scheduling')
}

// ============ 收藏夹弹窗（卡片爱心与详情弹窗共用） ============
const favRecipeId = ref<number | null>(null)


function openFavModal(recipeId: number) {
  favRecipeId.value = recipeId
}

function closeFavModal() {
  favRecipeId.value = null
}

function gotoSettings() {
  navigateTo('/settings')
}
</script>

<template>
  <main class="mx-auto max-w-6xl px-4 py-8">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">菜谱</h1>
      <button
        type="button"
        class="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700"
        @click="openCreate"
      >+ 新建菜谱</button>
    </div>

    <!-- 搜索框 -->
    <div class="mt-4">
      <input
        v-model="searchInput"
        type="search"
        placeholder="搜索菜谱名称或简介…"
        class="w-full rounded-md border border-zinc-300 bg-white px-4 py-2 text-sm outline-none focus:border-green-600"
      >
    </div>

    <!-- 标签栏：单选，默认"全部" -->
    <div class="mt-4 flex flex-wrap gap-2">
      <button
        type="button"
        class="rounded-full border px-3 py-1 text-sm transition-colors"
        :class="activeTagId === null
          ? 'border-green-600 bg-green-600 text-white'
          : 'border-zinc-300 bg-white text-zinc-700 hover:border-green-500'"
        @click="activeTagId = null"
      >
        全部
      </button>
      <button
        v-for="tag in tagList"
        :key="tag.id"
        type="button"
        class="rounded-full border px-3 py-1 text-sm transition-colors"
        :class="activeTagId === tag.id
          ? 'border-green-600 bg-green-600 text-white'
          : 'border-zinc-300 bg-white text-zinc-700 hover:border-green-500'"
        @click="selectTag(tag.id!)"
      >
        {{ tag.name }}
      </button>
    </div>

    <!-- 用餐时间栏：单选，默认"全部" -->
    <div class="mt-2 flex flex-wrap gap-2">
      <button
        type="button"
        class="rounded-full border px-3 py-1 text-sm transition-colors"
        :class="activeMeal === null
          ? 'border-amber-600 bg-amber-600 text-white'
          : 'border-zinc-300 bg-white text-zinc-700 hover:border-amber-500'"
        @click="activeMeal = null"
      >
        全部时段
      </button>
      <button
        v-for="opt in mealOptions"
        :key="opt.value"
        type="button"
        class="rounded-full border px-3 py-1 text-sm transition-colors"
        :class="activeMeal === opt.value
          ? 'border-amber-600 bg-amber-600 text-white'
          : 'border-zinc-300 bg-white text-zinc-700 hover:border-amber-500'"
        @click="selectMeal(opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>

    <!-- 筛选状态行 -->
    <p class="mt-4 text-sm text-zinc-500">
      共 {{ total }} 道菜谱
      <button
        v-if="keywords || activeTagId !== null || activeMeal !== null"
        type="button"
        class="ml-2 text-green-600 hover:underline"
        @click="clearFilters"
      >
        清除筛选
      </button>
    </p>

    <!-- 卡片网格：正常横向封面（4:3） -->
    <div v-if="items.length" class="mt-3 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
      <article
        v-for="item in items"
        :key="item.id"
        class="cursor-pointer overflow-hidden rounded-lg border border-zinc-200 bg-white transition-shadow hover:shadow-md"
        @click="openDetail(item)"
        @contextmenu="openCtxMenu($event, item)"
      >
        <div class="relative w-full" style="aspect-ratio: 650 / 370;overflow: hidden;">
          <img
            v-if="item.coverUrl"
            :src="thumbUrl(item.coverUrl)"
            :alt="item.title"
            class="absolute inset-0 h-full w-full object-cover object-center"
            loading="lazy"
          >
          <div v-else class="absolute inset-0 flex items-center justify-center bg-zinc-100 text-zinc-400">暂无封面</div>
          <span
            v-if="item.difficulty"
            class="absolute left-2 top-2 rounded bg-black/50 px-1.5 py-0.5 text-xs text-white"
          >{{ difficultyText[item.difficulty] }}</span>
          <button
            type="button"
            class="absolute right-2 top-2 z-10 flex items-center gap-1 rounded-full px-2 py-1 text-xs backdrop-blur-sm transition-colors"
            :class="item.favoriteCount
              ? 'bg-red-600/85 text-white hover:bg-red-500'
              : 'bg-black/50 text-white hover:bg-black/70'"
            :title="item.favoriteCount ? `已收藏 ${item.favoriteCount} 个收藏夹` : '收藏'"
            @click.stop="openFavModal(item.id)"
          >
            <span aria-hidden="true">♥</span>
            <span v-if="item.favoriteCount">{{ item.favoriteCount }}</span>
          </button>
        </div>
        <div class="p-3">
          <h2 class="truncate text-sm font-semibold">{{ item.title }}</h2>
          <p class="mt-1 line-clamp-2 min-h-8 text-xs text-zinc-500">{{ item.summary }}</p>
          <div class="mt-2 flex flex-wrap gap-1">
            <span
              v-for="tag in item.tags"
              :key="tag.id"
              class="rounded bg-zinc-100 px-1.5 py-0.5 text-xs text-zinc-600"
            >{{ tag.name }}</span>
          </div>
        </div>
      </article>
    </div>

    <p v-else-if="status === 'done' && !error" class="mt-8 text-center text-sm text-zinc-400">没有找到匹配的菜谱</p>
    <p v-if="error" class="mt-8 text-center text-sm text-red-500">{{ error }}</p>

    <!-- 上拉加载哨兵 -->
    <div ref="sentinel" class="h-4" />
    <p v-if="status === 'loading'" class="py-4 text-center text-sm text-zinc-400">加载中…</p>
    <p v-else-if="status === 'done' && items.length" class="py-4 text-center text-sm text-zinc-400">已全部加载</p>
  </main>

  <!-- 食谱详情弹窗（点击卡片） -->
  <Teleport to="body">
    <div v-if="detailOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" @click.self="closeDetail">
      <div class="flex h-[90vh] w-full max-w-5xl flex-col overflow-hidden rounded-lg bg-white shadow-lg">
        <div class="flex items-center justify-between border-b border-zinc-200 px-5 py-3">
          <h3 class="text-lg font-semibold">{{ detailLoading ? '加载中…' : detail?.title }}</h3>
          <button type="button" class="text-zinc-400 hover:text-zinc-600" @click="closeDetail">✕</button>
        </div>

        <p v-if="detailError" class="m-5 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ detailError }}</p>

        <div v-else-if="detail" class="detail-body min-h-0 flex-1 overflow-y-auto p-5">
          <p class="text-sm leading-relaxed text-zinc-600">{{ detail.summary }}</p>
            <div class="mt-3 flex flex-wrap gap-2 text-xs text-zinc-600">
              <span v-if="detail.difficulty" class="rounded bg-zinc-100 px-2 py-1">难度：{{ difficultyText[detail.difficulty] || '未填' }}</span>
              <span v-if="detail.cookMinutes" class="rounded bg-zinc-100 px-2 py-1">耗时：{{ detail.cookMinutes }} 分钟</span>
              <span v-if="detail.calories" class="rounded bg-zinc-100 px-2 py-1">热量：{{ detail.calories }} kcal/份</span>
              <span v-if="detail.servings" class="rounded bg-zinc-100 px-2 py-1">{{ detail.servings }} 人份</span>
            </div>
            <div class="mt-3 flex flex-wrap gap-2">
              <span
                v-if="detail.mealMask"
                class="rounded bg-amber-50 px-2 py-1 text-xs text-amber-700"
              >{{ mealText(detail.mealMask) }}</span>
              <span
                v-for="tag in detail.tags"
                :key="tag.id"
                class="rounded bg-green-50 px-2 py-1 text-xs text-green-700"
              >{{ tag.name }}</span>
            </div>
            <!-- 食材 -->
            <div v-if="detail.ingredients?.length" class="mt-5">
              <h4 class="text-sm font-semibold">食材</h4>
              <ul class="mt-2 grid grid-cols-2 gap-x-4 gap-y-1 text-sm text-zinc-700">
                <li v-for="(ing, i) in detail.ingredients" :key="i" class="flex justify-between border-b border-dashed border-zinc-200 py-1">
                  <span>{{ ing.name }}<span v-if="ing.optional" class="ml-1 text-xs text-zinc-400">（可选）</span></span>
                  <span class="text-zinc-500">{{ ing.amount }}</span>
                </li>
              </ul>
            </div>

            <!-- 工具 -->
            <div v-if="detail.tools?.length" class="mt-5">
              <h4 class="text-sm font-semibold">工具</h4>
              <div class="mt-2 flex flex-wrap gap-2">
                <span v-for="(t, i) in detail.tools" :key="i" class="rounded bg-zinc-100 px-2 py-1 text-xs text-zinc-600">{{ t }}</span>
              </div>
            </div>

            <!-- 步骤 -->
            <div v-if="detail.steps?.length" class="mt-5">
              <h4 class="text-sm font-semibold">步骤</h4>
              <ol class="mt-2 space-y-4">
                <li v-for="(step, i) in detail.steps" :key="i" class="flex gap-3">
                  <span class="flex size-6 shrink-0 items-center justify-center rounded-full bg-green-600 text-xs font-semibold text-white">{{ i + 1 }}</span>
                  <div class="min-w-0 flex-1">
                    <p v-if="step.title" class="text-sm font-semibold text-zinc-800">{{ step.title }}</p>
                    <p
                      class="whitespace-pre-wrap text-sm leading-relaxed text-zinc-700"
                      :class="step.title ? 'mt-1' : ''"
                    >{{ step.content }}</p>
                    <div v-if="step.media?.length" class="mt-2 flex flex-wrap gap-2">
                      <img
                        v-for="mediaId in step.media"
                        :key="mediaId"
                        :src="attachmentUrl(mediaId)"
                        :alt="`步骤图 ${mediaId}`"
                        class="size-24 rounded border border-zinc-200 object-cover"
                        loading="lazy"
                      >
                    </div>
                  </div>
                </li>
              </ol>
            </div>

            <!-- 注意事项 -->
            <div v-if="detail.tips" class="mt-5 rounded-md bg-amber-50 p-3">
              <h4 class="text-sm font-semibold text-amber-800">小贴士</h4>
              <p class="mt-1 whitespace-pre-wrap text-sm text-amber-700">{{ detail.tips }}</p>
            </div>
        </div>

        <!-- 操作：收藏（收藏夹弹窗）+ 编辑 -->
        <div v-if="detail" class="flex justify-end gap-3 border-t border-zinc-200 p-5">
          <button
            type="button"
            class="rounded-md border border-zinc-300 px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50"
            @click="openFavModal(detail.id)"
          >
            收藏
          </button>
          <button
            type="button"
            class="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700"
            @click="openEditFromDetail"
          >
            编辑
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 右键菜单 -->
  <Teleport to="body">
    <div v-if="ctxMenu.item" class="fixed inset-0 z-50" @click="closeCtxMenu" @contextmenu.prevent="closeCtxMenu">
      <div
        class="absolute min-w-32 rounded-md border border-zinc-200 bg-white py-1 shadow-lg"
        :style="{ left: `${ctxMenu.x}px`, top: `${ctxMenu.y}px` }"
      >
        <button
          type="button"
          class="block w-full px-4 py-2 text-left text-sm text-zinc-700 hover:bg-zinc-100"
          :disabled="copyPromptBusyId !== null"
          @click.stop="copyGenPrompt(ctxMenu.item!)"
        >
          {{ copyPromptBusyId === ctxMenu.item!.id ? '渲染中…' : '复制生图提示词' }}
        </button>
        <button
          type="button"
          class="block w-full px-4 py-2 text-left text-sm text-zinc-700 hover:bg-zinc-100"
          @click.stop="openEdit(ctxMenu.item!)"
        >
          编辑
        </button>
        <button
          type="button"
          class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50"
          @click.stop="askDelete(ctxMenu.item!)"
        >
          删除
        </button>
      </div>
    </div>
  </Teleport>

  <!-- 删除确认弹窗 -->
  <Teleport to="body">
    <div v-if="deleteConfirm" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4" @click.self="cancelDelete">
      <div class="w-full max-w-sm rounded-lg bg-white p-6 shadow-lg">
        <p class="mt-3 text-sm text-zinc-600">
          确定要删除食谱
          <span class="font-medium text-red-600">「{{ deleteConfirm.item.title }}」</span>
        </p>
        <p v-if="deleteConfirm.error" class="mt-2 text-sm text-red-500">{{ deleteConfirm.error }}</p>
        <div class="mt-5 flex justify-end gap-3">
          <button
            type="button"
            class="rounded-md border border-zinc-300 px-4 py-2 text-sm hover:bg-zinc-50"
            :disabled="deleteConfirm.deleting"
            @click="cancelDelete"
          >
            取消
          </button>
          <button
            type="button"
            class="rounded-md bg-red-600 px-4 py-2 text-sm text-white hover:bg-red-700 disabled:opacity-50"
            :disabled="deleteConfirm.deleting"
            @click="confirmDelete"
          >
            {{ deleteConfirm.deleting ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- 编辑弹窗 -->
  <Teleport to="body">
    <div v-if="editOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" @click.self="closeEdit">
      <div class="max-h-[90vh] w-full max-w-5xl overflow-y-auto rounded-lg bg-white p-6 shadow-lg lg:flex lg:flex-col lg:overflow-hidden">
        <div class="flex shrink-0 items-center justify-between">
          <h3 class="text-lg font-semibold">{{ (editForm?.id ?? 0) > 0 ? '编辑食谱' : '新建食谱' }}</h3>
          <button type="button" class="text-zinc-400 hover:text-zinc-600" @click="closeEdit">✕</button>
        </div>

        <p v-if="editLoading" class="py-10 text-center text-sm text-zinc-400">加载中…</p>
        <template v-else-if="editForm">
          <p v-if="editError" class="mt-3 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ editError }}</p>

          <div class="mt-4 grid gap-6 lg:grid-cols-2 lg:min-h-0 lg:flex-1">
            <!-- 左列：基础信息（桌面端独立滚动，不随右列联动） -->
            <div class="space-y-4 lg:min-h-0 lg:overflow-y-auto">
              <!-- 封面图（左列首行） -->
              <div>
                <label class="mb-1 block text-sm font-medium">封面图</label>
                <div class="flex items-center gap-4">
                  <img
                    v-if="coverPreview"
                    :src="coverPreview"
                    alt="封面预览"
                    class="h-24 w-16 rounded border border-zinc-200 object-cover"
                  >
                  <div v-else class="flex h-24 w-16 items-center justify-center rounded border border-dashed border-zinc-300 text-xs text-zinc-400">无封面</div>
                  <label class="cursor-pointer rounded-md border border-zinc-300 px-3 py-1.5 text-sm text-zinc-700 hover:bg-zinc-50">
                    选择文件
                    <input type="file" accept="image/*" class="hidden" @change="onCoverChange">
                  </label>
                </div>
              </div>

              <!-- 名称 -->
              <div>
                <label class="mb-1 block text-sm font-medium">菜谱名称</label>
                <input
                  v-model="editForm.title"
                  class="w-full resize-y rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-600"
                />
              </div>

              <!-- 分类（多选标签） -->
              <div>
                <label class="mb-1 block text-sm font-medium">分类</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="tag in tagList"
                    :key="tag.id"
                    type="button"
                    class="rounded-full border px-3 py-1 text-sm transition-colors"
                    :class="editForm.tagIds.includes(tag.id!)
                      ? 'border-green-600 bg-green-600 text-white'
                      : 'border-zinc-300 bg-white text-zinc-700 hover:border-green-500'"
                    @click="toggleEditTag(tag.id!)"
                  >
                    {{ tag.name }}
                  </button>
                </div>
              </div>

              <!-- 用餐时段（多选） -->
              <div>
                <label class="mb-1 block text-sm font-medium">适合时段（可多选）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="opt in mealOptions"
                    :key="opt.value"
                    type="button"
                    class="rounded-full border px-3 py-1 text-sm transition-colors"
                    :class="(editForm.mealMask & opt.value) !== 0
                      ? 'border-amber-600 bg-amber-600 text-white'
                      : 'border-zinc-300 bg-white text-zinc-700 hover:border-amber-500'"
                    @click="toggleEditMeal(opt.value)"
                  >
                    {{ opt.label }}
                  </button>
                </div>
              </div>

              <!-- 炊具（多选，固定选项） -->
              <div>
                <label class="mb-1 block text-sm font-medium">所需工具（可多选）</label>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="tool in toolOptions"
                    :key="tool"
                    type="button"
                    class="rounded-full border px-3 py-1 text-sm transition-colors"
                    :class="editForm.tools.includes(tool)
                      ? 'border-sky-600 bg-sky-600 text-white'
                      : 'border-zinc-300 bg-white text-zinc-700 hover:border-sky-500'"
                    @click="toggleEditTool(tool)"
                  >
                    {{ tool }}
                  </button>
                </div>
              </div>

              <!-- 难度 / 耗时 / 热量 -->
              <div class="grid grid-cols-3 gap-3">
                <div>
                  <label class="mb-1 block text-sm font-medium">难度</label>
                  <select
                    v-model.number="editForm.difficulty"
                    class="w-full rounded-md border border-zinc-300 px-2 py-2 text-sm outline-none focus:border-green-600"
                  >
                    <option :value="0">未填</option>
                    <option :value="1">简单</option>
                    <option :value="2">中等</option>
                    <option :value="3">较难</option>
                  </select>
                </div>
                <div>
                  <label class="mb-1 block text-sm font-medium">耗时(分)</label>
                  <input
                    v-model.number="editForm.cookMinutes"
                    type="number"
                    min="0"
                    class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-600"
                  >
                </div>
                <div>
                  <label class="mb-1 block text-sm font-medium">热量(kcal)</label>
                  <input
                    v-model.number="editForm.calories"
                    type="number"
                    min="0"
                    class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-600"
                  >
                </div>
              </div>
            </div>

            <!-- 右列：食材 + 步骤（桌面端独立滚动） -->
            <div class="space-y-5 lg:min-h-0 lg:overflow-y-auto">
              <!-- 食材（动态表单） -->
              <div>
                <div class="mb-1 flex items-center justify-between">
                  <label class="text-sm font-medium">食材</label>
                  <button type="button" class="text-sm text-green-600 hover:underline" @click="addIngredient">+ 添加食材</button>
                </div>
                <div
                  v-for="(ing, idx) in editForm.ingredients"
                  :key="idx"
                  class="mb-2 flex items-center gap-2"
                  draggable="true"
                  :class="idx === dragFromIngredient ? 'opacity-50' : ''"
                  @dragstart="onDragStart($event, idx, 'ingredient')"
                  @dragover="onDragOver($event, 'ingredient')"
                  @drop="onDrop($event, idx, 'ingredient')"
                  @dragend="dragFromIngredient = -1"
                >
                  <span class="cursor-grab text-zinc-400 active:cursor-grabbing" title="拖拽排序">⠿</span>
                  <input
                    v-model="ing.name"
                    type="text"
                    placeholder="食材名"
                    class="w-40 rounded border border-zinc-300 px-2 py-1.5 text-sm outline-none focus:border-green-600"
                  >
                  <input
                    v-model="ing.amount"
                    type="text"
                    placeholder="用量（如 500g）"
                    class="w-32 rounded border border-zinc-300 px-2 py-1.5 text-sm outline-none focus:border-green-600"
                  >
                  <button type="button" class="text-sm text-red-500 hover:underline" @click="removeIngredient(idx)">删除</button>
                </div>
              </div>

              <!-- 步骤（可多图） -->
              <div>
                <div class="mb-1 flex items-center justify-between">
                  <label class="text-sm font-medium">步骤</label>
                  <button type="button" class="text-sm text-green-600 hover:underline" @click="addStep">+ 添加步骤</button>
                </div>

                <div
                  v-for="(step, idx) in editForm.steps"
                  :key="idx"
                  class="mb-2 rounded-lg border border-zinc-200 p-3"
                  draggable="true"
                  :class="idx === dragFromStep ? 'opacity-50' : ''"
                  @dragstart="onDragStart($event, idx, 'step')"
                  @dragover="onDragOver($event, 'step')"
                  @drop="onDrop($event, idx, 'step')"
                  @dragend="dragFromStep = -1"
                >
                  <div class="mb-2 flex items-center justify-between">
                    <span class="flex items-center gap-1.5 text-sm font-medium text-zinc-600">
                      <span class="cursor-grab text-zinc-400 active:cursor-grabbing" title="拖拽排序">⠿</span>
                      第 {{ idx + 1 }} 步
                    </span>
                    <button
                      type="button"
                      class="text-xs text-red-500 hover:underline"
                      @click="removeStep(idx)"
                    >删除步骤</button>
                  </div>

                  <!-- 步骤标题（可选） -->
                  <input
                    v-model="step.title"
                    type="text"
                    placeholder="步骤标题（可选），如 焯水去腥"
                    class="mb-2 w-full rounded border border-zinc-300 px-2 py-1.5 text-sm outline-none focus:border-green-600"
                  >

                  <!-- 步骤内容 -->
                  <textarea
                    v-model="step.content"
                    rows="3"
                    placeholder="步骤内容"
                    class="w-full resize-y rounded border border-zinc-300 px-2 py-1.5 text-sm outline-none focus:border-green-600"
                  />

                  <!-- 步骤图片：已有缩略图 + 添加 -->
                  <div class="mt-2 flex flex-wrap items-center gap-2">
                    <div
                      v-for="mediaId in step.media"
                      :key="mediaId"
                      class="group relative size-16 overflow-hidden rounded border border-zinc-200"
                    >
                      <img
                        :src="attachmentUrl(mediaId)"
                        :alt="`步骤图 ${mediaId}`"
                        class="size-full object-cover"
                        loading="lazy"
                      >
                      <button
                        type="button"
                        class="absolute right-0.5 top-0.5 flex size-5 items-center justify-center rounded-full bg-black/60 text-xs text-white hover:bg-black/80"
                        @click="removeStepImage(idx, mediaId)"
                      >×</button>
                    </div>
                    <label
                      class="flex size-16 cursor-pointer items-center justify-center rounded border border-dashed border-zinc-300 text-xl text-zinc-400 hover:border-green-500 hover:text-green-600"
                      :class="stepUploading.includes(idx) ? 'pointer-events-none opacity-60' : ''"
                    >
                      <span v-if="stepUploading.includes(idx)" class="text-xs">上传中…</span>
                      <span v-else>+</span>
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

                <button
                  type="button"
                  class="w-full rounded-md border border-dashed border-zinc-300 py-2 text-sm text-zinc-500 hover:border-green-500 hover:text-green-600"
                  @click="addStep"
                >+ 添加步骤</button>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="mt-6 flex shrink-0 justify-end gap-3">
            <button
              type="button"
              class="rounded-md border border-zinc-300 px-4 py-2 text-sm hover:bg-zinc-50"
              :disabled="editSaving"
              @click="closeEdit"
            >
              取消
            </button>
            <button
              type="button"
              class="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-50"
              :disabled="editSaving"
              @click="saveEdit"
            >
              {{ editSaving ? '保存中…' : '保存' }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
  <!-- 右下角常驻菜单：收藏 / 成员 / 食谱编排，竖排图标按钮，始终可见 -->
  <Teleport to="body">
    <nav
      aria-label="快捷导航"
      class="fixed bottom-6 right-6 z-40 flex flex-col gap-2"
    >
      <button
        type="button"
        class="group flex items-center gap-2 rounded-full border border-zinc-200 bg-white px-4 py-2.5 text-sm text-zinc-700 shadow-lg transition-colors hover:border-green-500 hover:text-green-600"
        @click="gotoFavorites"
      >
        <span aria-hidden="true">★</span>
        <span>收藏</span>
      </button>
      <button
        type="button"
        class="group flex items-center gap-2 rounded-full border border-zinc-200 bg-white px-4 py-2.5 text-sm text-zinc-700 shadow-lg transition-colors hover:border-green-500 hover:text-green-600"
        @click="gotoMembers"
      >
        <span aria-hidden="true">👤</span>
        <span>成员</span>
      </button>
      <button
        type="button"
        class="group flex items-center gap-2 rounded-full border border-zinc-200 bg-white px-4 py-2.5 text-sm text-zinc-700 shadow-lg transition-colors hover:border-green-500 hover:text-green-600"
        @click="gotoScheduling"
      >
        <span aria-hidden="true">📅</span>
        <span>食谱编排</span>
      </button>
      <button
        type="button"
        class="group flex items-center gap-2 rounded-full border border-zinc-200 bg-white px-4 py-2.5 text-sm text-zinc-700 shadow-lg transition-colors hover:border-green-500 hover:text-green-600"
        @click="gotoSettings"
      >
        <span aria-hidden="true">⚙️</span>
        <span>设置</span>
      </button>
    </nav>
  </Teleport>

  <!-- 复制生图提示词结果提示 -->
  <Teleport to="body">
    <div
      v-if="copyPromptTip"
      class="fixed left-1/2 top-6 z-[80] -translate-x-1/2 rounded-md bg-zinc-900/90 px-4 py-2 text-sm text-white shadow-lg"
    >
      {{ copyPromptTip }}
    </div>
  </Teleport>
  <!-- 收藏夹弹窗（卡片爱心 / 详情弹窗收藏按钮共用） -->
  <FavoriteFolderModal :open="favRecipeId !== null" :recipe-id="favRecipeId" @close="closeFavModal" />
</template>
