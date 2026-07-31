import { QueryClient } from "@tanstack/react-query"

import { AtlasApp } from "@/app/atlas-app"
import { createAtlasRouter } from "@/app/router"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: true,
    },
    mutations: { retry: false },
  },
})
const router = createAtlasRouter()

export default function App() {
  return <AtlasApp queryClient={queryClient} router={router} />
}
