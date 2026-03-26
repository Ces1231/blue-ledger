import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  getScholarships,
  createScholarship,
  updateScholarship,
  type ScholarshipApplication,
  type CreateScholarshipPayload,
} from '../../api/scholarships'

const STATUSES = ['Researching', 'In Progress', 'Submitted', 'Awarded', 'Rejected', 'Withdrawn']

const STATUS_COLORS: Record<string, { bg: string; fg: string }> = {
  Researching: { bg: '#F2EDE4', fg: '#6B6657' },
  'In Progress':{ bg: '#EAF0FB', fg: '#003087' },
  Submitted:   { bg: '#FFF3DC', fg: '#7A4A00' },
  Awarded:     { bg: '#EAF5EE', fg: '#1A6B3A' },
  Rejected:    { bg: '#FCE4EC', fg: '#8B1A1A' },
  Withdrawn:   { bg: '#F2EDE4', fg: '#6B6657' },
}

function StatusBadge({ status }: { status: string }) {
  const colors = STATUS_COLORS[status] ?? STATUS_COLORS.Researching
  return (
    <span
      style={{
        background: colors.bg, color: colors.fg,
        borderRadius: 99, padding: '2px 10px', fontSize: '.7rem', fontWeight: 600, whiteSpace: 'nowrap',
      }}
    >
      {status}
    </span>
  )
}

const EMPTY_FORM: CreateScholarshipPayload = {
  title: '', amount_cents: undefined, provider: '', deadline: '', notes: '',
}

export function ScholarshipsPage() {
  const { isAdmin, role } = useAuth()
  const isPIA = role === 'pia' || isAdmin
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showCreate, setShowCreate] = useState(false)
  const [editTarget, setEditTarget] = useState<ScholarshipApplication | null>(null)
  const [form, setForm] = useState<CreateScholarshipPayload>(EMPTY_FORM)
  const [newStatus, setNewStatus] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['scholarships'],
    queryFn: () => getScholarships({ per_page: 100 }),
  })

  const createMutation = useMutation({
    mutationFn: createScholarship,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['scholarships'] })
      showToast('Scholarship application added.', 'success')
      setShowCreate(false)
      setForm(EMPTY_FORM)
    },
    onError: () => showToast('Failed to add application.', 'error'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      updateScholarship(id, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['scholarships'] })
      showToast('Status updated.', 'success')
      setEditTarget(null)
    },
    onError: () => showToast('Failed to update status.', 'error'),
  })

  const scholarships = data?.data ?? []

  return (
    <>
      <Topbar title="Scholarships" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            {scholarships.length} application{scholarships.length !== 1 ? 's' : ''} tracked
          </p>
          {isPIA && (
            <Button size="sm" variant="gold" onClick={() => setShowCreate(true)}>
              + Add Application
            </Button>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading scholarships...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && scholarships.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No scholarship applications tracked yet.
              </p>
            </Card.Body>
          </Card>
        )}

        {scholarships.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Title</th>
                  <th>Member</th>
                  <th>Provider</th>
                  <th>Amount</th>
                  <th>Deadline</th>
                  <th>Status</th>
                  {isPIA && <th>Actions</th>}
                </tr>
              </thead>
              <tbody>
                {scholarships.map((s) => (
                  <tr key={s.id}>
                    <td style={{ fontWeight: 600 }}>{s.title}</td>
                    <td style={{ fontSize: '.82rem' }}>
                      {s.member_first_name ? `${s.member_first_name} ${s.member_last_name}` : '—'}
                    </td>
                    <td style={{ fontSize: '.82rem', color: 'var(--muted)' }}>{s.provider ?? '—'}</td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem' }}>
                      {s.amount_cents ? `$${(s.amount_cents / 100).toLocaleString()}` : '—'}
                    </td>
                    <td style={{ fontSize: '.78rem', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                      {s.deadline
                        ? new Date(s.deadline).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })
                        : '—'}
                    </td>
                    <td>
                      <StatusBadge status={s.status} />
                    </td>
                    {isPIA && (
                      <td>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            setEditTarget(s)
                            setNewStatus(s.status)
                          }}
                        >
                          Update
                        </Button>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>

      {/* Create modal */}
      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Add Scholarship Application">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Scholarship Title</label>
            <input
              className="form-input"
              value={form.title}
              onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              placeholder="e.g. NPHC Leadership Scholarship"
            />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">Provider</label>
              <input
                className="form-input"
                value={form.provider ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, provider: e.target.value }))}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Amount ($)</label>
              <input
                className="form-input"
                type="number"
                value={form.amount_cents ? form.amount_cents / 100 : ''}
                onChange={(e) =>
                  setForm((f) => ({ ...f, amount_cents: e.target.value ? Math.round(Number(e.target.value) * 100) : undefined }))
                }
                placeholder="0.00"
              />
            </div>
          </div>
          <div className="form-group">
            <label className="form-label">Deadline</label>
            <input
              className="form-input"
              type="date"
              value={form.deadline ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, deadline: e.target.value }))}
            />
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
            disabled={!form.title}
          >
            Add Application
          </Button>
        </div>
      </Modal>

      {/* Status update modal */}
      <Modal isOpen={!!editTarget} onClose={() => setEditTarget(null)} title="Update Status" size="sm">
        {editTarget && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>{editTarget.title}</p>
            <div className="form-group">
              <label className="form-label">New Status</label>
              <select
                className="form-input"
                value={newStatus}
                onChange={(e) => setNewStatus(e.target.value)}
              >
                {STATUSES.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
            </div>
            <Button
              variant="gold"
              onClick={() => updateMutation.mutate({ id: editTarget.id, status: newStatus })}
              loading={updateMutation.isPending}
            >
              Save
            </Button>
          </div>
        )}
      </Modal>
    </>
  )
}
