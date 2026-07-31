import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  BellIcon,
  CheckCircle2Icon,
  ChevronRightIcon,
  LayoutListIcon,
  LogOutIcon,
  SearchIcon,
  ShieldCheckIcon,
} from "lucide-react"

import { AtlasMark } from "@/components/atlas-mark"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import {
  adminSessionQuery,
  deleteAdminSession,
} from "@/features/admin/api"
import { eventListQuery } from "@/features/public/api"
import type { AdminSession } from "@/lib/api/types"

export function AdminDashboard({ session }: { session: AdminSession }) {
  const queryClient = useQueryClient()
  const eventsQuery = useQuery(eventListQuery)
  const signOutMutation = useMutation({
    mutationFn: deleteAdminSession,
    onSuccess: async () => {
      await queryClient.resetQueries({
        queryKey: adminSessionQuery.queryKey,
        exact: true,
      })
    },
  })

  return (
    <div className="min-h-svh bg-muted/35">
      <header className="border-b bg-background">
        <div className="flex h-16 items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-3">
            <AtlasMark />
            <div>
              <p className="text-sm font-semibold">Atlas</p>
              <p className="text-xs text-muted-foreground">Administration</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" aria-label="Notifications">
              <BellIcon aria-hidden="true" />
            </Button>
            <div className="hidden text-right sm:block">
              <p className="text-sm font-medium">{session.user.displayName}</p>
              <p className="text-xs text-muted-foreground capitalize">
                {session.user.role}
              </p>
            </div>
            <Avatar>
              <AvatarFallback>AA</AvatarFallback>
            </Avatar>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Sign out"
              onClick={() => signOutMutation.mutate()}
              disabled={signOutMutation.isPending}
            >
              <LogOutIcon aria-hidden="true" />
            </Button>
          </div>
        </div>
      </header>

      <div className="grid md:grid-cols-[14rem_1fr]">
        <aside className="hidden min-h-[calc(100svh-4rem)] border-r bg-background p-3 md:block">
          <nav aria-label="Administration" className="flex flex-col gap-1">
            <a
              href="/admin"
              aria-current="page"
              className="flex h-10 items-center gap-2 rounded-lg bg-accent px-3 text-sm font-medium text-accent-foreground"
            >
              <LayoutListIcon aria-hidden="true" />
              Review queue
            </a>
            <span className="mt-5 px-3 text-xs font-medium tracking-wide text-muted-foreground uppercase">
              Workspace
            </span>
            <span className="flex h-10 items-center gap-2 px-3 text-sm text-muted-foreground">
              <ShieldCheckIcon aria-hidden="true" />
              Published events
            </span>
          </nav>
          <Button
            variant="ghost"
            className="mt-8 w-full justify-start"
            onClick={() => signOutMutation.mutate()}
            disabled={signOutMutation.isPending}
          >
            <LogOutIcon data-icon="inline-start" aria-hidden="true" />
            Sign out
          </Button>
        </aside>

        <main className="min-w-0 p-4 sm:p-7">
          <div className="mx-auto max-w-6xl">
            {signOutMutation.isError && (
              <p role="alert" className="mb-4 text-sm text-destructive">
                Atlas could not sign out this session. Try again.
              </p>
            )}
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
              <div>
                <p className="text-sm font-medium text-primary">
                  Editorial workspace
                </p>
                <h1 className="mt-1 text-3xl font-semibold tracking-tight">
                  Review queue
                </h1>
                <p className="mt-2 text-sm text-muted-foreground">
                  Verify source-backed changes before they reach the public
                  view.
                </p>
              </div>
              <Badge variant="outline" className="h-7">
                <CheckCircle2Icon data-icon="inline-start" aria-hidden="true" />
                System healthy
              </Badge>
            </div>

            <div className="mt-7 flex flex-col gap-3 border-y bg-background p-3 sm:flex-row sm:items-center">
              <div className="relative flex-1">
                <SearchIcon
                  aria-hidden="true"
                  className="absolute top-1/2 left-2.5 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                  aria-label="Search review queue"
                  placeholder="Search incidents"
                  className="pl-8"
                />
              </div>
              <Button variant="outline">All review states</Button>
            </div>

            <section
              aria-label="Events awaiting review"
              className="bg-background"
            >
              {eventsQuery.isPending ? (
                <div className="flex flex-col gap-2 p-4">
                  <Skeleton className="h-20 w-full" />
                  <Skeleton className="h-20 w-full" />
                </div>
              ) : eventsQuery.isError ? (
                <p className="p-6 text-sm text-destructive">
                  The review queue could not be loaded.
                </p>
              ) : eventsQuery.data.events.length === 0 ? (
                <p className="p-10 text-center text-sm text-muted-foreground">
                  There are no incidents awaiting review.
                </p>
              ) : (
                <div className="divide-y">
                  {eventsQuery.data.events.map((event) => (
                    <article
                      key={event.id}
                      className="grid gap-4 p-4 transition-colors hover:bg-muted/40 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
                    >
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <Badge variant="secondary" className="capitalize">
                            {event.type}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            Revision {event.revision} · {event.sourceCount}{" "}
                            source
                          </span>
                        </div>
                        <h2 className="mt-2 font-semibold">{event.title}</h2>
                        <p className="mt-1 line-clamp-1 text-sm text-muted-foreground">
                          {event.summary}
                        </p>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="text-right">
                          <p className="text-sm font-medium">
                            Needs verification
                          </p>
                          <p className="text-xs text-muted-foreground">
                            New source evidence
                          </p>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label={`Review ${event.title}`}
                        >
                          <ChevronRightIcon aria-hidden="true" />
                        </Button>
                      </div>
                    </article>
                  ))}
                </div>
              )}
            </section>
          </div>
        </main>
      </div>
    </div>
  )
}
