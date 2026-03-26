import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useCurrentMember } from '../../hooks/useCurrentMember'
import { getMemberXPHistory } from '../../api/members'

const XP_LEVELS = [
  { key: 'neo',    min: 0,    max: 249,   label: 'Neophyte' },
  { key: 'bronze', min: 250,  max: 749,   label: 'Bronze Varsity' },
  { key: 'silver', min: 750,  max: 1499,  label: 'Silver Elite' },
  { key: 'gold',   min: 1500, max: 2499,  label: 'Gold Legend' },
  { key: 'icon',   min: 2500, max: 99999, label: 'Chapter Icon' },
]

function getXPProgress(xp: number) {
  const current = XP_LEVELS.find((l) => xp >= l.min && xp <= l.max) ?? XP_LEVELS[0]
  const next = XP_LEVELS[XP_LEVELS.indexOf(current) + 1]
  if (!next) return { pct: 100, current, next: null, xpToNext: 0 }
  const pct = Math.round(((xp - current.min) / (next.min - current.min)) * 100)
  return { pct, current, next, xpToNext: next.min - xp }
}

const SOURCE_LABELS: Record<string, string> = {
  checkin: 'Event Check-in',
  rsvp: 'RSVP',
  admin: 'Admin Award',
  system: 'System',
  quiz: 'Quiz',
  service: 'Service',
  dues: 'Dues Paid',
  badge: 'Badge Earned',
  props: 'Props',
  mentorship: 'Mentorship',
}

export function ProfilePage() {
  const { member, isLoading } = useCurrentMember()

  const { data: historyData } = useQuery({
    queryKey: ['xp-history', member?.id],
    queryFn: () => getMemberXPHistory(member!.id, { per_page: 10 }),
    enabled: !!member,
  })

  if (isLoading) {
    return (
      <>
        <Topbar title="My Profile" />
        <main className="page-body">
          <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading profile...</p>
        </main>
      </>
    )
  }

  if (!member) return null

  const xpData = getXPProgress(member.xp_total)
  const initials = `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()
  const history = historyData?.data ?? []

  const ROLE_COLORS: Record<string, { bg: string; fg: string }> = {
    member:   { bg: '#EAF0FB', fg: '#003087' },
    chair:    { bg: '#F5E6C8', fg: '#7A4A00' },
    pia:      { bg: '#EAF5EE', fg: '#1A6B3A' },
    admin:    { bg: '#001A4D', fg: '#C9A84C' },
    sysadmin: { bg: '#FCE4EC', fg: '#8B1A1A' },
  }
  const roleColor = ROLE_COLORS[member.role] ?? ROLE_COLORS.member

  return (
    <>
      <Topbar title="My Profile" />
      <main className="page-body">
        {/* Hero section */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Body>
            <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'flex-start', flexWrap: 'wrap' }}>
              {/* Avatar */}
              <div
                style={{
                  width: 80, height: 80, borderRadius: '50%', flexShrink: 0,
                  background: member.avatar_bg ?? 'var(--navy)',
                  color: member.avatar_fg ?? 'var(--gold)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: '1.6rem', fontWeight: 700,
                }}
              >
                {initials}
              </div>

              {/* Info */}
              <div style={{ flex: 1, minWidth: 200 }}>
                <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
                  <h1 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.5rem', color: 'var(--navy)' }}>
                    Bro. {member.first_name} {member.last_name}
                  </h1>
                  <span
                    style={{
                      background: roleColor.bg, color: roleColor.fg,
                      borderRadius: 99, padding: '2px 10px', fontSize: '.7rem', fontWeight: 600,
                      textTransform: 'capitalize',
                    }}
                  >
                    {member.role}
                  </span>
                </div>
                <div style={{ fontSize: '.82rem', color: 'var(--muted)', marginTop: 4 }}>
                  {member.member_display_id}
                  {member.inducted_year && ` • Inducted ${member.inducted_year}`}
                  {member.city && ` • ${member.city}`}
                </div>
                {member.employer && (
                  <div style={{ fontSize: '.82rem', color: 'var(--ink2)', marginTop: 4 }}>
                    {member.job_title ? `${member.job_title} at ` : ''}{member.employer}
                  </div>
                )}
                {member.linkedin_url && (
                  <a
                    href={member.linkedin_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{ fontSize: '.78rem', color: 'var(--navy)', textDecoration: 'none', display: 'inline-block', marginTop: 4 }}
                  >
                    LinkedIn
                  </a>
                )}

                {/* XP bar */}
                <div style={{ marginTop: '0.75rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.75rem', color: 'var(--muted)', marginBottom: 4 }}>
                    <span>{xpData.current.label}</span>
                    {xpData.next && <span>{xpData.xpToNext} XP to {xpData.next.label}</span>}
                  </div>
                  <div style={{ height: 8, background: 'var(--cream2)', borderRadius: 99, overflow: 'hidden' }}>
                    <div
                      style={{
                        height: '100%', width: `${xpData.pct}%`,
                        background: 'var(--navy)', borderRadius: 99, transition: 'width 0.4s ease',
                      }}
                    />
                  </div>
                </div>
              </div>

              <Link to="/profile/edit">
                <Button size="sm" variant="outline">Edit Profile</Button>
              </Link>
            </div>
          </Card.Body>
        </Card>

        {/* Stats row */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: '0.75rem', marginBottom: '1.25rem' }}>
          {[
            { label: 'Total XP', value: member.xp_total.toLocaleString() },
            { label: 'Semester XP', value: member.xp_semester.toLocaleString() },
            { label: 'Level', value: member.level },
            { label: 'Dues Status', value: member.dues_status.charAt(0).toUpperCase() + member.dues_status.slice(1) },
          ].map((stat) => (
            <div
              key={stat.label}
              className="card"
              style={{ padding: '0.85rem', textAlign: 'center' }}
            >
              <div style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.3rem', color: 'var(--navy)' }}>
                {stat.value}
              </div>
              <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 2 }}>{stat.label}</div>
            </div>
          ))}
        </div>

        {/* Recent XP activity */}
        <Card>
          <Card.Header
            title="Recent XP Activity"
            action={<span style={{ fontSize: '.78rem', color: 'var(--muted)' }}>{history.length} entries</span>}
          />
          <Card.Body>
            {history.length === 0 && (
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>No activity yet.</p>
            )}
            <div className="table-wrap">
              {history.length > 0 && (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Activity</th>
                      <th>Source</th>
                      <th>XP</th>
                      <th>Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {history.map((entry) => (
                      <tr key={entry.id}>
                        <td>{entry.activity}</td>
                        <td>
                          <span
                            style={{
                              background: 'var(--info-bg)', color: 'var(--info)',
                              borderRadius: 99, padding: '1px 8px', fontSize: '.68rem', fontWeight: 600,
                            }}
                          >
                            {SOURCE_LABELS[entry.source] ?? entry.source}
                          </span>
                        </td>
                        <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem', color: 'var(--success)', fontWeight: 600 }}>
                          +{entry.xp_awarded}
                        </td>
                        <td style={{ fontSize: '.78rem', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                          {new Date(entry.created_at).toLocaleDateString('en', { month: 'short', day: 'numeric' })}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
