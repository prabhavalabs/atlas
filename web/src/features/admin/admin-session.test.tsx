import { HttpResponse, http } from "msw"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it } from "vitest"

import { adminUser, eventListResponse } from "@/test/fixtures"
import { renderApp } from "@/test/render-app"
import { server } from "@/test/server"

describe("administrator session", () => {
  it("presents a labelled sign-in form when no session exists", async () => {
    server.use(
      http.get("*/api/v1/admin/session", () =>
        HttpResponse.json(
          { title: "Administrator authentication is required" },
          { status: 401 }
        )
      )
    )

    renderApp("/admin")

    expect(
      await screen.findByRole("heading", { name: "Sign in to Atlas" })
    ).toBeInTheDocument()
    expect(screen.getByLabelText("Email address")).toBeInTheDocument()
    expect(screen.getByLabelText("Password")).toBeInTheDocument()
  })

  it("uses a mutation to sign in and opens the review queue", async () => {
    let authenticated = false
    server.use(
      http.get("*/api/v1/admin/session", () =>
        authenticated
          ? HttpResponse.json({
              user: adminUser,
              expiresAt: "2026-08-01T08:00:00Z",
            })
          : HttpResponse.json({}, { status: 401 })
      ),
      http.post("*/api/v1/admin/session", async ({ request }) => {
        const body = (await request.json()) as Record<string, string>
        if (
          body.email !== "admin@example.test" ||
          body.password !== "correct horse battery staple"
        ) {
          return HttpResponse.json({}, { status: 401 })
        }
        authenticated = true
        return HttpResponse.json({
          user: adminUser,
          csrfToken: "csrf-test-token",
          expiresAt: "2026-08-01T08:00:00Z",
        })
      }),
      http.get("*/api/v1/events", () => HttpResponse.json(eventListResponse))
    )
    const user = userEvent.setup()
    renderApp("/admin")

    await user.type(
      await screen.findByLabelText("Email address"),
      "admin@example.test"
    )
    await user.type(
      screen.getByLabelText("Password"),
      "correct horse battery staple"
    )
    await user.click(screen.getByRole("button", { name: "Sign in" }))

    expect(
      await screen.findByRole("heading", { name: "Review queue" })
    ).toBeInTheDocument()
    expect(
      screen.getByText(eventListResponse.events[0].title)
    ).toBeInTheDocument()
    expect(screen.getByText("Atlas Administrator")).toBeInTheDocument()
  })

  it("signs out with CSRF protection and returns to the login form", async () => {
    let authenticated = true
    let csrfHeader = ""
    document.cookie = "atlas_admin_csrf=csrf-test-token; path=/"
    server.use(
      http.get("*/api/v1/admin/session", () =>
        authenticated
          ? HttpResponse.json({
              user: adminUser,
              expiresAt: "2026-08-01T08:00:00Z",
            })
          : HttpResponse.json({}, { status: 401 })
      ),
      http.delete("*/api/v1/admin/session", ({ request }) => {
        csrfHeader = request.headers.get("X-CSRF-Token") ?? ""
        authenticated = false
        return new HttpResponse(null, { status: 204 })
      }),
      http.get("*/api/v1/events", () => HttpResponse.json(eventListResponse))
    )
    const user = userEvent.setup()
    renderApp("/admin")

    await screen.findByRole("heading", { name: "Review queue" })
    await user.click(screen.getAllByRole("button", { name: "Sign out" })[0])

    expect(
      await screen.findByRole("heading", { name: "Sign in to Atlas" })
    ).toBeInTheDocument()
    expect(csrfHeader).toBe("csrf-test-token")
    document.cookie =
      "atlas_admin_csrf=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT"
  })
})
