import "maplibre-gl/dist/maplibre-gl.css"

import { useEffect, useRef } from "react"
import type { Map as MapLibreMap, Marker as MapLibreMarker } from "maplibre-gl"

import type { DisasterEvent } from "@/lib/api/types"

const OPEN_FREE_MAP_STYLE = "https://tiles.openfreemap.org/styles/liberty"

type EventMapProps = {
  events: DisasterEvent[]
  selectedEventID: string | null
  onSelect: (eventID: string) => void
}

export function EventMap({ events, selectedEventID, onSelect }: EventMapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<MapLibreMap | null>(null)
  const markersRef = useRef<MapLibreMarker[]>([])

  useEffect(() => {
    if (
      !containerRef.current ||
      typeof window.WebGLRenderingContext === "undefined"
    ) {
      return
    }
    let cancelled = false

    void import("maplibre-gl").then(({ Map, Marker, NavigationControl }) => {
      if (cancelled || !containerRef.current) return
      const selected = events[0]
      const map = new Map({
        container: containerRef.current,
        style: OPEN_FREE_MAP_STYLE,
        center: selected ? [selected.longitude, selected.latitude] : [12, 15],
        zoom: selected ? 4.2 : 1.5,
        attributionControl: { compact: true },
      })
      map.addControl(new NavigationControl({ showCompass: false }), "top-right")
      mapRef.current = map

      markersRef.current = events.map((event) => {
        const markerButton = document.createElement("button")
        markerButton.type = "button"
        markerButton.className = "atlas-map-marker"
        markerButton.setAttribute("aria-label", event.title)
        markerButton.dataset.priority = event.priority
        markerButton.addEventListener("click", () => onSelect(event.id))
        return new Marker({ element: markerButton })
          .setLngLat([event.longitude, event.latitude])
          .addTo(map)
      })
    })

    return () => {
      cancelled = true
      for (const marker of markersRef.current) marker.remove()
      markersRef.current = []
      mapRef.current?.remove()
      mapRef.current = null
    }
  }, [events, onSelect])

  useEffect(() => {
    const selected = events.find((event) => event.id === selectedEventID)
    if (!selected || !mapRef.current) return
    mapRef.current.flyTo({
      center: [selected.longitude, selected.latitude],
      zoom: Math.max(mapRef.current.getZoom(), 4.2),
      essential: false,
    })
  }, [events, selectedEventID])

  return (
    <section
      aria-label="Event map"
      className="relative min-h-80 overflow-hidden rounded-lg border bg-muted"
    >
      <div ref={containerRef} className="absolute inset-0" />
      <p className="absolute bottom-3 left-3 rounded-md border bg-background/95 px-2 py-1 text-xs text-muted-foreground shadow-sm">
        OpenStreetMap data · OpenFreeMap tiles
      </p>
      <noscript>The event list provides the same incident locations.</noscript>
    </section>
  )
}
