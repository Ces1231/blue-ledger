import { Link } from 'react-router-dom'
import { Topbar } from '../../components/Topbar'
import { Card, StatCard } from '../../components/Card'
import { useDashboard } from './useDashboard'
import { useAuth } from '../../hooks/useAuth'

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

const QUICK_ACTIONS = [
  { icon: '📅', title: 'Events',     sub: 'RSVP & check in',    path: '/events',      bg: '#EAF0FB', fg: '#003087' },
  { icon: '🏆', title: 'Leaderboard',sub: 'See your rank',      path: '/leaderboard', bg: '#F5E6C8', fg: '#7A4A00' },
  { icon: '🤝', title: 'Service',    sub: 'Log service hours',  path: '/service',     bg: '#EAF5EE', fg: '#1A6B3A' },
  { icon: '💳', title: 'Dues',       sub: 'Pay your dues',      path: '/dues',        bg: '#FCE4EC', fg: '#8B1A1A' },
  { icon: '👥', title: 'Members',    sub: 'Browse directory',   path: '/members',     bg: '#EAF0FB', fg: '#003087' },
  { icon: '⭐', title: 'Quests',     sub: 'Track progress',     path: '/quests',      bg: '#F5E6C8', fg: '#7A4A00' },
]

export function DashboardPage() {
  const { user } = useAuth()
  const {
    member,
    totalMembers,
    upcomingEvents,
    leaderboard,
    myRank,
    unpaidDues,
    isLoading,
  } = useDashboard()

  const xpData = member ? getXPProgress(member.xp_total) : null

  return (
    <>
      <Topbar title="Dashboard" />
      <main className="page-body">
        {/* Hero XP section */}
        {member && xpData && (
          <div className="dash-hero fade-in">
            <div className="dash-greeting">Welcome back</div>
            <div className="dash-name">
              Bro. {user?.first_name} {user?.last_name}
            </div>
            <div className="dash-meta-row">
              <div>
                <div className="dash-xp-big">{member.xp_total.toLocaleString()}</div>
                <div className="dash-xp-label">Total XP</div>
              </div>
              <div className="dash-level-badge">{member.level}</div>
              {xpData.next && (
                <div className="dash-xp-bar-wrap">
                  <div className="dash-xp-bar-label">
                    <span>Level Progress</span>
                    <span>{xpData.xpToNext} XP to {xpData.next.label}</span>
                  </div>
                  <div className="dash-xp-bar">
                    <div className="dash-xp-fill" style={{ width: `${xpData.pct}%` }} />
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Quick actions */}
        <div className="quick-actions">
          {QUICK_ACTIONS.map((a) => (
            <Link key={a.path} to={a.path} style={{ textDecoration: 'none' }}>
              <div className="quick-card">
                <div className="quick-icon" style={{ background: a.bg, color: a.fg }}>
                  {a.icon}
                </div>
                <div>
                  <div className="quick-title">{a.title}</div>
                  <div className="quick-sub">{a.sub}</div>
                </div>
              </div>
            </Link>
          ))}
        </div>

        {/* Dashboard grid */}
        <div className="dash-grid">
          {/* Left column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {/* Stats row */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: '10px' }}>
              <StatCard
                label="Chapter Rank"
                value={myRank > 0 ? `#${myRank}` : '—'}
                sub="of leaderboard"
              />
              <StatCard
                label="Semester XP"
                value={(member?.xp_semester ?? 0).toLocaleString()}
                sub="this semester"
              />
              <StatCard
                label="Dues Status"
                value={unpaidDues.length === 0 ? 'Paid' : 'Outstanding'}
                sub={unpaidDues.length === 0 ? 'All clear' : `${unpaidDues.length} unpaid`}
              />
            </div>

            {/* Upcoming events */}
            <Card>
              <Card.Header
                title="Upcoming Events"
                action={<Link to="/events" style={{ fontSize: '.78rem', color: 'var(--navy)' }}>View all</Link>}
              />
              <Card.Body>
                {isLoading && (
                  <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>
                )}
                {!isLoading && upcomingEvents.length === 0 && (
                  <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>No upcoming events</p>
                )}
                {upcomingEvents.map((ev) => (
                  <Link
                    key={ev.id}
                    to={`/events`}
                    style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 0', borderBottom: '1px solid var(--border)', textDecoration: 'none', color: 'inherit' }}
                  >
                    <div style={{ background: 'var(--navy)', color: 'var(--gold)', borderRadius: '8px', padding: '6px 10px', fontSize: '.72rem', fontFamily: 'DM Mono, monospace', textAlign: 'center', minWidth: '52px' }}>
                      <div style={{ fontSize: '1rem', fontWeight: 700 }}>
                        {new Date(ev.event_date).getDate()}
                      </div>
                      <div style={{ textTransform: 'uppercase', letterSpacing: '.06em' }}>
                        {new Date(ev.event_date).toLocaleDateString('en', { month: 'short' })}
                      </div>
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontSize: '.85rem', fontWeight: 600, color: 'var(--ink)' }}>{ev.name}</div>
                      <div style={{ fontSize: '.72rem', color: 'var(--muted)' }}>
                        {ev.location ?? ev.event_type} • +{ev.xp_attend} XP
                      </div>
                    </div>
                  </Link>
                ))}
              </Card.Body>
            </Card>
          </div>

          {/* Right column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {/* Chapter stats */}
            <Card>
              <Card.Header title="Chapter Stats" />
              <Card.Body>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span style={{ fontSize: '.83rem', color: 'var(--muted)' }}>Active Members</span>
                    <span style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.2rem', color: 'var(--navy)' }}>
                      {totalMembers}
                    </span>
                  </div>
                </div>
              </Card.Body>
            </Card>

            {/* Leaderboard preview */}
            <Card>
              <Card.Header
                title="Top 5"
                action={<Link to="/leaderboard" style={{ fontSize: '.78rem', color: 'var(--navy)' }}>Full board</Link>}
              />
              <div>
                {leaderboard.map((e) => (
                  <div key={e.member_id} className="lb-row">
                    <span className="lb-rank">{e.rank}</span>
                    <div
                      className="avatar-circle"
                      style={{
                        width: 32, height: 32,
                        background: e.avatar_bg ?? '#001A4D',
                        color: e.avatar_fg ?? '#C9A84C',
                        fontSize: '.7rem',
                      }}
                    >
                      {e.first_name[0]}{e.last_name[0]}
                    </div>
                    <div className="lb-name" style={{ flex: 1 }}>
                      {e.first_name} {e.last_name}
                    </div>
                    <span className="lb-xp">{e.xp_total.toLocaleString()}</span>
                  </div>
                ))}
              </div>
            </Card>
          </div>
        </div>
      </main>
    </>
  )
}
