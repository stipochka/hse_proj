import { create } from 'zustand'
import type { User } from '@/shared/lib/keycloak'

interface AuthState {
  isAuthenticated: boolean
  user: User | null
  setUser: (user: User | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  setUser: (user) => {
    set({
      user,
      isAuthenticated: !!user,
    })
  },
  logout: () => {
    set({
      user: null,
      isAuthenticated: false,
    })
  },
}))
