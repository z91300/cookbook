<script setup lang="ts">
// 设置页：通用 key-value 设置项管理（如 AI 生图提示词模板）
import { apis } from '~/api'

interface SettingItem {
  key: string
  value: string
}

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const savedTip = ref('')
// 编辑态：key -> value（key 不可改，value 可编辑）
const entries = ref<SettingItem[]>([])

// 提示词模板可用变量说明（与后端 renderTemplate 支持的 {recipe.*} 一致）
const templateVars = [  { name: '{recipe.title}', desc: '菜谱标题' },
  { name: '{recipe.summary}', desc: '一句话简介' },
  { name: '{recipe.ingredients}', desc: '食材列表（逐行「名称 用量」）' },
  { name: '{recipe.steps}', desc: '步骤列表（逐行 ①②③ 编号）' },
  { name: '{recipe.tools}', desc: '工具列表（逐行）' },
  { name: '{recipe.tips}', desc: '小贴士（空时输出省略提示）' },
  { name: '{recipe.stepCount}', desc: '步骤条数，如 5' },
  { name: '{recipe.ingredientCount}', desc: '食材条数，如 11' },
  { name: '{recipe.infoLine}', desc: '信息条，如「约 100 分钟 · 3 人份 · 难度中等」' },
  { name: '{recipe.calories}', desc: '热量，如 160 kcal/份（未填为空）' },
  { name: '{recipe.difficulty}', desc: '难度：简单/中等/较难（未填为空）' },
  { name: '{recipe.servings}', desc: '份量，如 3 人份（未填为空）' },
  { name: '{recipe.cookMinutes}', desc: '耗时，如 90 分钟（未填为空）' },
  { name: '{recipe.mealMask}', desc: '用餐时段，如 早餐、午餐（未填为空）' },
]

