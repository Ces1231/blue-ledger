import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { getEngagementLog } from '../../api/xp'

const SOURCE_LABELS: Record<string, string> = {
  checkin:    'Event Check-in',
  rsvp:       'RSVP',
  admin:      'Admin Award',
  system:     'System',
  quiz:       'Quiz',
  service:    'Service',
  dues:       'Dues Paid',
  badge:      'Badge Earned',
  props:      'Props',
  mentorship: 'Mentorship',
}

const SOURCE_COLORS: Record<string, { bg: string; fg: string }> = {
  checkin:    { bg: '#EAF5EE', fg: '#1A6B3A' },
  rsvp:       { bg: '#EAF0FB', fg: '#003087' },
  admin:      { bg: '#F5E6C8', fg: '#7A4A00' },
  system:     { bg: '#F2EDE4', fg: '#6B6657' },
  quiz:       { bg: '#FFF3DC', fg: '#7A4A00' },
  service:    { bg: '#EAF5EE', fg: '#1A6B3A' },
  dues:       { bg: '#FCE4EC', fg: '#8B1A1A' },
  badge:      { bg: '#F5E6C8', fg: '#7A4A00' },
  props:      { bg: '#EAF5EE', fg: '#1A6B3A' },
  mentorship: { bg: '#EAF0FB', fg: '#003087' },
}

export function EngagementLogPage() {
  const { isAdmin } = useAuth()
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['engagement-log', page],
    queryFn: () => getEngagementLog({ page, per_page: 50 }),
    enabled: isAdmin,
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Engagement Log" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--danger)' }}>Admin access required.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  const entries = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const pages = data?.meta?.pages ?? 1

  return (
    <>
      <Topbar title="Engagement Log" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <span style={{ fontSize: '.85rem', color: 'var(--muted)' }}>{total} total entries</span>
          <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
            <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Prev
            </Button>
            <span style={{ fontSize: '.8rem', color: 'var(--muted)', padding: '0 4px' }}>
              {page} / {pages}
            </span>
            <Button size="sm" variant="outline" disabled={page >= pages} onClick={() => setPage((p) => p + 1)}>
              Next
            </Button>
          </div>
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        {entries.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Member</th>
                  <th>Activity</th>
                  <th>Source</th>
                  <th>XP</th>
                  <th>Date</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => {
                  const colors = SOURCE_COLORS[e.source] ?? SOURCE_COLORS.system
                  return (
                    <tr key={e.id}>
                      <td style={{ fontSize: '.82rem' }}>
                        {e.first_name ? `${e.first_name} ${e.last_name}` : e.member_id}
                      </td>
                      <td>{e.activity}</td>
                      <td>
                        <span
                          style={{
                            background: colors.bg, color: colors.fg,
                            borderRadius: 99, padding: '1px 8px', fontSize: '.68rem', fontWeight: 600,
                          }}
                        >
                          {SOURCE_LABELS[e.source] ?? e.source}
                        </span>
                      </td>
                      <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem', color: 'var(--success)', fontWeight: 600 }}>
                        +{e.xp_awarded}
                      </td>
                      <td style={{ fontSize: '.75rem', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                        {new Date(e.created_at).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </>
  )
}
