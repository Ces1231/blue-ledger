import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { getServiceLog, logService, type LogServicePayload } from '../../api/service'

export function ServiceLogPage() {
  const qc = useQueryClient()
  const [showModal, setShowModal] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['service'],
    queryFn: () => getServiceLog({ per_page: 50 }),
  })

  const { register, handleSubmit, reset, formState: { errors } } = useForm<LogServicePayload>()

  const mutation = useMutation({
    mutationFn: logService,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['service'] })
      setShowModal(false)
      reset()
    },
  })

  const entries = data?.data ?? []
  const totalHours = entries.reduce((s, e) => s + e.hours, 0)
  const verifiedHours = entries.filter((e) => e.verified).reduce((s, e) => s + e.hours, 0)

  return (
    <>
      <Topbar title="Service Log" />
      <main className="page-body">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 10, marginBottom: '1.25rem' }}>
          <div className="stat-card">
            <div className="stat-label">Total Hours</div>
            <div className="stat-value">{totalHours.toFixed(1)}</div>
            <div className="stat-sub">logged</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Verified</div>
            <div className="stat-value">{verifiedHours.toFixed(1)}</div>
            <div className="stat-sub">hours confirmed</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Pending</div>
            <div className="stat-value">{entries.filter((e) => !e.verified).length}</div>
            <div className="stat-sub">awaiting verification</div>
          </div>
        </div>

        <Card>
          <Card.Header
            title="My Service Log"
            action={
              <Button size="sm" variant="gold" onClick={() => setShowModal(true)}>
                + Log Hours
              </Button>
            }
          />
          <Card.Body>
            {isLoading && <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>}
            {!isLoading && entries.length === 0 && (
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '1.5rem 0' }}>
                No service hours logged yet. Click "Log Hours" to get started.
              </p>
            )}
            {entries.map((entry) => (
              <div
                key={entry.id}
                style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid var(--border)' }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: '.88rem' }}>{entry.event_name}</div>
                  <div style={{ fontSize: '.73rem', color: 'var(--muted)' }}>
                    {new Date(entry.service_date).toLocaleDateString()}
                    {entry.organization && ` • ${entry.organization}`}
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
                  <span style={{ fontFamily: 'DM Mono, monospace', fontSize: '.88rem', fontWeight: 600 }}>
                    {entry.hours}h
                  </span>
                  <span style={{
                    background: entry.verified ? 'var(--success-bg)' : 'var(--warn-bg)',
                    color: entry.verified ? 'var(--success)' : 'var(--warn)',
                    borderRadius: 99, padding: '2px 8px', fontSize: '.7rem', fontWeight: 600,
                  }}>
                    {entry.verified ? `Verified (+${entry.xp_awarded} XP)` : 'Pending'}
                  </span>
                </div>
              </div>
            ))}
          </Card.Body>
        </Card>
      </main>

      <Modal isOpen={showModal} onClose={() => { setShowModal(false); reset() }} title="Log Service Hours">
        <form onSubmit={handleSubmit((d) => mutation.mutate(d))} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Event / Organization Name</label>
            <input
              className="form-input"
              {...register('event_name', { required: 'Event name is required' })}
              placeholder="Community cleanup"
            />
            {errors.event_name && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.event_name.message}</p>}
          </div>

          <div className="form-group">
            <label className="form-label">Organization (optional)</label>
            <input className="form-input" {...register('organization')} placeholder="Red Cross" />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
            <div className="form-group">
              <label className="form-label">Date</label>
              <input type="date" className="form-input" {...register('service_date', { required: true })} />
            </div>
            <div className="form-group">
              <label className="form-label">Hours</label>
              <input
                type="number"
                step="0.5"
                min="0.5"
                max="24"
                className="form-input"
                {...register('hours', { required: true, valueAsNumber: true })}
                placeholder="2.0"
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Notes (optional)</label>
            <textarea className="form-input" rows={2} {...register('notes')} />
          </div>

          {mutation.isError && (
            <p style={{ color: 'var(--danger)', fontSize: '.83rem' }}>Failed to log hours. Try again.</p>
          )}

          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <Button variant="ghost" type="button" onClick={() => { setShowModal(false); reset() }}>
              Cancel
            </Button>
            <Button variant="gold" type="submit" loading={mutation.isPending}>
              Log Hours
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
