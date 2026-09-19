<script setup lang="ts">
// 收藏面板：收藏夹列表 + 夹内食谱管理（右下角浮动面板内嵌）
import { apis } from '~/api'
import type { Cookbook_internal_model_favorite_folder, Cookbook_internal_model_recipe_list_item } from '~/api/components'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const folders = ref<Cookbook_internal_model_favorite_folder[]>([])
const foldersError = ref('')
const foldersLoading = ref(false)

const expandedId = ref<number | null>(null) // 展开查看夹内食谱
const recipes = ref<Cookbook_internal_model_recipe_list_item[]>([])
const recipesLoading = ref(false)
const recipesError = ref('')

const newFolderName = ref('')
const creating = ref(false)
const createError = ref('')

const renamingId = ref<number | null>(null)
const renameValue = ref('')
const renaming = ref(false)

const removingRecipeId = ref<number | null>(null)
const deletingId = ref<number | null>(null)

async function loadFolders() {
  foldersLoading.value = true
  foldersError.value = ''
  try {
    const res = await apis.favorite.getList()
    folders.value = res.list ?? []
    // 当前展开的夹被删除后收起
    if (expandedId.value && !folders.value.some(f => f.id === expandedId.value)) {
      expandedId.value = null
      recipes.value = []
    }
  }
  catch (e) {
    foldersError.value = e instanceof Error ? e.message : '加载收藏夹失败'
  }
  foldersLoading.value = false
}

watch(() => props.open, (open) => {
  if (open) {
    expandedId.value = null
    recipes.value = []
    loadFolders()
  }
}, { immediate: true })

async function toggleFolder(id: number) {
  if (expandedId.value === id) {
    expandedId.value = null
    recipes.value = []
    return
  }
  expandedId.value = id
  recipesLoading.value = true
  recipesError.value = ''
  try {
    const res = await apis.favorite.getRecipes({ pathParams: { id } })
    recipes.value = res.list ?? []
  }
  catch (e) {
    recipesError.value = e instanceof Error ? e.message : '加载收藏失败'
  }
  recipesLoading.value = false
}

async function createFolder() {
  const name = newFolderName.value.trim()
  if (!name) return
  creating.value = true
  createError.value = ''
  try {
    await apis.favorite.create({ body: { name } })
    newFolderName.value = ''
    await loadFolders()
  }
  catch (e) {
    createError.value = e instanceof Error ? e.message : '创建失败'
  }
  creating.value = false
}

function startRename(folder: Cookbook_internal_model_favorite_folder) {
  renamingId.value = folder.id!
  renameValue.value = folder.name ?? ''
}

async function confirmRename() {
  const id = renamingId.value
  const name = renameValue.value.trim()
  if (id == null || !name) return
  renaming.value = true
  try {
    await apis.favorite.update({ pathParams: { id }, body: { name } })
    renamingId.value = null
    await loadFolders()
  }
  catch (e) {
    createError.value = e instanceof Error ? e.message : '重命名失败'
  }
  renaming.value = false
}

async function removeFolder(id: number) {
  deletingId.value = id
  try {
    await apis.favorite.delete({ pathParams: { id } })
    await loadFolders()
  }
  catch (e) {
    foldersError.value = e instanceof Error ? e.message : '删除失败'
  }
  deletingId.value = null
}

async function removeRecipe(folderId: number, recipeId: number) {
  removingRecipeId.value = recipeId
  try {
    await apis.favorite.removeItem({ pathParams: { id: folderId, recipeId } })
    recipes.value = recipes.value.filter(r => r.id !== recipeId)
    await loadFolders() // 刷新 recipeCount
  }
  catch (e) {
    recipesError.value = e instanceof Error ? e.message : '移除失败'
  }
  removingRecipeId.value = null
}

/** 通知父页面（详情收藏浮层勾选后）刷新计数 */
defineExpose({ loadFolders })
</script>

