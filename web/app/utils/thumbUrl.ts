/**
 * 附件缩略图 URL：在 /attachments/{id}/content 上追加 ?w=<width>。
 * 后端按宽度等比缩小（不放大），结果长缓存 + ETag。
 * 非 /attachments/ URL（外链等）原样返回。
 */
export function thumbUrl(url: string, width = 360): string {
  if (!url) return url
  const m = url.match(/^\/attachments\/(\d+)\/content$/)
  if (!m) return url
  return `${url}?w=${width}`
}