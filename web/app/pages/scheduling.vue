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
const loading = ref(false)
const loadError = ref('')

async function loadWeek() {
  loading.value = true
  loadError.value = ''
  const start = weekDates.value[0].yyyymmdd
  const end = weekDates.value[6].yyyymmdd
  try {
    const res = await apis.scheduling.getList({ params: { startDate: start, endDate: end } })
    entries.value = res.list ?? []
  }
  catch (e) {
    loadError.value = e instanceof Error ? e.message : '加载编排失败'
  }
  loading.value = false
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
    await loadWeek()
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
  searchRecipes()
}

function closePicker() {
  pickerOpen.value = false
  pickerCell.value = null
  addingId.value = null
}

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
    await loadWeek()
    closePicker()
  }
  catch (e) {
    pickerError.value = e instanceof Error ? e.message : '编排失败'
  }
}
</script>

<template>
  <main class="mx-auto max-w-6xl px-4 py-8">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">食谱编排</h1>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-md border border-zinc-300 px-3 py-1.5 text-sm hover:bg-zinc-50"
          @click="prevWeek"
        >上一周</button>
        <button
          type="button"
          class="rounded-md border border-zinc-300 px-3 py-1.5 text-sm hover:bg-zinc-50"
          :disabled="weekOffset === 0"
          @click="thisWeek"
        >本周</button>
        <button
          type="button"
          class="rounded-md border border-zinc-300 px-3 py-1.5 text-sm hover:bg-zinc-50"
          @click="nextWeek"
        >下一周</button>
      </div>
    </div>

    <p class="mt-2 text-sm text-zinc-500">{{ weekLabel }}</p>

    <p v-if="loadError" class="mt-4 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ loadError }}</p>
    <p v-if="loading && !entries.length" class="mt-4 text-center text-sm text-zinc-400">加载中…</p>

    <!-- 周历网格：行=用餐时段，列=周一~周日 -->
    <div class="mt-4 overflow-x-auto">
      <table class="w-full min-w-[760px] border-collapse text-sm">
        <thead>
          <tr>
            <th class="w-20 border border-zinc-200 bg-zinc-50 px-2 py-2 text-left text-xs font-medium text-zinc-500">
              时段 / 日期
            </th>
            <th
              v-for="d in weekDates"
              :key="d.key"
              class="border border-zinc-200 bg-zinc-50 px-2 py-2 text-left text-xs font-medium"
              :class="isToday(d.date) ? 'text-green-700' : 'text-zinc-600'"
            >
              {{ d.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="slot in mealSlots" :key="slot.key">
            <td class="border border-zinc-200 bg-zinc-50 px-2 py-3 text-xs font-medium text-amber-700">
              {{ slot.label }}
            </td>
            <td
              v-for="d in weekDates"
              :key="d.key"
              class="border border-zinc-200 align-top"
            >
              <div class="min-h-16 p-1.5">
                <!-- 已编排条目（一餐可多菜） -->
                <div
                  v-for="e in cellEntries(d.yyyymmdd, slot.value)"
                  :key="e.id"
                  class="mb-1 flex items-start gap-2 rounded bg-zinc-50 p-1"
                >
                  <img
                    v-if="e.coverUrl"
                    :src="thumbUrl(e.coverUrl)"
                    :alt="e.recipeTitle"
                    class="size-8 shrink-0 rounded object-cover"
                    loading="lazy"
                  >
                  <div v-else class="flex size-8 shrink-0 items-center justify-center rounded bg-zinc-200 text-[10px] text-zinc-500">无图</div>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-xs font-medium text-zinc-700">{{ e.recipeTitle }}</p>
                    <button
                      v-if="e.id"
                      type="button"
                      class="text-[10px] text-zinc-400 hover:text-red-500"
                      :disabled="removingId === e.id"
                      @click="removeEntry(e.id)"
                    >{{ removingId === e.id ? '移除中…' : '移除' }}</button>
                  </div>
                </div>
                <!-- 添加按钮 -->
                <button
                  type="button"
                  class="flex h-7 w-full items-center justify-center rounded text-zinc-300 hover:bg-zinc-100 hover:text-green-600"
                  @click="openPicker(d.yyyymmdd, slot.value, d.label, slot.label)"
                >+ 添加</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <NuxtLink to="/" class="mt-6 inline-block text-sm text-zinc-500 hover:text-zinc-700">← 返回菜谱</NuxtLink>

    <!-- 菜谱选择器：搜索 + 选中即编排 -->
    <Teleport to="body">
      <div v-if="pickerOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" @click.self="closePicker">
        <div class="max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-lg bg-white p-5 shadow-lg">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">添加菜谱</h3>
            <button type="button" class="text-zinc-400 hover:text-zinc-600" @click="closePicker">✕</button>
          </div>
          <p v-if="pickerCell" class="mt-1 text-sm text-zinc-500">
            编排到 <span class="font-medium text-green-700">{{ pickerCell.dayLabel }}</span>
            · <span class="font-medium text-amber-700">{{ pickerCell.mealLabel }}</span>
          </p>

          <div class="mt-3">
            <input
              v-model="pickerSearch"
              type="search"
              placeholder="搜索菜谱名称或简介…"
              class="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-600"
            >
          </div>

          <p v-if="pickerError" class="mt-3 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ pickerError }}</p>
          <p v-else-if="pickerLoading" class="mt-4 text-center text-sm text-zinc-400">加载中…</p>
          <p v-else-if="!pickerResults.length" class="mt-4 text-center text-sm text-zinc-400">没有匹配的菜谱</p>

          <ul v-else class="mt-3 space-y-1.5">
            <li
              v-for="r in pickerResults"
              :key="r.id"
              class="flex cursor-pointer items-center gap-3 rounded-md border border-zinc-200 p-2 hover:border-green-500 hover:bg-green-50"
              :class="addingId === r.id ? 'opacity-60' : ''"
              @click="pickRecipe(r)"
            >
              <img
                v-if="r.coverUrl"
                :src="thumbUrl(r.coverUrl)"
                :alt="r.title"
                class="size-10 shrink-0 rounded object-cover"
                loading="lazy"
              >
              <div v-else class="flex size-10 shrink-0 items-center justify-center rounded bg-zinc-100 text-xs text-zinc-400">无图</div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-zinc-800">{{ r.title }}</p>
                <p class="truncate text-xs text-zinc-500">{{ r.summary }}</p>
              </div>
              <span v-if="addingId === r.id" class="text-xs text-zinc-400">编排中…</span>
            </li>
          </ul>
        </div>
      </div>
    </Teleport>
  </main>
</template>
