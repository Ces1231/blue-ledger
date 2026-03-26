import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { getLeaderboard } from '../../api/xp'
import { useAuth } from '../../hooks/useAuth'
import type { LeaderboardEntry } from '../../types'

type Mode = 'alltime' | 'semester'

function PodiumItem({
  entry,
  rank,
  height,
  crownColor,
}: {
  entry: LeaderboardEntry
  rank: number
  height: number
  crownColor: string
}) {
  const RANK_EMOJI = ['', '🥇', '🥈', '🥉']
  const avatarSize = rank === 1 ? 60 : 48

  return (
    <div className="lb-podium-item">
      <div className="lb-podium-rank">{RANK_EMOJI[rank] ?? rank}</div>
      <div
        className="lb-podium-avatar"
        style={{
          width: avatarSize,
          height: avatarSize,
          background: entry.avatar_bg ?? '#001A4D',
          color: entry.avatar_fg ?? '#C9A84C',
          fontSize: rank === 1 ? '.9rem' : '.75rem',
          fontWeight: 700,
          borderColor: crownColor,
        }}
      >
        {entry.first_name[0]}{entry.last_name[0]}
      </div>
      <div className="lb-podium-name">
        {entry.first_name} {entry.last_name}
      </div>
      <div className="lb-podium-xp">{entry.xp_total.toLocaleString()} XP</div>
      <div
        className="lb-podium-plinth"
        style={{
          background: crownColor,
          height,
          opacity: 0.8,
        }}
      >
        <span style={{ fontFamily: 'DM Mono, monospace', fontWeight: 700, color: '#fff', fontSize: '.78rem' }}>
          #{rank}
        </span>
      </div>
    </div>
  )
}

