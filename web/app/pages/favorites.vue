<script setup lang="ts">
// 收藏页：收藏夹管理 + 夹内食谱查看/移除
import { apis } from '~/api'
import type { Cookbook_internal_model_favorite_folder, Cookbook_internal_model_recipe_list_item } from '~/api/components'

const folders = ref<Cookbook_internal_model_favorite_folder[]>([])
const foldersError = ref('')
const foldersLoading = ref(false)

const expandedId = ref<number | null>(null)
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
</script>

<template>
  <main class="page page--narrow">
    <h1 class="page-title">收藏</h1>

    <p v-if="foldersError" class="error-alert mt-4">{{ foldersError }}</p>
    <p v-else-if="foldersLoading && !folders.length" class="empty-note mt-8 text-sm">加载中…</p>
    <p v-else-if="!folders.length" class="empty-note mt-8 text-sm">暂无收藏夹</p>

    <div v-else class="mt-4 space-y-3">
      <div v-for="folder in folders" :key="folder.id" class="card p-4">
        <!-- 夹头 -->
        <div class="folder-card__head">
          <button type="button" class="folder-card__toggle" @click="toggleFolder(folder.id!)">
            <span class="folder-card__name">{{ folder.name }}</span>
            <span class="folder-card__count">{{ folder.recipeCount }} 个菜谱</span>
          </button>
          <button type="button" class="text-btn text-btn--edit" @click="startRename(folder)">重命名</button>
          <button
            type="button"
            class="text-btn text-btn--delete"
            :disabled="deletingId === folder.id"
            @click="removeFolder(folder.id!)"
          >{{ deletingId === folder.id ? '删除中…' : '删除' }}</button>
        </div>
        <!-- 重命名行 -->
        <div v-if="renamingId === folder.id" class="folder-card__rename">
          <input
            v-model="renameValue"
            type="text"
            class="input input--sm flex-1"
            @keyup.enter="confirmRename"
          >
          <button
            type="button"
            class="btn btn--primary btn--sm"
            :disabled="renaming"
            @click="confirmRename"
          >{{ renaming ? '保存中…' : '保存' }}</button>
          <button type="button" class="text-btn text-btn--muted" @click="renamingId = null">取消</button>
        </div>
        <!-- 夹内食谱网格 -->
        <div v-if="expandedId === folder.id" class="mt-3">
          <p v-if="recipesLoading" class="empty-note py-3 text-sm">加载中…</p>
          <p v-else-if="recipesError" class="error-text py-3">{{ recipesError }}</p>
          <p v-else-if="!recipes.length" class="empty-note py-3 text-sm">还没有收藏菜谱</p>
          <div v-else class="folder-card__grid">
            <div v-for="recipe in recipes" :key="recipe.id" class="folder-recipe">
              <div class="folder-recipe__thumb">
                <img
                  v-if="recipe.coverUrl"
                  :src="thumbUrl(recipe.coverUrl)"
                  :alt="recipe.title"
                  class="folder-recipe__img"
                  loading="lazy"
                >
                <div v-else class="folder-recipe__empty">暂无封面</div>
              </div>
              <div class="folder-recipe__body">
                <h3 class="folder-recipe__title">{{ recipe.title }}</h3>
                <button
                  type="button"
                  class="text-btn text-btn--delete text-btn--xs mt-1"
                  :disabled="removingRecipeId === recipe.id"
                  @click="removeRecipe(folder.id!, recipe.id!)"
                >{{ removingRecipeId === recipe.id ? '移除中…' : '移除收藏' }}</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建收藏夹 -->
    <div class="mt-6 flex gap-2">
      <p v-if="createError" class="error-text mr-2 self-center">{{ createError }}</p>
      <input
        v-model="newFolderName"
        type="text"
        placeholder="新收藏夹名称"
        class="input flex-1"
        @keyup.enter="createFolder"
      >
      <button
        type="button"
        class="btn btn--primary"
        :disabled="creating || !newFolderName.trim()"
        @click="createFolder"
      >{{ creating ? '创建中…' : '新建收藏夹' }}</button>
    </div>

    <NuxtLink to="/" class="back-link">← 返回菜谱</NuxtLink>
  </main>
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 收藏夹 ---- */
.folder-card__head { @apply flex items-center gap-2; }
.folder-card__toggle { @apply min-w-0 flex-1 text-left; }
.folder-card__name { @apply text-base font-medium text-zinc-800 dark:text-zinc-200; }
.folder-card__count { @apply ml-2 text-sm text-zinc-400 dark:text-zinc-500; }
.folder-card__rename { @apply mt-2 flex gap-2; }
.folder-card__grid { @apply grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4; }

/* ---- 夹内食谱小卡 ---- */
.folder-recipe { @apply overflow-hidden rounded-lg border border-zinc-200 dark:border-zinc-800; }
.folder-recipe__thumb { @apply relative w-full; aspect-ratio: 650 / 370; }
.folder-recipe__img { @apply absolute inset-0 h-full w-full object-cover object-center; }
.folder-recipe__empty { @apply absolute inset-0 flex items-center justify-center bg-zinc-100 text-xs text-zinc-400 dark:bg-zinc-800 dark:text-zinc-500; }
.folder-recipe__body { @apply p-2; }
.folder-recipe__title { @apply truncate text-sm font-semibold; }
</style>
