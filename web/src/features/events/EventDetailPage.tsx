import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { getEvent, rsvpEvent } from '../../api/events'

export function EventDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()

  const { data: event, isLoading } = useQuery({
    queryKey: ['event', id],
    queryFn: () => getEvent(id!),
    enabled: !!id,
  })

  const rsvpMutation = useMutation({
    mutationFn: (status: 'yes' | 'no' | 'maybe') => rsvpEvent(id!, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['event', id] }),
  })

  if (isLoading) {
    return (
      <>
        <Topbar title="Event" />
        <main className="page-body">
          <p style={{ color: 'var(--muted)' }}>Loading...</p>
        </main>
      </>
    )
  }

  if (!event) {
    return (
      <>
        <Topbar title="Event" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--muted)' }}>Event not found.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  const date = new Date(event.event_date)

  return (
    <>
      <Topbar title={event.name} />
      <main className="page-body">
        <div style={{ marginBottom: '1rem' }}>
          <Link to="/events" style={{ fontSize: '.83rem', color: 'var(--navy)' }}>← Back to Events</Link>
        </div>

        <Card>
          <Card.Header title={event.name} />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
                <div>
                  <div style={{ fontSize: '.72rem', color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em' }}>Date</div>
                  <div style={{ fontWeight: 600 }}>
                    {date.toLocaleDateString('en', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
                    {event.event_time && ` at ${event.event_time}`}
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: '.72rem', color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em' }}>Type</div>
                  <div style={{ fontWeight: 600 }}>{event.event_type}</div>
                </div>
                {event.location && (
                  <div>
                    <div style={{ fontSize: '.72rem', color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em' }}>Location</div>
                    <div style={{ fontWeight: 600 }}>{event.location}</div>
                  </div>
                )}
                <div>
                  <div style={{ fontSize: '.72rem', color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em' }}>XP Reward</div>
                  <div style={{ fontWeight: 600, color: 'var(--gold)' }}>+{event.xp_attend} XP (attend)</div>
                </div>
              </div>

              {event.description && (
                <p style={{ fontSize: '.9rem', lineHeight: 1.65, borderTop: '1px solid var(--border)', paddingTop: 12 }}>
                  {event.description}
                </p>
              )}

              <div style={{ display: 'flex', gap: 8, borderTop: '1px solid var(--border)', paddingTop: 12 }}>
                <Button
                  variant="gold"
                  onClick={() => rsvpMutation.mutate('yes')}
                  loading={rsvpMutation.isPending}
                >
                  RSVP Yes (+{event.xp_rsvp} XP)
                </Button>
                <Button variant="outline" onClick={() => rsvpMutation.mutate('maybe')}>
                  Maybe
                </Button>
                <Button variant="ghost" onClick={() => rsvpMutation.mutate('no')}>
                  Can't go
                </Button>
              </div>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
