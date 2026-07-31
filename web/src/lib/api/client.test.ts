import { afterEach, describe, expect, it, vi } from "vitest"

import { requestJSON } from "@/lib/api/client"

describe("administrator request protection", () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    document.cookie = "atlas_admin_csrf=; Max-Age=0; Path=/"
  })

  it("copies the readable CSRF cookie into unsafe admin requests", async () => {
    document.cookie = "atlas_admin_csrf=csrf-from-cookie; Path=/"
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ event: { id: "event-id" } }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    )
    vi.stubGlobal("fetch", fetchMock)

    await requestJSON("/api/v1/admin/events/event-id/title", {
      method: "PATCH",
      body: JSON.stringify({ title: "Updated title" }),
    })

    const requestInit = fetchMock.mock.calls[0][1] as RequestInit
    expect(new Headers(requestInit.headers).get("X-CSRF-Token")).toBe(
      "csrf-from-cookie",
    )
  })
})
