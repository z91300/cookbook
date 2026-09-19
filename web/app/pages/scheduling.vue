<script setup lang="ts">
// 食谱编排页：按周 × 用餐时段编排菜谱，数据走后端 /schedulings
import { apis } from '~/api'
import type { Cookbook_internal_model_scheduling_item, Cookbook_internal_model_recipe_list_item } from '~/api/components'

// 用餐时段（位掩码值与后端 model 常量对齐，一菜一行 meal 取单值）
const mealSlots = [
  { key: 'breakfast', label: '早餐', value: 1 },
  { key: 'lunch', label: '午餐', value: 2 },
  { key: 'dinner', label: '晚餐', value: 4 },
  { key: 'extra', label: '加餐', value: 8 },
] as const

const weekDays = [
  { key: 'mon', label: '周一' },
  { key: 'tue', label: '周二' },
  { key: 'wed', label: '周三' },
  { key: 'thu', label: '周四' },
  { key: 'fri', label: '周五' },
  { key: 'sat', label: '周六' },
  { key: 'sun', label: '周日' },
] as const

type DayKey = typeof weekDays[number]['key']

// 周起始（默认本周一）；weekOffset 切换周
const weekOffset = ref(0)

function mondayOf(offset: number): Date {
  const d = new Date()
  const day = d.getDay() // 0=周日 ... 6=周六
  const diffToMonday = day === 0 ? -6 : 1 - day
  d.setDate(d.getDate() + diffToMonday + offset * 7)
  d.setHours(0, 0, 0, 0)
  return d
}

function yyyymmdd(d: Date): number {
  return d.getFullYear() * 10000 + (d.getMonth() + 1) * 100 + d.getDate()
}

const weekDates = computed(() => {
  const start = mondayOf(weekOffset.value)
  return weekDays.map((w, i) => {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    return { ...w, date: d, yyyymmdd: yyyymmdd(d), label: `${w.label} ${d.getMonth() + 1}/${d.getDate()}` }
  })
})

const weekLabel = computed(() => {
  const a = weekDates.value[0].date
  const b = weekDates.value[6].date
  const fmt = (d: Date) => `${d.getFullYear()}.${d.getMonth() + 1}.${d.getDate()}`
  return `${fmt(a)} - ${fmt(b)}`
})

function prevWeek() { weekOffset.value-- }
function nextWeek() { weekOffset.value++ }
function thisWeek() { weekOffset.value = 0 }

function isToday(date: Date): boolean {
  if (weekOffset.value !== 0) return false
  const today = new Date()
  return today.getDate() === date.getDate()
    && today.getMonth() === date.getMonth()
    && today.getFullYear() === date.getFullYear()
}

// ============ 编排数据：按周拉取 ============
const entries = ref<Cookbook_internal_model_scheduling_item[]>([])
const loadError = ref('')
// 首屏/切周加载：遮罩周历网格，避免误以为空白直接编辑；增删改后的刷新为静默模式
const gridLoading = ref(true)
let loadSeq = 0

async function loadWeek(opts?: { background?: boolean }) {
  const seq = ++loadSeq
  if (!opts?.background) gridLoading.value = true
  loadError.value = ''
  const start = weekDates.value[0].yyyymmdd
  const end = weekDates.value[6].yyyymmdd
  try {
    const res = await apis.scheduling.getList({ params: { startDate: start, endDate: end } })
    if (seq !== loadSeq) return // 已切到其他周，丢弃过期响应
    entries.value = res.list ?? []
  }
  catch (e) {
    if (seq !== loadSeq) return
    loadError.value = e instanceof Error ? e.message : '加载编排失败'
  }
  if (seq === loadSeq) gridLoading.value = false
}

watch(weekOffset, () => loadWeek())
onMounted(() => loadWeek())

// 单元格内的编排项：按 planDate+meal 分组
function cellEntries(dayYyyymmdd: number, mealValue: number): Cookbook_internal_model_scheduling_item[] {
  return entries.value.filter(e => e.planDate === dayYyyymmdd && e.meal === mealValue)
}

// ============ 移除编排条目 ============
const removingId = ref<number | null>(null)

async function removeEntry(id: number) {
  removingId.value = id
  try {
    await apis.scheduling.delete({ pathParams: { id } })
    await loadWeek({ background: true })
  }
  catch (e) {
    loadError.value = e instanceof Error ? e.message : '删除失败'
  }
  removingId.value = null
}

