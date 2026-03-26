import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getMilestones,
  createMilestone,
  deleteMilestone,
  type CreateMilestonePayload,
} from '../../api/milestones'

const TYPE_EMOJI: Record<string, string> = {
  birthday:   '🎂',
  graduation: '🎓',
  new_job:    '💼',
  engagement: '💍',
  other:      '🏅',
}

const MILESTONE_TYPES = ['birthday', 'graduation', 'new_job', 'engagement', 'other']

export function MilestonesPage() {
  const { isAdmin, isChair, memberID } = useAuth()
  const canManage = isAdmin || isChair
  const [modalOpen, setModalOpen] = useState(false)
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['milestones'],
    queryFn: () => getMilestones(),
  })

  const { register, handleSubmit, reset, formState: { errors } } = useForm<CreateMilestonePayload>({
    defaultValues: { member_id: memberID },
  })

  const createMut = useMutation({
    mutationFn: createMilestone,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones'] })
      setModalOpen(false)
      reset({ member_id: memberID })
    },
  })

  const deleteMut = useMutation({
    mutationFn: deleteMilestone,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  })

  const closeModal = () => {
    setModalOpen(false)
    reset({ member_id: memberID })
  }

  return (
    <>
      <Topbar title="Milestones" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.25rem' }}>
          <Button size="sm" variant="gold" onClick={() => setModalOpen(true)}>
            + Add Milestone
          </Button>
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        {!isLoading && items.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                🏅 No milestones yet — be the first to celebrate a brother!
              </p>
            </Card.Body>
          </Card>
        )}

        {items.map((m) => (
          <div key={m.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
            <div style={{ padding: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
                <span style={{ fontSize: '2rem', lineHeight: 1 }}>
                  {TYPE_EMOJI[m.type] ?? '🏅'}
                </span>
                <div>
                  <div style={{ fontWeight: 600, fontSize: '.93rem' }}>{m.title}</div>
                  {m.member_name && (
                    <div style={{ fontSize: '.8rem', color: 'var(--muted)', marginTop: 2 }}>
                      Bro. {m.member_name}
                    </div>
                  )}
                  <div style={{ display: 'flex', gap: 8, marginTop: 4, flexWrap: 'wrap' }}>
                    <span style={{ fontSize: '.7rem', background: 'var(--cream2)', borderRadius: 99, padding: '1px 8px', color: 'var(--ink2)' }}>
                      {m.type.replace('_', ' ')}
                    </span>
                    {m.date && (
                      <span style={{ fontSize: '.7rem', color: 'var(--muted)' }}>
                        {new Date(m.date + 'T00:00:00').toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}
                      </span>
                    )}
                  </div>
                  {m.description && (
                    <p style={{ fontSize: '.82rem', color: 'var(--ink2)', marginTop: 6, lineHeight: 1.6 }}>
                      {m.description}
                    </p>
                  )}
                </div>
              </div>
              {canManage && (
                <button
                  onClick={() => deleteMut.mutate(m.id)}
                  style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--muted)', fontSize: '1.1rem', padding: '0 0 0 8px', flexShrink: 0 }}
                  title="Delete milestone"
                  aria-label="Delete milestone"
                >
                  ✕
                </button>
              )}
            </div>
          </div>
        ))}
      </main>

      <Modal isOpen={modalOpen} onClose={closeModal} title="Add Milestone">
        <form
          onSubmit={handleSubmit((d) => createMut.mutate(d))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: 14 }}
        >
          <div className="form-group">
            <label className="form-label">Member ID</label>
            <input
              className="form-input"
              {...register('member_id', { required: 'Member ID is required' })}
              placeholder="UUID of the member being celebrated"
            />
            {errors.member_id && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.member_id.message}</p>}
          </div>

          <div className="form-group">
            <label className="form-label">Type</label>
            <select className="form-input" {...register('type', { required: 'Type is required' })}>
              <option value="">Select type...</option>
              {MILESTONE_TYPES.map((t) => (
                <option key={t} value={t}>{TYPE_EMOJI[t]} {t.replace('_', ' ')}</option>
              ))}
            </select>
            {errors.type && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.type.message}</p>}
          </div>

          <div className="form-group">
            <label className="form-label">Title</label>
            <input
              className="form-input"
              {...register('title', { required: 'Title is required', minLength: { value: 2, message: 'Min 2 characters' } })}
              placeholder="e.g. Graduated with honors from Howard University"
            />
            {errors.title && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.title.message}</p>}
          </div>

          <div className="form-group">
            <label className="form-label">
              Date{' '}
              <span style={{ color: 'var(--muted)', fontWeight: 400 }}>(optional)</span>
            </label>
            <input type="date" className="form-input" {...register('date')} />
          </div>

          <div className="form-group">
            <label className="form-label">
              Description{' '}
              <span style={{ color: 'var(--muted)', fontWeight: 400 }}>(optional)</span>
            </label>
            <textarea
              className="form-input"
              rows={3}
              {...register('description')}
              placeholder="Add a note..."
              style={{ resize: 'vertical' }}
            />
          </div>

          <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end', marginTop: 4 }}>
            <Button type="button" variant="outline" size="sm" onClick={closeModal}>
              Cancel
            </Button>
            <Button type="submit" variant="gold" size="sm" loading={createMut.isPending}>
              Save Milestone
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
