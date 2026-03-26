import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getCampaigns,
  createCampaign,
  contribute,
  type Campaign,
} from '../../api/fundraising'

function centsToDisplay(cents: number): string {
  return `$${(cents / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function ProgressBar({ raised, goal }: { raised: number; goal: number }) {
  const pct = goal > 0 ? Math.min(100, Math.round((raised / goal) * 100)) : 0
  return (
    <div style={{ background: 'var(--cream2)', borderRadius: 99, height: 8, overflow: 'hidden', margin: '8px 0 4px' }}>
      <div
        style={{
          height: '100%',
          width: `${pct}%`,
          background: pct >= 100 ? 'var(--success, #27ae60)' : 'var(--gold)',
          borderRadius: 99,
          transition: 'width .4s ease',
        }}
      />
    </div>
  )
}

interface CreateCampaignForm {
  title: string
  description: string
  goal_dollars: string
  deadline: string
}

interface DonateForm {
  amount_dollars: string
  note: string
}

export function FundraisingPage() {
  const { isAdmin, isChair } = useAuth()
  const canManage = isAdmin || isChair
  const [campaignModal, setCampaignModal] = useState(false)
  const [donateTarget, setDonateTarget] = useState<Campaign | null>(null)
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['fundraising'],
    queryFn: getCampaigns,
  })

  const createForm = useForm<CreateCampaignForm>()
  const donateForm = useForm<DonateForm>()

  const createMut = useMutation({
    mutationFn: (vals: CreateCampaignForm) =>
      createCampaign({
        title: vals.title,
        description: vals.description,
        goal_cents: Math.round(parseFloat(vals.goal_dollars || '0') * 100),
        deadline: vals.deadline || undefined,
        active: true,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['fundraising'] })
      setCampaignModal(false)
      createForm.reset()
    },
  })

  const donateMut = useMutation({
    mutationFn: (vals: DonateForm) =>
      contribute(
        donateTarget!.id,
        Math.round(parseFloat(vals.amount_dollars || '0') * 100),
        vals.note || undefined,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['fundraising'] })
      setDonateTarget(null)
      donateForm.reset()
    },
  })

  const active = items.filter((c) => c.active)
  const closed = items.filter((c) => !c.active)

  const renderCard = (c: Campaign) => {
    const pct = c.goal_cents > 0 ? Math.min(100, Math.round((c.raised_cents / c.goal_cents) * 100)) : 0
    return (
      <div key={c.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
        <div style={{ padding: '1.25rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 8 }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={{ fontWeight: 700, fontSize: '.96rem' }}>{c.title}</span>
                <span
                  style={{
                    fontSize: '.68rem',
                    fontWeight: 600,
                    borderRadius: 99,
                    padding: '1px 8px',
                    background: c.active ? 'var(--gold-pale)' : 'var(--cream2)',
                    color: c.active ? 'var(--gold-dark, #7a5a00)' : 'var(--muted)',
                  }}
                >
                  {c.active ? 'Active' : 'Closed'}
                </span>
              </div>
              {c.deadline && (
                <div style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 2 }}>
                  📅 Deadline: {new Date(c.deadline).toLocaleDateString()}
                </div>
              )}
            </div>
            {c.active && (
              <Button size="sm" variant="gold" onClick={() => setDonateTarget(c)}>Donate</Button>
            )}
          </div>

          {c.description && (
            <p style={{ fontSize: '.84rem', color: 'var(--ink2)', marginTop: 6, lineHeight: 1.6 }}>{c.description}</p>
          )}

          <ProgressBar raised={c.raised_cents} goal={c.goal_cents} />
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.75rem', color: 'var(--ink2)' }}>
            <span>{centsToDisplay(c.raised_cents)} raised</span>
            <span style={{ color: pct >= 100 ? 'var(--success, #27ae60)' : 'var(--muted)' }}>
              {pct}% of {centsToDisplay(c.goal_cents)}
            </span>
          </div>
        </div>
      </div>
    )
  }

  return (
    <>
      <Topbar title="Fundraising" />
      <main className="page-body">

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.25rem' }}>
          {canManage && (
            <Button size="sm" variant="gold" onClick={() => setCampaignModal(true)}>+ New Campaign</Button>
          )}
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading…</p></Card.Body></Card>
        )}

        {!isLoading && items.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                💰 No campaigns yet.{canManage && ' Start one!'}
              </p>
            </Card.Body>
          </Card>
        )}

        {active.length > 0 && (
          <>
            <div style={{ fontSize: '.72rem', fontWeight: 700, color: 'var(--muted)', letterSpacing: '.08em', textTransform: 'uppercase', marginBottom: 8 }}>Active</div>
            {active.map(renderCard)}
          </>
        )}

        {closed.length > 0 && (
          <>
            <div style={{ fontSize: '.72rem', fontWeight: 700, color: 'var(--muted)', letterSpacing: '.08em', textTransform: 'uppercase', margin: '1.25rem 0 8px' }}>Closed</div>
            {closed.map(renderCard)}
          </>
        )}
      </main>

      {/* New Campaign Modal */}
      <Modal isOpen={campaignModal} onClose={() => { setCampaignModal(false); createForm.reset() }} title="New Campaign" size="md">
        <form
          onSubmit={createForm.handleSubmit((v) => createMut.mutate(v))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}
        >
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Title <span style={{ color: 'var(--danger,#c0392b)' }}>*</span></label>
            <input className="input" placeholder="e.g. Step Show Fundraiser" {...createForm.register('title', { required: 'Required' })} />
            {createForm.formState.errors.title && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.75rem', marginTop: 2 }}>{createForm.formState.errors.title.message}</p>}
          </div>
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Description</label>
            <textarea className="input" rows={3} style={{ resize: 'vertical' }} {...createForm.register('description')} />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
            <div>
              <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Goal ($) <span style={{ color: 'var(--danger,#c0392b)' }}>*</span></label>
              <input className="input" type="number" min="1" step="0.01" placeholder="500.00" {...createForm.register('goal_dollars', { required: 'Required' })} />
            </div>
            <div>
              <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Deadline</label>
              <input className="input" type="date" {...createForm.register('deadline')} />
            </div>
          </div>
          {createMut.isError && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.78rem' }}>Failed to create campaign.</p>}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" size="sm" onClick={() => { setCampaignModal(false); createForm.reset() }}>Cancel</Button>
            <Button type="submit" variant="gold" size="sm" disabled={createMut.isPending}>{ createMut.isPending ? 'Creating…' : 'Create' }</Button>
          </div>
        </form>
      </Modal>

      {/* Donate Modal */}
      <Modal isOpen={!!donateTarget} onClose={() => { setDonateTarget(null); donateForm.reset() }} title={`Donate to ${donateTarget?.title ?? ''}`} size="sm">
        <form
          onSubmit={donateForm.handleSubmit((v) => donateMut.mutate(v))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}
        >
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Amount ($) <span style={{ color: 'var(--danger,#c0392b)' }}>*</span></label>
            <input className="input" type="number" min="1" step="0.01" placeholder="25.00" {...donateForm.register('amount_dollars', { required: 'Required', min: { value: 1, message: 'Min $1' } })} />
            {donateForm.formState.errors.amount_dollars && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.75rem', marginTop: 2 }}>{donateForm.formState.errors.amount_dollars.message}</p>}
          </div>
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Note (optional)</label>
            <input className="input" placeholder="Keep up the great work!" {...donateForm.register('note')} />
          </div>
          {donateMut.isError && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.78rem' }}>Failed to submit donation.</p>}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" size="sm" onClick={() => { setDonateTarget(null); donateForm.reset() }}>Cancel</Button>
            <Button type="submit" variant="gold" size="sm" disabled={donateMut.isPending}>{ donateMut.isPending ? 'Submitting…' : 'Donate' }</Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
