import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import apiClient from '../../api/client'
import type { Badge, APIResponse } from '../../types'

async function getBadges(): Promise<Badge[]> {
  const { data } = await apiClient.get<APIResponse<Badge[]>>('/badges')
  return data.data
}

async function createBadge(payload: Partial<Badge>): Promise<Badge> {
  const { data } = await apiClient.post<APIResponse<Badge>>('/badges', payload)
  return data.data
}

const RARITY_OPTIONS = ['common', 'uncommon', 'rare', 'legendary'] as const

const RARITY_COLORS: Record<string, { bg: string; fg: string }> = {
  common:    { bg: '#F2EDE4', fg: '#6B6657' },
  uncommon:  { bg: '#EAF5EE', fg: '#1A6B3A' },
  rare:      { bg: '#EAF0FB', fg: '#003087' },
  legendary: { bg: '#F5E6C8', fg: '#7A4A00' },
}

export function AdminBadgesPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({
    name: '', description: '', icon: '', rarity: 'common' as typeof RARITY_OPTIONS[number],
    xp_reward: 50, requirement: '',
  })

  const { data: badges = [], isLoading } = useQuery({
    queryKey: ['admin-badges'],
    queryFn: getBadges,
  })

  const createMutation = useMutation({
    mutationFn: () => createBadge({ ...form, criteria: { type: 'manual' } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-badges'] })
      showToast('Badge created.', 'success')
      setShowCreate(false)
    },
    onError: () => showToast('Failed to create badge.', 'error'),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Badges — Admin" />
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

  return (
    <>
      <Topbar title="Badges — Admin" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <span style={{ fontSize: '.85rem', color: 'var(--muted)' }}>{badges.length} badges</span>
          <Button size="sm" variant="gold" onClick={() => setShowCreate(true)}>+ Create Badge</Button>
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: '0.75rem' }}>
          {badges.map((badge) => {
            const colors = RARITY_COLORS[badge.rarity] ?? RARITY_COLORS.common
            return (
              <div key={badge.id} className="card fade-in" style={{ padding: '1rem', textAlign: 'center' }}>
                <div style={{ fontSize: '1.8rem', marginBottom: '0.4rem' }}>{badge.icon ?? '🎖️'}</div>
                <div style={{ fontWeight: 700, fontSize: '.85rem', color: 'var(--ink)', marginBottom: 4 }}>{badge.name}</div>
                <span
                  style={{
                    background: colors.bg, color: colors.fg,
                    borderRadius: 99, padding: '1px 8px', fontSize: '.65rem', fontWeight: 600,
                    textTransform: 'capitalize',
                  }}
                >
                  {badge.rarity}
                </span>
                <div style={{ marginTop: 6, fontSize: '.72rem', color: 'var(--muted)' }}>{badge.description}</div>
                <div style={{ marginTop: 4, fontSize: '.7rem', color: 'var(--warn)', fontWeight: 600 }}>+{badge.xp_reward} XP</div>
              </div>
            )
          })}
        </div>
      </main>

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Create Badge">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">Name</label>
              <input className="form-input" value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
            </div>
            <div className="form-group">
              <label className="form-label">Icon (emoji)</label>
              <input className="form-input" value={form.icon} onChange={(e) => setForm((f) => ({ ...f, icon: e.target.value }))} placeholder="🎖️" />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Description</label>
            <textarea className="form-input" rows={2} value={form.description} onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))} style={{ resize: 'vertical' }} />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">Rarity</label>
              <select className="form-input" value={form.rarity} onChange={(e) => setForm((f) => ({ ...f, rarity: e.target.value as typeof form.rarity }))}>
                {RARITY_OPTIONS.map((r) => <option key={r} value={r}>{r}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">XP Reward</label>
              <input className="form-input" type="number" value={form.xp_reward} onChange={(e) => setForm((f) => ({ ...f, xp_reward: Number(e.target.value) }))} />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Requirement</label>
            <input className="form-input" value={form.requirement} onChange={(e) => setForm((f) => ({ ...f, requirement: e.target.value }))} placeholder="e.g. Attend 10 events" />
          </div>
          <Button variant="gold" onClick={() => createMutation.mutate()} loading={createMutation.isPending} disabled={!form.name}>
            Create Badge
          </Button>
        </div>
      </Modal>
    </>
  )
}
