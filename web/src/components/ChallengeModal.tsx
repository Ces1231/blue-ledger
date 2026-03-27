import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from './Toast'
import { sendChallenge, type SendChallengeInput } from '../api/challenges'

// ── Types ──────────────────────────────────────────────────────────────────────

export interface ChallengeTarget {
  id: string
  name: string
  level?: string
  xp_total?: number
}

interface ChallengeModalProps {
  isOpen: boolean
  onClose: () => void
  target: ChallengeTarget
}

// ── Challenge type definitions ─────────────────────────────────────────────────

const CHALLENGE_TYPES: {
  key: SendChallengeInput['type']
  icon: string
  label: string
  desc: string
  color: string
}[] = [
  {
    key: 'trivia',
    icon: '🧠',
    label: 'Trivia Battle',
    desc: 'Answer 5 PBS / Blue Ledger questions. Most correct wins.',
    color: '#f5c842',
  },
  {
    key: 'xp_duel',
    icon: '⚡',
    label: 'XP Duel',
    desc: 'Whoever earns more XP in the next 24 hours takes the stake.',
    color: '#4a9eff',
  },
  {
    key: 'service_race',
    icon: '🏃',
    label: 'Service Race',
    desc: 'Log the most service hours in 7 days. Ties go to time of last log.',
    color: '#00cc66',
  },
  {
    key: 'streak_showdown',
    icon: '🔥',
    label: 'Streak Showdown',
    desc: 'Maintain a longer daily app-login streak over the next 30 days.',
    color: '#ff6b35',
  },
]

// ── Component ──────────────────────────────────────────────────────────────────

export function ChallengeModal({ isOpen, onClose, target }: ChallengeModalProps) {
  const { showToast } = useToast()
  const qc = useQueryClient()

  const [selectedType, setSelectedType] = useState<SendChallengeInput['type']>('trivia')
  const [xpStake, setXpStake] = useState(50)

  const mutation = useMutation({
    mutationFn: (input: SendChallengeInput) => sendChallenge(input),
    onSuccess: () => {
      showToast(`⚔️ Challenge sent to ${target.name}!`, 'success')
      qc.invalidateQueries({ queryKey: ['challenges'] })
      onClose()
    },
    onError: (err: Error) => {
      const msg = err?.message ?? 'Failed to send challenge.'
      showToast(msg.includes('409') ? 'Already an active challenge with this brother.' : msg, 'error')
    },
  })

  if (!isOpen) return null

  const selected = CHALLENGE_TYPES.find(t => t.key === selectedType)!

  const handleSubmit = () => {
    mutation.mutate({
      challenged_id: target.id,
      type: selectedType,
      xp_stake: xpStake,
    })
  }

  return (
    <div
      className="modal-overlay"
      onClick={onClose}
      style={{ zIndex: 9999 }}
    >
      <div
        className="modal-box"
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: '480px', width: '95%' }}
      >
        {/* Header */}
        <div className="modal-header">
          <span className="modal-title">⚔️ Challenge a Brother</span>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          {/* Target member */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
              padding: '12px 16px',
              background: 'rgba(74,158,255,.08)',
              borderRadius: '8px',
              border: '1px solid rgba(74,158,255,.2)',
              marginBottom: '1.2rem',
            }}
          >
            <div
              className="avatar-circle"
              style={{ width: 44, height: 44, fontSize: '.85rem', background: '#001A4D', color: '#C9A84C', flexShrink: 0 }}
            >
              {target.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()}
            </div>
            <div>
              <div style={{ fontWeight: 700, color: 'var(--ink)', fontSize: '.9rem' }}>Bro. {target.name}</div>
              {target.level && (
                <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '2px' }}>
                  {target.level}
                  {target.xp_total != null ? ` · ${target.xp_total.toLocaleString()} XP` : ''}
                </div>
              )}
            </div>
          </div>

          {/* Challenge type picker */}
          <div style={{ marginBottom: '1.2rem' }}>
            <div style={{ fontSize: '.78rem', fontWeight: 600, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: '.6rem' }}>
              Select Challenge Type
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
              {CHALLENGE_TYPES.map(t => (
                <button
                  key={t.key}
                  onClick={() => setSelectedType(t.key)}
                  style={{
                    padding: '10px 12px',
                    borderRadius: '8px',
                    border: `2px solid ${selectedType === t.key ? t.color : 'var(--border)'}`,
                    background: selectedType === t.key ? `${t.color}18` : 'transparent',
                    cursor: 'pointer',
                    textAlign: 'left',
                    transition: 'all .15s',
                  }}
                >
                  <div style={{ fontSize: '1.25rem', marginBottom: '4px' }}>{t.icon}</div>
                  <div style={{ fontSize: '.78rem', fontWeight: 700, color: 'var(--ink)' }}>{t.label}</div>
                </button>
              ))}
            </div>
            {/* Selected type description */}
            <div
              style={{
                marginTop: '10px',
                padding: '8px 12px',
                borderRadius: '6px',
                background: `${selected.color}12`,
                border: `1px solid ${selected.color}30`,
                fontSize: '.75rem',
                color: 'var(--muted)',
                lineHeight: 1.5,
              }}
            >
              {selected.desc}
            </div>
          </div>

          {/* XP stake slider */}
          <div style={{ marginBottom: '1rem' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '.4rem',
              }}
            >
              <span style={{ fontSize: '.78rem', fontWeight: 600, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em' }}>
                XP Stake
              </span>
              <span
                style={{
                  fontFamily: 'DM Mono, monospace',
                  fontWeight: 700,
                  color: 'var(--gold)',
                  fontSize: '1rem',
                }}
              >
                ⚡ {xpStake} XP
              </span>
            </div>
            <input
              type="range"
              min={10}
              max={500}
              step={10}
              value={xpStake}
              onChange={(e) => setXpStake(Number(e.target.value))}
              style={{ width: '100%', accentColor: 'var(--gold)', cursor: 'pointer' }}
            />
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.68rem', color: 'var(--muted)', marginTop: '2px' }}>
              <span>10 XP</span>
              <span>500 XP</span>
            </div>
          </div>

          {/* Stakes summary */}
          <div
            style={{
              padding: '10px 14px',
              borderRadius: '6px',
              background: 'rgba(201,168,76,.1)',
              border: '1px solid rgba(201,168,76,.25)',
              fontSize: '.78rem',
              color: 'var(--ink)',
              textAlign: 'center',
            }}
          >
            🏆 Win → <strong>+{xpStake} XP</strong> &nbsp;&nbsp;|&nbsp;&nbsp;
            💀 Lose → <strong>-{xpStake} XP</strong> applied by chapter admin
          </div>
        </div>

        {/* Footer */}
        <div className="modal-footer">
          <button
            className="btn btn-outline btn-sm"
            onClick={onClose}
            disabled={mutation.isPending}
          >
            Cancel
          </button>
          <button
            className="btn btn-primary btn-sm"
            onClick={handleSubmit}
            disabled={mutation.isPending}
            style={{ gap: '6px' }}
          >
            {mutation.isPending ? 'Sending…' : '⚔️ Send Challenge'}
          </button>
        </div>
      </div>
    </div>
  )
}
