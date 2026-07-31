import { queryOptions } from "@tanstack/react-query"

import { requestJSON } from "@/lib/api/client"
import type { EventListResponse } from "@/lib/api/types"

export const eventListQuery = queryOptions({
  queryKey: ["public-events"],
  queryFn: () => requestJSON<EventListResponse>("/api/v1/events"),
  staleTime: 60_000,
  refetchInterval: 120_000,
})
