import { onBeforeUnmount, watch } from 'vue'

/**
 * 弹窗打开期间锁定页面滚动：
 * - overflow hidden 阻止滚轮/键盘滚动穿透到首页
 * - overscroll-behavior none 阻止触屏链式滑动（横向滑到页面边缘不再带动页面）
 * 注意不要在根元素设 touch-action: none——它会关闭后代元素的手势处理起点，
 * 导致弹窗内部横向滑动（snap 翻页）在移动端完全失效。
 * 锁定期间保留 scrollbar-gutter 槽位（main.css 常驻），页面不发生横向跳动。
 * 传入若干布尔源，任一为真即锁定；组件卸载时自动恢复。
 */
export function useBodyScrollLock(sources: Array<() => boolean>) {
  if (typeof window !== 'undefined') console.log('[scroll-lock] composable executed, client, sources=', sources.length)
  if (typeof document === 'undefined') return

  const html = document.documentElement
  let applied = false

  function sync() {
    const locked = sources.some(src => src())
    if (locked === applied) return
    if (locked) {
      html.style.overflow = 'hidden'
      html.style.overscrollBehavior = 'none'
    }
    else {
      html.style.overflow = ''
      html.style.overscrollBehavior = ''
    }
  }

  const stops = sources.map(src => watch(src, sync))
  sync()
  onBeforeUnmount(() => {
    for (const stop of stops) stop()
    if (applied) {
      html.style.overflow = ''
      html.style.overscrollBehavior = ''
      applied = false
    }
  })
}