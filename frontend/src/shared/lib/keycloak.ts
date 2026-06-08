import Keycloak from 'keycloak-js'
import { setToken } from './api'

let keycloakInstance: Keycloak | null = null

export interface User {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  roles: string[]
}

export const initKeycloak = async (): Promise<User | null> => {
  const keycloak = new Keycloak({
    url: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8180',
    realm: import.meta.env.VITE_KEYCLOAK_REALM || 'edu',
    clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'edu-frontend',
  })

  keycloakInstance = keycloak

  try {
    const authenticated = await keycloak.init({
      onLoad: 'login-required',
      pkceMethod: 'S256',
      checkLoginIframe: false, // prevents infinite redirect loop in Keycloak 24
    })

    if (authenticated && keycloak.token) {
      setToken(keycloak.token)

      const realmRoles: string[] =
        (keycloak.tokenParsed?.realm_access as { roles?: string[] })?.roles || []

      const user: User = {
        id: keycloak.subject || '',
        username: keycloak.tokenParsed?.preferred_username || '',
        email: keycloak.tokenParsed?.email || '',
        firstName: keycloak.tokenParsed?.given_name || '',
        lastName: keycloak.tokenParsed?.family_name || '',
        roles: realmRoles,
      }

      // Refresh token periodically
      setInterval(() => {
        keycloak.updateToken(70).catch(() => {
          keycloak.logout()
        })
      }, 60000)

      return user
    }

    return null
  } catch (error) {
    console.error('Failed to initialize Keycloak:', error)
    return null
  }
}

export const getKeycloak = (): Keycloak | null => {
  return keycloakInstance
}

export const logout = () => {
  if (keycloakInstance) {
    keycloakInstance.logout()
    setToken(null)
  }
}