const KEY_IMAGE_GEN_PROMPT = 'image_gen_prompt'
const IMAGE_GEN_PROMPT_HINT = [
  '帮我生成一张 {recipe.title} 的菜谱卡片图片，最终必须输出一张完整的 PNG 格式单张图片。',
  '**提示词目标：** 生成一张严格 9:16 纵向比例的简约信息图表式菜谱卡片；单张图像一次性直出（不要分镜、不要多图、不要拼图、不要前后对比、不要附带文字说明页），输出 PNG 格式。',
  '**0. 画布与高度分配（硬约束，最高优先级）**',
  '* 单图直出：仅输出这一张 9:16 卡片，不允许输出多张、分镜或对比图。',
  '* 画布：纵向 9:16，按总高 1920px、总宽 1080px 构图，整图高度记为 100%。',
  '* 上部实拍图区：固定 32% 高度（约 614px），不得压缩、不得放大、不得留白；照片满幅铺满 1080px 宽，盘中菜品主体完整不裁切、不偏心、不歪斜。',
  '* 下部信息区：占剩余 68% 高度（约 1306px）；其中正文排版区不得少于 60%（约 1152px），余下约 8%（约 154px）作为底部留白与边框内边距。',
  '* 全图必须 100% 铺满，底部不许出现大片空白，内容不许溢出或裁切；元素间距按内容多少均匀自适应拉伸，禁止靠留白补齐。',
  '**1. 版式与视觉风格**',
  '* 背景：整张卡片使用米白色（#FAF7F0 附近）带细腻纸张纹理的底色。',
  '* 上下切分：实拍图与纸张区之间是一条清晰的横向硬切分，直接接边，不加阴影、不加圆角、不留缝隙。',
  '* 边框：信息区四周一圈极细的浅灰绿色（#8A9A7B 附近）矩形边框，四角点缀简约植物叶片/藤蔓线稿装饰。',
  '* 字体：标题用优雅的中文衬线字体（宋体/思源宋体类），正文用清晰细衬线；标题下方配一条极简波浪线装饰。',
  '* 插画：全部图示为黑色细线稿 + 单色淡彩填充的手绘风插画，线条粗细统一、笔触干净，不要扁平色块图标、不要写实插画。',
  '* 配色：米白底 + 黑灰线稿 + 菜品主色调的少量点睛色，整体低饱和、干净不杂乱。',
  '**2. 上部实拍图区（真实度要求最高，优先保证）**',
  '* 必须是真实美食摄影照片：单反/微单实拍质感，绝不是插画、3D 渲染、CG 或 AI 涂抹感。',
  '* 菜品：{recipe.title} 成品特写——完整呈现这道菜的关键质感（炖菜汤汁浓稠挂勺带自然油润高光、炒菜油亮入味、汤品清澈或浓醇按实际情况），盛在浅口米白色粗陶汤盘里。',
  '* 器物：哑光米白色粗陶浅口汤盘，放在浅色原木桌面上。',
  '* 摄影参数：等效 50-85mm 焦距，f/2.8-f/4，菜品主体全清晰、背景柔和虚化；自然光从左后上方 45° 侧逆光射入，桌面有柔和自然的投影。',
  '* 色彩：自然真实、低饱和耐看；严禁过饱和、荧光色、HDR 过曝、塑料蜡质感。',
  '* 画面内不得出现任何文字、水印、logo、手指、人手、多余餐具或无关食材。',
  '**3. 下部信息区（自上而下固定顺序，不得增删模块）**',
  '* 标题区：居中大字显示菜名「{recipe.title}」；下方一行小字简介「{recipe.summary}」；再下方一条波浪线装饰。',
  '* 原料区：标题「原料」，下方按图标在上、文字在下的方式等宽排列，每行最多 6 项，按原料总条数（本道 {recipe.ingredientCount} 项）自动换行，项间距均匀；每项写「名称 + 用量」，用量可换行。',
  '* 原料内容（逐字照抄，不得增删改写）：',
  '{recipe.ingredients}',
  '* 做法区：标题「做法」，按步骤实际条数依次排列（本道 {recipe.stepCount} 步，不得增删）；每条 = 左侧黑色圆圈数字序号（①②③…）+ 中间手绘线稿插画（画该步骤的厨具或动作）+ 右侧两行以内的动作文字；条目之间用极细虚线分隔。',
  '* 做法内容（逐字照抄，不得增删改写、不得删减或自行概括）：',
  '{recipe.steps}',
  '* 小贴士区：底部一个细边框圆角矩形，标题「小贴士」，内含 1-2 行提示文字；若小贴士内容为空，则整块省略，其高度并入做法区间距，整图仍保持 100% 铺满。',
  '{recipe.tips}',
  '* 信息条：{recipe.infoLine}。',
  '**4. 一致性锁定（每次生成必须完全一致，只允许内容变）**',
  '* 固定不变：画布比例、32%/68% 高度分配、米白纸张底色、上下硬切分、细边框与四角叶片装饰、字体体系与标题波浪线、线稿插画风格与线条粗细、圆圈数字样式、虚线分隔样式、小贴士框样式、整体低饱和配色。',
  '* 允许变化：菜名、简介、原料内容与条数、步骤内容与条数及对应插画题材、小贴士文字。',
  '* 版式必须稳定：不得改变分区顺序、不得新增模块、不得移动标题位置。',
  '* 文字完整性：做法与原料文字必须逐字照抄、完整展示，不得缩写、合并、省略任何步骤或任一字句；步骤数严格等于所给条数，不得自行删减。',
  '**5. 菜谱内容（严格照抄所给文字，不得增删改写、不得删减或自行概括）**',
  '* 原料：',
  '{recipe.ingredients}',
  '* 做法：',
  '{recipe.steps}',
  '* 小贴士：{recipe.tips}',
  '* 信息条：{recipe.infoLine}',
  '**6. 负向约束 (Negative Prompt)**',
  '* 禁止：画面杂乱、深色背景、背景过暗、文字重叠或错位、错别字、英文、乱码、过于写实的烹饪过程图（过程只做线稿）、分区比例失调、上部图片被压缩拉伸、下部内容溢出裁切、底部大片空白、多余模块、水印与 logo、文字删减/步骤合并/自行概括、多图输出/分镜/拼图/前后对比图、非 PNG 格式。',
].join('\n')

function hintText(key: string): string {
  return key === KEY_IMAGE_GEN_PROMPT ? '模板支持 {recipe.*} 变量（见下方说明），复制到菜谱右键菜单时会替换为对应菜谱内容' : ''
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await apis.setting.getList()
    entries.value = (res.settings ?? []).map(s => ({ key: s.key, value: s.value ?? '' }))
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  }
  loading.value = false
}

async function save() {
  saving.value = true
  error.value = ''
  savedTip.value = ''
  try {
    await apis.setting.upsert({ body: { items: entries.value.map(e => ({ key: e.key, value: e.value })) } })
    savedTip.value = '已保存'
    setTimeout(() => (savedTip.value = ''), 2000)
  }
  catch (e) {
    error.value = e instanceof Error ? e.message : '保存失败'
  }
  saving.value = false
}

