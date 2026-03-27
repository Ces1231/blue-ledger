import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAuth } from '../hooks/useAuth'
import apiClient from '../api/client'

// ── Types ──────────────────────────────────────────────────────────────────────

interface PresenceContextValue {
  onlineMembers: string[]
  isOnline: (memberID: string) => boolean
}

const PresenceContext = createContext<PresenceContextValue>({
  onlineMembers: [],
  isOnline: () => false,
})

export function usePresence(): PresenceContextValue {
  return useContext(PresenceContext)
}

// ── Provider ───────────────────────────────────────────────────────────────────

interface PresenceProviderProps {
  children: ReactNode
}

export function PresenceProvider({ children }: PresenceProviderProps) {
  const { isAuthenticated } = useAuth()
  const { on, isConnected } = useWebSocket()
  const [onlineMembers, setOnlineMembers] = useState<string[]>([])
  const pollTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Subscribe to live PRESENCE_UPDATE messages from the WebSocket
  useEffect(() => {
    const unsub = on('PRESENCE_UPDATE', (payload) => {
      const p = payload as { online_members?: string[] }
      if (Array.isArray(p?.online_members)) {
        setOnlineMembers(p.online_members)
      }
    })
    return unsub
  }, [on])

  // Poll REST fallback when WS is disconnected
  useEffect(() => {
    if (!isAuthenticated) return

    const fetchPresence = async () => {
      try {
        const res = await apiClient.get<{ data: { online_members: string[] } }>('/v1/presence')
        if (Array.isArray(res.data?.data?.online_members)) {
          setOnlineMembers(res.data.data.online_members)
        }
      } catch {
        // ignore
      }
    }

    if (!isConnected) {
      fetchPresence()
      pollTimerRef.current = setInterval(fetchPresence, 60000)
    } else {
      if (pollTimerRef.current) {
        clearInterval(pollTimerRef.current)
        pollTimerRef.current = null
      }
    }

    return () => {
      if (pollTimerRef.current) clearInterval(pollTimerRef.current)
    }
  }, [isConnected, isAuthenticated])

  const isOnline = (memberID: string) => onlineMembers.includes(memberID)

  return (
    <PresenceContext.Provider value={{ onlineMembers, isOnline }}>
      {children}
    </PresenceContext.Provider>
  )
}
