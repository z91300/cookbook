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

// 弹窗打开期间锁定页面滚动 + 移动端侧滑返回关闭弹窗
useBodyScrollLock([() => props.open])
useModalBackClose(() => props.open, () => emit('close'))

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
    <div v-if="open" class="modal-overlay modal-overlay--top" @click.self="emit('close')">
      <div class="modal-panel max-w-sm">
        <div class="fav-modal__header">
          <h3 class="modal-title--sm">收藏到收藏夹</h3>
          <button type="button" class="modal-close" @click="emit('close')">✕</button>
        </div>

        <div class="fav-modal__list">
          <p v-if="loading" class="empty-note py-8 text-sm">加载中…</p>
          <template v-else>
            <p v-if="error" class="error-alert m-2 text-xs">{{ error }}</p>
            <p v-if="!folders.length" class="empty-note py-6 text-xs">暂无收藏夹</p>
            <label
              v-for="folder in folders"
              :key="folder.id"
              class="fav-modal__folder"
            >
              <input
                type="checkbox"
                class="accent-green-600"
                :checked="folder.checked"
                :disabled="folder.busy"
                @change="toggleFolder(folder)"
              >
              <span class="fav-modal__folder-name">{{ folder.name }}</span>
              <span v-if="folder.busy" class="fav-modal__folder-count">…</span>
              <span v-else class="fav-modal__folder-count">{{ folder.recipeCount }}</span>
            </label>
          </template>
        </div>

        <!-- 新建收藏夹：点击弹窗其他区域或 ESC 取消 -->
        <div class="fav-modal__footer">
          <template v-if="creating">
            <p v-if="createError" class="error-text mb-2 text-xs">{{ createError }}</p>
            <div class="flex gap-2">
              <input
                ref="newNameInput"
                v-model="newName"
                type="text"
                placeholder="新收藏夹名称"
                class="input input--sm flex-1"
                @keyup.enter="createFolder"
                @blur="cancelCreating"
              >
              <button
                type="button"
                class="btn btn--primary btn--sm"
                :disabled="creatingFolder || !newName.trim()"
                @click="createFolder"
              >{{ creatingFolder ? '…' : '创建' }}</button>
            </div>
          </template>
          <button v-else type="button" class="fav-modal__create" @click="toggleCreating">
            <PlusIcon class="size-3.5" />
            新建收藏夹
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.fav-modal__header { @apply flex items-center justify-between border-b border-zinc-200 px-4 py-3 dark:border-zinc-800; }
.fav-modal__list { @apply max-h-72 min-h-24 overflow-y-auto px-2 py-1; }
.fav-modal__folder { @apply flex cursor-pointer items-center gap-2 rounded px-2 py-2 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800; }
.fav-modal__folder-name { @apply min-w-0 flex-1 truncate text-sm text-zinc-700 dark:text-zinc-300; }
.fav-modal__folder-count { @apply text-xs text-zinc-400 dark:text-zinc-500; }
.fav-modal__footer { @apply border-t border-zinc-200 p-3 dark:border-zinc-800; }
.fav-modal__create { @apply flex w-full items-center justify-center gap-1 rounded border border-dashed border-zinc-300 py-2 text-sm text-zinc-500 transition-colors hover:border-green-500 hover:text-green-600 dark:border-zinc-700 dark:text-zinc-400 dark:hover:border-green-500 dark:hover:text-green-400; }
</style>
