export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export type User = {
  id: number
  name: string
  email: string
  role: string
  created_at: string
  updated_at: string
}

export type Container = {
  id: string
  name: string
  image: string
  state: string
  status: string
  created: number
}

export type ContainerStats = {
  container_id: string
  name: string
  read: string
  cpu_percent: number
  memory_usage: number
  memory_limit: number
  memory_percent: number
  network_rx: number
  network_tx: number
  block_read: number
  block_write: number
}

export type DockerImage = {
  Id: string
  RepoTags?: string[]
  RepoDigests?: string[]
  Size: number
  Created: number
}

export type AuditLog = {
  id: number
  user_id: number | null
  action: string
  resource: string
  resource_id: string
  ip_address: string
  user_agent: string
  created_at: string
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = localStorage.getItem('cloudforge_token')
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (options.body) headers.set('Content-Type', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const response = await fetch(`${API_URL}${path}`, { ...options, headers })
  const body = await response.json().catch(() => ({ error: 'Invalid server response' }))

  if (!response.ok) {
    throw new Error(body.error || `Request failed (${response.status})`)
  }

  return body as T
}
