import { useState, useRef, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { useWebSocket } from '../../hooks/useWebSocket'
import { getMembers } from '../../api/members'
import {
  getThreads,
  createThread,
  getThread,
  sendMessage,
  type MessageThread,
  type Message,
} from '../../api/messages'

function formatRelativeTime(iso: string) {
  const date = new Date(iso)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return date.toLocaleDateString('en', { month: 'short', day: 'numeric' })
}

export function MessagesPage() {
  const { memberID } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const { on } = useWebSocket()
  const [activeThread, setActiveThread] = useState<MessageThread | null>(null)
  const [msgInput, setMsgInput] = useState('')
  const [showNewThread, setShowNewThread] = useState(false)
  const [subject, setSubject] = useState('')
  const [initMsg, setInitMsg] = useState('')
  const [selectedMembers, setSelectedMembers] = useState<string[]>([])
  const bottomRef = useRef<HTMLDivElement>(null)

  // Real-time: invalidate message cache on new WS message event
  useEffect(() => {
    const unsub = on('MESSAGE_NEW', (payload) => {
      const p = payload as { conversation_id?: string } | null
      qc.invalidateQueries({ queryKey: ['message-threads'] })
      if (activeThread && p?.conversation_id === activeThread.id) {
        qc.invalidateQueries({ queryKey: ['thread-messages', activeThread.id] })
      }
    })
    return unsub
  }, [on, qc, activeThread])

  const { data: threads = [], isLoading: threadsLoading } = useQuery({
    queryKey: ['message-threads'],
    queryFn: getThreads,
    refetchInterval: 15000,
  })

  const { data: messages = [], isLoading: msgsLoading } = useQuery({
    queryKey: ['thread-messages', activeThread?.id],
    queryFn: () => getThread(activeThread!.id),
    enabled: !!activeThread,
    refetchInterval: 5000,
  })

  const { data: membersData } = useQuery({
    queryKey: ['members-list'],
    queryFn: () => getMembers({ per_page: 100 }),
    enabled: showNewThread,
  })

  const sendMutation = useMutation({
    mutationFn: () => sendMessage(activeThread!.id, msgInput.trim()),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['thread-messages', activeThread?.id] })
      qc.invalidateQueries({ queryKey: ['message-threads'] })
      setMsgInput('')
    },
    onError: () => showToast('Failed to send message.', 'error'),
  })

  const createMutation = useMutation({
    mutationFn: () =>
      createThread({
        subject,
        initial_message: initMsg,
        participant_ids: selectedMembers,
        is_group_chat: selectedMembers.length > 1,
      }),
    onSuccess: (thread) => {
      qc.invalidateQueries({ queryKey: ['message-threads'] })
      setActiveThread(thread)
      setShowNewThread(false)
      setSubject('')
      setInitMsg('')
      setSelectedMembers([])
    },
    onError: () => showToast('Failed to create thread.', 'error'),
  })

  // Scroll to bottom when messages load
  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages.length])

  const members = membersData?.data ?? []

  return (
    <>
      <Topbar title="Messages" />
      <main className="page-body" style={{ padding: 0, display: 'flex', height: 'calc(100vh - 60px)' }}>
        {/* Thread list */}
        <div
          style={{
            width: 280, flexShrink: 0, borderRight: '1px solid var(--border)',
            display: 'flex', flexDirection: 'column', background: 'var(--white)',
          }}
        >
          <div style={{ padding: '0.75rem 1rem', borderBottom: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontWeight: 700, fontSize: '.88rem', color: 'var(--navy)' }}>Threads</span>
            <Button size="sm" variant="gold" onClick={() => setShowNewThread(true)}>
              + New
            </Button>
          </div>

          <div style={{ flex: 1, overflowY: 'auto' }}>
            {threadsLoading && (
              <p style={{ padding: '1rem', color: 'var(--muted)', fontSize: '.82rem' }}>Loading...</p>
            )}
            {!threadsLoading && threads.length === 0 && (
              <p style={{ padding: '1rem', color: 'var(--muted)', fontSize: '.82rem' }}>No messages yet.</p>
            )}
            {threads.map((t) => (
              <div
                key={t.id}
                onClick={() => setActiveThread(t)}
                style={{
                  padding: '0.85rem 1rem',
                  borderBottom: '1px solid var(--border)',
                  cursor: 'pointer',
                  background: activeThread?.id === t.id ? 'var(--info-bg)' : 'transparent',
                  transition: 'background 0.1s',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 6 }}>
                  <span style={{ fontWeight: 600, fontSize: '.83rem', color: 'var(--ink)' }} className="truncate">
                    {t.subject || `Thread with ${t.author_first_name}`}
                  </span>
                  {t.last_message_at && (
                    <span style={{ fontSize: '.68rem', color: 'var(--faint)', flexShrink: 0 }}>
                      {formatRelativeTime(t.last_message_at)}
                    </span>
                  )}
                </div>
                {t.last_message_body && (
                  <p style={{ fontSize: '.75rem', color: 'var(--muted)', marginTop: 2 }} className="truncate">
                    {t.last_message_body}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Message thread */}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
          {!activeThread ? (
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--muted)', fontSize: '.85rem' }}>
              Select a thread or start a new conversation
            </div>
          ) : (
            <>
              {/* Thread header */}
              <div style={{ padding: '0.75rem 1.25rem', borderBottom: '1px solid var(--border)', background: 'var(--white)' }}>
                <div style={{ fontWeight: 700, fontSize: '.92rem', color: 'var(--navy)' }}>
                  {activeThread.subject || 'Conversation'}
                </div>
                <div style={{ fontSize: '.72rem', color: 'var(--muted)' }}>
                  {activeThread.message_count} message{activeThread.message_count !== 1 ? 's' : ''}
                </div>
              </div>

              {/* Messages area */}
              <div style={{ flex: 1, overflowY: 'auto', padding: '1rem 1.25rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {msgsLoading && (
                  <p style={{ color: 'var(--muted)', fontSize: '.82rem' }}>Loading messages...</p>
                )}
                {messages.map((m) => {
                  const isOwn = m.sender_id === memberID
                  return (
                    <div
                      key={m.id}
                      style={{
                        display: 'flex',
                        flexDirection: isOwn ? 'row-reverse' : 'row',
                        gap: '0.5rem',
                        alignItems: 'flex-end',
                      }}
                    >
                      <div
                        style={{
                          width: 30, height: 30, borderRadius: '50%', flexShrink: 0,
                          background: isOwn ? 'var(--navy)' : 'var(--cream2)',
                          color: isOwn ? 'var(--gold)' : 'var(--muted)',
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          fontSize: '.65rem', fontWeight: 700,
                        }}
                      >
                        {(m.sender_first_name?.[0] ?? '') + (m.sender_last_name?.[0] ?? '')}
                      </div>
                      <div style={{ maxWidth: '65%' }}>
                        {!isOwn && (
                          <div style={{ fontSize: '.68rem', color: 'var(--muted)', marginBottom: 2 }}>
                            {m.sender_first_name} {m.sender_last_name}
                          </div>
                        )}
                        <div
                          style={{
                            padding: '0.6rem 0.9rem',
                            borderRadius: isOwn ? '12px 12px 2px 12px' : '12px 12px 12px 2px',
                            background: isOwn ? 'var(--navy)' : 'var(--white)',
                            color: isOwn ? '#fff' : 'var(--ink)',
                            fontSize: '.85rem',
                            lineHeight: 1.5,
                            boxShadow: 'var(--shadow)',
                            border: isOwn ? 'none' : '1px solid var(--border)',
                          }}
                        >
                          {m.body}
                        </div>
                        <div style={{ fontSize: '.65rem', color: 'var(--faint)', marginTop: 2, textAlign: isOwn ? 'right' : 'left' }}>
                          {formatRelativeTime(m.sent_at)}
                        </div>
                      </div>
                    </div>
                  )
                })}
                <div ref={bottomRef} />
              </div>

              {/* Input area */}
              <div style={{ padding: '0.75rem 1.25rem', borderTop: '1px solid var(--border)', background: 'var(--white)' }}>
                <div style={{ display: 'flex', gap: 8 }}>
                  <input
                    className="form-input"
                    style={{ flex: 1 }}
                    placeholder="Type a message..."
                    value={msgInput}
                    onChange={(e) => setMsgInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey && msgInput.trim()) {
                        e.preventDefault()
                        sendMutation.mutate()
                      }
                    }}
                  />
                  <Button
                    variant="gold"
                    onClick={() => sendMutation.mutate()}
                    loading={sendMutation.isPending}
                    disabled={!msgInput.trim()}
                  >
                    Send
                  </Button>
                </div>
              </div>
            </>
          )}
        </div>
      </main>

      {/* New thread modal */}
      <Modal isOpen={showNewThread} onClose={() => setShowNewThread(false)} title="New Message">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Subject</label>
            <input
              className="form-input"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="What's this about?"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Recipients</label>
            <select
              className="form-input"
              multiple
              size={5}
              value={selectedMembers}
              onChange={(e) => {
                const vals = Array.from(e.target.selectedOptions).map((o) => o.value)
                setSelectedMembers(vals)
              }}
            >
              {members
                .filter((m) => m.id !== memberID)
                .map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.first_name} {m.last_name}
                  </option>
                ))}
            </select>
            <p style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 4 }}>
              Hold Cmd/Ctrl to select multiple
            </p>
          </div>
          <div className="form-group">
            <label className="form-label">Message</label>
            <textarea
              className="form-input"
              rows={3}
              value={initMsg}
              onChange={(e) => setInitMsg(e.target.value)}
              style={{ resize: 'vertical' }}
            />
          </div>
          <Button
            variant="gold"
            onClick={() => createMutation.mutate()}
            loading={createMutation.isPending}
            disabled={!subject || !initMsg || selectedMembers.length === 0}
          >
            Start Conversation
          </Button>
        </div>
      </Modal>
    </>
  )
}
