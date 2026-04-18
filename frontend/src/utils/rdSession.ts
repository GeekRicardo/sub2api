/**
 * 从 AI 请求的 metadata.user_id 字段中提取 session_id。
 *
 * 兼容两种常见格式：
 *   1) Claude Code: JSON 字符串 {"device_id":"...","session_id":"<uuid>"}
 *   2) LangSmith: user_<hash>_account__session_<uuid>
 */
export function extractSessionIdFromString(userId: unknown): string | null {
  if (!userId || typeof userId !== 'string') return null
  const trimmed = userId.trim()
  if (trimmed.startsWith('{')) {
    try {
      const obj = JSON.parse(trimmed)
      if (obj && typeof obj.session_id === 'string') return obj.session_id
    } catch {
      /* fallthrough */
    }
  }
  const m = trimmed.match(/_session_([a-zA-Z0-9-]+)/)
  return m ? m[1] : null
}

export function extractSessionId(request: Record<string, any> | null | undefined): string | null {
  if (!request) return null
  const userId = request?.metadata?.user_id
  return extractSessionIdFromString(userId)
}
