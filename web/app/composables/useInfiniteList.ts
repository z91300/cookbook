import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'

/** 后端分页响应形状（与 model.PageRes 对齐） */
export interface PagedResult<T> {
  list?: T[] | null
  total?: number
}

export type InfiniteStatus = 'idle' | 'loading' | 'error' | 'done'

/**
 * 分页无限加载（上拉加载）：首次进入加载第 1 页，之后滚动到哨兵附近加载下一页。
 * - fetchPage 闭包内读取当前筛选参数；筛选变化时调用 reset() 重新从第 1 页加载
 * - target 为哨兵元素 ref；IntersectionObserver 进入视口即触发
 */
export function useInfiniteList<T>(
  fetchPage: (page: number) => Promise<PagedResult<T>>,
  opts: { target?: Ref<Element | null> } = {},
) {
  const items = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const status = ref<InfiniteStatus>('idle')
  const error = ref('')
  const page = ref(0)

  let observer: IntersectionObserver | null = null
  let watchStop: (() => void) | null = null

  function sentinelVisible(): boolean {
    const el = opts.target?.value
    if (!el || typeof window === 'undefined') return false
    return el.getBoundingClientRect().top < window.innerHeight + 400
  }

  async function loadNext() {
    if (status.value === 'loading' || status.value === 'done') return
    status.value = 'loading'
    error.value = ''
    try {
      const res = await fetchPage(page.value + 1)
      const list = res.list ?? []
      items.value.push(...list)
      total.value = res.total ?? items.value.length
      page.value += 1
      status.value = list.length === 0 || items.value.length >= total.value ? 'done' : 'idle'
      // 哨兵仍在视口内（一页数据不足铺满屏幕），继续加载
      if (status.value === 'idle' && sentinelVisible()) void loadNext()
    }
    catch (e) {
      status.value = 'error'
      error.value = e instanceof Error ? e.message : '加载失败'
    }
  }

  function reset() {
    items.value = []
    total.value = 0
    page.value = 0
    status.value = 'idle'
    error.value = ''
    void loadNext()
  }

  onMounted(() => {
    if (opts.target && typeof IntersectionObserver !== 'undefined') {
      observer = new IntersectionObserver(
        (entries) => {
          if (entries.some(e => e.isIntersecting)) void loadNext()
        },
        { rootMargin: '400px 0px' },
      )
      if (opts.target.value) observer.observe(opts.target.value)
      // 哨兵可能随条件渲染重新挂载，重挂后继续观察
      watchStop = watch(opts.target, (el) => {
        observer?.disconnect()
        if (el) observer?.observe(el)
      })
    }
    void loadNext()
  })

  onBeforeUnmount(() => {
    watchStop?.()
    observer?.disconnect()
  })

  return { items, total, status, error, loadNext, reset }
}
