import { useQuery } from "@tanstack/react-query"

import { Spinner } from "@/components/ui/spinner"
import { AdminDashboard } from "@/features/admin/admin-dashboard"
import { adminSessionQuery } from "@/features/admin/api"
import { LoginForm } from "@/features/admin/login-form"
import { APIError } from "@/lib/api/client"

export function AdminRoute() {
  const sessionQuery = useQuery(adminSessionQuery)

  if (sessionQuery.isPending) {
    return (
      <main
        className="grid min-h-svh place-items-center"
        aria-label="Loading administrator session"
      >
        <Spinner className="size-6" />
      </main>
    )
  }
  if (
    sessionQuery.isError &&
    sessionQuery.error instanceof APIError &&
    sessionQuery.error.status === 401
  ) {
    return <LoginForm />
  }
  if (sessionQuery.isError) {
    return (
      <main className="grid min-h-svh place-items-center px-4 text-center">
        <div>
          <h1 className="text-xl font-semibold">
            Administration is unavailable
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Atlas could not verify the administrator session. Try again shortly.
          </p>
        </div>
      </main>
    )
  }
  return <AdminDashboard session={sessionQuery.data} />
}
