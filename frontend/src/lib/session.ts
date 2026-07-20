import { getAccessToken, keycloakEnabled } from './keycloak'

export function getAuthToken(): string | undefined {
  if (keycloakEnabled) {
    return getAccessToken()
  }
  return undefined
}
