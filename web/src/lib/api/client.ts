const configuredBaseURL =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") ?? ""

export class APIError extends Error {
  readonly status: number
  readonly code: string

  constructor(
    status: number,
    code: string,
    message: string
  ) {
    super(message)
    this.name = "APIError"
    this.status = status
    this.code = code
  }
}

export async function requestJSON<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }
  const method = (init.method ?? "GET").toUpperCase()
  if (
    path.startsWith("/api/v1/admin/") &&
    !["GET", "HEAD", "OPTIONS"].includes(method) &&
    !headers.has("X-CSRF-Token")
  ) {
    const csrfToken = readCookie("atlas_admin_csrf")
    if (csrfToken) headers.set("X-CSRF-Token", csrfToken)
  }

  const response = await fetch(`${configuredBaseURL}${path}`, {
    ...init,
    credentials: "include",
    headers,
  })
  if (!response.ok) {
    const problem = (await response.json().catch(() => null)) as {
      code?: string
      title?: string
    } | null
    throw new APIError(
      response.status,
      problem?.code ?? "request_failed",
      problem?.title ?? "The request could not be completed"
    )
  }
  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null
  const prefix = `${encodeURIComponent(name)}=`
  const entry = document.cookie
    .split(";")
    .map((value) => value.trim())
    .find((value) => value.startsWith(prefix))
  if (!entry) return null
  try {
    return decodeURIComponent(entry.slice(prefix.length))
  } catch {
    return null
  }
}
