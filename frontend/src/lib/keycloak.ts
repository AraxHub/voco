import Keycloak from 'keycloak-js'

const url = import.meta.env.VITE_KEYCLOAK_URL?.toString().trim() ?? ''
const realm = import.meta.env.VITE_KEYCLOAK_REALM?.toString().trim() || 'voco'
const clientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID?.toString().trim() || 'voco-frontend'

export const keycloakEnabled = Boolean(url)

if (import.meta.env.DEV && !keycloakEnabled) {
  console.warn('[voco] VITE_KEYCLOAK_URL не задан — auth отключён, гейт на главной не покажется')
}

export const keycloak = keycloakEnabled
  ? new Keycloak({
      url,
      realm,
      clientId,
    })
  : null

export function getAccessToken(): string | undefined {
  return keycloak?.token ?? undefined
}

/** Keep Keycloak login pages in sync with app light/dark theme. */
function syncThemeCookieForKeycloak() {
  try {
    const theme =
      document.documentElement.dataset.theme === 'light' ||
      document.documentElement.dataset.theme === 'dark'
        ? document.documentElement.dataset.theme
        : 'light'
    const host = window.location.hostname
    const domain =
      host === 'localhost' || host === '127.0.0.1'
        ? ''
        : host.endsWith('voco-online.ru')
          ? '; Domain=.voco-online.ru'
          : ''
    document.cookie = `voco_theme=${theme}; Path=/; Max-Age=2592000; SameSite=Lax${domain}`
  } catch {
    // ignore
  }
}

let initPromise: Promise<boolean> | null = null

function hasOAuthCallback(): boolean {
  const hash = window.location.hash
  return hash.includes('code=') || hash.includes('error=') || hash.includes('state=')
}

function readOAuthHashError(): string | null {
  const hash = window.location.hash.replace(/^#/, '')
  if (!hash) return null
  const params = new URLSearchParams(hash)
  const err = params.get('error')
  if (!err) return null
  const desc = params.get('error_description')
  return desc ? `${err}: ${decodeURIComponent(desc.replace(/\+/g, ' '))}` : err
}

export function initKeycloak(): Promise<boolean> {
  if (!keycloak) {
    return Promise.resolve(false)
  }

  // After redirect from Keycloak the URL carries #code=… — must not reuse a stale init promise.
  if (hasOAuthCallback()) {
    initPromise = null
  }

  initPromise ??= Promise.race([
    keycloak
      .init({
        onLoad: 'check-sso',
        pkceMethod: 'S256',
        checkLoginIframe: false,
      })
      .then((authenticated) => {
        const oauthErr = readOAuthHashError()
        if (oauthErr) {
          console.error('[voco] keycloak oauth error', oauthErr)
          return false
        }

        keycloak.onTokenExpired = () => {
          void keycloak.updateToken(30)
        }
        return authenticated
      }),
    new Promise<boolean>((_, reject) => {
      window.setTimeout(() => reject(new Error('keycloak init timeout')), 10_000)
    }),
  ]).catch((err) => {
    initPromise = null
    console.error('keycloak init failed', err)
    return false
  })

  return initPromise
}

export async function login(redirectUri?: string): Promise<void> {
  if (!keycloak) return
  syncThemeCookieForKeycloak()
  // Scopes come from Keycloak default client scopes (profile, email, account, …).
  await keycloak.login(redirectUri ? { redirectUri } : undefined)
}

export async function logout(): Promise<void> {
  if (!keycloak) return
  await keycloak.logout({ redirectUri: window.location.origin })
}

export async function register(redirectUri?: string): Promise<void> {
  if (!keycloak) return
  syncThemeCookieForKeycloak()
  await keycloak.register(redirectUri ? { redirectUri } : undefined)
}

/** Password change via Keycloak AIA (Account REST /credentials/password was removed). */
export async function changePassword(redirectUri?: string): Promise<void> {
  if (!keycloak) return
  syncThemeCookieForKeycloak()
  // Refresh token first so the browser SSO session is more likely still warm.
  try {
    await keycloak.updateToken(-1)
  } catch {
    // ignore — login() below will re-auth if needed
  }
  await keycloak.login({
    action: 'UPDATE_PASSWORD',
    redirectUri: redirectUri ?? `${window.location.origin}/account`,
  })
}

/** Email change via Keycloak AIA — sends verification to the new address. */
export async function changeEmail(redirectUri?: string): Promise<void> {
  if (!keycloak) return
  syncThemeCookieForKeycloak()
  try {
    await keycloak.updateToken(-1)
  } catch {
    // ignore
  }
  await keycloak.login({
    action: 'UPDATE_EMAIL',
    redirectUri: redirectUri ?? `${window.location.origin}/account`,
  })
}
