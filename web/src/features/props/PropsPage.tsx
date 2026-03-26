import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getProps, giveProps } from '../../api/props'
import { getMembers } from '../../api/members'
import type { PropsCategory } from '../../types'

const CATEGORIES: PropsCategory[] = [
  'Leadership', 'Brotherhood', 'Service', 'Academic', 'Professionalism', 'Other',
]

const CATEGORY_COLORS: Record<PropsCategory, { bg: string; fg: string }> = {
  Leadership:     { bg: '#EAF0FB', fg: '#003087' },
  Brotherhood:    { bg: '#001A4D', fg: '#C9A84C' },
  Service:        { bg: '#EAF5EE', fg: '#1A6B3A' },
  Academic:       { bg: '#F5E6C8', fg: '#7A4A00' },
  Professionalism:{ bg: '#F2EDE4', fg: '#6B6657' },
  Other:          { bg: '#FCE4EC', fg: '#8B1A1A' },
}

function Avatar({ bg, fg, name }: { bg?: string; fg?: string; name: string }) {
  const initials = name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()
  return (
    <div
      className="avatar-circle"
      style={{
        width: 36, height: 36, flexShrink: 0,
        background: bg ?? '#001A4D',
        color: fg ?? '#C9A84C',
        fontSize: '.72rem', fontWeight: 700,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        borderRadius: '50%',
      }}
    >
      {initials}
    </div>
  )
}

export function PropsPage() {
  const { memberID } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showModal, setShowModal] = useState(false)
  const [toMemberId, setToMemberId] = useState('')
  const [category, setCategory] = useState<PropsCategory>('Brotherhood')
  const [message, setMessage] = useState('')

  const { data: propsData, isLoading } = useQuery({
    queryKey: ['props'],
    queryFn: () => getProps({ per_page: 50 }),
  })

  const { data: membersData } = useQuery({
    queryKey: ['members-list'],
    queryFn: () => getMembers({ per_page: 100 }),
    enabled: showModal,
  })

  const giveMutation = useMutation({
    mutationFn: () => giveProps({ to_member_id: toMemberId, category, message }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['props'] })
      showToast('Props given! +10 XP awarded.', 'success')
      setShowModal(false)
      setToMemberId('')
      setMessage('')
    },
    onError: () => showToast('Failed to give props. Try again.', 'error'),
  })

  const props = propsData?.data ?? []
  const members = membersData?.data ?? []

  return (
    <>
      <Topbar title="Props" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            Recognize your brothers. Props award +10 XP to the recipient.
          </p>
          <Button variant="gold" size="sm" onClick={() => setShowModal(true)}>
            + Give Props
          </Button>
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading props...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && props.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No props yet. Be the first to recognize a brother!
              </p>
            </Card.Body>
          </Card>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {props.map((p) => {
            const colors = CATEGORY_COLORS[p.category] ?? CATEGORY_COLORS.Other
            const fromName = p.from_member
              ? `${p.from_member.first_name} ${p.from_member.last_name}`
              : 'A brother'
            const toName = p.to_member
              ? `${p.to_member.first_name} ${p.to_member.last_name}`
              : 'A member'
            const isOwn = p.from_id === memberID

            return (
              <div
                key={p.id}
                className="card fade-in"
                style={{ padding: '1rem 1.25rem', display: 'flex', gap: '0.75rem', alignItems: 'flex-start' }}
              >
                <Avatar
                  bg={p.from_member?.avatar_bg}
                  fg={p.from_member?.avatar_fg}
                  name={fromName}
                />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <span style={{ fontWeight: 600, fontSize: '.88rem', color: 'var(--ink)' }}>
                      {isOwn ? 'You' : fromName}
                    </span>
                    <span style={{ fontSize: '.8rem', color: 'var(--muted)' }}>gave props to</span>
                    <span style={{ fontWeight: 600, fontSize: '.88rem', color: 'var(--navy)' }}>
                      Bro. {toName}
                    </span>
                    <span
                      style={{
                        background: colors.bg, color: colors.fg,
                        borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
                      }}
                    >
                      {p.category}
                    </span>
                    <span
                      style={{
                        marginLeft: 'auto', background: 'var(--gold-bg)', color: 'var(--warn)',
                        borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
                      }}
                    >
                      +{p.xp_awarded} XP
                    </span>
                  </div>
                  {p.message && (
                    <p style={{ marginTop: 4, fontSize: '.82rem', color: 'var(--ink2)', lineHeight: 1.5 }}>
                      "{p.message}"
                    </p>
                  )}
                  <div style={{ marginTop: 4, fontSize: '.72rem', color: 'var(--faint)' }}>
                    {new Date(p.created_at).toLocaleDateString('en', {
                      month: 'short', day: 'numeric', year: 'numeric',
                    })}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </main>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title="Give Props">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="form-group">
            <label className="form-label">Recognize a Brother</label>
            <select
              className="form-input"
              value={toMemberId}
              onChange={(e) => setToMemberId(e.target.value)}
            >
              <option value="">Select a member...</option>
              {members
                .filter((m) => m.id !== memberID)
                .map((m) => (
                  <option key={m.id} value={m.id}>
                    Bro. {m.first_name} {m.last_name}
                  </option>
                ))}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Category</label>
            <select
              className="form-input"
              value={category}
              onChange={(e) => setCategory(e.target.value as PropsCategory)}
            >
              {CATEGORIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Message (optional)</label>
            <textarea
              className="form-input"
              rows={3}
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="What did they do to deserve recognition?"
              style={{ resize: 'vertical' }}
            />
          </div>

          <Button
            variant="gold"
            onClick={() => giveMutation.mutate()}
            loading={giveMutation.isPending}
            disabled={!toMemberId}
          >
            Give Props (+10 XP)
          </Button>
        </div>
      </Modal>
    </>
  )
}
