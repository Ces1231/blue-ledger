import axios from 'axios'
import type { AxiosError } from 'axios'
import { useAuthStore } from '../stores/authStore'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

export const apiClient = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ── Request interceptor: attach access token ─────────────────────────────────
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ── Response interceptor: handle 401 → token refresh ────────────────────────
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value: string) => void
  reject: (reason: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else if (token) {
      resolve(token)
    }
  })
  failedQueue = []
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as typeof error.config & { _retry?: boolean }

    if (error.response?.status !== 401 || originalRequest?._retry) {
      return Promise.reject(error)
    }

    const { refreshToken, logout, setAccessToken, login, user } = useAuthStore.getState()

    if (!refreshToken) {
      logout()
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        if (originalRequest) {
          originalRequest.headers = originalRequest.headers ?? {}
          originalRequest.headers.Authorization = `Bearer ${token}`
          originalRequest._retry = true
        }
        return apiClient(originalRequest!)
      })
    }

    isRefreshing = true
    originalRequest._retry = true

    try {
      const response = await axios.post(`${BASE_URL}/auth/refresh`, {
        refresh_token: refreshToken,
      })

      const { access_token, refresh_token, user: refreshedUser } = response.data.data

      // Update store with new tokens
      login(refreshedUser ?? user!, access_token, refresh_token)
      setAccessToken(access_token)

      processQueue(null, access_token)

      if (originalRequest) {
        originalRequest.headers = originalRequest.headers ?? {}
        originalRequest.headers.Authorization = `Bearer ${access_token}`
      }

      return apiClient(originalRequest!)
    } catch (refreshError) {
      processQueue(refreshError, null)
      logout()
      window.location.href = '/login'
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export default apiClient
