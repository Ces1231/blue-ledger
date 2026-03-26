import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getStudyGroups,
  createStudyGroup,
  joinStudyGroup,
  deleteStudyGroup,
  type CreateStudyGroupPayload,
} from '../../api/study-groups'

export function StudyGroupsPage() {
  const { isAdmin, isChair, memberID } = useAuth()
  const canManage = isAdmin || isChair
  const [modalOpen, setModalOpen] = useState(false)
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['study-groups'],
    queryFn: getStudyGroups,
  })

  const { register, handleSubmit, reset, formState: { errors } } = useForm<CreateStudyGroupPayload>({
    defaultValues: { xp_reward: 0 },
  })

  const createMut = useMutation({
    mutationFn: createStudyGroup,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['study-groups'] })
      setModalOpen(false)
      reset({ xp_reward: 0 })
    },
  })

  const joinMut = useMutation({
    mutationFn: joinStudyGroup,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['study-groups'] }),
  })

  const deleteMut = useMutation({
    mutationFn: deleteStudyGroup,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['study-groups'] }),
  })

  const closeModal = () => {
    setModalOpen(false)
    reset({ xp_reward: 0 })
  }

  return (
    <>
      <Topbar title="Study Groups" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.25rem' }}>
          <Button size="sm" variant="gold" onClick={() => setModalOpen(true)}>
            + Create Study Group
          </Button>
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        {!isLoading && items.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                📖 No study groups yet — start one and earn XP together!
              </p>
            </Card.Body>
          </Card>
        )}

        {items.map((sg) => {
          const hasJoined = sg.member_ids.includes(memberID)
          const isHost = sg.host_id === memberID
          const canDelete = canManage || isHost

          return (
            <div key={sg.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
              <div style={{ padding: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                      <span style={{ fontWeight: 700, fontSize: '.95rem', fontFamily: 'DM Serif Display, serif' }}>
                        {sg.topic}
                      </span>
                      {isHost && (
                        <span style={{ fontSize: '.68rem', background: 'var(--gold)', color: '#fff', borderRadius: 99, padding: '1px 8px', fontWeight: 700 }}>
                          HOST
                        </span>
                      )}
                      {hasJoined && !isHost && (
                        <span style={{ fontSize: '.68rem', background: '#2e7d32', color: '#fff', borderRadius: 99, padding: '1px 8px' }}>
                          Joined ✓
                        </span>
                      )}
                    </div>
                    <div style={{ display: 'flex', gap: 12, marginTop: 5, flexWrap: 'wrap', fontSize: '.8rem', color: 'var(--muted)' }}>
                      {sg.host_name && <span>Host: Bro. {sg.host_name}</span>}
                      <span>📅 {new Date(sg.date + 'T00:00:00').toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' })}</span>
                      {sg.location && <span>📍 {sg.location}</span>}
                      <span>👥 {sg.member_ids.length} attending</span>
                      {sg.xp_reward > 0 && (
                        <span style={{ color: 'var(--gold)', fontWeight: 600 }}>+{sg.xp_reward} XP</span>
                      )}
                    </div>
                  </div>

                  <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexShrink: 0, marginLeft: 12 }}>
                    {!hasJoined && (
                      <Button
                        size="sm"
                        variant="outline"
                        loading={joinMut.isPending && joinMut.variables === sg.id}
                        onClick={() => joinMut.mutate(sg.id)}
                      >
                        Join
                      </Button>
                    )}
                    {canDelete && (
                      <button
                        onClick={() => deleteMut.mutate(sg.id)}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--muted)', fontSize: '1.1rem' }}
                        title="Delete study group"
                        aria-label="Delete study group"
                      >
                        ✕
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </main>

      <Modal isOpen={modalOpen} onClose={closeModal} title="Create Study Group">
        <form
          onSubmit={handleSubmit((d) => createMut.mutate(d))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: 14 }}
        >
          <div className="form-group">
            <label className="form-label">Topic</label>
            <input
              className="form-input"
              {...register('topic', { required: 'Topic is required', minLength: { value: 2, message: 'Min 2 characters' } })}
              placeholder="e.g. Physics Final Exam Review"
            />
            {errors.topic && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.topic.message}</p>}
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">Date</label>
              <input type="date" className="form-input" {...register('date', { required: 'Date is required' })} />
              {errors.date && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.date.message}</p>}
            </div>
            <div className="form-group">
              <label className="form-label">XP Reward</label>
              <input type="number" min={0} className="form-input" {...register('xp_reward', { valueAsNumber: true })} />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">
              Location{' '}
              <span style={{ color: 'var(--muted)', fontWeight: 400 }}>(optional)</span>
            </label>
            <input className="form-input" {...register('location')} placeholder="Library Room 204, Zoom link, etc." />
          </div>

          <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end', marginTop: 4 }}>
            <Button type="button" variant="outline" size="sm" onClick={closeModal}>
              Cancel
            </Button>
            <Button type="submit" variant="gold" size="sm" loading={createMut.isPending}>
              Create Group
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
