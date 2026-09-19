<script setup lang="ts">
// 收藏夹弹窗：勾选收藏夹（复选框 + 数量）+ 新建收藏夹；ESC / 点弹窗外区域取消新建
import { apis } from '~/api'

const props = defineProps<{ open: boolean, recipeId: number | null }>()
const emit = defineEmits<{ close: [] }>()

interface FolderRow {
  id: number
  name: string
  recipeCount: number
  checked: boolean
  busy: boolean
}

const folders = ref<FolderRow[]>([])
const loading = ref(false)
const error = ref('')

// 新建收藏夹（内联输入行）
const creating = ref(false)
const newName = ref('')
const creatingFolder = ref(false)
const createError = ref('')

async function load() {
  if (!props.recipeId) return
  loading.value = true
  error.value = ''
  try {
    const [foldersRes, checkedRes] = await Promise.all([
      apis.favorite.getList(),
      apis.favorite.getByRecipe({ pathParams: { id: props.recipeId } }),
    ])
    const checked = new Set(checkedRes.favoriteIds ?? [])
    folders.value = (foldersRes.list ?? []).map(f => ({
      id: f.id!,
      name: f.name ?? '',
      recipeCount: f.recipeCount ?? 0,
      checked: checked.has(f.id!),
      busy: false,
    }))
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '加载收藏夹失败'
  }
  loading.value = false
}

watch(() => props.open, (open) => {
  if (open) {
    creating.value = false
    newName.value = ''
    createError.value = ''
    load()
  }
})

function toggleCreating() {
  creating.value = !creating.value
  if (creating.value) createError.value = ''
  else newName.value = ''
}

function cancelCreating() {
  creating.value = false
  newName.value = ''
  createError.value = ''
}

async function createFolder() {
  const name = newName.value.trim()
  if (!name || creatingFolder.value) return
  creatingFolder.value = true
  createError.value = ''
  try {
    const res = await apis.favorite.create({ body: { name } })
    newName.value = ''
    creating.value = false
    // 新夹追加进列表（新收藏的夹自然未勾选当前菜谱）
    folders.value.push({ id: res.id!, name, recipeCount: 0, checked: false, busy: false })
  }
  catch (e) {
    createError.value = e instanceof Error ? e.message : '创建失败'
  }
  creatingFolder.value = false
}

async function toggleFolder(folder: FolderRow) {
  if (!props.recipeId || folder.busy) return
  const prev = folder.checked
  folder.busy = true
  try {
    if (prev) {
      await apis.favorite.removeItem({ pathParams: { id: folder.id, recipeId: props.recipeId } })
      folder.checked = false
      folder.recipeCount = Math.max(0, folder.recipeCount - 1)
    }
    else {
      await apis.favorite.addItem({ pathParams: { id: folder.id }, body: { recipeId: props.recipeId } })
      folder.checked = true
      folder.recipeCount += 1
    }
  }
  catch (e) {
    folder.checked = prev // 失败回滚勾选态
    error.value = e instanceof Error ? e.message : '操作失败'
  }
  folder.busy = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && creating.value) cancelCreating()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-black/40 p-4"
      @click.self="emit('close')"
    >
      <div class="w-full max-w-sm rounded-lg bg-white shadow-lg">
        <div class="flex items-center justify-between border-b border-zinc-200 px-4 py-3">
          <h3 class="text-base font-semibold">收藏到收藏夹</h3>
          <button type="button" class="text-zinc-400 hover:text-zinc-600" @click="emit('close')">✕</button>
        </div>

        <div class="max-h-72 min-h-24 overflow-y-auto px-2 py-1">
          <p v-if="loading" class="py-8 text-center text-sm text-zinc-400">加载中…</p>
          <template v-else>
            <p v-if="error" class="m-2 rounded bg-red-50 px-3 py-2 text-xs text-red-600">{{ error }}</p>
            <p v-if="!folders.length" class="py-6 text-center text-xs text-zinc-400">暂无收藏夹</p>
            <label
              v-for="folder in folders"
              :key="folder.id"
              class="flex cursor-pointer items-center gap-2 rounded px-2 py-2 hover:bg-zinc-50"
            >
              <input
                type="checkbox"
                class="accent-green-600"
                :checked="folder.checked"
                :disabled="folder.busy"
                @change="toggleFolder(folder)"
              >
              <span class="min-w-0 flex-1 truncate text-sm text-zinc-700">{{ folder.name }}</span>
              <span v-if="folder.busy" class="text-xs text-zinc-400">…</span>
              <span v-else class="text-xs text-zinc-400">{{ folder.recipeCount }}</span>
            </label>
          </template>
        </div>

        <!-- 新建收藏夹：点击弹窗其他区域或 ESC 取消 -->
        <div class="border-t border-zinc-200 p-3">
          <template v-if="creating">
            <p v-if="createError" class="mb-2 text-xs text-red-500">{{ createError }}</p>
            <div class="flex gap-2">
              <input
                ref="newNameInput"
                v-model="newName"
                type="text"
                placeholder="新收藏夹名称"
                class="flex-1 rounded border border-zinc-300 px-2 py-1.5 text-sm outline-none focus:border-green-600"
                @keyup.enter="createFolder"
                @blur="cancelCreating"
              >
              <button
                type="button"
                class="rounded bg-green-600 px-3 py-1.5 text-sm text-white hover:bg-green-700 disabled:opacity-50"
                :disabled="creatingFolder || !newName.trim()"
                @click="createFolder"
              >{{ creatingFolder ? '…' : '创建' }}</button>
            </div>
          </template>
          <button
            v-else
            type="button"
            class="w-full rounded border border-dashed border-zinc-300 py-2 text-sm text-zinc-500 hover:border-green-500 hover:text-green-600"
            @click="toggleCreating"
          >+ 新建收藏夹</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>