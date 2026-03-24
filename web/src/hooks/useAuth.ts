import { useAuthStore } from '../stores/authStore'
import { logout as logoutAPI } from '../api/auth'
import type { UserSummary } from '../types'

export function useAuth() {
  const { user, accessToken, isAuthenticated, login, logout, setAccessToken } = useAuthStore()

  const handleLogin = (user: UserSummary, accessToken: string, refreshToken: string) => {
    login(user, accessToken, refreshToken)
  }

  const handleLogout = async () => {
    const { refreshToken } = useAuthStore.getState()
    await logoutAPI(refreshToken ?? undefined)
    logout()
  }

  return {
    user,
    accessToken,
    isAuthenticated,
    login: handleLogin,
    logout: handleLogout,
    setAccessToken,

    // Convenience accessors
    chapterID: user?.chapter_id ?? '',
    memberID: user?.member_id ?? '',
    role: user?.role ?? 'member',
    isSysadmin: user?.is_sysadmin ?? false,
    isAdmin: user?.role === 'admin' || user?.role === 'sysadmin' || false,
    isChair: user?.role === 'chair' || user?.role === 'admin' || user?.role === 'sysadmin' || false,
    displayName: user ? `${user.first_name} ${user.last_name}` : '',
  }
}
