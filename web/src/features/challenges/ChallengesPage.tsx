import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { useWebSocket } from '../../hooks/useWebSocket'
import { TriviaGame } from './TriviaGame'
import {
  listChallenges,
  acceptChallenge,
  declineChallenge,
  type Challenge,
} from '../../api/challenges'

// ── Constants ──────────────────────────────────────────────────────────────────

const TYPE_META: Record<string, { icon: string; label: string; color: string }> = {
  xp_duel:         { icon: '⚡', label: 'XP Duel',         color: '#4a9eff' },
  service_race:    { icon: '🏃', label: 'Service Race',     color: '#00cc66' },
  trivia:          { icon: '🧠', label: 'Trivia Battle',    color: '#f5c842' },
  streak_showdown: { icon: '🔥', label: 'Streak Showdown',  color: '#ff6b35' },
}

const STATUS_LABEL: Record<string, string> = {
  pending:   '⏳ Pending',
  accepted:  '✅ Accepted',
  declined:  '❌ Declined',
  active:    '🔵 Active',
  completed: '🏆 Completed',
  expired:   '💀 Expired',
}

function formatRelative(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(ms / 60000)
  if (mins < 1)  return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

function timeUntil(iso: string) {
  const ms = new Date(iso).getTime() - Date.now()
  if (ms <= 0) return 'expired'
  const hours = Math.floor(ms / 3600000)
  if (hours < 24) return `${hours}h remaining`
  return `${Math.floor(hours / 24)}d remaining`
}

// ── Challenge Card ─────────────────────────────────────────────────────────────

function ChallengeCard({
  ch,
  myID,
  onAccept,
  onDecline,
  onPlay,
}: {
  ch: Challenge
  myID: string
  onAccept: (id: string) => void
  onDecline: (id: string) => void
  onPlay?: (ch: Challenge) => void
}) {
  const meta = TYPE_META[ch.type] ?? { icon: '⚔️', label: ch.type, color: '#4a9eff' }
  const isIncoming = ch.challenged_id === myID && ch.status === 'pending'
  const isWinner = ch.winner_id === myID
  const opponentName = ch.challenger_id === myID ? ch.challenged_name : ch.challenger_name
  const canPlay = ch.status === 'active' && ch.type === 'trivia' && onPlay

  return (
    <div
      style={{
        background: 'rgba(74,158,255,.06)',
        border: `1px solid ${meta.color}33`,
        borderLeft: `4px solid ${meta.color}`,
        borderRadius: '8px',
        padding: '1rem 1.2rem',
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
        flexWrap: 'wrap',
      }}
    >
      {/* Type icon */}
      <div style={{ fontSize: '1.8rem', lineHeight: 1 }}>{meta.icon}</div>

      {/* Info */}
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
          <span style={{ fontWeight: 700, color: meta.color, fontSize: '.85rem' }}>
            {meta.label.toUpperCase()}
          </span>
          <span
            style={{
              fontSize: '.7rem',
              fontFamily: 'DM Mono, monospace',
              background: 'rgba(255,255,255,.08)',
              padding: '2px 6px',
              borderRadius: '4px',
              color: 'var(--muted)',
            }}
          >
            {STATUS_LABEL[ch.status]}
          </span>
        </div>
        <div style={{ fontSize: '.84rem', color: 'var(--ink)', marginTop: '3px' }}>
          {ch.challenger_id === myID
            ? `You challenged ${opponentName ?? '—'}`
            : `${opponentName ?? '—'} challenges you`}
        </div>
        <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '3px', display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <span>⚡ Stake: <strong>{ch.xp_stake} XP</strong></span>
          {ch.status === 'pending' && <span>⏱ {timeUntil(ch.expires_at)}</span>}
          <span>{formatRelative(ch.created_at)}</span>
        </div>
      </div>

      {/* Win/loss result */}
      {ch.status === 'completed' && (
        <div
          style={{
            padding: '4px 14px',
            borderRadius: '20px',
            fontWeight: 700,
            fontSize: '.78rem',
            background: isWinner ? 'rgba(0,255,136,.15)' : 'rgba(255,80,80,.12)',
            color: isWinner ? '#00ff88' : '#ff5050',
            border: `1px solid ${isWinner ? '#00ff8844' : '#ff505044'}`,
          }}
        >
          {isWinner ? `✅ WON +${ch.xp_stake} XP` : `❌ LOST -${Math.floor(ch.xp_stake / 2)} XP`}
        </div>
      )}

      {/* Accept / Decline actions */}
      {isIncoming && (
        <div style={{ display: 'flex', gap: '8px', flexShrink: 0 }}>
          <Button size="sm" onClick={() => onAccept(ch.id)}>✅ Accept</Button>
          <Button size="sm" variant="ghost" onClick={() => onDecline(ch.id)}>❌ Decline</Button>
        </div>
      )}

      {/* Play button for active trivia */}
      {canPlay && (
        <button
          className="btn btn-primary btn-sm"
          onClick={() => onPlay(ch)}
          style={{ flexShrink: 0 }}
        >
          🧠 Play Now!
        </button>
      )}
    </div>
  )
}

// ── Main Page ──────────────────────────────────────────────────────────────────

export function ChallengesPage() {
  const { memberID } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const { on } = useWebSocket()
  const [activeTab, setActiveTab] = useState<'incoming' | 'sent' | 'active' | 'history'>('incoming')
  const [triviaChallenge, setTriviaChallenge] = useState<Challenge | null>(null)

  const { data: challenges = [], isLoading } = useQuery({
    queryKey: ['challenges'],
    queryFn: listChallenges,
    refetchInterval: 30000,
  })

  // Real-time updates via WebSocket
  useEffect(() => {
    const unsubInvite = on('CHALLENGE_INVITE', () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
      showToast('⚔️ New challenge received!', 'info')
    })
    const unsubAccepted = on('CHALLENGE_ACCEPTED', () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
      showToast('✅ Your challenge was accepted!', 'success')
    })
    const unsubResult = on('CHALLENGE_RESULT', () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
    })
    return () => { unsubInvite(); unsubAccepted(); unsubResult() }
  }, [on, qc, showToast])

  const acceptMutation = useMutation({
    mutationFn: acceptChallenge,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
      showToast('✅ Challenge accepted! Get ready.', 'success')
    },
    onError: () => showToast('Failed to accept challenge.', 'error'),
  })

  const declineMutation = useMutation({
    mutationFn: declineChallenge,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
      showToast('Challenge declined.', 'info')
    },
    onError: () => showToast('Failed to decline challenge.', 'error'),
  })

  // Segment challenges
  const incoming = challenges.filter(c => c.challenged_id === memberID && c.status === 'pending')
  const sent     = challenges.filter(c => c.challenger_id === memberID && c.status === 'pending')
  const active   = challenges.filter(c => c.status === 'active' && (c.challenger_id === memberID || c.challenged_id === memberID))
  const history  = challenges.filter(c => ['completed', 'declined', 'expired'].includes(c.status))

  const wins   = history.filter(c => c.winner_id === memberID).length
  const losses = history.filter(c => c.status === 'completed' && c.winner_id && c.winner_id !== memberID).length
  const xpWon  = history.filter(c => c.winner_id === memberID).reduce((s, c) => s + c.xp_stake, 0)

  const tabs = [
    { key: 'incoming', label: `📨 Incoming`, count: incoming.length },
    { key: 'sent',     label: `📤 Sent`,     count: sent.length },
    { key: 'active',   label: `🔵 Active`,   count: active.length },
    { key: 'history',  label: `📜 History`,  count: history.length },
  ] as const

  const tabData: Record<string, Challenge[]> = { incoming, sent, active, history }
  const currentList = tabData[activeTab] ?? []

  return (
    <div>
      <Topbar title="⚔️ Challenges" />

      {/* Trivia game overlay */}
      {triviaChallenge && (
        <TriviaGame
          challenge={triviaChallenge}
          myID={memberID}
          onClose={() => setTriviaChallenge(null)}
        />
      )}

      <div className="page-content">

        {/* Stats bar */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: '1rem',
            marginBottom: '1.5rem',
          }}
        >
          {[
            { label: '🏆 Wins',        value: wins },
            { label: '💀 Losses',      value: losses },
            { label: '⚡ XP Won',      value: `+${xpWon}` },
          ].map(stat => (
            <Card key={stat.label} style={{ textAlign: 'center', padding: '.9rem' }}>
              <div style={{ fontSize: '1.6rem', fontWeight: 700, color: 'var(--gold)' }}>{stat.value}</div>
              <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '4px' }}>{stat.label}</div>
            </Card>
          ))}
        </div>

        {/* Tab nav */}
        <div style={{ display: 'flex', gap: '6px', marginBottom: '1.2rem', flexWrap: 'wrap' }}>
          {tabs.map(tab => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              style={{
                padding: '6px 14px',
                borderRadius: '20px',
                border: `1px solid ${activeTab === tab.key ? 'var(--neon-blue)' : 'var(--border)'}`,
                background: activeTab === tab.key ? 'rgba(74,158,255,.15)' : 'transparent',
                color: activeTab === tab.key ? 'var(--neon-blue)' : 'var(--muted)',
                cursor: 'pointer',
                fontSize: '.8rem',
                fontWeight: activeTab === tab.key ? 700 : 400,
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
              }}
            >
              {tab.label}
              {tab.count > 0 && (
                <span
                  style={{
                    background: activeTab === tab.key ? 'var(--neon-blue)' : 'var(--muted)',
                    color: '#060D1A',
                    borderRadius: '50%',
                    width: '18px',
                    height: '18px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '.65rem',
                    fontWeight: 700,
                  }}
                >
                  {tab.count}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Challenge list */}
        <Card>
          {isLoading ? (
            <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--muted)' }}>Loading challenges…</div>
          ) : currentList.length === 0 ? (
            <div style={{ padding: '2.5rem', textAlign: 'center' }}>
              <div style={{ fontSize: '2.5rem', marginBottom: '.5rem' }}>⚔️</div>
              <div style={{ color: 'var(--muted)', fontSize: '.85rem' }}>
                {activeTab === 'incoming'
                  ? 'No pending challenges. Show brothers what you\'re made of!'
                  : activeTab === 'sent'
                  ? 'No sent challenges yet. Challenge a brother from their profile!'
                  : activeTab === 'active'
                  ? 'No active battles right now.'
                  : 'No challenge history yet.'}
              </div>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '.75rem', padding: '1rem' }}>
              {currentList.map(ch => (
                <ChallengeCard
                  key={ch.id}
                  ch={ch}
                  myID={memberID}
                  onAccept={(id) => acceptMutation.mutate(id)}
                  onDecline={(id) => declineMutation.mutate(id)}
                  onPlay={ch.type === 'trivia' ? setTriviaChallenge : undefined}
                />
              ))}
            </div>
          )}
        </Card>

      </div>
    </div>
  )
}
