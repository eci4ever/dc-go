const BASE = '/api/v1'

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

interface LoginData {
  accessToken: string
  refreshToken: string
  user: User
  session: Session
}

interface RegisterData {
  accessToken: string
  refreshToken: string
  user: User
  session: Session
}

interface SessionData {
  user: User
  session: Session | null
}

interface User {
  id: string
  name: string
  email: string
  email_verified: boolean
  image: string | null
  role: string | null
  banned: boolean
  created_at: string
  updated_at: string
}

interface Session {
  id: string
  expires_at: string
  created_at: string
  user_id: string
}

function getToken(): string | null {
  return localStorage.getItem('accessToken')
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem('accessToken', access)
  localStorage.setItem('refreshToken', refresh)
}

function clearTokens() {
  localStorage.removeItem('accessToken')
  localStorage.removeItem('refreshToken')
}

async function request<T>(path: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }
  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers })
  return res.json()
}

export async function login(email: string, password: string): Promise<ApiResponse<LoginData>> {
  const res = await request<LoginData>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  if (res.success && res.data) {
    setTokens(res.data.accessToken, res.data.refreshToken)
  }
  return res
}

export async function register(name: string, email: string, password: string): Promise<ApiResponse<RegisterData>> {
  const res = await request<RegisterData>('/auth/register', {
    method: 'POST',
    body: JSON.stringify({ name, email, password }),
  })
  if (res.success && res.data) {
    setTokens(res.data.accessToken, res.data.refreshToken)
  }
  return res
}

export async function logout(): Promise<void> {
  const refreshToken = localStorage.getItem('refreshToken')
  if (refreshToken) {
    await request('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    })
  }
  clearTokens()
}

export async function getSession(): Promise<ApiResponse<SessionData>> {
  return request<SessionData>('/auth/session')
}

export type { User, Session, LoginData, RegisterData, SessionData }