<template>
  <div class="flex max-h-[60vh] flex-col">
    <p v-if="foldersError" class="error-alert m-3">{{ foldersError }}</p>
    <p v-else-if="foldersLoading && !folders.length" class="empty-note py-6 text-sm">加载中…</p>
    <p v-else-if="!folders.length" class="empty-note py-6 text-sm">暂无收藏夹</p>

    <div v-else class="flex-1 overflow-y-auto px-2 py-1">
      <div v-for="folder in folders" :key="folder.id" class="fav-panel__folder">
        <!-- 夹行 -->
        <div class="fav-panel__row">
          <button type="button" class="fav-panel__toggle" @click="toggleFolder(folder.id!)">
            <span class="fav-panel__name">{{ folder.name }}</span>
            <span class="fav-panel__count">{{ folder.recipeCount }} 个菜谱</span>
          </button>
          <button type="button" class="text-btn text-btn--edit text-btn--xs" @click="startRename(folder)">重命名</button>
          <button
            type="button"
            class="text-btn text-btn--delete text-btn--xs"
            :disabled="deletingId === folder.id"
            @click="removeFolder(folder.id!)"
          >{{ deletingId === folder.id ? '删除中…' : '删除' }}</button>
          <span class="fav-panel__arrow">{{ expandedId === folder.id ? '▾' : '▸' }}</span>
        </div>
        <!-- 重命名行 -->
        <div v-if="renamingId === folder.id" class="flex gap-2 pb-2">
          <input
            v-model="renameValue"
            type="text"
            class="input input--sm flex-1"
            @keyup.enter="confirmRename"
          >
          <button
            type="button"
            class="btn btn--primary btn--xs"
            :disabled="renaming"
            @click="confirmRename"
          >{{ renaming ? '保存中…' : '保存' }}</button>
          <button type="button" class="text-btn text-btn--muted text-btn--xs" @click="renamingId = null">取消</button>
        </div>
        <!-- 夹内食谱 -->
        <div v-if="expandedId === folder.id" class="pb-2">
          <p v-if="recipesLoading" class="empty-note py-2 text-xs">加载中…</p>
          <p v-else-if="recipesError" class="error-text py-2 text-xs">{{ recipesError }}</p>
          <p v-else-if="!recipes.length" class="empty-note py-2 text-xs">还没有收藏菜谱</p>
          <ul v-else class="space-y-1 pb-1">
            <li v-for="recipe in recipes" :key="recipe.id" class="fav-panel__recipe">
              <img
                v-if="recipe.coverUrl"
                :src="thumbUrl(recipe.coverUrl)"
                :alt="recipe.title"
                class="fav-panel__recipe-img"
                loading="lazy"
              >
              <div v-else class="fav-panel__recipe-empty" />
              <span class="fav-panel__recipe-name">{{ recipe.title }}</span>
              <button
                type="button"
                class="text-btn text-btn--delete text-btn--xs"
                :disabled="removingRecipeId === recipe.id"
                @click="removeRecipe(folder.id!, recipe.id!)"
              >{{ removingRecipeId === recipe.id ? '移除中…' : '移除' }}</button>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 新建收藏夹 -->
    <div class="fav-panel__footer">
      <p v-if="createError" class="error-text mb-2 text-xs">{{ createError }}</p>
      <div class="flex gap-2">
        <input
          v-model="newFolderName"
          type="text"
          placeholder="新收藏夹名称"
          class="input input--sm flex-1"
          @keyup.enter="createFolder"
        >
        <button
          type="button"
          class="btn btn--primary btn--sm"
          :disabled="creating || !newFolderName.trim()"
          @click="createFolder"
        >{{ creating ? '…' : '添加' }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.fav-panel__folder { @apply border-b border-zinc-100 last:border-0 dark:border-zinc-800; }
.fav-panel__row { @apply flex items-center gap-2 py-2.5; }
.fav-panel__toggle { @apply min-w-0 flex-1 text-left; }
.fav-panel__name { @apply text-sm font-medium text-zinc-800 dark:text-zinc-200; }
.fav-panel__count { @apply ml-2 text-xs text-zinc-400 dark:text-zinc-500; }
.fav-panel__arrow { @apply text-xs text-zinc-300 dark:text-zinc-600; }
.fav-panel__recipe { @apply flex items-center gap-2 rounded px-1 py-1 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800; }
.fav-panel__recipe-img { @apply h-8 w-6 rounded-sm border border-zinc-200 object-cover dark:border-zinc-700; }
.fav-panel__recipe-empty { @apply h-8 w-6 rounded-sm border border-dashed border-zinc-300 dark:border-zinc-700; }
.fav-panel__recipe-name { @apply min-w-0 flex-1 truncate text-sm text-zinc-700 dark:text-zinc-300; }
.fav-panel__footer { @apply border-t border-zinc-200 p-3 dark:border-zinc-800; }
</style>
