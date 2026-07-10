const BASE = "/api/v1";

interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
}
interface User {
  id: string;
  name: string;
  email: string;
  email_verified: boolean;
  image: string | null;
  role: string | null;
  banned: boolean;
  created_at: string;
  updated_at: string;
}
interface Session {
  id: string;
  expires_at: string;
  created_at: string;
  user_id: string;
}
interface SessionData {
  user: User;
  session: Session | null;
}

function csrfToken() {
  return document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token="))
    ?.split("=")[1];
}
let refreshing: Promise<boolean> | null = null;

async function refresh(): Promise<boolean> {
  if (!refreshing)
    refreshing = fetch(`${BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "X-CSRF-Token": csrfToken() ?? "" },
    })
      .then((r) => r.ok)
      .catch(() => false)
      .finally(() => {
        refreshing = null;
      });
  return refreshing;
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retryAuth = true,
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = { ...(options.headers as Record<string, string>) };
  if (options.body) headers["Content-Type"] = "application/json";
  if (!["GET", "HEAD", "OPTIONS"].includes((options.method ?? "GET").toUpperCase()))
    headers["X-CSRF-Token"] = csrfToken() ?? "";
  try {
    const res = await fetch(`${BASE}${path}`, { ...options, headers, credentials: "include" });
    if (res.status === 401 && retryAuth && path !== "/auth/refresh" && (await refresh()))
      return request(path, options, false);
    const body = await res.json().catch(() => null);
    if (body) return body as ApiResponse<T>;
    return { success: res.ok, message: res.ok ? undefined : `Request failed (${res.status})` };
  } catch {
    return { success: false, message: "Unable to reach the server. Please try again." };
  }
}

export const login = (email: string, password: string) =>
  request<SessionData>(
    "/auth/login",
    {
      method: "POST",
      body: JSON.stringify({ email, password }),
    },
    false,
  );
export const register = (name: string, email: string, password: string) =>
  request<SessionData>(
    "/auth/register",
    {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    },
    false,
  );
export async function logout() {
  return request("/auth/logout", { method: "POST" });
}
export const getSession = () => request<SessionData>("/auth/session");
export type { User, Session, SessionData, ApiResponse };
