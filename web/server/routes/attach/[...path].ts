/**
 * 开发期附件转发：/attach/** → 后端 :8000/attach/**
 *
 * 说明：nitro devProxy 会剥掉挂载前缀（/attach/demo/x.svg 转发成 /demo/x.svg），
 * 与后端静态路由冲突；server route 手动转发可完整保留路径。
 * 仅 dev 生效：生产由 nginx/CDN 直接映射附件目录。
 */
export default defineEventHandler(async (event) => {
  const path = event.path // 保留完整路径，如 /attach/demo/x.svg
  const target = `http://127.0.0.1:8000${path}`
  return proxyRequest(event, target)
})