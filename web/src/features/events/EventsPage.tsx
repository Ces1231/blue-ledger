import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { getEvents, rsvpEvent } from '../../api/events'
import type { Event } from '../../types'

const EVENT_TYPE_COLORS: Record<string, { bg: string; fg: string }> = {
  Meeting:   { bg: '#EAF0FB', fg: '#003087' },
  Service:   { bg: '#EAF5EE', fg: '#1A6B3A' },
  Social:    { bg: '#F5E6C8', fg: '#7A4A00' },
  Conference:{ bg: '#FCE4EC', fg: '#8B1A1A' },
  Committee: { bg: '#F2EDE4', fg: '#6B6657' },
  SBC:       { bg: '#001A4D', fg: '#C9A84C' },
  Other:     { bg: '#F2EDE4', fg: '#6B6657' },
}

function EventCard({ event }: { event: Event }) {
  const qc = useQueryClient()
  const { isAdmin } = useAuth()
  const colors = EVENT_TYPE_COLORS[event.event_type] ?? EVENT_TYPE_COLORS.Other

  const rsvpMutation = useMutation({
    mutationFn: (status: 'yes' | 'no' | 'maybe') => rsvpEvent(event.id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  })

  const date = new Date(event.event_date)

  return (
    <div className="card fade-in" style={{ marginBottom: '1rem' }}>
      <div style={{ display: 'flex', gap: '1rem', alignItems: 'flex-start', padding: '1.25rem' }}>
        <div style={{
          background: colors.bg, color: colors.fg,
          borderRadius: 10, padding: '8px 12px', textAlign: 'center',
          minWidth: 56, flexShrink: 0, fontFamily: 'DM Mono, monospace',
        }}>
          <div style={{ fontSize: '1.3rem', fontWeight: 700, lineHeight: 1 }}>
            {date.getDate()}
          </div>
          <div style={{ fontSize: '.65rem', textTransform: 'uppercase', letterSpacing: '.07em' }}>
            {date.toLocaleDateString('en', { month: 'short' })}
          </div>
        </div>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
            <div>
              <Link
                to={`/events/${event.id}`}
                style={{ fontWeight: 600, fontSize: '.95rem', color: 'var(--ink)', textDecoration: 'none' }}
              >
                {event.name}
              </Link>
              <div style={{ fontSize: '.78rem', color: 'var(--muted)', marginTop: 2 }}>
                {event.event_type}
                {event.location && ` • ${event.location}`}
                {event.event_time && ` • ${event.event_time}`}
              </div>
            </div>
            <span
              style={{
                background: colors.bg, color: colors.fg,
                borderRadius: 99, padding: '2px 10px', fontSize: '.7rem',
                fontWeight: 600, whiteSpace: 'nowrap',
              }}
            >
              +{event.xp_attend} XP
            </span>
          </div>

          {event.description && (
            <p style={{ fontSize: '.8rem', color: 'var(--muted)', marginTop: 6, lineHeight: 1.5 }}>
              {event.description}
            </p>
          )}

          <div style={{ display: 'flex', gap: 8, marginTop: 10, flexWrap: 'wrap' }}>
            <Button
              size="sm"
              variant="gold"
              onClick={() => rsvpMutation.mutate('yes')}
              loading={rsvpMutation.isPending}
            >
              RSVP Yes (+{event.xp_rsvp} XP)
            </Button>
            <Button size="sm" variant="outline" onClick={() => rsvpMutation.mutate('maybe')}>
              Maybe
            </Button>
            <Button size="sm" variant="ghost" onClick={() => rsvpMutation.mutate('no')}>
              Can't go
            </Button>
            {isAdmin && (
              <Link to="/scanner" style={{ marginLeft: 'auto' }}>
                <Button size="sm" variant="outline">Scan QR</Button>
              </Link>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export function EventsPage() {
  const { isAdmin } = useAuth()
  const [filter, setFilter] = useState<'upcoming' | 'all'>('upcoming')

  const { data, isLoading } = useQuery({
    queryKey: ['events', filter],
    queryFn: () => getEvents({ upcoming: filter === 'upcoming', per_page: 50 }),
  })

  const events = data?.data ?? []

  return (
    <>
      <Topbar title="Events" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              size="sm"
              variant={filter === 'upcoming' ? 'primary' : 'outline'}
              onClick={() => setFilter('upcoming')}
            >
              Upcoming
            </Button>
            <Button
              size="sm"
              variant={filter === 'all' ? 'primary' : 'outline'}
              onClick={() => setFilter('all')}
            >
              All Events
            </Button>
          </div>
          {isAdmin && (
            <Link to="/events/create">
              <Button size="sm" variant="gold">+ Create Event</Button>
            </Link>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading events...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && events.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No {filter === 'upcoming' ? 'upcoming' : ''} events found.
              </p>
            </Card.Body>
          </Card>
        )}

        {events.map((event) => (
          <EventCard key={event.id} event={event} />
        ))}
      </main>
    </>
  )
}
