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
const templateVars = [
  { name: '{recipe.title}', desc: '菜谱标题' },
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

onMounted(load)
</script>

<template>
  <main class="mx-auto max-w-3xl px-4 py-8">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">设置</h1>
      <NuxtLink to="/" class="text-sm text-green-600 hover:underline">← 返回首页</NuxtLink>
    </div>

    <p v-if="error" class="mt-4 rounded bg-red-50 px-3 py-2 text-sm text-red-600">{{ error }}</p>

    <div v-if="loading" class="mt-8 text-center text-sm text-zinc-400">加载中…</div>

    <template v-else>
      <div v-if="!entries.length" class="mt-8 rounded-md border border-dashed border-zinc-300 p-8 text-center text-sm text-zinc-400">
        暂无设置项
      </div>

      <div v-for="item in entries" :key="item.key" class="mt-6">
        <div class="flex items-center justify-between">
          <label :for="`setting-${item.key}`" class="text-sm font-semibold">{{ item.key }}</label>
          <div class="flex items-center gap-3">
            <span v-if="savedTip" class="text-xs text-green-600">{{ savedTip }}</span>
            <button
              v-if="item.key === KEY_IMAGE_GEN_PROMPT"
              type="button"
              class="text-xs text-zinc-500 hover:text-green-600 hover:underline"
              @click="fillExample(item)"
            >
              填入推荐模板
            </button>
          </div>
        </div>
        <p v-if="hintText(item.key)" class="mt-1 text-xs text-zinc-400">{{ hintText(item.key) }}</p>
        <textarea
          :id="`setting-${item.key}`"
          v-model="item.value"
          rows="16"
          class="mt-2 w-full rounded-md border border-zinc-300 bg-white p-3 font-mono text-sm leading-relaxed focus:border-green-500 focus:outline-none"
          :placeholder="item.key === KEY_IMAGE_GEN_PROMPT ? '输入生图提示词模板，可用 {recipe.title} 等变量…' : '设置值'"
        />
      </div>

      <div v-if="entries.length" class="mt-6 flex justify-end">
        <button
          type="button"
          class="rounded-md bg-green-600 px-6 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-50"
          :disabled="saving"
          @click="save"
        >
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>

      <!-- 模板变量说明 -->
      <section class="mt-10 rounded-md bg-zinc-50 p-4">
        <h2 class="text-sm font-semibold">提示词模板变量</h2>
        <p class="mt-1 text-xs text-zinc-500">模板中以 <code class="rounded bg-zinc-200 px-1">{recipe.xxx}</code> 形式引用菜谱字段，复制提示词时自动替换为对应菜谱的内容。</p>
        <ul class="mt-3 grid grid-cols-1 gap-x-6 gap-y-1 text-xs sm:grid-cols-2">
          <li v-for="v in templateVars" :key="v.name" class="flex justify-between gap-2 border-b border-dashed border-zinc-200 py-1">
            <code class="shrink-0 text-green-700">{{ v.name }}</code>
            <span class="text-right text-zinc-500">{{ v.desc }}</span>
          </li>
        </ul>
      </section>
    </template>
  </main>
</template>