// ============ 菜谱选择器：搜索 + 选中即编排 ============
const pickerOpen = ref(false)
const pickerCell = ref<{ yyyymmdd: number, meal: number, dayLabel: string, mealLabel: string } | null>(null)
const pickerSearch = ref('')
const pickerKeywords = ref('')
const pickerResults = ref<Cookbook_internal_model_recipe_list_item[]>([])
const pickerLoading = ref(false)
const pickerError = ref('')
const addingId = ref<number | null>(null) // 正在添加的 recipeId

function openPicker(dayYyyymmdd: number, mealValue: number, dayLabel: string, mealLabel: string) {
  pickerCell.value = { yyyymmdd: dayYyyymmdd, meal: mealValue, dayLabel, mealLabel }
  pickerOpen.value = true
  pickerSearch.value = ''
  pickerKeywords.value = ''
  pickerError.value = ''
  lastPicked.value = null
  searchRecipes()
}

function closePicker() {
  pickerOpen.value = false
  pickerCell.value = null
  addingId.value = null
  lastPicked.value = null
}

// 选择器弹窗：锁定页面滚动 + 移动端侧滑返回关闭弹窗
useBodyScrollLock([() => pickerOpen.value])
useModalBackClose(() => pickerOpen.value, closePicker)

// ESC 关闭选择器（与收藏夹弹窗一致）
function onPickerKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && pickerOpen.value) closePicker()
}
onMounted(() => window.addEventListener('keydown', onPickerKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onPickerKeydown))

async function searchRecipes() {
  pickerLoading.value = true
  pickerError.value = ''
  try {
    const res = await apis.recipe.getList({
      params: { page: 1, pageSize: 20, keywords: pickerKeywords.value || undefined },
    })
    pickerResults.value = res.list ?? []
  }
  catch (e) {
    pickerError.value = e instanceof Error ? e.message : '加载菜谱失败'
  }
  pickerLoading.value = false
}

// 搜索防抖 300ms
let debounceTimer: ReturnType<typeof setTimeout> | undefined
watch(pickerSearch, (val) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    const trimmed = val.trim()
    if (trimmed === pickerKeywords.value) return
    pickerKeywords.value = trimmed
    searchRecipes()
  }, 300)
})

// 连续编排：选完不关弹窗，展示最近一次成功结果，方便同餐多菜
const lastPicked = ref<{ title: string, dayLabel: string, mealLabel: string } | null>(null)

async function pickRecipe(recipe: Cookbook_internal_model_recipe_list_item) {
  if (!pickerCell.value) return
  addingId.value = recipe.id!
  try {
    await apis.scheduling.create({
      body: {
        planDate: pickerCell.value.yyyymmdd,
        meal: pickerCell.value.meal as 1 | 2 | 4 | 8,
        recipeId: recipe.id!,
      },
    })
    await loadWeek({ background: true })
    lastPicked.value = {
      title: recipe.title ?? '菜谱',
      dayLabel: pickerCell.value.dayLabel,
      mealLabel: pickerCell.value.mealLabel,
    }
    addingId.value = null
  }
  catch (e) {
    pickerError.value = e instanceof Error ? e.message : '编排失败'
    addingId.value = null
  }
}

// ============ 拖拽移动：把已编排条目拖到其他日期/时段（走 scheduling_update 全量覆盖） ============
const dragEntry = ref<Cookbook_internal_model_scheduling_item | null>(null)
const dragOverCell = ref('') // `${yyyymmdd}-${meal}` 高亮落点
const moveTargetId = ref<number | null>(null) // 正在移动的条目

function cellKey(dayYyyymmdd: number, mealValue: number): string {
  return `${dayYyyymmdd}-${mealValue}`
}

function onEntryDragStart(e: DragEvent, entry: Cookbook_internal_model_scheduling_item) {
  if (!e.dataTransfer) return
  dragEntry.value = entry
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', `scheduling:${entry.id}`)
}

function onEntryDragEnd() {
  dragEntry.value = null
  dragOverCell.value = ''
}

function onCellDragOver(e: DragEvent, dayYyyymmdd: number, mealValue: number) {
  if (!dragEntry.value) return
  e.preventDefault()
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOverCell.value = cellKey(dayYyyymmdd, mealValue)
}

