import { render, waitFor } from "@testing-library/react"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { EventMap } from "@/features/public/event-map"
import type { DisasterEvent } from "@/lib/api/types"

const maplibre = vi.hoisted(() => ({
  addControl: vi.fn(),
  fitBounds: vi.fn(),
  flyTo: vi.fn(),
  getZoom: vi.fn(() => 1.5),
  removeMap: vi.fn(),
  removeMarker: vi.fn(),
}))

vi.mock("maplibre-gl", () => ({
  Map: class {
    addControl = maplibre.addControl
    fitBounds = maplibre.fitBounds
    flyTo = maplibre.flyTo
    getZoom = maplibre.getZoom
    remove = maplibre.removeMap
  },
  Marker: class {
    setLngLat() {
      return this
    }

    addTo() {
      return this
    }

    remove = maplibre.removeMarker
  },
  NavigationControl: class {},
}))

const earthquake: DisasterEvent = {
  id: "30e0b917-4247-44f7-b771-f0550b0f0c9b",
  slug: "magnitude-6-2-earthquake-near-sulawesi",
  type: "earthquake",
  lifecycle: "active",
  priority: "high",
  confidence: "high",
  title: "Magnitude 6.2 earthquake near Sulawesi",
  summary: "Initial assessments remain in progress.",
  locationName: "Sulawesi, Indonesia",
  latitude: -1.43,
  longitude: 120.01,
  startedAt: "2026-07-31T19:41:00Z",
  verifiedAt: "2026-07-31T20:04:00Z",
  sourceCount: 1,
  revision: 1,
}

const wildfire: DisasterEvent = {
  ...earthquake,
  id: "e9ef0e3d-1e61-4ea3-a5a3-bd3bbd49b596",
  slug: "wildfire-near-valparaiso",
  type: "wildfire",
  priority: "critical",
  confidence: "medium",
  title: "Wildfire near Valparaiso",
  locationName: "Valparaiso, Chile",
  latitude: -33.0472,
  longitude: -71.6127,
}

describe("event map", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(window, "WebGLRenderingContext", {
      configurable: true,
      value: class {},
    })
  })

  afterEach(() => {
    Object.defineProperty(window, "WebGLRenderingContext", {
      configurable: true,
      value: undefined,
    })
  })

  it("fits every published incident into the initial viewport", async () => {
    render(
      <EventMap
        events={[earthquake, wildfire]}
        selectedEventID={earthquake.id}
        onSelect={() => undefined}
      />
    )

    await waitFor(() =>
      expect(maplibre.fitBounds).toHaveBeenCalledWith(
        [
          [-71.6127, -33.0472],
          [120.01, -1.43],
        ],
        { duration: 0, maxZoom: 5, padding: 48 }
      )
    )
  })
})
