import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getEvents } from '../../api/events'
import type { Event } from '../../types'

export function AdminEventsPage() {
  const { isAdmin } = useAuth()
  const { showToast } = useToast()

  const { data, isLoading } = useQuery({
    queryKey: ['admin-events'],
    queryFn: () => getEvents({ per_page: 100 }),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Events — Admin" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Admin access required.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  const events = data?.data ?? []

  return (
    <>
      <Topbar title="Events — Admin" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <span style={{ fontSize: '.85rem', color: 'var(--muted)' }}>{events.length} events</span>
          <Link to="/events/create">
            <Button size="sm" variant="gold">+ Create Event</Button>
          </Link>
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading events...</p>
            </Card.Body>
          </Card>
        )}

        {events.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Date</th>
                  <th>XP</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {events.map((ev) => (
                  <tr key={ev.id}>
                    <td style={{ fontWeight: 600 }}>{ev.name}</td>
                    <td>
                      <span
                        style={{
                          background: 'var(--info-bg)', color: 'var(--info)',
                          borderRadius: 99, padding: '1px 8px', fontSize: '.68rem', fontWeight: 600,
                        }}
                      >
                        {ev.event_type}
                      </span>
                    </td>
                    <td style={{ fontSize: '.78rem', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                      {new Date(ev.event_date).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })}
                    </td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem' }}>
                      +{ev.xp_attend}
                    </td>
                    <td>
                      <span
                        style={{
                          background: ev.is_active ? 'var(--success-bg)' : 'var(--cream2)',
                          color: ev.is_active ? 'var(--success)' : 'var(--muted)',
                          borderRadius: 99, padding: '1px 8px', fontSize: '.68rem', fontWeight: 600,
                        }}
                      >
                        {ev.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </>
  )
}
