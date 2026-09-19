<script setup lang="ts">
// 成员管理页：家庭成员（角色/备注）查看与编辑
import { apis } from '~/api'
import type { Cookbook_internal_model_member_item } from '~/api/components'

// 预置角色（可自定义，后端 role 为自由文本）
const presetRoles = ['爸爸', '妈妈', '儿子', '妻子', '妹妹']

interface MemberForm {
  id: number // 0=新增
  name: string
  role: string
  customRole: string
  note: string
}

const members = ref<Cookbook_internal_model_member_item[]>([])
const listError = ref('')
const listLoading = ref(false)

const formOpen = ref(false)
const form = ref<MemberForm | null>(null)
const saving = ref(false)
const saveError = ref('')

const deletingId = ref<number | null>(null)

async function loadMembers() {
  listLoading.value = true
  listError.value = ''
  try {
    const res = await apis.member.getList()
    members.value = res.list ?? []
  }
  catch (e) {
    listError.value = e instanceof Error ? e.message : '加载成员失败'
  }
  listLoading.value = false
}

onMounted(loadMembers)

function openCreate() {
  saveError.value = ''
  form.value = { id: 0, name: '', role: '', customRole: '', note: '' }
  formOpen.value = true
}

function openEdit(m: Cookbook_internal_model_member_item) {
  saveError.value = ''
  const isPreset = presetRoles.includes(m.role ?? '')
  form.value = {
    id: m.id!,
    name: m.name ?? '',
    role: isPreset ? m.role! : '',
    customRole: isPreset ? '' : (m.role ?? ''),
    note: m.note ?? '',
  }
  formOpen.value = true
}

function closeForm() {
  formOpen.value = false
  form.value = null
  saveError.value = ''
}

function toggleRole(role: string) {
  if (!form.value) return
  form.value.role = form.value.role === role ? '' : role
  if (form.value.role) form.value.customRole = '' // 预置优先，选预置清自定义
}

async function saveForm() {
  if (!form.value) return
  if (!form.value.name.trim()) {
    saveError.value = '成员名字不能为空'
    return
  }
  saving.value = true
  saveError.value = ''
  const role = (form.value.customRole.trim() || form.value.role).trim()
  const body = {
    name: form.value.name.trim(),
    role,
    note: form.value.note,
  }
  try {
    if (form.value.id > 0) {
      await apis.member.update({ pathParams: { id: form.value.id }, body })
    }
    else {
      await apis.member.create({ body })
    }
    formOpen.value = false
    form.value = null
    await loadMembers()
  }
  catch (e) {
    saveError.value = e instanceof Error ? e.message : '保存失败'
  }
  saving.value = false
}

async function removeMember(m: Cookbook_internal_model_member_item) {
  listError.value = ''
  deletingId.value = m.id!
  try {
    await apis.member.delete({ pathParams: { id: m.id! } })
    await loadMembers()
  }
  catch (e) {
    listError.value = e instanceof Error ? e.message : '删除失败'
  }
  deletingId.value = null
}
</script>

<template>
  <main class="page page--narrow">
    <div class="page-header">
      <h1 class="page-title">家庭成员</h1>
      <button type="button" class="btn btn--primary" @click="openCreate">+ 新增成员</button>
    </div>

    <!-- 表单 -->
    <div v-if="formOpen && form" class="card mt-4 p-4">
      <p v-if="saveError" class="error-alert mb-3">{{ saveError }}</p>

      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="form-label">名字 <span class="text-red-500 dark:text-red-400">*</span></label>
          <input
            v-model="form.name"
            type="text"
            placeholder="如 小明"
            class="input input--sm w-full"
          >
        </div>
        <div>
          <label class="form-label">自定义角色 <span class="form-hint">（填写则优先于下面选择）</span></label>
          <input
            v-model="form.customRole"
            type="text"
            class="input input--sm w-full"
          >
        </div>
      </div>

      <div class="mt-3">
        <label class="form-label">角色</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="role in presetRoles"
            :key="role"
            type="button"
            class="chip chip--green"
            :class="{ 'chip--active': form.role === role }"
            @click="toggleRole(role)"
          >{{ role }}</button>
        </div>
      </div>

      <div class="mt-3">
        <label class="form-label">备注 <span class="form-hint">（口味、忌口、过敏等）</span></label>
        <textarea
          v-model="form.note"
          rows="2"
          class="input input--sm w-full"
        />
      </div>

      <div class="member-form__footer">
        <button type="button" class="btn btn--outline" @click="closeForm">取消</button>
        <button type="button" class="btn btn--primary" :disabled="saving" @click="saveForm">{{ saving ? '保存中…' : '保存' }}</button>
      </div>
    </div>

    <!-- 列表 -->
    <p v-if="listError" class="error-alert mt-4">{{ listError }}</p>
    <p v-else-if="listLoading && !members.length" class="empty-note mt-8 text-sm">加载中…</p>
    <p v-else-if="!members.length" class="empty-note mt-8 text-sm">还没有添加成员</p>
    <ul v-else class="member-list">
      <li v-for="m in members" :key="m.id" class="card p-4">
        <div class="flex items-center gap-2">
          <span class="tag-badge tag-badge--green">{{ m.role || '成员' }}</span>
          <span class="member-card__name">{{ m.name }}</span>
          <button type="button" class="text-btn text-btn--edit text-btn--xs" @click="openEdit(m)">编辑</button>
          <button
            type="button"
            class="text-btn text-btn--delete text-btn--xs"
            :disabled="deletingId === m.id"
            @click="removeMember(m)"
          >{{ deletingId === m.id ? '删除中…' : '删除' }}</button>
        </div>
        <p v-if="m.note" class="member-card__note">{{ m.note }}</p>
      </li>
    </ul>

    <NuxtLink to="/" class="back-link">← 返回菜谱</NuxtLink>
  </main>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.member-form__footer { @apply mt-4 flex justify-end gap-2; }
.member-list { @apply mt-4 space-y-2; }
.member-card__name { @apply min-w-0 flex-1 truncate text-sm font-medium text-zinc-800 dark:text-zinc-200; }
.member-card__note { @apply mt-1.5 text-xs text-zinc-500 dark:text-zinc-400; }
</style>
