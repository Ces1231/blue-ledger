import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { createEvent, type CreateEventPayload } from '../../api/events'

const EVENT_TYPES = ['Meeting', 'Service', 'Social', 'Conference', 'Committee', 'SBC', 'Other']

export function CreateEventPage() {
  const { isAdmin } = useAuth()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { register, handleSubmit, formState: { errors } } = useForm<CreateEventPayload>({
    defaultValues: { xp_attend: 25, xp_rsvp: 5 },
  })

  const mutation = useMutation({
    mutationFn: createEvent,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events'] })
      navigate('/events')
    },
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Create Event" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--danger)' }}>Admin access required.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Create Event" />
      <main className="page-body" style={{ maxWidth: 640 }}>
        <Card>
          <Card.Header title="New Event" />
          <Card.Body>
            <form onSubmit={handleSubmit((data) => mutation.mutate(data))}
              style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

              <div className="form-group">
                <label className="form-label">Event Name</label>
                <input
                  className="form-input"
                  {...register('name', { required: 'Name is required' })}
                  placeholder="Chapter Meeting"
                />
                {errors.name && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.name.message}</p>}
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Date</label>
                  <input
                    type="date"
                    className="form-input"
                    {...register('event_date', { required: 'Date is required' })}
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Time (optional)</label>
                  <input type="time" className="form-input" {...register('event_time')} />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Event Type</label>
                  <select className="form-input" {...register('event_type', { required: true })}>
                    {EVENT_TYPES.map((t) => <option key={t}>{t}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">Location</label>
                  <input className="form-input" {...register('location')} placeholder="Room 101" />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">XP for Attending</label>
                  <input type="number" className="form-input" {...register('xp_attend', { valueAsNumber: true })} />
                </div>
                <div className="form-group">
                  <label className="form-label">XP for RSVP</label>
                  <input type="number" className="form-input" {...register('xp_rsvp', { valueAsNumber: true })} />
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">Description (optional)</label>
                <textarea className="form-input" rows={3} {...register('description')} />
              </div>

              {mutation.isError && (
                <p style={{ color: 'var(--danger)', fontSize: '.85rem' }}>
                  Failed to create event. Please try again.
                </p>
              )}

              <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                <Button variant="ghost" type="button" onClick={() => navigate('/events')}>
                  Cancel
                </Button>
                <Button variant="gold" type="submit" loading={mutation.isPending}>
                  Create Event
                </Button>
              </div>
            </form>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