function fillExample(item: SettingItem) {
  item.value = IMAGE_GEN_PROMPT_HINT
}

// ============ 标签管理：列表（含菜谱数）/ 新增 / 重命名 / 删除 ============
interface TagRow {
  id: number
  name: string
  sort: number
  recipeCount: number
}

// 挂载即拉取，初始置 true 避免首帧闪现「暂无标签」
const tagsLoading = ref(true)
const tagRows = ref<TagRow[]>([])
const tagError = ref('')
const tagSavedTip = ref('')

// 新增
const newTagName = ref('')
const addingTag = ref(false)

// 重命名（行内编辑）
const renamingId = ref<number | null>(null)
const renameValue = ref('')
const renamingBusy = ref(false)

// 删除确认弹窗
const deleteTagTarget = ref<TagRow | null>(null)
const deletingTag = ref(false)
const deleteTagError = ref('')

// 删除确认弹窗期间锁定页面滚动 + 侧滑返回关闭确认框
useBodyScrollLock([() => deleteTagTarget.value !== null])
useModalBackClose(() => deleteTagTarget.value !== null, () => (deleteTagTarget.value = null))

function showTagTip(text: string) {
  tagSavedTip.value = text
  setTimeout(() => (tagSavedTip.value = ''), 2500)
}

async function loadTags() {
  tagsLoading.value = true
  tagError.value = ''
  try {
    const res = await apis.tag.getManageList()
    tagRows.value = (res.list ?? []).map(t => ({
      id: t.id!,
      name: t.name ?? '',
      sort: t.sort ?? 0,
      recipeCount: t.recipeCount ?? 0,
    }))
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '加载标签失败'
  }
  tagsLoading.value = false
}

// 变更后刷新首页标签栏缓存（useAsyncData key 与首页一致）
async function refreshHomeTags() {
  try {
    await refreshNuxtData('tag-list')
  }
  catch {
    // 首页缓存刷新失败不影响本页
  }
}

async function addTag() {
  const name = newTagName.value.trim()
  if (!name || addingTag.value) return
  addingTag.value = true
  tagError.value = ''
  try {
    await apis.tag.create({ body: { name } })
    newTagName.value = ''
    await loadTags()
    await refreshHomeTags()
    showTagTip('标签已添加')
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '添加失败'
  }
  addingTag.value = false
}

function startRename(tag: TagRow) {
  renamingId.value = tag.id
  renameValue.value = tag.name
  tagError.value = ''
}

function cancelRename() {
  renamingId.value = null
  renameValue.value = ''
}

async function confirmRename() {
  const id = renamingId.value
  const name = renameValue.value.trim()
  if (id === null || !name || renamingBusy.value) return
  renamingBusy.value = true
  tagError.value = ''
  try {
    await apis.tag.update({ pathParams: { id }, body: { name } })
    renamingId.value = null
    renameValue.value = ''
    await loadTags()
    await refreshHomeTags()
    showTagTip('标签已重命名')
  }
  catch (e) {
    tagError.value = e instanceof Error ? e.message : '重命名失败'
  }
  renamingBusy.value = false
}

function askDeleteTag(tag: TagRow) {
  deleteTagError.value = ''
  deleteTagTarget.value = tag
}

function cancelDeleteTag() {
  deleteTagTarget.value = null
}

async function confirmDeleteTag() {
  const tag = deleteTagTarget.value
  if (!tag || deletingTag.value) return
  deletingTag.value = true
  deleteTagError.value = ''
  try {
    const res = await apis.tag.delete({ pathParams: { id: tag.id } })
    deleteTagTarget.value = null
    await loadTags()
    await refreshHomeTags()
    const affected = res.affectedRecipes ?? 0
    showTagTip(affected > 0 ? `已删除标签，并从 ${affected} 道菜谱移除该标签` : '标签已删除')
  }
  catch (e) {
    deleteTagError.value = e instanceof Error ? e.message : '删除失败'
  }
  deletingTag.value = false
}

onMounted(() => {
  load()
  loadTags()
})

// ============ 页签 ============
// 标签管理与通用设置分页签管理，避免单页过长
type SettingsTab = 'tags' | 'general'
const activeTab = ref<SettingsTab>('tags')
const settingsTabs: Array<{ key: SettingsTab, label: string }> = [
  { key: 'tags', label: '标签管理' },
  { key: 'general', label: '通用设置' },
]
</script>

