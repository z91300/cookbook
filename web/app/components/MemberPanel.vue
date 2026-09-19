<script setup lang="ts">
// 成员面板：家庭成员管理（角色/备注），供人工筛选与后续 AI 排每周食谱参考
import { apis } from '~/api'
import type { Cookbook_internal_model_member_item } from '~/api/components'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

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

watch(() => props.open, (open) => {
  if (open) {
    closeForm()
    loadMembers()
  }
}, { immediate: true })

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
  <div class="flex max-h-[60vh] flex-col">
    <div class="member-panel__top">
      <button
        v-if="!formOpen"
        type="button"
        class="text-btn text-btn--accent"
        @click="openCreate"
      >+ 新增成员</button>
      <span v-else class="member-panel__form-title">{{ form?.id ? '编辑成员' : '新增成员' }}</span>
      <button v-if="formOpen" type="button" class="text-btn text-btn--muted text-btn--xs" @click="closeForm">返回列表</button>
    </div>

    <!-- 表单 -->
    <div v-if="formOpen && form" class="flex-1 space-y-3 overflow-y-auto p-3">
      <p v-if="saveError" class="error-alert">{{ saveError }}</p>

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
        <input
          v-model="form.customRole"
          type="text"
          placeholder="自定义角色（填写则优先于上面选择）"
          class="input input--sm mt-2 w-full"
        >
      </div>

      <div>
        <label class="form-label">备注 <span class="form-hint">（口味、忌口、过敏等）</span></label>
        <textarea
          v-model="form.note"
          rows="2"
          class="input input--sm w-full"
        />
      </div>

      <div class="member-panel__footer">
        <button type="button" class="btn btn--outline btn--sm" @click="closeForm">取消</button>
        <button type="button" class="btn btn--primary btn--sm" :disabled="saving" @click="saveForm">{{ saving ? '保存中…' : '保存' }}</button>
      </div>
    </div>

    <!-- 列表 -->
    <div v-else class="flex-1 overflow-y-auto px-2 py-1">
      <p v-if="listError" class="error-alert m-3">{{ listError }}</p>
      <p v-else-if="listLoading && !members.length" class="empty-note py-6 text-sm">加载中…</p>
      <p v-else-if="!members.length" class="empty-note py-6 text-sm">还没有添加成员</p>
      <ul v-else>
        <li v-for="m in members" :key="m.id" class="member-panel__item">
          <div class="flex items-center gap-2">
            <span class="tag-badge tag-badge--green">{{ m.role || '成员' }}</span>
            <span class="member-panel__name">{{ m.name }}</span>
            <button type="button" class="text-btn text-btn--edit text-btn--xs" @click="openEdit(m)">编辑</button>
            <button
              type="button"
              class="text-btn text-btn--delete text-btn--xs"
              :disabled="deletingId === m.id"
              @click="removeMember(m)"
            >{{ deletingId === m.id ? '删除中…' : '删除' }}</button>
          </div>
          <p v-if="m.note" class="member-panel__note">{{ m.note }}</p>
        </li>
      </ul>

      <!-- 列表态底部添加入口 -->
      <div v-if="!formOpen" class="p-3">
        <button type="button" class="member-panel__add-bottom" @click="openCreate">+ 新增成员</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "~/assets/css/main.css";

.member-panel__top { @apply flex items-center justify-between px-3 pt-2; }
.member-panel__form-title { @apply text-sm font-medium text-zinc-600 dark:text-zinc-400; }
.member-panel__footer { @apply flex justify-end gap-2 pb-1; }
.member-panel__item { @apply border-b border-zinc-100 px-1 py-2.5 last:border-0 dark:border-zinc-800; }
.member-panel__name { @apply min-w-0 flex-1 truncate text-sm font-medium text-zinc-800 dark:text-zinc-200; }
.member-panel__note { @apply mt-1 text-xs text-zinc-500 dark:text-zinc-400; }
.member-panel__add-bottom { @apply w-full rounded-md border border-dashed border-zinc-300 py-2 text-sm text-zinc-500 transition-colors hover:border-green-500 hover:text-green-600 dark:border-zinc-700 dark:text-zinc-400 dark:hover:border-green-500 dark:hover:text-green-400; }
</style>
