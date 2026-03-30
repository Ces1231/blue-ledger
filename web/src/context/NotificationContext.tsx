import { createContext, useContext, ReactNode, useEffect, useState, useCallback } from 'react'
import { getNotifications, markRead, type Notification } from '../api/notifications'

interface NotificationContextValue {
  notifications: Notification[]
  unreadCount: number
  isLoading: boolean
  error: string | null
  markAsRead: (id: string) => Promise<void>
  markAllAsRead: () => Promise<void>
  refreshNotifications: () => Promise<void>
}

export const NotificationContext = createContext<NotificationContextValue | undefined>(undefined)

interface NotificationProviderProps {
  children: ReactNode
}

export function NotificationProvider({ children }: NotificationProviderProps) {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const refreshNotifications = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await getNotifications()
      setNotifications(data)
    } catch (err) {
      console.error('Failed to fetch notifications:', err)
      setError('Failed to load notifications')
    } finally {
      setIsLoading(false)
    }
  }, [])

  const markAsRead = useCallback(
    async (id: string) => {
      try {
        await markRead(id)
        setNotifications((prev) =>
          prev.map((n) => (n.id === id ? { ...n, read: true } : n))
        )
      } catch (err) {
        console.error('Failed to mark notification as read:', err)
      }
    },
    []
  )

  const markAllAsRead = useCallback(async () => {
    try {
      // Note: This should call the backend markAllRead endpoint
      // For now, we'll mark all locally and call the API
      setNotifications((prev) => prev.map((n) => ({ ...n, read: true })))
      // await markAllRead() // uncomment when API is ready
    } catch (err) {
      console.error('Failed to mark all as read:', err)
    }
  }, [])

  // Load notifications on mount and set up polling
  useEffect(() => {
    refreshNotifications()

    // Poll for new notifications every 30 seconds
    const interval = setInterval(refreshNotifications, 30000)

    return () => clearInterval(interval)
  }, [refreshNotifications])

  const unreadCount = notifications.filter((n) => !n.read).length

  const value: NotificationContextValue = {
    notifications,
    unreadCount,
    isLoading,
    error,
    markAsRead,
    markAllAsRead,
    refreshNotifications,
  }

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  )
}

/** Hook to access notifications context */
export function useNotifications() {
  const context = useContext(NotificationContext)
  if (context === undefined) {
    throw new Error('useNotifications must be used within NotificationProvider')
  }
  return context
}
