import { beforeEach, describe, expect, it, vi } from "vitest";

import { getSession, updateProfile, uploadAvatar } from "@/lib/api";

const jsonResponse = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

describe("API client", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.stubGlobal("document", { cookie: "csrf_token=test-csrf-token" });
  });

  it("requests the current session with credentials", async () => {
    const payload = { success: true, data: { user: { id: "user-id" }, session: {} } };
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(payload));
    vi.stubGlobal("fetch", fetchMock);

    await expect(getSession()).resolves.toEqual(payload);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/session",
      expect.objectContaining({ credentials: "include" }),
    );
  });

  it("adds JSON and CSRF headers to state-changing requests", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ success: true, data: {} }));
    vi.stubGlobal("fetch", fetchMock);

    await updateProfile("Updated User");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/users/me",
      expect.objectContaining({
        method: "PUT",
        credentials: "include",
        body: JSON.stringify({ name: "Updated User" }),
        headers: expect.objectContaining({
          "Content-Type": "application/json",
          "X-CSRF-Token": "test-csrf-token",
        }),
      }),
    );
  });

  it("uploads avatars as FormData without overriding the multipart content type", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ success: true, data: {} }));
    vi.stubGlobal("fetch", fetchMock);
    const file = new File(["avatar"], "avatar.png", { type: "image/png" });

    await uploadAvatar(file);

    const [, options] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(options.body).toBeInstanceOf(FormData);
    expect((options.body as FormData).get("avatar")).toBe(file);
    expect(options.headers).toEqual({ "X-CSRF-Token": "test-csrf-token" });
  });

  it("refreshes once and retries an unauthorized request", async () => {
    const success = { success: true, data: { user: { id: "user-id" }, session: {} } };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ success: false }, 401))
      .mockResolvedValueOnce(jsonResponse({ success: true }))
      .mockResolvedValueOnce(jsonResponse(success));
    vi.stubGlobal("fetch", fetchMock);

    await expect(getSession()).resolves.toEqual(success);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(fetchMock.mock.calls[1]?.[0]).toBe("/api/v1/auth/refresh");
    expect(fetchMock.mock.calls[2]?.[0]).toBe("/api/v1/auth/session");
  });

  it("returns a stable error when the server cannot be reached", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));

    await expect(getSession()).resolves.toEqual({
      success: false,
      message: "Unable to reach the server. Please try again.",
    });
  });
});