async function onCellDrop(dayYyyymmdd: number, mealValue: number) {
  const entry = dragEntry.value
  dragEntry.value = null
  dragOverCell.value = ''
  if (!entry?.id) return
  if (entry.planDate === dayYyyymmdd && entry.meal === mealValue) return
  moveTargetId.value = entry.id
  try {
    await apis.scheduling.update({
      pathParams: { id: entry.id },
      body: {
        planDate: dayYyyymmdd,
        meal: mealValue as 1 | 2 | 4 | 8,
        recipeId: entry.recipeId!,
        servings: entry.servings ?? 0,
        note: entry.note ?? '',
        sort: entry.sort ?? 0,
      },
    })
    await loadWeek({ background: true })
  }
  catch (e) {
    loadError.value = e instanceof Error ? e.message : '移动失败'
  }
  moveTargetId.value = null
}
</script>

<template>
  <main class="page">
    <div class="page-header page-header--wrap">
      <div>
        <h1 class="page-title">食谱编排</h1>
        <p class="schedule-week__label mt-1">
          {{ weekLabel }}
          <button
            v-if="weekOffset !== 0"
            type="button"
            class="text-btn text-btn--accent text-btn--xs"
            @click="thisWeek"
          >回到本周</button>
        </p>
      </div>
      <div class="flex items-center gap-1.5">
        <button type="button" class="btn btn--outline btn--sm" @click="prevWeek">‹ 上一周</button>
        <button
          type="button"
          class="week-toggle"
          :class="{ 'week-toggle--active': weekOffset === 0 }"
          @click="thisWeek"
        >本周</button>
        <button type="button" class="btn btn--outline btn--sm" @click="nextWeek">下一周 ›</button>
      </div>
    </div>

    <p v-if="loadError" class="error-alert mt-4">{{ loadError }}</p>

    <!-- 周历网格：行=用餐时段，列=周一~周日；条目可拖拽换槽；加载中整体遮罩 -->
    <div class="week-panel relative">
      <table class="week-table">
        <thead>
          <tr>
            <th class="week-table__head w-20">
              时段 / 日期
            </th>
            <th
              v-for="d in weekDates"
              :key="d.key"
              class="week-table__head"
              :class="{ 'week-table__head--today': isToday(d.date) }"
            >
              <span class="inline-flex items-center gap-1.5">{{ d.label }}
                <span v-if="isToday(d.date)" class="week-table__today-badge">今天</span>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="slot in mealSlots" :key="slot.key">
            <td class="week-table__meal">
              {{ slot.label }}
            </td>
            <td
              v-for="d in weekDates"
              :key="d.key"
              class="week-table__cell"
              :class="{ 'week-table__cell--today': isToday(d.date) }"
              @dragover="onCellDragOver($event, d.yyyymmdd, slot.value)"
              @drop.prevent="onCellDrop(d.yyyymmdd, slot.value)"
            >
              <div
                class="week-table__cell-body"
                :class="{ 'week-table__cell-body--drop': dragOverCell === cellKey(d.yyyymmdd, slot.value) }"
                @click.self="openPicker(d.yyyymmdd, slot.value, d.label, slot.label)"
              >
                <!-- 已编排条目（一餐可多菜）：卡片化，可拖拽换槽 -->
                <div
                  v-for="e in cellEntries(d.yyyymmdd, slot.value)"
                  :key="e.id"
                  draggable="true"
                  class="schedule-entry group"
                  :class="dragEntry?.id === e.id || moveTargetId === e.id ? 'opacity-50' : ''"
                  :title="`${e.recipeTitle}（拖动可换日期/时段）`"
                  @dragstart="onEntryDragStart($event, e)"
                  @dragend="onEntryDragEnd"
                >
                  <img
                    v-if="e.coverUrl"
                    :src="thumbUrl(e.coverUrl)"
                    :alt="e.recipeTitle"
                    class="schedule-entry__img"
                    loading="lazy"
                  >
                  <div v-else class="schedule-entry__noimg">无图</div>
                  <p class="schedule-entry__title">{{ e.recipeTitle }}</p>
                  <button
                    v-if="e.id"
                    type="button"
                    class="icon-btn schedule-entry__remove"
                    :disabled="removingId === e.id"
                    :title="removingId === e.id ? '移除中…' : '移除'"
                    @click.stop="removeEntry(e.id)"
                  >✕</button>
                </div>
                <!-- 添加按钮 -->
                <button
                  type="button"
                  class="schedule-cell__add"
                  @click="openPicker(d.yyyymmdd, slot.value, d.label, slot.label)"
                >+ 添加</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 加载遮罩：数据就绪前盖住网格，禁止编辑 -->
      <div v-if="gridLoading" class="week-panel__loading">
        <span class="week-panel__spinner" aria-hidden="true" />
        <span>加载中…</span>
      </div>
    </div>
    <p class="schedule-tips">提示：拖动已编排的菜谱可以移动到其他日期或时段；点击格子空白处或「+ 添加」加菜。</p>

    <NuxtLink to="/" class="back-link">← 返回菜谱</NuxtLink>

    <!-- 菜谱选择器：搜索 + 选中即编排 -->
    <Teleport to="body">
      <div v-if="pickerOpen" class="modal-overlay" @click.self="closePicker">
        <div class="picker-modal__panel">
          <div class="picker-modal__header">
            <h3 class="modal-title">添加菜谱</h3>
            <div class="flex items-center gap-3">
              <button type="button" class="btn btn--primary btn--sm" @click="closePicker">完成</button>
              <button type="button" class="modal-close" @click="closePicker">✕</button>
            </div>
          </div>
          <p v-if="pickerCell" class="picker-modal__target">
            编排到 <span class="picker-modal__day">{{ pickerCell.dayLabel }}</span>
            · <span class="picker-modal__meal">{{ pickerCell.mealLabel }}</span>
          </p>

          <!-- 连续编排：最近一次成功提示，可继续添加 -->
          <p v-if="lastPicked" class="picker-modal__picked">✓ 已将「{{ lastPicked.title }}」编排到 {{ lastPicked.dayLabel }} {{ lastPicked.mealLabel }}，可继续添加</p>

          <div class="mt-3">
            <input
              v-model="pickerSearch"
              type="search"
              placeholder="搜索菜谱名称或简介…"
              class="input w-full"
            >
          </div>

          <p v-if="pickerError" class="error-alert mt-3">{{ pickerError }}</p>
          <p v-else-if="pickerLoading" class="empty-note mt-4 text-sm">加载中…</p>
          <p v-else-if="!pickerResults.length" class="empty-note mt-4 text-sm">没有匹配的菜谱</p>

          <ul v-else class="picker-modal__list">
            <li
              v-for="r in pickerResults"
              :key="r.id"
              class="picker-item"
              :class="addingId === r.id ? 'opacity-60' : ''"
              @click="pickRecipe(r)"
            >
              <img
                v-if="r.coverUrl"
                :src="thumbUrl(r.coverUrl)"
                :alt="r.title"
                class="picker-item__img"
                loading="lazy"
              >
              <div v-else class="picker-item__noimg">无图</div>
              <div class="min-w-0 flex-1">
                <p class="picker-item__title">{{ r.title }}</p>
                <p class="picker-item__summary">{{ r.summary }}</p>
              </div>
              <span v-if="addingId === r.id" class="picker-item__busy">编排中…</span>
            </li>
          </ul>
        </div>
      </div>
    </Teleport>
  </main>
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 页头 / 周切换 ---- */
.schedule-week__label { @apply flex items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400; }
.week-toggle { @apply rounded-md border border-zinc-300 px-3 py-1.5 text-sm transition-colors hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-800; }
.week-toggle--active { @apply border-green-600 bg-green-600 text-white hover:border-green-600 hover:bg-green-600 dark:border-green-600 dark:hover:bg-green-600; }

