import { QueryClientProvider, type QueryClient } from "@tanstack/react-query"
import { RouterProvider } from "@tanstack/react-router"
import { ErrorBoundary } from "react-error-boundary"

import { TooltipProvider } from "@/components/ui/tooltip"
import type { AtlasRouter } from "@/app/router"

function ApplicationError() {
  return (
    <main className="grid min-h-svh place-items-center px-4 text-center">
      <div>
        <h1 className="text-xl font-semibold">Atlas encountered a problem</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Reload the page to restore the monitoring view.
        </p>
      </div>
    </main>
  )
}

export function AtlasApp({
  queryClient,
  router,
}: {
  queryClient: QueryClient
  router: AtlasRouter
}) {
  return (
    <ErrorBoundary fallback={<ApplicationError />}>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          <RouterProvider router={router} />
        </TooltipProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
