import { useEffect, useMemo, useRef, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import {
  ArrowUpRightIcon,
  CircleAlertIcon,
  Clock3Icon,
  ListFilterIcon,
  MapIcon,
  RefreshCwIcon,
  ShieldCheckIcon,
} from "lucide-react"

import { AtlasMark } from "@/components/atlas-mark"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { Skeleton } from "@/components/ui/skeleton"
import { eventListQuery } from "@/features/public/api"
import { EventMap } from "@/features/public/event-map"
import { useMediaQuery } from "@/hooks/use-media-query"
import type { DisasterEvent } from "@/lib/api/types"
import { cn } from "@/lib/utils"
import { useMapStore } from "@/stores/map-store"

const dateFormatter = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "UTC",
})

function formatUTC(value: string) {
  return `${dateFormatter.format(new Date(value))} UTC`
}

function EventRow({
  event,
  selected,
  onSelect,
}: {
  event: DisasterEvent
  selected: boolean
  onSelect: () => void
}) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onSelect}
      className={cn(
        "group w-full border-b px-4 py-4 text-left transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none focus-visible:ring-inset",
        selected ? "bg-accent" : "hover:bg-muted/60"
      )}
    >
      <span className="flex items-start gap-3">
        <span
          aria-hidden="true"
          className={cn(
            "mt-1.5 size-2.5 shrink-0 rounded-full",
            event.priority === "critical" || event.priority === "high"
              ? "bg-hazard-severe"
              : "bg-hazard-moderate"
          )}
        />
        <span className="min-w-0 flex-1">
          <span className="mb-1 flex flex-wrap items-center gap-2 text-xs font-medium tracking-wide text-muted-foreground uppercase">
            <span>{event.type}</span>
            <span aria-hidden="true">·</span>
            <span>{event.priority}</span>
          </span>
          <span className="block text-sm font-semibold text-foreground group-hover:text-primary">
            {event.title}
          </span>
          <span className="mt-1 block text-sm text-muted-foreground">
            {event.locationName}
          </span>
          <span className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
            <Clock3Icon aria-hidden="true" />
            Verified {formatUTC(event.verifiedAt)}
          </span>
        </span>
      </span>
    </button>
  )
}

function SelectedEvent({ event }: { event: DisasterEvent }) {
  return (
    <article className="border-t px-5 py-5 lg:border-t-0 lg:border-l">
      <div className="flex flex-wrap items-center gap-2">
        <Badge variant="outline" className="capitalize">
          {event.lifecycle}
        </Badge>
        <Badge variant="secondary" className="capitalize">
          {event.confidence} confidence
        </Badge>
      </div>
      <h2 className="mt-4 text-xl font-semibold tracking-tight">
        {event.title}
      </h2>
      <p className="mt-1 text-sm text-muted-foreground">{event.locationName}</p>
      <p className="mt-4 text-sm leading-6 text-muted-foreground">
        {event.summary}
      </p>
      <dl className="mt-5 grid grid-cols-2 gap-x-4 gap-y-5 border-y py-4 text-sm">
        <div>
          <dt className="text-xs tracking-wide text-muted-foreground uppercase">
            Started
          </dt>
          <dd className="mt-1 font-medium">{formatUTC(event.startedAt)}</dd>
        </div>
        <div>
          <dt className="text-xs tracking-wide text-muted-foreground uppercase">
            Sources
          </dt>
          <dd className="mt-1 font-medium">{event.sourceCount} connected</dd>
        </div>
      </dl>
      <Button variant="link" className="mt-3 px-0" disabled>
        Full situation page
        <ArrowUpRightIcon data-icon="inline-end" aria-hidden="true" />
      </Button>
    </article>
  )
}

