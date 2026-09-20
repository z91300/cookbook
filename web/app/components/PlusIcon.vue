<script setup lang="ts">
// 通用「新增 / 添加」加号图标（heroicons outline 风格，与 QuickNav 图标同一套线条语言）。
// 用图标替代文案里的 "+" 符号：字符 "+" 受字体影响，横竖笔画粗细/居中与界面文字不统一，看着突兀。
//
// 尺寸：默认 width/height = 1em（跟随字号），使用方可用 class 覆盖（如 class="size-4"）。
// 颜色：继承 currentColor，自动跟随按钮文字色（含主题与 hover 态）。
</script>

<template>
  <svg
    viewBox="0 0 24 24"
    width="1em"
    height="1em"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    class="shrink-0"
    aria-hidden="true"
  >
    <path d="M12 4.5v15m7.5-7.5h-15" />
  </svg>
</template>