export function LeaderboardPage() {
  const { isAuthenticated, memberID } = useAuth()
  const [mode, setMode] = useState<Mode>('alltime')

  const { data: leaderboard = [], isLoading } = useQuery({
    queryKey: ['leaderboard', mode],
    queryFn: () => getLeaderboard(mode),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 2,
  })

  const myEntry = leaderboard.find((e) => e.member_id === memberID)
  const myRank = myEntry ? leaderboard.indexOf(myEntry) + 1 : null

  const [first, second, third] = leaderboard
  const rest = leaderboard.slice(3)

  return (
    <>
      <Topbar title="Leaderboard" />
      <main className="page-body">
        {/* Mode toggle */}
        <div style={{ display: 'flex', gap: '8px', marginBottom: '1.5rem' }}>
          <button
            className={`btn btn-sm ${mode === 'alltime' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setMode('alltime')}
          >
            All Time
          </button>
          <button
            className={`btn btn-sm ${mode === 'semester' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setMode('semester')}
          >
            This Semester
          </button>
        </div>

        {/* My rank callout */}
        {myEntry && myRank && (
          <div
            className="fade-in"
            style={{
              background: 'var(--navy)',
              borderRadius: 'var(--radius-lg)',
              padding: '1rem 1.25rem',
              marginBottom: '1.5rem',
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
            }}
          >
            <div
              className="avatar-circle"
              style={{
                width: 44,
                height: 44,
                background: myEntry.avatar_bg ?? '#C9A84C',
                color: myEntry.avatar_fg ?? '#001A4D',
                fontSize: '.78rem',
                fontWeight: 700,
                flexShrink: 0,
              }}
            >
              {myEntry.first_name[0]}{myEntry.last_name[0]}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ color: 'rgba(255,255,255,.5)', fontSize: '.62rem', fontFamily: 'DM Mono, monospace', textTransform: 'uppercase', letterSpacing: '.1em', marginBottom: '2px' }}>
                Your Standing
              </div>
              <div style={{ color: '#fff', fontSize: '.88rem', fontWeight: 600 }}>
                Rank #{myRank} of {leaderboard.length}
              </div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.6rem', color: 'var(--gold)', lineHeight: 1 }}>
                {myEntry.xp_total.toLocaleString()}
              </div>
              <div style={{ fontSize: '.62rem', color: 'rgba(255,255,255,.35)', fontFamily: 'DM Mono, monospace', textTransform: 'uppercase', letterSpacing: '.08em' }}>
                Total XP
              </div>
            </div>
          </div>
        )}

        {isLoading && (
          <div style={{ textAlign: 'center', paddingTop: '3rem', color: 'var(--muted)' }}>Loading...</div>
        )}

        {!isLoading && leaderboard.length === 0 && (
          <div className="empty-state">
            <div className="empty-state-icon">🏆</div>
            <div className="empty-state-text">No leaderboard data yet</div>
          </div>
        )}

        {/* Podium (top 3) */}
        {!isLoading && leaderboard.length >= 3 && (
          <div className="lb-podium fade-in">
            {second && (
              <PodiumItem entry={second} rank={2} height={80} crownColor="#9E9E9E" />
            )}
            {first && (
              <PodiumItem entry={first} rank={1} height={110} crownColor="#C9A84C" />
            )}
            {third && (
              <PodiumItem entry={third} rank={3} height={60} crownColor="#CD7F32" />
            )}
          </div>
        )}

        {/* Full table */}
        {!isLoading && leaderboard.length > 0 && (
          <Card>
            <Card.Header title={`Full Rankings — ${mode === 'alltime' ? 'All Time' : 'This Semester'}`} />
            <div>
              {leaderboard.map((entry, idx) => {
                const rank = idx + 1
                const isMe = entry.member_id === memberID
                return (
                  <Link
                    key={entry.member_id}
                    to={`/members/${entry.member_id}`}
                    style={{ textDecoration: 'none', color: 'inherit' }}
                  >
                    <div
                      className="lb-row"
                      style={isMe ? { background: 'rgba(0,26,77,.04)' } : undefined}
                    >
                      <span className="lb-rank" style={rank <= 3 ? { color: ['#C9A84C', '#9E9E9E', '#CD7F32'][rank - 1] } : undefined}>
                        {rank}
                      </span>
                      <div
                        className="avatar-circle"
                        style={{
                          width: 32,
                          height: 32,
                          background: entry.avatar_bg ?? '#001A4D',
                          color: entry.avatar_fg ?? '#C9A84C',
                          fontSize: '.7rem',
                        }}
                      >
                        {entry.first_name[0]}{entry.last_name[0]}
                      </div>
                      <div className="lb-name" style={isMe ? { fontWeight: 700 } : undefined}>
                        {entry.first_name} {entry.last_name}
                        {isMe && (
                          <span style={{ marginLeft: '6px', fontSize: '.65rem', color: 'var(--navy)', fontFamily: 'DM Mono, monospace', fontWeight: 400 }}>
                            (you)
                          </span>
                        )}
                      </div>
                      <span className={`lb-level badge badge-${entry.level_key}`}>{entry.level}</span>
                      <span className="lb-xp">{entry.xp_total.toLocaleString()}</span>
                    </div>
                  </Link>
                )
              })}
            </div>
          </Card>
        )}

        {/* Rest of board after podium (shown alongside card above — card is the authoritative list) */}
        {rest.length > 0 && false /* rendered in the card above */ && (
          <div style={{ marginTop: '1rem' }}>
            {rest.map((entry, i) => (
              <div key={entry.member_id} className="lb-row">
                <span className="lb-rank">{i + 4}</span>
                <div
                  className="avatar-circle"
                  style={{
                    width: 32,
                    height: 32,
                    background: entry.avatar_bg ?? '#001A4D',
                    color: entry.avatar_fg ?? '#C9A84C',
                    fontSize: '.7rem',
                  }}
                >
                  {entry.first_name[0]}{entry.last_name[0]}
                </div>
                <div className="lb-name">{entry.first_name} {entry.last_name}</div>
                <span className={`lb-level badge badge-${entry.level_key}`}>{entry.level}</span>
                <span className="lb-xp">{entry.xp_total.toLocaleString()}</span>
              </div>
            ))}
          </div>
        )}
      </main>
    </>
  )
}
