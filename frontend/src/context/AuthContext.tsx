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
  token: string | undefined
  login: (redirectUri?: string) => Promise<void>
  logout: () => Promise<void>
  register: (redirectUri?: string) => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

function readKeycloakUsername(): string | null {
  return keycloak?.tokenParsed?.preferred_username?.toString() ?? null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(!keycloakEnabled)
  const [authenticated, setAuthenticated] = useState(false)
  const [username, setUsername] = useState<string | null>(null)
  const [token, setToken] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (!keycloakEnabled) {
      setReady(true)
      return
    }

    if (keycloak) {
      keycloak.onAuthSuccess = () => {
        setAuthenticated(true)
        setUsername(readKeycloakUsername())
        setToken(getAccessToken())
      }
      keycloak.onAuthLogout = () => {
        setAuthenticated(false)
        setUsername(null)
        setToken(undefined)
      }
    }

    void initKeycloak()
      .then((ok) => {
        setAuthenticated(ok)
        setUsername(readKeycloakUsername())
        setToken(getAccessToken())
      })
      .finally(() => setReady(true))
  }, [])

  const handleLogin = useCallback(async (redirectUri?: string) => {
    await login(redirectUri)
  }, [])

  const handleLogout = useCallback(async () => {
    await keycloakLogout()
    setAuthenticated(false)
    setUsername(null)
    setToken(undefined)
  }, [])

  const handleRegister = useCallback(async (redirectUri?: string) => {
    await register(redirectUri)
  }, [])

  const value = useMemo<AuthState>(
    () => ({
      enabled: keycloakEnabled,
      ready,
      authenticated,
      username,
      token,
      login: handleLogin,
      logout: handleLogout,
      register: handleRegister,
    }),
    [ready, authenticated, username, token, handleLogin, handleLogout, handleRegister],
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
