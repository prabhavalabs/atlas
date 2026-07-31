import { QueryClient } from "@tanstack/react-query"
import { createMemoryHistory } from "@tanstack/react-router"
import { render } from "@testing-library/react"

import { AtlasApp } from "@/app/atlas-app"
import { createAtlasRouter } from "@/app/router"

export function renderApp(path: string) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  const router = createAtlasRouter({
    history: createMemoryHistory({ initialEntries: [path] }),
  })

  return render(<AtlasApp queryClient={queryClient} router={router} />)
}
