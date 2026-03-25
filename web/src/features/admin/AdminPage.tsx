import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getMembers } from '../../api/members'
import { awardXP, type AwardXPPayload } from '../../api/xp'

const ACTION_CARDS = [
  {
    icon: '⭐',
    title: 'Award XP',
    description: 'Manually award XP to any member for any activity.',
    action: 'award-xp',
    bg: '#F5E6C8', fg: '#7A4A00',
  },
  {
    icon: '📥',
    title: 'Import Members',
    description: 'Bulk import members via CSV file.',
    action: 'import',
    bg: '#EAF0FB', fg: '#003087',
  },
  {
    icon: '🎯',
    title: 'Manage Quests',
    description: 'Create and manage chapter quests and milestones.',
    action: 'quests',
    bg: '#EAF5EE', fg: '#1A6B3A',
  },
  {
    icon: '📊',
    title: 'Export PIA Report',
    description: 'Export chapter health data for PIA reporting.',
    action: 'export-pia',
    bg: '#FCE4EC', fg: '#8B1A1A',
  },
]

export function AdminPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showAwardXP, setShowAwardXP] = useState(false)
  const [xpForm, setXpForm] = useState<AwardXPPayload>({
    member_id: '', xp_amount: 0, activity: '', note: '',
  })

  const { data: membersData } = useQuery({
    queryKey: ['members-admin'],
    queryFn: () => getMembers({ per_page: 200 }),
    enabled: isAdmin,
  })

  const awardMutation = useMutation({
    mutationFn: awardXP,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['members-admin'] })
      showToast(`Awarded ${xpForm.xp_amount} XP successfully.`, 'success')
      setShowAwardXP(false)
      setXpForm({ member_id: '', xp_amount: 0, activity: '', note: '' })
    },
    onError: () => showToast('Failed to award XP.', 'error'),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Admin" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Admin access required.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  const members = membersData?.data ?? []
  const activeMembers = members.filter((m) => m.status === 'active')
  const paidMembers = members.filter((m) => m.dues_status === 'paid')

  const handleAction = (action: string) => {
    if (action === 'award-xp') { setShowAwardXP(true); return }
    if (action === 'import') { showToast('Member import is available in the full release.', 'info'); return }
    if (action === 'quests') { showToast('Quest management is available in the full release.', 'info'); return }
    if (action === 'export-pia') { showToast('PIA export is available in the full release.', 'info'); return }
  }

  return (
    <>
      <Topbar title="Admin Panel" />
      <main className="page-body">
        {/* Member stats row */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: '0.75rem', marginBottom: '1.5rem' }}>
          {[
            { label: 'Total Members', value: members.length },
            { label: 'Active', value: activeMembers.length },
            { label: 'Dues Paid', value: paidMembers.length },
            { label: 'Paid Rate', value: members.length > 0 ? `${Math.round((paidMembers.length / members.length) * 100)}%` : '—' },
          ].map((stat) => (
            <div key={stat.label} className="card" style={{ padding: '1rem', textAlign: 'center' }}>
              <div style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.6rem', color: 'var(--navy)' }}>
                {stat.value}
              </div>
              <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 2 }}>{stat.label}</div>
            </div>
          ))}
        </div>

        {/* Quick links */}
        <div style={{ display: 'flex', gap: 8, marginBottom: '1.5rem', flexWrap: 'wrap' }}>
          <Link to="/admin/members">
            <Button size="sm" variant="outline">Manage Members</Button>
          </Link>
          <Link to="/admin/dues">
            <Button size="sm" variant="outline">Dues</Button>
          </Link>
          <Link to="/admin/badges">
            <Button size="sm" variant="outline">Badges</Button>
          </Link>
          <Link to="/admin/events">
            <Button size="sm" variant="outline">Events</Button>
          </Link>
          <Link to="/admin/engagement-log">
            <Button size="sm" variant="outline">Engagement Log</Button>
          </Link>
          <Link to="/admin/point-economy">
            <Button size="sm" variant="outline">Point Economy</Button>
          </Link>
          <Link to="/admin/settings">
            <Button size="sm" variant="outline">Settings</Button>
          </Link>
        </div>

        {/* Action cards grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: '1rem' }}>
          {ACTION_CARDS.map((card) => (
            <button
              key={card.action}
              onClick={() => handleAction(card.action)}
              style={{
                background: '#fff', border: '1.5px solid var(--border)',
                borderRadius: 'var(--radius-lg)', padding: '1.25rem',
                textAlign: 'left', cursor: 'pointer',
                display: 'flex', flexDirection: 'column', gap: '0.5rem',
                transition: 'box-shadow 0.15s, border-color 0.15s',
              }}
              onMouseEnter={(e) => {
                ;(e.currentTarget as HTMLElement).style.boxShadow = 'var(--shadow-md)'
                ;(e.currentTarget as HTMLElement).style.borderColor = 'var(--border2)'
              }}
              onMouseLeave={(e) => {
                ;(e.currentTarget as HTMLElement).style.boxShadow = 'none'
                ;(e.currentTarget as HTMLElement).style.borderColor = 'var(--border)'
              }}
            >
              <div
                style={{
                  width: 44, height: 44, borderRadius: 10,
                  background: card.bg, color: card.fg,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: '1.4rem',
                }}
              >
                {card.icon}
              </div>
              <div style={{ fontWeight: 700, fontSize: '.88rem', color: 'var(--ink)' }}>{card.title}</div>
              <div style={{ fontSize: '.75rem', color: 'var(--muted)', lineHeight: 1.4 }}>{card.description}</div>
            </button>
          ))}
        </div>
      </main>

      {/* Award XP Modal */}
      <Modal isOpen={showAwardXP} onClose={() => setShowAwardXP(false)} title="Award XP">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Member</label>
            <select
              className="form-input"
              value={xpForm.member_id}
              onChange={(e) => setXpForm((f) => ({ ...f, member_id: e.target.value }))}
            >
              <option value="">Select a member...</option>
              {members.map((m) => (
                <option key={m.id} value={m.id}>
                  Bro. {m.first_name} {m.last_name} ({m.member_display_id})
                </option>
              ))}
            </select>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">XP Amount</label>
              <input
                className="form-input"
                type="number"
                value={xpForm.xp_amount}
                onChange={(e) => setXpForm((f) => ({ ...f, xp_amount: Number(e.target.value) }))}
                min={1}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Activity</label>
              <input
                className="form-input"
                value={xpForm.activity}
                onChange={(e) => setXpForm((f) => ({ ...f, activity: e.target.value }))}
                placeholder="e.g. Community Award"
              />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Note (optional)</label>
            <textarea
              className="form-input"
              rows={2}
              value={xpForm.note ?? ''}
              onChange={(e) => setXpForm((f) => ({ ...f, note: e.target.value }))}
              style={{ resize: 'vertical' }}
            />
          </div>
          <Button
            variant="gold"
            onClick={() => awardMutation.mutate(xpForm)}
            loading={awardMutation.isPending}
            disabled={!xpForm.member_id || xpForm.xp_amount <= 0 || !xpForm.activity}
          >
            Award XP
          </Button>
        </div>
      </Modal>
    </>
  )
}