<template>
  <main class="page page--narrow">
    <div class="page-header">
      <h1 class="page-title">设置</h1>
      <NuxtLink to="/" class="text-btn text-btn--accent">← 返回首页</NuxtLink>
    </div>

    <p v-if="error" class="error-alert mt-4">{{ error }}</p>

    <!-- 页签：标签管理 / 通用设置 -->
    <div class="settings-tabs">
      <button
        v-for="t in settingsTabs"
        :key="t.key"
        type="button"
        class="settings-tab"
        :class="{ 'settings-tab--active': activeTab === t.key }"
        @click="activeTab = t.key"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- 标签管理 -->
    <section v-show="activeTab === 'tags'" class="mt-6">
      <div class="flex items-center justify-between">
        <h2 class="section-title">标签管理</h2>
        <span v-if="tagSavedTip" class="saved-tip">{{ tagSavedTip }}</span>
      </div>
      <p class="form-hint mt-1">新增标签会排在末尾；删除标签会把它从使用中的菜谱上移除，菜谱本身保留。</p>

      <p v-if="tagError" class="error-alert mt-2">{{ tagError }}</p>

      <div v-if="tagsLoading" class="empty-note mt-3 text-sm">加载中…</div>
      <ul v-else class="settings-list">
        <li v-if="!tagRows.length" class="empty-note px-3 py-4 text-sm">暂无标签</li>
        <li v-for="tag in tagRows" :key="tag.id" class="settings-list__item">
          <template v-if="renamingId === tag.id">
            <input
              v-model="renameValue"
              type="text"
              maxlength="20"
              class="input input--sm min-w-0 flex-1"
              @keyup.enter="confirmRename"
              @keyup.esc="cancelRename"
            >
            <button
              type="button"
              class="btn btn--primary btn--xs"
              :disabled="renamingBusy || !renameValue.trim()"
              @click="confirmRename"
            >{{ renamingBusy ? '保存中…' : '保存' }}</button>
            <button type="button" class="btn btn--outline btn--xs" @click="cancelRename">取消</button>
          </template>
          <template v-else>
            <span class="settings-list__name">{{ tag.name }}</span>
            <span class="settings-list__count">{{ tag.recipeCount }} 道菜谱</span>
            <button type="button" class="text-btn text-btn--edit text-btn--xs" @click="startRename(tag)">重命名</button>
            <button
              type="button"
              class="text-btn text-btn--delete text-btn--xs"
              @click="askDeleteTag(tag)"
            >删除</button>
          </template>
        </li>
      </ul>

      <div class="mt-3 flex gap-2">
        <input
          v-model="newTagName"
          type="text"
          maxlength="20"
          placeholder="新标签名（最长 20 字）"
          class="input w-56"
          @keyup.enter="addTag"
        >
        <button
          type="button"
          class="btn btn--primary"
          :disabled="addingTag || !newTagName.trim()"
          @click="addTag"
        >
          {{ addingTag ? '添加中…' : '添加标签' }}
        </button>
      </div>
    </section>

    <!-- 通用设置 -->
    <div v-show="activeTab === 'general'">
      <div v-if="loading" class="empty-note mt-8 text-sm">加载中…</div>

      <template v-else>
        <div v-if="!entries.length" class="settings-empty">
          暂无设置项
        </div>

        <div v-for="item in entries" :key="item.key" class="mt-6">
          <div class="flex items-center justify-between">
            <label :for="`setting-${item.key}`" class="section-title">{{ item.key }}</label>
            <div class="flex items-center gap-3">
              <span v-if="savedTip" class="saved-tip">{{ savedTip }}</span>
              <button
                v-if="item.key === KEY_IMAGE_GEN_PROMPT"
                type="button"
                class="text-btn text-btn--edit text-btn--xs"
                @click="fillExample(item)"
              >
                填入推荐模板
              </button>
            </div>
          </div>
          <p v-if="hintText(item.key)" class="form-hint mt-1">{{ hintText(item.key) }}</p>
          <textarea
            :id="`setting-${item.key}`"
            v-model="item.value"
            rows="16"
            class="input mt-2 w-full p-3 font-mono leading-relaxed"
            :placeholder="item.key === KEY_IMAGE_GEN_PROMPT ? '输入生图提示词模板，可用 {recipe.title} 等变量…' : '设置值'"
          />
        </div>

        <div v-if="entries.length" class="mt-6 flex justify-end">
          <button
            type="button"
            class="btn btn--primary"
            :disabled="saving"
            @click="save"
          >
            {{ saving ? '保存中…' : '保存' }}
          </button>
        </div>

        <!-- 模板变量说明 -->
        <section class="settings-vars">
          <h2 class="section-title">提示词模板变量</h2>
          <p class="form-hint mt-1">模板中以 <code class="inline-code">{recipe.xxx}</code> 形式引用菜谱字段，复制提示词时自动替换为对应菜谱的内容。</p>
          <ul class="settings-vars__list">
            <li v-for="v in templateVars" :key="v.name" class="settings-vars__row">
              <code class="settings-vars__name">{{ v.name }}</code>
              <span class="settings-vars__desc">{{ v.desc }}</span>
            </li>
          </ul>
        </section>
      </template>
    </div>
  </main>

  <!-- 删除标签确认弹窗 -->
  <Teleport to="body">
    <div
      v-if="deleteTagTarget"
      class="modal-overlay modal-overlay--confirm"
      @click.self="cancelDeleteTag"
    >
      <div class="modal-panel max-w-sm p-6">
        <h3 class="modal-title--sm">删除标签</h3>
        <p class="delete-tag-modal__text">
          确定要删除标签
          <span class="delete-tag-modal__highlight">「{{ deleteTagTarget.name }}」</span>？
        </p>
        <p
          v-if="deleteTagTarget.recipeCount > 0"
          class="warn-note mt-2"
        >
          该标签正被 {{ deleteTagTarget.recipeCount }} 道菜谱使用，删除后将从这些菜谱中移除该标签，菜谱本身不受影响。
        </p>
        <p v-if="deleteTagError" class="error-text mt-2">{{ deleteTagError }}</p>
        <div class="delete-tag-modal__actions">
          <button
            type="button"
            class="btn btn--outline"
            :disabled="deletingTag"
            @click="cancelDeleteTag"
          >
            取消
          </button>
          <button
            type="button"
            class="btn btn--danger"
            :disabled="deletingTag"
            @click="confirmDeleteTag"
          >
            {{ deletingTag ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@reference "~/assets/css/main.css";

/* ---- 页签 ---- */
.settings-tabs { @apply mt-6 flex gap-6 border-b border-zinc-200 dark:border-zinc-800; }
.settings-tab { @apply -mb-px border-b-2 border-transparent px-1 pb-2 text-sm text-zinc-500 transition-colors hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200; }
.settings-tab--active { @apply border-green-600 font-medium text-green-700 hover:text-green-700 dark:border-green-500 dark:text-green-400 dark:hover:text-green-400; }

/* ---- 标签管理 ---- */
.saved-tip { @apply text-xs text-green-600 dark:text-green-400; }
.settings-list { @apply mt-3 divide-y divide-zinc-100 rounded-md border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900; }
.settings-list__item { @apply flex items-center gap-3 px-3 py-2; }
.settings-list__name { @apply min-w-0 flex-1 truncate text-sm text-zinc-800 dark:text-zinc-200; }
.settings-list__count { @apply shrink-0 text-xs text-zinc-400 dark:text-zinc-500; }

/* ---- 通用设置 ---- */
.settings-empty { @apply mt-8 rounded-md border border-dashed border-zinc-300 p-8 text-center text-sm text-zinc-400 dark:border-zinc-700 dark:text-zinc-500; }
.settings-vars { @apply mt-10 rounded-md bg-zinc-50 p-4 dark:bg-zinc-900; }
.inline-code { @apply rounded bg-zinc-200 px-1 dark:bg-zinc-800; }
.settings-vars__list { @apply mt-3 grid grid-cols-1 gap-x-6 gap-y-1 text-xs sm:grid-cols-2; }
.settings-vars__row { @apply flex justify-between gap-2 border-b border-dashed border-zinc-200 py-1 dark:border-zinc-800; }
.settings-vars__name { @apply shrink-0 text-green-700 dark:text-green-400; }
.settings-vars__desc { @apply text-right text-zinc-500 dark:text-zinc-400; }

/* ---- 删除标签确认弹窗 ---- */
.delete-tag-modal__text { @apply mt-3 text-sm text-zinc-600 dark:text-zinc-400; }
.delete-tag-modal__highlight { @apply font-medium text-red-600 dark:text-red-400; }
.delete-tag-modal__actions { @apply mt-5 flex justify-end gap-3; }
</style>
