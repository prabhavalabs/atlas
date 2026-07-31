import axe from "axe-core"
import { HttpResponse, http } from "msw"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"

import { event, eventListResponse } from "@/test/fixtures"
import { renderApp } from "@/test/render-app"
import { server } from "@/test/server"

describe("public disaster overview", () => {
  it("shows active disasters without touching administrator identity", async () => {
    const adminSessionRequest = vi.fn()
    server.use(
      http.get("*/api/v1/events", () => HttpResponse.json(eventListResponse)),
      http.get("*/api/v1/admin/session", () => {
        adminSessionRequest()
        return HttpResponse.json({}, { status: 401 })
      })
    )

    renderApp("/")

    expect(
      await screen.findByRole("heading", { name: "Active disasters" })
    ).toBeInTheDocument()
    expect(
      await screen.findByRole("button", {
        name: /Magnitude 6.2 earthquake near Sulawesi/,
      })
    ).toBeInTheDocument()
    expect(
      screen.getByRole("region", { name: "Event map" })
    ).toBeInTheDocument()
    expect(adminSessionRequest).not.toHaveBeenCalled()
  })

  it("supports keyboard selection and has no detectable accessibility violations", async () => {
    server.use(
      http.get("*/api/v1/events", () => HttpResponse.json(eventListResponse))
    )
    const user = userEvent.setup()
    renderApp("/")

    await screen.findByRole("heading", { name: "Active disasters" })
    await user.click(screen.getByRole("link", { name: "Skip to disasters" }))
    expect(screen.getByRole("main")).toHaveFocus()

    const result = await axe.run(document.body, {
      rules: { "color-contrast": { enabled: false } },
    })
    expect(result.violations).toEqual([])
  })

  it("filters the incident list and map by disaster type", async () => {
    const wildfire = {
      ...event,
      id: "e9ef0e3d-1e61-4ea3-a5a3-bd3bbd49b596",
      slug: "wildfire-near-valparaiso",
      type: "wildfire",
      title: "Wildfire near Valparaiso",
      locationName: "Valparaiso, Chile",
      latitude: -33.0472,
      longitude: -71.6127,
    }
    server.use(
      http.get("*/api/v1/events", () =>
        HttpResponse.json({ ...eventListResponse, events: [event, wildfire] })
      )
    )
    const user = userEvent.setup()
    renderApp("/")

    await screen.findByRole("button", { name: /Wildfire near Valparaiso/ })
    await user.click(
      screen.getByRole("combobox", { name: "Filter by disaster type" })
    )
    await user.click(screen.getByRole("option", { name: "Earthquakes" }))

    expect(
      screen.getByRole("button", {
        name: /Magnitude 6.2 earthquake near Sulawesi/,
      })
    ).toBeInTheDocument()
    expect(
      screen.queryByRole("button", { name: /Wildfire near Valparaiso/ })
    ).not.toBeInTheDocument()
    expect(screen.getByText("1 published incident")).toBeInTheDocument()
  })
})
