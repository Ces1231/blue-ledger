import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { UserSummary } from '../types'

interface AuthState {
  user: UserSummary | null
  // access_token is kept in-memory only (NOT persisted) to reduce XSS attack surface
  accessToken: string | null
  // refresh_token is persisted in localStorage (survives tab close)
  refreshToken: string | null
  isAuthenticated: boolean

  // Actions
  login: (user: UserSummary, accessToken: string, refreshToken: string) => void
  logout: () => void
  setAccessToken: (token: string) => void
  setUser: (user: UserSummary) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      login: (user, accessToken, refreshToken) => {
        set({
          user,
          accessToken,
          refreshToken,
          isAuthenticated: true,
        })
      },

      logout: () => {
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
        })
      },

      setAccessToken: (token) => set({ accessToken: token }),

      setUser: (user) => set({ user }),
    }),
    {
      name: 'bl-auth',
      // Only persist refresh_token and user — NOT the access token
      partialize: (state) => ({
        refreshToken: state.refreshToken,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
