export type CreateRoomRequest = {
  title?: string
}

export type CreateRoomResponse = {
  roomId: string
  joinUrl?: string
}

export type IssueTokenRequest = {
  name: string
}

export type IssueTokenResponse = {
  token: string
  livekitUrl?: string
  message?: string
}

import { getAuthToken } from './session'

const API_BASE = import.meta.env.VITE_API_BASE_URL?.toString().trim() || ''

function authHeaders(): HeadersInit {
  const token = getAuthToken()
  if (!token) {
    return {}
  }
  return { Authorization: `Bearer ${token}` }
}

async function http<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
      ...(init?.headers || {}),
    },
  })

  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string; message?: string }
      message = body.error || body.message || message
    } catch {
      // ignore
    }
    throw new Error(message)
  }

  return (await res.json()) as T
}

export async function createRoom(req: CreateRoomRequest = {}): Promise<CreateRoomResponse> {
  return http<CreateRoomResponse>('/api/v1/rooms', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function issueToken(roomId: string, req: IssueTokenRequest): Promise<IssueTokenResponse> {
  return http<IssueTokenResponse>(`/api/v1/rooms/${encodeURIComponent(roomId)}/token`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

