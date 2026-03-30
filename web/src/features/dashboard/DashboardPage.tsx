import { Link } from 'react-router-dom'
import { QRCodeSVG } from 'qrcode.react'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { StreakWidget } from '../../components/StreakWidget'
import { useDashboard } from './useDashboard'
import { useAuth } from '../../hooks/useAuth'
import type { DuesRecord } from '../../types'
import type { Goal } from '../../api/goals'

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

const ACTION_TILES = [
  { icon: '🎖️', label: 'Badge Room',         sub: 'EARN REWARDS',    path: '/quests' },
  { icon: '📜', label: 'Quest Log',           sub: 'TRACK PROGRESS',  path: '/quests' },
  { icon: '🛍️', label: 'Paraphernalia Floor', sub: 'BROWSE STORE',    path: '/store'  },
]

// ── Dues bar chart ──────────────────────────────────────────────────
function DuesChart({ dues }: { dues: DuesRecord[] }) {
  const items = [...dues].reverse()
  const maxAmt = Math.max(...items.map((d) => d.amount_cents), 1)

  function shortLabel(semester: string) {
    const parts = semester.split(' ')
    if (parts.length >= 2) {
      const map: Record<string, string> = { Spring: 'SPR', Fall: 'FAL', Summer: 'SUM' }
      return [map[parts[0]] ?? parts[0].slice(0, 3), `'${parts[1].slice(2)}`]
    }
    return [semester.slice(0, 4)]
  }

  if (items.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: '1rem 0', color: 'var(--g-muted)', fontSize: '.75rem' }}>
        No dues records
      </div>
    )
  }

  return (
    <>
      <div className="dues-chart-wrap">
        {items.map((d) => {
          const pct = Math.max(8, Math.round((d.amount_cents / maxAmt) * 100))
          const paid = d.status === 'paid' || d.status === 'waived'
          return (
            <div
              key={d.id}
              className={`dues-bar ${paid ? 'paid' : 'unpaid'}`}
              style={{ height: `${pct}%` }}
              title={`${d.semester}: $${(d.amount_cents / 100).toFixed(0)} — ${d.status}`}
            />
          )
        })}
      </div>
      <div className="dues-bar-labels">
        {items.map((d) => (
          <div key={d.id} className="dues-bar-label">
            {shortLabel(d.semester).map((t, i) => <div key={i}>{t}</div>)}
          </div>
        ))}
      </div>
    </>
  )
}

// ── Achievement row ─────────────────────────────────────────────────
function AchievementRow({ goal }: { goal: Goal }) {
  const pct =
    goal.target_value > 0
      ? Math.min(100, Math.round((goal.current_value / goal.target_value) * 100))
      : 0
  const done = pct >= 100
  return (
    <div className="achievement-row">
      <div className={`achievement-check ${done ? 'done' : 'pending'}`}>
        {done ? '✓' : '○'}
      </div>
      <div className="achievement-bar-wrap">
        <div className="achievement-title">{goal.title}</div>
        <div className="achievement-bar">
          <div className={`achievement-bar-fill ${done ? 'done' : ''}`} style={{ width: `${pct}%` }} />
        </div>
      </div>
      <div style={{ fontSize: '.7rem', color: 'var(--g-muted)', minWidth: 34, textAlign: 'right' }}>
        {pct}%
      </div>
    </div>
  )
}

