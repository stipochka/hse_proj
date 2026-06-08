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

const parseJWT = (token: string): Record<string, any> => {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) throw new Error('Invalid JWT')
    const decoded = JSON.parse(atob(parts[1]))
    return decoded
  } catch {
    return {}
  }
}

export const initKeycloak = async (): Promise<User | null> => {
  const keycloak = new Keycloak({
    url: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080/auth',
    realm: import.meta.env.VITE_KEYCLOAK_REALM || 'master',
    clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'admin-frontend',
  })

  keycloakInstance = keycloak

  try {
    const authenticated = await keycloak.init({
      onLoad: 'check-sso',
      silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
    })

    if (authenticated && keycloak.token) {
      setToken(keycloak.token)
      
      const tokenData = parseJWT(keycloak.token)
      const user: User = {
        id: keycloak.subject || '',
        username: keycloak.tokenParsed?.preferred_username || '',
        email: keycloak.tokenParsed?.email || '',
        firstName: keycloak.tokenParsed?.given_name || '',
        lastName: keycloak.tokenParsed?.family_name || '',
        roles: keycloak.tokenParsed?.roles || [],
      }

      // Refresh token periodically
      setInterval(() => {
        keycloak.refreshToken(70).catch(() => {
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
