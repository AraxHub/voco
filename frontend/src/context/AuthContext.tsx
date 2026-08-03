import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import {
  getAccessToken,
  initKeycloak,
  keycloak,
  keycloakEnabled,
  login,
  logout as keycloakLogout,
  register,
} from '../lib/keycloak'

type AuthState = {
  enabled: boolean
  ready: boolean
  authenticated: boolean
  username: string | null
  /** Given name / display name for rooms (not login). */
  displayName: string | null
  token: string | undefined
  login: (redirectUri?: string) => Promise<void>
  logout: () => Promise<void>
  register: (redirectUri?: string) => Promise<void>
  /** Force-refresh access token claims (e.g. after profile edit). */
  refreshProfile: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

function readKeycloakUsername(): string | null {
  return keycloak?.tokenParsed?.preferred_username?.toString() ?? null
}

function readKeycloakDisplayName(): string | null {
  const t = keycloak?.tokenParsed
  if (!t) return null
  const given = t.given_name?.toString().trim()
  if (given) return given
  const full = t.name?.toString().trim()
  if (full) return full
  return null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(!keycloakEnabled)
  const [authenticated, setAuthenticated] = useState(false)
  const [username, setUsername] = useState<string | null>(null)
  const [displayName, setDisplayName] = useState<string | null>(null)
  const [token, setToken] = useState<string | undefined>(undefined)

  const syncFromToken = useCallback(() => {
    setUsername(readKeycloakUsername())
    setDisplayName(readKeycloakDisplayName())
    setToken(getAccessToken())
  }, [])

  useEffect(() => {
    if (!keycloakEnabled) {
      setReady(true)
      return
    }

    let cancelled = false

    if (keycloak) {
      keycloak.onAuthSuccess = () => {
        if (cancelled) return
        setAuthenticated(true)
        syncFromToken()
      }
      keycloak.onAuthRefreshSuccess = () => {
        if (cancelled) return
        syncFromToken()
      }
      keycloak.onAuthLogout = () => {
        if (cancelled) return
        setAuthenticated(false)
        setUsername(null)
        setDisplayName(null)
        setToken(undefined)
      }
    }

    void initKeycloak()
      .then((ok) => {
        if (cancelled) return
        // Set auth before ready so RequireAuth never sees ready&&!authenticated mid-init.
        setAuthenticated(ok || Boolean(keycloak?.authenticated))
        syncFromToken()
        setReady(true)
      })
      .catch(() => {
        if (cancelled) return
        setAuthenticated(Boolean(keycloak?.authenticated))
        syncFromToken()
        setReady(true)
      })

    return () => {
      cancelled = true
    }
  }, [syncFromToken])

  const handleLogin = useCallback(async (redirectUri?: string) => {
    await login(redirectUri)
  }, [])

  const handleLogout = useCallback(async () => {
    await keycloakLogout()
    setAuthenticated(false)
    setUsername(null)
    setDisplayName(null)
    setToken(undefined)
  }, [])

  const handleRegister = useCallback(async (redirectUri?: string) => {
    await register(redirectUri)
  }, [])

  const refreshProfile = useCallback(async () => {
    if (!keycloak?.authenticated) return
    // -1 forces a new access token so given_name / name claims update.
    await keycloak.updateToken(-1)
    syncFromToken()
  }, [syncFromToken])

  const value = useMemo<AuthState>(
    () => ({
      enabled: keycloakEnabled,
      ready,
      authenticated,
      username,
      displayName,
      token,
      login: handleLogin,
      logout: handleLogout,
      register: handleRegister,
      refreshProfile,
    }),
    [
      ready,
      authenticated,
      username,
      displayName,
      token,
      handleLogin,
      handleLogout,
      handleRegister,
      refreshProfile,
    ],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