// ── Dashboard ────────────────────────────────────────────────────────
export function DashboardPage() {
  const { user } = useAuth()
  const { member, totalMembers, upcomingEvents, leaderboard, myRank, unpaidDues, myDues, chapterGoals, isLoading } =
    useDashboard()

  const xpData       = member ? getXPProgress(member.xp_total) : null
  const yearsOfService = member?.inducted_year ? new Date().getFullYear() - member.inducted_year : null
  const paidCount    = myDues.filter((d) => d.status === 'paid' || d.status === 'waived').length
  const duesPct      = myDues.length > 0 ? Math.round((paidCount / myDues.length) * 100) : 100
  const qrValue      = member
    ? `BL:${member.member_display_id}:${member.first_name} ${member.last_name}`
    : 'BL:ID'

  const initials = user ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase() : '??'

  return (
    <>
      <Topbar title="Chapter HQ" />
      <main className="page-body" style={{ maxWidth: 1100, margin: '0 auto' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>

          {/* ── PLAYER CARD ── */}
          <div className="player-card fade-in">
            <div className="player-avatar-ring">
              {member?.avatar_url
                ? <img src={member.avatar_url} alt="avatar" />
                : <span>{initials}</span>
              }
            </div>

            <div className="player-stats">
              <div className="player-name">
                {user ? `${user.first_name} ${user.last_name}` : '—'}
              </div>
              {member && (
                <>
                  <div className="player-stat-row">
                    <span>🪪</span>
                    <span>Member #:</span>
                    <span className="val">{member.member_display_id}</span>
                  </div>
                  {yearsOfService !== null && (
                    <div className="player-stat-row">
                      <span>📅</span>
                      <span>Years of Service:</span>
                      <span className="val">{yearsOfService}</span>
                    </div>
                  )}
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <div className="player-xp">
                      <span style={{ color: 'var(--gold)', fontSize: '1rem' }}>✦</span>
                      <span className="player-xp-value">{member.xp_total.toLocaleString()}</span>
                      <span className="player-xp-label">Experience Points</span>
                    </div>
                    <span className="player-level">{member.level}</span>
                  </div>
                  {xpData?.next && (
                    <div style={{ marginTop: '.5rem', maxWidth: 300 }}>
                      <div style={{ fontSize: '.64rem', color: 'var(--g-muted)', marginBottom: 4 }}>
                        {xpData.xpToNext.toLocaleString()} XP → {xpData.next.label}
                      </div>
                      <div className="progress-bar" style={{ height: 5 }}>
                        <div
                          className="progress-fill"
                          style={{ width: `${xpData.pct}%`, background: 'var(--g-accent)' }}
                        />
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>

            {/* Right side: rank + QR */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', alignItems: 'center', flexShrink: 0 }}>
              {myRank > 0 && (
                <div className="rank-badge">
                  <div className="rank-number">#{myRank}</div>
                  <div className="rank-label">Chapter Rank</div>
                </div>
              )}
              <Link to="/digital-id" className="qr-teaser">
                <QRCodeSVG value={qrValue} size={60} bgColor="transparent" fgColor="#60A5FA" />
                <div className="qr-teaser-sub">Scan to Share Profile</div>
              </Link>
            </div>
          </div>

          {/* ── ACTION TILES ── */}
          <div className="action-tiles fade-in">
            {ACTION_TILES.map((tile) => (
              <Link key={tile.label} to={tile.path} className="action-tile">
                <div className="action-tile-icon">{tile.icon}</div>
                <div className="action-tile-label">{tile.label}</div>
                <div className="action-tile-sub">{tile.sub}</div>
              </Link>
            ))}
          </div>

          {/* ── STREAK WIDGET ── */}
          {member && (
            <StreakWidget memberId={member.id} className="fade-in" showAnimation={true} />
          )}

          {/* ── MIDDLE ROW: Achievements + Financial ── */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>

            {/* Chapter Achievements */}
            <Card>
              <Card.Header
                title="Chapter Achievements"
                action={
                  <Link to="/goals" style={{ fontSize: '.75rem', color: 'var(--g-accent)', textDecoration: 'none' }}>
                    View all
                  </Link>
                }
              />
              <Card.Body>
                {isLoading ? (
                  <p style={{ color: 'var(--g-muted)', fontSize: '.83rem' }}>Loading…</p>
                ) : chapterGoals.length === 0 ? (
                  <p style={{ color: 'var(--g-muted)', fontSize: '.83rem' }}>No active chapter goals</p>
                ) : (
                  <div className="achievement-list">
                    {chapterGoals.slice(0, 5).map((goal) => (
                      <AchievementRow key={goal.id} goal={goal} />
                    ))}
                  </div>
                )}
              </Card.Body>
            </Card>

            {/* Financial Status */}
            <Card>
              <Card.Header title="Financial Status" />
              <Card.Body>
                <div className="fin-budget-label">Budget Status</div>
                <div className="fin-budget-bar">
                  <div className="fin-budget-fill" style={{ width: `${duesPct}%` }} />
                </div>
                <div className="fin-budget-label">
                  Dues Collection — {paidCount}/{myDues.length} paid
                </div>
                <DuesChart dues={myDues} />
                {unpaidDues.length > 0 && (
                  <div style={{ marginTop: '.85rem' }}>
                    <Link
                      to="/dues"
                      style={{
                        display: 'block', textAlign: 'center',
                        background: 'rgba(139,26,26,0.22)',
                        border: '1px solid rgba(200,50,50,0.35)',
                        borderRadius: 8, padding: '6px 12px',
                        fontSize: '.75rem', color: '#FF7575', textDecoration: 'none',
                      }}
                    >
                      ⚠️ {unpaidDues.length} outstanding {unpaidDues.length === 1 ? 'balance' : 'balances'}
                    </Link>
                  </div>
                )}
              </Card.Body>
            </Card>
          </div>

          {/* ── BOTTOM ROW: Events + Leaderboard ── */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem' }}>

            {/* Upcoming Events */}
            <Card>
              <Card.Header
                title="Recent Chapter Events"
                action={
                  <Link to="/events" style={{ fontSize: '.75rem', color: 'var(--g-accent)', textDecoration: 'none' }}>
                    View all
                  </Link>
                }
              />
              <Card.Body>
                {isLoading ? (
                  <p style={{ color: 'var(--g-muted)', fontSize: '.83rem' }}>Loading…</p>
                ) : upcomingEvents.length === 0 ? (
                  <p style={{ color: 'var(--g-muted)', fontSize: '.83rem' }}>No upcoming events</p>
                ) : (
                  upcomingEvents.map((ev) => {
                    const d = new Date(ev.event_date)
                    return (
                      <Link key={ev.id} to="/events" className="game-event-row">
                        <div className="game-event-date">
                          <div className="game-event-day">{d.getDate()}</div>
                          <div className="game-event-mon">
                            {d.toLocaleDateString('en', { month: 'short' })}
                          </div>
                        </div>
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div className="game-event-name">{ev.name}</div>
                          <div className="game-event-meta">{ev.location ?? ev.event_type}</div>
                        </div>
                        <span className="xp-tag">+{ev.xp_attend} XP</span>
                      </Link>
                    )
                  })
                )}
              </Card.Body>
            </Card>

            {/* Leaderboard */}
            <Card>
              <Card.Header
                title="Member Leaderboard"
                action={
                  <Link to="/leaderboard" style={{ fontSize: '.75rem', color: 'var(--g-accent)', textDecoration: 'none' }}>
                    Full board
                  </Link>
                }
              />
              <div>
                {leaderboard.map((e) => (
                  <div key={e.member_id} className="lb-row">
                    <span className="lb-rank">{e.rank}</span>
                    <div
                      className="avatar-circle"
                      style={{
                        width: 30, height: 30,
                        background: e.avatar_bg ?? 'var(--navy)',
                        color: e.avatar_fg ?? 'var(--gold)',
                        fontSize: '.68rem',
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

              {/* Bottom stat row */}
              <div
                style={{
                  display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px',
                  padding: '1rem', borderTop: '1px solid var(--border)',
                }}
              >
                <div className="stat-card">
                  <div className="stat-label">Active Members</div>
                  <div className="stat-value">{totalMembers}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Semester XP</div>
                  <div className="stat-value">{(member?.xp_semester ?? 0).toLocaleString()}</div>
                </div>
              </div>
            </Card>

          </div>
        </div>
      </main>
    </>
  )
}
