import {
  Outlet,
  createRootRoute,
  createRoute,
  createRouter,
  type RouterHistory,
} from "@tanstack/react-router"

import { LazyAdminRoute, NotFoundPage } from "@/app/route-components"
import { PublicOverview } from "@/features/public/public-overview"

const rootRoute = createRootRoute({
  component: Outlet,
  notFoundComponent: NotFoundPage,
})

const publicRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: PublicOverview,
})

const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/admin",
  component: LazyAdminRoute,
})

const routeTree = rootRoute.addChildren([publicRoute, adminRoute])

export function createAtlasRouter(options: { history?: RouterHistory } = {}) {
  return createRouter({
    routeTree,
    history: options.history,
    defaultPreload: "intent",
    scrollRestoration: true,
  })
}

export type AtlasRouter = ReturnType<typeof createAtlasRouter>
