import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getAIConfig, getAIChatHistory, streamChat } from '../../api/ai'
import type { AIMessage, ChatMessage } from '../../api/ai'

export default function AssistantPage() {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<{ role: 'user' | 'assistant'; content: string }[]>([])
  const [isStreaming, setIsStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  const { data: config } = useQuery({
    queryKey: ['ai-config'],
    queryFn: getAIConfig,
  })

  const { data: history } = useQuery({
    queryKey: ['ai-history'],
    queryFn: getAIChatHistory,
  })

  // Seed from history on load
  useEffect(() => {
    if (history && history.length > 0 && messages.length === 0) {
      setMessages(history.map((m: AIMessage) => ({ role: m.role, content: m.content })))
    }
  }, [history])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, isStreaming])

  const handleSend = async () => {
    const text = input.trim()
    if (!text || isStreaming) return
    setInput('')
    setError(null)

    const userMsg = { role: 'user' as const, content: text }
    const updatedMessages = [...messages, userMsg]
    setMessages(updatedMessages)

    // Add placeholder assistant message
    setMessages((prev) => [...prev, { role: 'assistant', content: '' }])
    setIsStreaming(true)

    const historyForAPI: ChatMessage[] = updatedMessages.map((m) => ({
      role: m.role,
      content: m.content,
    }))

    await streamChat(
      text,
      historyForAPI.slice(0, -1), // exclude the user message just sent (API adds it)
      (delta) => {
        setMessages((prev) => {
          const updated = [...prev]
          updated[updated.length - 1] = {
            role: 'assistant',
            content: updated[updated.length - 1].content + delta,
          }
          return updated
        })
      },
      () => setIsStreaming(false),
      (err) => {
        setError(err)
        setIsStreaming(false)
      }
    )
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const providerLabel = config?.provider === 'openai' ? 'ChatGPT' : 'Claude'
  const providerColor = config?.provider === 'openai' ? '#10a37f' : '#c9a84c'

  return (
    <div className="flex flex-col h-full" style={{ maxHeight: 'calc(100vh - 120px)' }}>
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 flex items-center gap-3">
        <div>
          <h1 className="text-xl font-bold text-navy-900" style={{ color: '#001A4D' }}>
            Chapter Assistant
          </h1>
          <p className="text-sm text-gray-500">Ask anything about chapter operations, XP, events, or fraternity history</p>
        </div>
        <span
          className="ml-auto px-2 py-1 rounded text-xs font-semibold text-white"
          style={{ backgroundColor: providerColor }}
        >
          {providerLabel}
        </span>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-400 mt-12">
            <div className="text-4xl mb-3">✨</div>
            <p className="text-base">Ask me anything about chapter operations, XP, events, or fraternity history.</p>
          </div>
        )}

        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div
              className="max-w-[75%] px-4 py-3 rounded-2xl text-sm leading-relaxed"
              style={
                msg.role === 'user'
                  ? { backgroundColor: '#001A4D', color: '#ffffff' }
                  : { backgroundColor: '#f5f0e8', color: '#1a1a2e' }
              }
            >
              {msg.content || (isStreaming && i === messages.length - 1 ? (
                <span className="inline-flex gap-1">
                  <span className="animate-bounce delay-0">•</span>
                  <span className="animate-bounce delay-75">•</span>
                  <span className="animate-bounce delay-150">•</span>
                </span>
              ) : '')}
            </div>
          </div>
        ))}

        {error && (
          <div className="text-center text-red-500 text-sm py-2">{error}</div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="px-4 py-3 border-t border-gray-200 flex gap-2 items-end">
        <textarea
          className="flex-1 resize-none rounded-xl border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          style={{ minHeight: '44px', maxHeight: '120px' }}
          placeholder="Ask your chapter assistant..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={isStreaming}
          rows={1}
        />
        <button
          onClick={handleSend}
          disabled={!input.trim() || isStreaming}
          className="px-4 py-2 rounded-xl text-white text-sm font-semibold disabled:opacity-50 transition-opacity"
          style={{ backgroundColor: '#001A4D' }}
        >
          {isStreaming ? '...' : 'Send'}
        </button>
      </div>
    </div>
  )
}