export function PublicOverview() {
  const mainRef = useRef<HTMLElement>(null)
  const eventQuery = useQuery(eventListQuery)
  const desktopMap = useMediaQuery("(min-width: 1024px)")
  const [typeFilter, setTypeFilter] = useState("all")
  const selectedEventID = useMapStore((state) => state.selectedEventID)
  const selectEvent = useMapStore((state) => state.selectEvent)
  const allEvents = useMemo(
    () => eventQuery.data?.events ?? [],
    [eventQuery.data?.events],
  )
  const events = useMemo(
    () =>
      typeFilter === "all"
        ? allEvents
        : allEvents.filter((event) => event.type === typeFilter),
    [allEvents, typeFilter],
  )
  const selected =
    events.find((event) => event.id === selectedEventID) ?? events[0]

  useEffect(() => {
    if (events[0] && !events.some((event) => event.id === selectedEventID)) {
      selectEvent(events[0].id)
    }
  }, [events, selectEvent, selectedEventID])

  return (
    <div className="min-h-svh bg-background text-foreground">
      <a
        href="#disasters"
        onClick={() => mainRef.current?.focus()}
        className="sr-only z-50 rounded-md bg-foreground px-3 py-2 text-background focus:not-sr-only focus:fixed focus:top-3 focus:left-3"
      >
        Skip to disasters
      </a>
      <header className="border-b">
        <div className="mx-auto flex h-16 max-w-screen-2xl items-center justify-between px-4 sm:px-6">
          <a href="/" className="flex items-center gap-3 font-semibold">
            <AtlasMark />
            <span>Atlas</span>
          </a>
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <span className="hidden items-center gap-1.5 sm:flex">
              <ShieldCheckIcon aria-hidden="true" />
              Evidence-first monitoring
            </span>
            <Badge variant="outline">Public beta</Badge>
          </div>
        </div>
      </header>

      <main
        ref={mainRef}
        id="disasters"
        tabIndex={-1}
        className="mx-auto max-w-screen-2xl outline-none"
      >
        <section className="border-b px-4 py-7 sm:px-6">
          <div className="flex flex-col justify-between gap-5 md:flex-row md:items-end">
            <div>
              <p className="text-sm font-medium text-primary">
                Global overview
              </p>
              <h1 className="mt-1 text-3xl font-semibold tracking-tight">
                Active disasters
              </h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
                Verified incident information, connected sources, and concise
                updates in one calm operational view.
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger aria-label="Filter by disaster type">
                  <ListFilterIcon aria-hidden="true" />
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="all">All disasters</SelectItem>
                    <SelectItem value="earthquake">Earthquakes</SelectItem>
                    <SelectItem value="wildfire">Wildfires</SelectItem>
                    <SelectItem value="flood">Floods</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
              <Button
                variant="outline"
                onClick={() => void eventQuery.refetch()}
                disabled={eventQuery.isFetching}
              >
                <RefreshCwIcon data-icon="inline-start" aria-hidden="true" />
                Refresh
              </Button>
              <Sheet>
                <SheetTrigger asChild>
                  <Button className="lg:hidden">
                    <MapIcon data-icon="inline-start" aria-hidden="true" />
                    View map
                  </Button>
                </SheetTrigger>
                <SheetContent side="bottom" className="h-[82svh]">
                  <SheetHeader>
                    <SheetTitle>Active disaster map</SheetTitle>
                    <SheetDescription>
                      Select a marker, then return to the incident list for
                      details.
                    </SheetDescription>
                  </SheetHeader>
                  <div className="min-h-0 flex-1 px-4 pb-4">
                    <EventMap
                      events={events}
                      selectedEventID={selected?.id ?? null}
                      onSelect={selectEvent}
                    />
                  </div>
                </SheetContent>
              </Sheet>
            </div>
          </div>
        </section>

        {eventQuery.isPending ? (
          <div className="grid gap-5 p-5 lg:grid-cols-[24rem_1fr]">
            <div className="flex flex-col gap-3">
              <Skeleton className="h-28 w-full" />
              <Skeleton className="h-28 w-full" />
            </div>
            <Skeleton className="min-h-[34rem] w-full" />
          </div>
        ) : eventQuery.isError ? (
          <div className="p-5">
            <Alert variant="destructive">
              <CircleAlertIcon aria-hidden="true" />
              <AlertTitle>Disaster data is temporarily unavailable</AlertTitle>
              <AlertDescription>
                The last request did not complete. Try refreshing in a moment.
              </AlertDescription>
            </Alert>
          </div>
        ) : allEvents.length === 0 ? (
          <div className="px-5 py-16 text-center">
            <h2 className="text-lg font-semibold">No published incidents</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Atlas has no active verified incidents to display right now.
            </p>
          </div>
        ) : events.length === 0 ? (
          <div className="px-5 py-16 text-center">
            <h2 className="text-lg font-semibold">No matching incidents</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              No active verified incidents match this disaster type.
            </p>
          </div>
        ) : (
          <div className="grid min-h-[38rem] lg:grid-cols-[23rem_minmax(24rem,1fr)_21rem]">
            <section aria-label="Disaster list" className="border-r">
              <div className="flex items-center justify-between border-b px-4 py-3 text-xs text-muted-foreground">
                <span>
                  {events.length} published incident
                  {events.length === 1 ? "" : "s"}
                </span>
                <span>Updated {formatUTC(eventQuery.data.generatedAt)}</span>
              </div>
              {events.map((event) => (
                <EventRow
                  key={event.id}
                  event={event}
                  selected={event.id === selected?.id}
                  onSelect={() => selectEvent(event.id)}
                />
              ))}
            </section>
            <div className="hidden p-4 lg:block">
              {desktopMap && (
                <EventMap
                  events={events}
                  selectedEventID={selected?.id ?? null}
                  onSelect={selectEvent}
                />
              )}
            </div>
            {selected && <SelectedEvent event={selected} />}
          </div>
        )}
      </main>
      <footer className="border-t px-4 py-5 text-center text-xs text-muted-foreground">
        Atlas is an open-source project by Prabhava Labs. Always follow local
        authorities for life-safety instructions.
      </footer>
    </div>
  )
}
