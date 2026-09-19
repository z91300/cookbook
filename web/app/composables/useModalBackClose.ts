import { onBeforeUnmount, watch } from 'vue'

/**
 * 移动端侧滑 / 浏览器返回键：关闭弹窗而不是退出页面。
 *
 * 原理：弹窗打开时压入一条同 URL 的 history 记录并注册到全局栈；
 * 返回手势/返回键触发 popstate 时关闭栈顶（最后打开的）弹窗；
 * 用户通过 ✕ / 遮罩等 UI 关闭时，延迟回退一条记录把多余的 history 条目消费掉。
 *
 * - 同一 tick 内「关旧开新」的转场（详情 → 编辑）：取消待执行的回退，
 *   旧弹窗压入的 history 条目直接移交给新弹窗，避免回退与新压入竞态。
 * - popstate 关闭的弹窗标记为已消费，其监听不会再触发额外回退。
 * - 所有弹窗（跨页面组件）共享一个协调器，LIFO 关闭最上层。
 * - 页面卸载不做 history 补偿：路由跳转本身会正常压栈，残留条目无害。
 */

interface ModalEntry {
  close: () => void
  /** 该弹窗的 history 条目已被 popstate 消费（返回手势关闭），无需再回退 */
  consumed: boolean
}

const stack: ModalEntry[] = []
let installed = false
/** UI 关闭触发的 history.back() 待执行定时器（转场时可取消） */
let pendingBack: ReturnType<typeof setTimeout> | null = null
/** 正在执行 UI 关闭的消费回退，期间到来的 popstate 不再关闭弹窗 */
let consumingBack = false
let lastPosition: number | null = null

function install() {
  if (installed || typeof window === 'undefined') return
  installed = true
  window.addEventListener('popstate', () => {
    // UI 关闭的消费回退：仅恢复基准位置，不关弹窗
    if (consumingBack) {
      consumingBack = false
      return
    }
    const pos = typeof history.state?.position === 'number' ? history.state.position as number : null
    const goingForward = lastPosition !== null && pos !== null && pos > lastPosition
    if (pos !== null) lastPosition = pos
    if (goingForward) return // 前进方向是正常导航，不干预
    const top = stack[stack.length - 1]
    if (top) {
      top.consumed = true
      top.close()
    }
  })
}

/**
 * 为一个弹窗接入手势返回关闭。
 * @param isOpen 弹窗打开状态
 * @param close  关闭弹窗的动作（由持有方改状态，如 emit('close')）
 */
export function useModalBackClose(isOpen: () => boolean, close: () => void) {
  if (typeof window === 'undefined') return
  install()

  const entry: ModalEntry = { close, consumed: false }
  let registered = false

  function register() {
    if (registered) return
    registered = true
    entry.consumed = false
    if (pendingBack !== null) {
      // 同 tick 转场：取消消费回退，旧条目移交给本弹窗
      clearTimeout(pendingBack)
      pendingBack = null
    }
    else {
      lastPosition = typeof history.state?.position === 'number' ? history.state.position as number : lastPosition
      history.pushState(history.state, '')
    }
    stack.push(entry)
  }

  function unregister() {
    const i = stack.indexOf(entry)
    if (i >= 0) stack.splice(i, 1)
    registered = false
    if (entry.consumed) return // 返回手势已消费条目
    if (pendingBack === null) {
      // UI 关闭：延迟回退消费条目（留出同 tick 转场移交的窗口）
      pendingBack = setTimeout(() => {
        pendingBack = null
        consumingBack = true
        history.back()
      }, 0)
    }
  }

  const stop = watch(isOpen, (open) => {
    if (open) register()
    else if (registered) unregister()
  })
  onBeforeUnmount(() => {
    stop()
    const i = stack.indexOf(entry)
    if (i >= 0) stack.splice(i, 1)
    registered = false
  })
}
