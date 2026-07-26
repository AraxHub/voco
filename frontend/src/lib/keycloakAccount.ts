import { getAccessToken, keycloak, keycloakEnabled } from './keycloak'

const url = import.meta.env.VITE_KEYCLOAK_URL?.toString().trim() ?? ''
const realm = import.meta.env.VITE_KEYCLOAK_REALM?.toString().trim() || 'voco'

export type AccountProfile = {
  username?: string
  firstName?: string
  lastName?: string
  email?: string
  emailVerified?: boolean
}

function accountBase(): string {
  return `${url.replace(/\/$/, '')}/realms/${realm}/account`
}

async function accountRequest<T>(path: string, init?: RequestInit): Promise<T> {
  if (!keycloakEnabled) {
    throw new Error('Keycloak не настроен')
  }

  const token = getAccessToken()
  if (!token) {
    throw new Error('Нужно войти в аккаунт')
  }

  const res = await fetch(`${accountBase()}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...(init?.headers ?? {}),
    },
  })

  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string; errorMessage?: string; message?: string }
      message = body.errorMessage || body.error || body.message || message
    } catch {
      // ignore
    }
    throw new Error(message)
  }

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  if (!text) {
    return undefined as T
  }

  return JSON.parse(text) as T
}

export async function fetchAccount(): Promise<AccountProfile> {
  return accountRequest<AccountProfile>('/?userProfileMetadata=false')
}

export async function updateAccount(profile: AccountProfile): Promise<void> {
  await accountRequest('/?userProfileMetadata=false', {
    method: 'POST',
    body: JSON.stringify(profile),
  })
}

export async function refreshKeycloakToken(): Promise<void> {
  if (!keycloak?.authenticated) return
  await keycloak.updateToken(30)
}
