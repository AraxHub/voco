import { ensureFreshToken, forceRefreshToken } from './keycloak'
import { getAuthToken } from './session'

const API_BASE = import.meta.env.VITE_API_BASE_URL?.toString().trim() || ''

export type CreateRoomRequest = { title?: string }
export type CreateRoomResponse = { roomId: string; joinUrl?: string }
export type IssueTokenRequest = { name: string }
export type IssueTokenResponse = { token: string; livekitUrl?: string; message?: string }

export type VocoUser = {
  id: string
  nickname: string
  email: string
  displayName: string
  lastSeenAt: string
}

export type Conversation = {
  ID: string
  id?: string
  Type: string
  type?: string
  Title: string
  title?: string
  CreatedBy: string
}

export type Message = {
  ID: string
  id?: string
  ConversationID: string
  SenderID: string
  Body: string
  CreatedAt: string
  EditedAt?: string
  DeletedForAllAt?: string
  Reactions?: { Emoji: string; UserID: string }[]
}

export type CalendarEvent = {
  ID: string
  Title: string
  Description: string
  StartsAt: string
  EndsAt: string
  Status: string
  RoomID?: string
  Timezone: string
  RRule?: string
}

function authHeaders(token?: string): HeadersInit {
  const t = token ?? getAuthToken()
  if (!t) return {}
  return { Authorization: `Bearer ${t}` }
}

function authErrorMessage(status: number, raw: string): string {
  const msg = raw.toLowerCase()
  if (status === 401 || msg.includes('invalid token') || msg.includes('authorization required') || msg.includes('unauthorized')) {
    return 'Сессия истекла или токен недействителен. Войдите снова.'
  }
  return raw
}

async function parseError(res: Response): Promise<string> {
  let message = `${res.status} ${res.statusText}`
  try {
    const body = (await res.json()) as { error?: string; message?: string }
    message = body.error || body.message || message
  } catch {
    /* ignore */
  }
  return authErrorMessage(res.status, message)
}

async function http<T>(path: string, init?: RequestInit, retried = false): Promise<T> {
  const token = await ensureFreshToken(60)
  const headers: HeadersInit = {
    ...authHeaders(token),
    ...(init?.headers || {}),
  }
  if (!(init?.body instanceof FormData)) {
    ;(headers as Record<string, string>)['Content-Type'] =
      (headers as Record<string, string>)['Content-Type'] || 'application/json'
  }
  const res = await fetch(`${API_BASE}${path}`, { ...init, headers })
  if (res.status === 401 && !retried) {
    const ok = await forceRefreshToken()
    if (ok) return http<T>(path, init, true)
    throw new Error(await parseError(res))
  }
  if (!res.ok) {
    throw new Error(await parseError(res))
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export async function createRoom(req: CreateRoomRequest = {}): Promise<CreateRoomResponse> {
  return http('/api/v1/rooms', { method: 'POST', body: JSON.stringify(req) })
}

export async function issueToken(roomId: string, req: IssueTokenRequest): Promise<IssueTokenResponse> {
  return http(`/api/v1/rooms/${encodeURIComponent(roomId)}/token`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function fetchMe(): Promise<VocoUser> {
  return http('/api/v1/users/me')
}

export async function updateMe(nickname: string, displayName?: string): Promise<VocoUser> {
  return http('/api/v1/users/me', {
    method: 'PATCH',
    body: JSON.stringify({ nickname, displayName }),
  })
}

export async function searchUsers(q: string): Promise<VocoUser[]> {
  return http(`/api/v1/users/search?q=${encodeURIComponent(q)}`)
}

export async function listConversations(): Promise<Conversation[]> {
  return http('/api/v1/conversations')
}

export async function openDirect(userId: string): Promise<{ conversation: Conversation }> {
  return http('/api/v1/conversations/direct', {
    method: 'POST',
    body: JSON.stringify({ userId }),
  })
}

export async function createGroup(title: string, memberIds: string[]): Promise<Conversation> {
  return http('/api/v1/conversations/groups', {
    method: 'POST',
    body: JSON.stringify({ title, memberIds }),
  })
}

export async function listMessages(conversationId: string): Promise<Message[]> {
  return http(`/api/v1/conversations/${conversationId}/messages`)
}

export async function sendMessage(conversationId: string, body: string): Promise<Message> {
  return http(`/api/v1/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ body }),
  })
}

export async function acceptRequest(conversationId: string): Promise<void> {
  return http(`/api/v1/conversations/${conversationId}/request/accept`, { method: 'POST' })
}

export async function blockRequest(conversationId: string): Promise<void> {
  return http(`/api/v1/conversations/${conversationId}/request/block`, { method: 'POST' })
}

export async function callFromChat(conversationId: string): Promise<{ roomId: string; joinUrl: string }> {
  return http(`/api/v1/conversations/${conversationId}/call`, { method: 'POST' })
}

export async function listEvents(from: string, to: string): Promise<CalendarEvent[]> {
  return http(`/api/v1/calendar/events?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`)
}

export async function createEvent(body: Record<string, unknown>): Promise<CalendarEvent> {
  return http('/api/v1/calendar/events', { method: 'POST', body: JSON.stringify(body) })
}

export async function rescheduleEvent(id: string, startsAt: string, endsAt: string): Promise<CalendarEvent> {
  return http(`/api/v1/calendar/events/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ startsAt, endsAt }),
  })
}

export async function cancelEvent(id: string): Promise<CalendarEvent> {
  return http(`/api/v1/calendar/events/${id}/cancel`, { method: 'POST' })
}

export async function getPushSettings(): Promise<{ pushEnabled: boolean }> {
  return http('/api/v1/push/settings')
}

export async function setPushSettings(pushEnabled: boolean): Promise<{ pushEnabled: boolean }> {
  return http('/api/v1/push/settings', {
    method: 'PUT',
    body: JSON.stringify({ pushEnabled }),
  })
}

export async function getVapidPublicKey(): Promise<{ publicKey: string }> {
  return http('/api/v1/push/vapidPublicKey')
}

export async function subscribePush(sub: PushSubscriptionJSON): Promise<void> {
  return http('/api/v1/push/subscribe', {
    method: 'POST',
    body: JSON.stringify({
      endpoint: sub.endpoint,
      keys: { p256dh: sub.keys?.p256dh, auth: sub.keys?.auth },
    }),
  })
}

export function apiBase(): string {
  return API_BASE
}

export function wsURL(token: string): string {
  const base = API_BASE || window.location.origin
  const u = new URL(base, window.location.origin)
  u.protocol = u.protocol === 'https:' ? 'wss:' : 'ws:'
  u.pathname = '/api/v1/ws'
  u.search = `token=${encodeURIComponent(token)}`
  return u.toString()
}
