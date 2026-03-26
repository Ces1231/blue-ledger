import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getCommittees,
  createCommittee,
  joinCommittee,
  type Committee,
} from '../../api/committees'

interface CreateCommitteeForm {
  name: string
  description: string
  meeting_schedule: string
}

export function CommitteesPage() {
  const { isAdmin, isChair } = useAuth()
  const canManage = isAdmin || isChair
  const [modalOpen, setModalOpen] = useState(false)
  const [joiningId, setJoiningId] = useState<string | null>(null)
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['committees'],
    queryFn: getCommittees,
  })

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } =
    useForm<CreateCommitteeForm>()

  const createMut = useMutation({
    mutationFn: (vals: CreateCommitteeForm) =>
      createCommittee({
        name: vals.name,
        description: vals.description,
        meeting_schedule: vals.meeting_schedule || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['committees'] })
      setModalOpen(false)
      reset()
    },
  })

  const joinMut = useMutation({
    mutationFn: (id: string) => joinCommittee(id),
    onMutate: (id) => setJoiningId(id),
    onSettled: () => setJoiningId(null),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['committees'] }),
  })

  const closeModal = () => { setModalOpen(false); reset() }

  return (
    <>
      <Topbar title="Committees" />
      <main className="page-body">

        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.25rem' }}>
          {canManage && (
            <Button size="sm" variant="gold" onClick={() => setModalOpen(true)}>+ New Committee</Button>
          )}
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading…</p></Card.Body></Card>
        )}

        {!isLoading && items.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                🏛️ No committees yet.{canManage && ' Create the first one!'}
              </p>
            </Card.Body>
          </Card>
        )}

        {items.map((c: Committee) => (
          <div key={c.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
            <div style={{ padding: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 8 }}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 700, fontSize: '.96rem', marginBottom: 2 }}>🏛️ {c.name}</div>
                {c.description && (
                  <p style={{ fontSize: '.84rem', color: 'var(--ink2)', lineHeight: 1.6, margin: '4px 0' }}>{c.description}</p>
                )}
                {c.meeting_schedule && (
                  <div style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 4 }}>
                    🗓️ {c.meeting_schedule}
                  </div>
                )}
              </div>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => joinMut.mutate(c.id)}
                disabled={joiningId === c.id}
              >
                {joiningId === c.id ? 'Joining…' : 'Join'}
              </Button>
            </div>
          </div>
        ))}
      </main>

      <Modal isOpen={modalOpen} onClose={closeModal} title="New Committee" size="md">
        <form
          onSubmit={handleSubmit((v) => createMut.mutate(v))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}
        >
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Name <span style={{ color: 'var(--danger,#c0392b)' }}>*</span></label>
            <input className="input" placeholder="e.g. Scholarship Committee" {...register('name', { required: 'Required' })} />
            {errors.name && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.name.message}</p>}
          </div>
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Description</label>
            <textarea className="input" rows={3} style={{ resize: 'vertical' }} placeholder="Purpose and responsibilities…" {...register('description')} />
          </div>
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Meeting Schedule</label>
            <input className="input" placeholder="e.g. Every 2nd Sunday at 4 PM" {...register('meeting_schedule')} />
          </div>
          {createMut.isError && <p style={{ color: 'var(--danger,#c0392b)', fontSize: '.78rem' }}>Failed to create committee.</p>}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" size="sm" onClick={closeModal}>Cancel</Button>
            <Button type="submit" variant="gold" size="sm" disabled={isSubmitting || createMut.isPending}>{createMut.isPending ? 'Creating…' : 'Create'}</Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