/* ---- 周历网格 ---- */
.week-panel { @apply mt-4 overflow-x-auto rounded-lg border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900; }
.week-panel__loading { @apply absolute inset-0 z-20 flex flex-col items-center justify-center gap-2 bg-white/80 text-sm text-zinc-500 dark:bg-zinc-900/80 dark:text-zinc-400; }
.week-panel__spinner { @apply size-6 animate-spin rounded-full border-2 border-zinc-200 border-t-green-600 dark:border-zinc-700 dark:border-t-green-500; }
.week-table { @apply w-full min-w-[760px] border-collapse text-sm; }
.week-table__head { @apply sticky left-0 z-10 border-b border-r border-zinc-100 bg-zinc-50 px-2 py-2 text-left text-xs font-medium text-zinc-600 dark:border-zinc-800 dark:bg-zinc-900 dark:text-zinc-400; }
.week-table__head--today { @apply bg-green-50 text-green-700 dark:bg-green-500/10 dark:text-green-400; }
.week-table__today-badge { @apply rounded bg-green-600 px-1 py-0.5 text-[10px] font-medium text-white; }
.week-table__meal { @apply sticky left-0 z-10 border-b border-r border-zinc-100 bg-zinc-50 px-2 py-3 text-xs font-medium text-amber-700 dark:border-zinc-800 dark:bg-zinc-900 dark:text-amber-400; }
.week-table__cell { @apply border-b border-zinc-100 align-top dark:border-zinc-800; }
.week-table__cell--today { @apply bg-green-50/40 dark:bg-green-500/5; }
.week-table__cell-body { @apply min-h-20 p-1.5 transition-colors; }
.week-table__cell-body--drop { @apply bg-green-50 ring-1 ring-inset ring-green-400 dark:bg-green-500/10 dark:ring-green-500; }

