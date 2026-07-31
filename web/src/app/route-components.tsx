import { lazy, Suspense } from "react"

import { Spinner } from "@/components/ui/spinner"

const AdminRoute = lazy(() =>
  import("@/features/admin/admin-route").then((module) => ({
    default: module.AdminRoute,
  })),
)

export function LazyAdminRoute() {
  return (
    <Suspense
      fallback={
        <main
          className="grid min-h-svh place-items-center"
          aria-label="Loading administration portal"
        >
          <Spinner className="size-6" />
        </main>
      }
    >
      <AdminRoute />
    </Suspense>
  )
}

export function NotFoundPage() {
  return (
    <main className="grid min-h-svh place-items-center px-4 text-center">
      <div>
        <p className="text-sm font-medium text-primary">404</p>
        <h1 className="mt-2 text-2xl font-semibold">Page not found</h1>
        <a
          className="mt-4 inline-block text-sm text-primary underline"
          href="/"
        >
          Return to active disasters
        </a>
      </div>
    </main>
  )
}
