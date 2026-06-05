const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token')
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  })
  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    throw new Error(await res.text())
  }
  return res.json()
}

export function apiLogin(email: string, password: string) {
  return request<{ token: string; user: any }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function apiGetDashboard() {
  return request<{
    total_users: number
    active_subs: number
    main_node: any
    server_time: string
  }>('/dashboard')
}

export function apiGetUsers() {
  return request<any[]>('/users')
}

export function apiCreateUser(data: { email: string; password_hash: string }) {
  return request<any>('/users', { method: 'POST', body: JSON.stringify(data) })
}

export function apiDeleteUser(id: number) {
  return request(`/users/${id}`, { method: 'DELETE' })
}

export function apiGetUserStats(id: number) {
  return request(`/users/${id}/stats`)
}

export function apiGetInbounds() {
  return request<any[]>('/inbounds')
}

export function apiCreateInbound(data: any) {
  return request<any>('/inbounds', { method: 'POST', body: JSON.stringify(data) })
}

export function apiDeleteInbound(id: number) {
  return request(`/inbounds/${id}`, { method: 'DELETE' })
}
