import { queryOptions } from "@tanstack/react-query"

import { requestJSON } from "@/lib/api/client"
import type { AdminSession } from "@/lib/api/types"

export const adminSessionQuery = queryOptions({
  queryKey: ["admin-session"],
  queryFn: () => requestJSON<AdminSession>("/api/v1/admin/session"),
  retry: false,
  staleTime: 30_000,
})

export function createAdminSession(input: { email: string; password: string }) {
  return requestJSON<AdminSession>("/api/v1/admin/session", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

export function deleteAdminSession() {
  return requestJSON<void>("/api/v1/admin/session", { method: "DELETE" })
}
