import { useEffect, useRef, useCallback, useState } from 'react'
import { useAuth } from './useAuth'

// ── Types ──────────────────────────────────────────────────────────────────────

export type WSMessageType =
  | 'PRESENCE_UPDATE'
  | 'CHALLENGE_INVITE'
  | 'CHALLENGE_ACCEPTED'
  | 'CHALLENGE_RESULT'
  | 'NOTIFICATION'
  | 'PROPS_RECEIVED'
  | 'MESSAGE_NEW'
  | 'PONG'

export interface WSMessage {
  type: WSMessageType | string
  payload?: unknown
}

type MessageHandler = (payload: unknown) => void

const WS_BASE_URL = import.meta.env.VITE_WS_URL ?? ''

function getWsUrl(accessToken: string): string {
  if (WS_BASE_URL) {
    return `${WS_BASE_URL}/v1/ws?token=${accessToken}`
  }
  // In Docker/production nginx proxies /api/ws → ws://api:8080/v1/ws
  // Use same origin so the nginx upgrade path is used automatically.
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${window.location.host}/api/ws?token=${accessToken}`
}

// ── Hook ───────────────────────────────────────────────────────────────────────

export interface UseWebSocketReturn {
  send: (type: string, payload?: unknown) => void
  on: (type: string, handler: MessageHandler) => () => void
  isConnected: boolean
}

export function useWebSocket(): UseWebSocketReturn {
  const { accessToken, isAuthenticated } = useAuth()
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map())
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const backoffRef = useRef(1000) // start at 1s, max 30s
  const [isConnected, setIsConnected] = useState(false)
  const mountedRef = useRef(true)

  const connect = useCallback(() => {
    if (!isAuthenticated || !accessToken || !mountedRef.current) return
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const url = getWsUrl(accessToken)
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      if (!mountedRef.current) return
      backoffRef.current = 1000 // reset backoff on successful connect
      setIsConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        // Messages may be newline-concatenated
        const lines = event.data.split('\n').filter(Boolean)
        for (const line of lines) {
          const msg: WSMessage = JSON.parse(line)
          const handlers = handlersRef.current.get(msg.type)
          if (handlers) {
            handlers.forEach((h) => h(msg.payload))
          }
          // Also dispatch to '*' catch-all handlers
          const all = handlersRef.current.get('*')
          if (all) {
            all.forEach((h) => h(msg))
          }
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      setIsConnected(false)
      wsRef.current = null
      // Exponential backoff reconnect
      const delay = Math.min(backoffRef.current, 30000)
      backoffRef.current = Math.min(backoffRef.current * 2, 30000)
      reconnectTimerRef.current = setTimeout(connect, delay)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [isAuthenticated, accessToken])

  useEffect(() => {
    mountedRef.current = true
    if (isAuthenticated && accessToken) {
      connect()
    }
    return () => {
      mountedRef.current = false
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current)
      wsRef.current?.close()
    }
  }, [isAuthenticated, accessToken, connect])

  // Ping every 30s to keep presence key alive
  useEffect(() => {
    if (!isConnected) return
    const timer = setInterval(() => {
      send('PING')
    }, 30000)
    return () => clearInterval(timer)
  }, [isConnected]) // eslint-disable-line react-hooks/exhaustive-deps

  const send = useCallback((type: string, payload?: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }))
    }
  }, [])

  const on = useCallback((type: string, handler: MessageHandler): (() => void) => {
    if (!handlersRef.current.has(type)) {
      handlersRef.current.set(type, new Set())
    }
    handlersRef.current.get(type)!.add(handler)
    return () => {
      handlersRef.current.get(type)?.delete(handler)
    }
  }, [])

  return { send, on, isConnected }
}
