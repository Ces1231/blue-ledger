import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { getAnnouncements } from '../../api/announcements'

export function AnnouncementsPage() {
  const { isAdmin } = useAuth()

  const { data: items, isLoading } = useQuery({
    queryKey: ['announcements'],
    queryFn: getAnnouncements,
  })

  const pinned = (items ?? []).filter((a) => a.is_pinned)
  const regular = (items ?? []).filter((a) => !a.is_pinned)

  return (
    <>
      <Topbar title="Announcements" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1.25rem' }}>
          {isAdmin && (
            <Link to="/announcements/create">
              <Button size="sm" variant="gold">+ Post Announcement</Button>
            </Link>
          )}
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        {!isLoading && (items ?? []).length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No announcements yet.
              </p>
            </Card.Body>
          </Card>
        )}

        {pinned.length > 0 && (
          <div style={{ marginBottom: '1.25rem' }}>
            {pinned.map((a) => (
              <div
                key={a.id}
                className="card fade-in"
                style={{
                  marginBottom: '0.75rem',
                  borderLeft: '4px solid var(--gold)',
                  background: 'var(--gold-pale)',
                }}
              >
                <div style={{ padding: '1.25rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                        <span style={{ fontSize: '.7rem', fontWeight: 700, color: 'var(--gold)', letterSpacing: '.08em', textTransform: 'uppercase' }}>
                          Pinned
                        </span>
                        {a.category && (
                          <span style={{ fontSize: '.7rem', color: 'var(--muted)', background: 'var(--cream2)', borderRadius: 99, padding: '1px 8px' }}>
                            {a.category}
                          </span>
                        )}
                      </div>
                      <div style={{ fontWeight: 700, fontFamily: 'DM Serif Display, serif', fontSize: '1rem' }}>
                        {a.title}
                      </div>
                    </div>
                    <span style={{ fontSize: '.72rem', color: 'var(--muted)', flexShrink: 0 }}>
                      {new Date(a.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <p style={{ fontSize: '.85rem', color: 'var(--ink2)', marginTop: 8, lineHeight: 1.65, whiteSpace: 'pre-wrap' }}>
                    {a.body}
                  </p>
                  {(a.poster_first_name || a.poster_last_name) && (
                    <div style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 8 }}>
                      — Bro. {a.poster_first_name} {a.poster_last_name}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {regular.map((a) => (
          <div key={a.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
            <div style={{ padding: '1.25rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  {a.category && (
                    <span style={{ fontSize: '.7rem', color: 'var(--muted)', background: 'var(--cream2)', borderRadius: 99, padding: '1px 8px', marginBottom: 4, display: 'inline-block' }}>
                      {a.category}
                    </span>
                  )}
                  <div style={{ fontWeight: 600, fontSize: '.92rem', marginTop: a.category ? 2 : 0 }}>
                    {a.title}
                  </div>
                </div>
                <span style={{ fontSize: '.72rem', color: 'var(--muted)', flexShrink: 0 }}>
                  {new Date(a.created_at).toLocaleDateString()}
                </span>
              </div>
              <p style={{ fontSize: '.83rem', color: 'var(--ink2)', marginTop: 8, lineHeight: 1.65, whiteSpace: 'pre-wrap' }}>
                {a.body}
              </p>
              {(a.poster_first_name || a.poster_last_name) && (
                <div style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 8 }}>
                  — Bro. {a.poster_first_name} {a.poster_last_name}
                </div>
              )}
            </div>
          </div>
        ))}
      </main>
    </>
  )
}
