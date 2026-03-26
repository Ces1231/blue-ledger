import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  getProspects,
  createProspect,
  updateProspectStage,
  archiveProspect,
  type Prospect,
  type CreateProspectPayload,
} from '../../api/intake'

const STAGES = ['Lead', 'Contacted', 'Interested', 'Rush', 'Offer', 'Pledging', 'Initiated', 'Declined']

const STAGE_COLORS: Record<string, { bg: string; fg: string }> = {
  Lead:      { bg: '#F2EDE4', fg: '#6B6657' },
  Contacted: { bg: '#EAF0FB', fg: '#003087' },
  Interested:{ bg: '#FFF3DC', fg: '#7A4A00' },
  Rush:      { bg: '#F5E6C8', fg: '#7A4A00' },
  Offer:     { bg: '#EAF5EE', fg: '#1A6B3A' },
  Pledging:  { bg: '#001A4D', fg: '#C9A84C' },
  Initiated: { bg: '#EAF5EE', fg: '#1A6B3A' },
  Declined:  { bg: '#FCE4EC', fg: '#8B1A1A' },
}

function StageBadge({ stage }: { stage: string }) {
  const colors = STAGE_COLORS[stage] ?? STAGE_COLORS.Lead
  return (
    <span
      style={{
        background: colors.bg, color: colors.fg,
        borderRadius: 99, padding: '2px 10px',
        fontSize: '.7rem', fontWeight: 600, whiteSpace: 'nowrap',
      }}
    >
      {stage}
    </span>
  )
}

const EMPTY_FORM: CreateProspectPayload = {
  first_name: '', last_name: '', email: '', phone: '',
  university: '', grad_year: undefined, notes: '',
}

export function IntakePage() {
  const { isAdmin, isChair } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState<CreateProspectPayload>(EMPTY_FORM)

  const { data: prospects = [], isLoading } = useQuery({
    queryKey: ['intake'],
    queryFn: getProspects,
  })

  const createMutation = useMutation({
    mutationFn: createProspect,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['intake'] })
      showToast('Prospect added to pipeline.', 'success')
      setShowModal(false)
      setForm(EMPTY_FORM)
    },
    onError: () => showToast('Failed to add prospect.', 'error'),
  })

  const stageMutation = useMutation({
    mutationFn: ({ id, stage }: { id: string; stage: string }) =>
      updateProspectStage(id, stage),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['intake'] }),
    onError: () => showToast('Failed to update stage.', 'error'),
  })

  const archiveMutation = useMutation({
    mutationFn: archiveProspect,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['intake'] })
      showToast('Prospect archived.', 'info')
    },
  })

  if (!isAdmin && !isChair) {
    return (
      <>
        <Topbar title="Intake Pipeline" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Chair or admin access required.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  const active = prospects.filter((p) => !p.archived_at)

  return (
    <>
      <Topbar title="Intake Pipeline" />
      <main className="page-body">
        <div
          style={{
            display: 'flex', justifyContent: 'space-between',
            alignItems: 'center', marginBottom: '1.25rem',
          }}
        >
          <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            {active.length} prospects in pipeline
          </p>
          <Button variant="gold" size="sm" onClick={() => setShowModal(true)}>
            + Add Prospect
          </Button>
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading pipeline...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && active.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No prospects in the pipeline yet.
              </p>
            </Card.Body>
          </Card>
        )}

        {active.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Grad Year</th>
                  <th>University</th>
                  <th>Stage</th>
                  <th>Notes</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {active.map((p) => (
                  <ProspectRow
                    key={p.id}
                    prospect={p}
                    onStageChange={(stage) => stageMutation.mutate({ id: p.id, stage })}
                    onArchive={() => archiveMutation.mutate(p.id)}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title="Add Prospect">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">First Name</label>
              <input
                className="form-input"
                value={form.first_name}
                onChange={(e) => setForm((f) => ({ ...f, first_name: e.target.value }))}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Last Name</label>
              <input
                className="form-input"
                value={form.last_name}
                onChange={(e) => setForm((f) => ({ ...f, last_name: e.target.value }))}
              />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Email</label>
            <input
              className="form-input"
              type="email"
              value={form.email ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
            />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">University</label>
              <input
                className="form-input"
                value={form.university ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, university: e.target.value }))}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Grad Year</label>
              <input
                className="form-input"
                type="number"
                value={form.grad_year ?? ''}
                onChange={(e) =>
                  setForm((f) => ({ ...f, grad_year: e.target.value ? Number(e.target.value) : undefined }))
                }
              />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Notes</label>
            <textarea
              className="form-input"
              rows={3}
              value={form.notes ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
              style={{ resize: 'vertical' }}
            />
          </div>
          <Button
            variant="gold"
            onClick={() => createMutation.mutate(form)}
            loading={createMutation.isPending}
            disabled={!form.first_name || !form.last_name}
          >
            Add to Pipeline
          </Button>
        </div>
      </Modal>
    </>
  )
}

function ProspectRow({
  prospect,
  onStageChange,
  onArchive,
}: {
  prospect: Prospect
  onStageChange: (stage: string) => void
  onArchive: () => void
}) {
  return (
    <tr>
      <td style={{ fontWeight: 600 }}>
        {prospect.first_name} {prospect.last_name}
        {prospect.email && (
          <div style={{ fontSize: '.72rem', color: 'var(--muted)', fontWeight: 400 }}>{prospect.email}</div>
        )}
      </td>
      <td>{prospect.grad_year ?? '—'}</td>
      <td>{prospect.university ?? '—'}</td>
      <td>
        <StageBadge stage={prospect.stage} />
      </td>
      <td style={{ maxWidth: 200 }}>
        <span style={{ fontSize: '.8rem', color: 'var(--muted)' }}>{prospect.notes ?? '—'}</span>
      </td>
      <td>
        <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
          <select
            className="form-input"
            style={{ padding: '4px 8px', fontSize: '.78rem', width: 'auto' }}
            value={prospect.stage}
            onChange={(e) => onStageChange(e.target.value)}
          >
            {STAGES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
          <Button size="sm" variant="ghost" onClick={onArchive}>
            Archive
          </Button>
        </div>
      </td>
    </tr>
  )
}
