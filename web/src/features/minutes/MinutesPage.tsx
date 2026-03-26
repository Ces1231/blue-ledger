import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  getMinutes,
  getMinutesById,
  createMinutes,
  finalizeMinutes,
  type Minutes,
  type CreateMinutesPayload,
} from '../../api/minutes'

function statusBadge(m: Minutes) {
  if (m.is_finalized) return { label: 'Finalized', bg: 'var(--success-bg)', fg: 'var(--success)' }
  return { label: 'Draft', bg: 'var(--warn-bg)', fg: 'var(--warn)' }
}

export function MinutesPage() {
  const { isAdmin, isChair } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Minutes | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState<CreateMinutesPayload>({ title: '', meeting_date: '', body: '' })

  const { data, isLoading } = useQuery({
    queryKey: ['minutes'],
    queryFn: () => getMinutes({ per_page: 100 }),
  })

  const createMutation = useMutation({
    mutationFn: createMinutes,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['minutes'] })
      showToast('Meeting minutes created.', 'success')
      setShowCreate(false)
      setForm({ title: '', meeting_date: '', body: '' })
    },
    onError: () => showToast('Failed to create minutes.', 'error'),
  })

  const finalizeMutation = useMutation({
    mutationFn: (id: string) => finalizeMinutes(id),
    onSuccess: (updated) => {
      qc.invalidateQueries({ queryKey: ['minutes'] })
      setSelected(updated)
      showToast('Minutes finalized.', 'success')
    },
    onError: () => showToast('Failed to finalize minutes.', 'error'),
  })

  const minutes = data?.data ?? []
  const filtered = minutes.filter(
    (m) =>
      m.title.toLowerCase().includes(search.toLowerCase()) ||
      m.meeting_date.includes(search)
  )

  if (selected) {
    const badge = statusBadge(selected)
    return (
      <>
        <Topbar title="Meeting Minutes" />
        <main className="page-body" style={{ maxWidth: 720 }}>
          <div style={{ marginBottom: '1rem' }}>
            <Button size="sm" variant="outline" onClick={() => setSelected(null)}>
              ← Back to list
            </Button>
          </div>
          <Card>
            <Card.Header
              title={selected.title}
              action={
                isAdmin && !selected.is_finalized ? (
                  <Button
                    size="sm"
                    variant="gold"
                    onClick={() => finalizeMutation.mutate(selected.id)}
                    loading={finalizeMutation.isPending}
                  >
                    Finalize
                  </Button>
                ) : undefined
              }
            />
            <Card.Body>
              <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap' }}>
                <span
                  style={{
                    background: badge.bg, color: badge.fg,
                    borderRadius: 99, padding: '2px 10px', fontSize: '.7rem', fontWeight: 600,
                  }}
                >
                  {badge.label}
                </span>
                <span style={{ fontSize: '.8rem', color: 'var(--muted)' }}>
                  {new Date(selected.meeting_date).toLocaleDateString('en', {
                    weekday: 'long', month: 'long', day: 'numeric', year: 'numeric',
                  })}
                </span>
                {selected.author_first_name && (
                  <span style={{ fontSize: '.8rem', color: 'var(--muted)' }}>
                    By {selected.author_first_name} {selected.author_last_name}
                  </span>
                )}
                {selected.finalized_at && (
                  <span style={{ fontSize: '.75rem', color: 'var(--success)' }}>
                    Finalized {new Date(selected.finalized_at).toLocaleDateString()}
                  </span>
                )}
              </div>
              <div
                style={{
                  whiteSpace: 'pre-wrap', lineHeight: 1.7, fontSize: '.88rem',
                  color: 'var(--ink)', fontFamily: 'Instrument Sans, sans-serif',
                }}
              >
                {selected.body}
              </div>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Meeting Minutes" />
      <main className="page-body">
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap' }}>
          <input
            className="form-input"
            style={{ flex: 1, minWidth: 200, maxWidth: 360 }}
            placeholder="Search by title or date..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          {(isAdmin || isChair) && (
            <Button size="sm" variant="gold" onClick={() => setShowCreate(true)}>
              + New Minutes
            </Button>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && filtered.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No minutes found.
              </p>
            </Card.Body>
          </Card>
        )}

        <div className="table-wrap">
          {filtered.length > 0 && (
            <table className="table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Title</th>
                  <th>Author</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((m) => {
                  const badge = statusBadge(m)
                  return (
                    <tr key={m.id}>
                      <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.78rem', whiteSpace: 'nowrap' }}>
                        {new Date(m.meeting_date).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })}
                      </td>
                      <td style={{ fontWeight: 600 }}>{m.title}</td>
                      <td style={{ fontSize: '.82rem', color: 'var(--muted)' }}>
                        {m.author_first_name ? `${m.author_first_name} ${m.author_last_name}` : '—'}
                      </td>
                      <td>
                        <span
                          style={{
                            background: badge.bg, color: badge.fg,
                            borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
                          }}
                        >
                          {badge.label}
                        </span>
                      </td>
                      <td>
                        <Button size="sm" variant="outline" onClick={() => setSelected(m)}>
                          View
                        </Button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      </main>

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="New Meeting Minutes">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Title</label>
            <input
              className="form-input"
              value={form.title}
              onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              placeholder="e.g. Chapter Meeting — March 2026"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Meeting Date</label>
            <input
              className="form-input"
              type="date"
              value={form.meeting_date}
              onChange={(e) => setForm((f) => ({ ...f, meeting_date: e.target.value }))}
            />
          </div>
          <div className="form-group">
            <label className="form-label">Body</label>
            <textarea
              className="form-input"
              rows={10}
              value={form.body}
              onChange={(e) => setForm((f) => ({ ...f, body: e.target.value }))}
              placeholder="Enter meeting notes, attendees, motions, votes..."
              style={{ resize: 'vertical' }}
            />
          </div>
          <Button
            variant="gold"
            onClick={() => createMutation.mutate(form)}
            loading={createMutation.isPending}
            disabled={!form.title || !form.meeting_date || !form.body}
          >
            Create Minutes
          </Button>
        </div>
      </Modal>
    </>
  )
}