.schedule-entry { @apply mb-1 flex cursor-grab items-center gap-1.5 rounded-md border border-zinc-200 bg-white p-1 transition-shadow hover:shadow-sm active:cursor-grabbing dark:border-zinc-800 dark:bg-zinc-800; }
.schedule-entry__img { @apply size-8 shrink-0 rounded object-cover; }
.schedule-entry__noimg { @apply flex size-8 shrink-0 items-center justify-center rounded bg-zinc-100 text-[10px] text-zinc-400 dark:bg-zinc-700 dark:text-zinc-500; }
.schedule-entry__title { @apply min-w-0 flex-1 truncate text-xs font-medium text-zinc-700 dark:text-zinc-300; }
.schedule-entry__remove { @apply size-4 shrink-0 rounded-full text-[10px] leading-none text-zinc-300 hover:bg-red-50 hover:text-red-500 disabled:opacity-50 dark:text-zinc-600 dark:hover:bg-red-500/10 dark:hover:text-red-400; }
.schedule-cell__add { @apply flex h-7 w-full items-center justify-center rounded-md border border-dashed border-zinc-200 text-xs text-zinc-400 transition-colors hover:border-green-400 hover:text-green-600 dark:border-zinc-700 dark:text-zinc-500 dark:hover:border-green-500 dark:hover:text-green-400; }
.schedule-tips { @apply mt-2 text-xs text-zinc-400 dark:text-zinc-500; }

/* ---- 菜谱选择器弹窗 ---- */
.picker-modal__panel { @apply max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-lg bg-white p-5 shadow-lg dark:bg-zinc-900; }
.picker-modal__header { @apply flex items-center justify-between; }
.picker-modal__target { @apply mt-1 text-sm text-zinc-500 dark:text-zinc-400; }
.picker-modal__day { @apply font-medium text-green-700 dark:text-green-400; }
.picker-modal__meal { @apply font-medium text-amber-700 dark:text-amber-400; }
.picker-modal__picked { @apply mt-2 rounded bg-green-50 px-3 py-2 text-sm text-green-700 dark:bg-green-500/10 dark:text-green-400; }
.picker-modal__list { @apply mt-3 space-y-1.5; }
.picker-item { @apply flex cursor-pointer items-center gap-3 rounded-md border border-zinc-200 p-2 transition-colors hover:border-green-500 hover:bg-green-50 dark:border-zinc-800 dark:hover:border-green-500 dark:hover:bg-green-500/10; }
.picker-item__img { @apply size-10 shrink-0 rounded object-cover; }
.picker-item__noimg { @apply flex size-10 shrink-0 items-center justify-center rounded bg-zinc-100 text-xs text-zinc-400 dark:bg-zinc-800 dark:text-zinc-500; }
.picker-item__title { @apply truncate text-sm font-medium text-zinc-800 dark:text-zinc-200; }
.picker-item__summary { @apply truncate text-xs text-zinc-500 dark:text-zinc-400; }
.picker-item__busy { @apply text-xs text-zinc-400 dark:text-zinc-500; }
</style>
