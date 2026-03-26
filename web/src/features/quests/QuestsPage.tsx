import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import apiClient from '../../api/client'
import type { Quest, MemberBadge, Badge } from '../../types'
import type { APIResponse } from '../../types'

async function getQuests(): Promise<Quest[]> {
  const { data } = await apiClient.get<APIResponse<Quest[]>>('/quests')
  return data.data
}

async function getMyBadges(): Promise<MemberBadge[]> {
  const { data } = await apiClient.get<APIResponse<MemberBadge[]>>('/badges/mine')
  return data.data
}

async function getAllBadges(): Promise<Badge[]> {
  const { data } = await apiClient.get<APIResponse<Badge[]>>('/badges')
  return data.data
}

const RARITY_COLORS: Record<string, { bg: string; fg: string; border: string }> = {
  common:    { bg: '#F2EDE4', fg: '#6B6657', border: '#CDC7BD' },
  uncommon:  { bg: '#EAF5EE', fg: '#1A6B3A', border: '#A8D5BC' },
  rare:      { bg: '#EAF0FB', fg: '#003087', border: '#BDD0F5' },
  legendary: { bg: '#F5E6C8', fg: '#7A4A00', border: '#C9A84C' },
}

function BadgeCard({ badge, earned }: { badge: Badge; earned: boolean }) {
  const colors = RARITY_COLORS[badge.rarity] ?? RARITY_COLORS.common
  return (
    <div
      className="card fade-in"
      style={{
        padding: '1rem',
        border: earned ? `2px solid var(--gold)` : `1.5px solid var(--border)`,
        opacity: earned ? 1 : 0.65,
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        textAlign: 'center',
        gap: '0.5rem',
      }}
    >
      {earned && (
        <span
          style={{
            position: 'absolute', top: 8, right: 8,
            background: 'var(--gold)', color: '#fff',
            fontSize: '.6rem', fontWeight: 700, borderRadius: 99, padding: '2px 6px',
          }}
        >
          EARNED
        </span>
      )}
      <div style={{ fontSize: '2rem' }}>{badge.icon ?? '🎖️'}</div>
      <div style={{ fontWeight: 700, fontSize: '.85rem', color: 'var(--ink)' }}>{badge.name}</div>
      <span
        style={{
          background: colors.bg, color: colors.fg,
          border: `1px solid ${colors.border}`,
          borderRadius: 99, padding: '1px 8px', fontSize: '.65rem', fontWeight: 600,
          textTransform: 'capitalize',
        }}
      >
        {badge.rarity}
      </span>
      {badge.description && (
        <p style={{ fontSize: '.75rem', color: 'var(--muted)', lineHeight: 1.4 }}>
          {badge.description}
        </p>
      )}
      <div
        style={{
          background: 'var(--gold-bg)', color: 'var(--warn)',
          borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
        }}
      >
        +{badge.xp_reward} XP
      </div>
      {badge.requirement && (
        <p style={{ fontSize: '.72rem', color: 'var(--faint)', fontStyle: 'italic' }}>
          {badge.requirement}
        </p>
      )}
    </div>
  )
}

function QuestCard({ quest }: { quest: Quest }) {
  const steps = quest.steps ?? []
  const progress = quest.progress ?? {}

  return (
    <div className="card fade-in" style={{ padding: '1.25rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8, marginBottom: '0.75rem' }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: '.92rem', color: 'var(--ink)' }}>{quest.title}</div>
          {quest.description && (
            <div style={{ fontSize: '.78rem', color: 'var(--muted)', marginTop: 2 }}>
              {quest.description}
            </div>
          )}
        </div>
        <span
          style={{
            background: 'var(--gold-bg)', color: 'var(--warn)',
            borderRadius: 99, padding: '2px 10px', fontSize: '.7rem', fontWeight: 700, whiteSpace: 'nowrap',
          }}
        >
          +{quest.xp_reward} XP
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        {steps.map((step, idx) => {
          const current = progress[String(idx)] ?? 0
          const pct = Math.min(100, Math.round((current / step.target_count) * 100))
          const done = current >= step.target_count

          return (
            <div key={idx}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.78rem', marginBottom: 4 }}>
                <span style={{ color: done ? 'var(--success)' : 'var(--ink2)', fontWeight: done ? 600 : 400 }}>
                  {done ? '✓ ' : ''}{step.title}
                </span>
                <span style={{ color: 'var(--muted)', fontFamily: 'DM Mono, monospace', fontSize: '.72rem' }}>
                  {current}/{step.target_count}
                </span>
              </div>
              <div
                style={{
                  height: 6, background: 'var(--cream2)', borderRadius: 99, overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    height: '100%',
                    width: `${pct}%`,
                    background: done ? 'var(--success)' : 'var(--navy)',
                    borderRadius: 99,
                    transition: 'width 0.4s ease',
                  }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export function QuestsPage() {
  const { data: quests = [], isLoading: questsLoading } = useQuery({
    queryKey: ['quests'],
    queryFn: getQuests,
  })

  const { data: myBadges = [] } = useQuery({
    queryKey: ['badges-mine'],
    queryFn: getMyBadges,
  })

  const { data: allBadges = [], isLoading: badgesLoading } = useQuery({
    queryKey: ['badges-all'],
    queryFn: getAllBadges,
  })

  const earnedIds = new Set(myBadges.map((mb) => mb.badge_id))
  const activeQuests = quests.filter((q) => q.is_active)

  return (
    <>
      <Topbar title="Quests & Badges" />
      <main className="page-body">
        {/* Badges section */}
        <Card style={{ marginBottom: '1.5rem' }}>
          <Card.Header title="Badge Collection" />
          <Card.Body>
            {badgesLoading && (
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading badges...</p>
            )}
            {!badgesLoading && allBadges.length === 0 && (
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>No badges configured for this chapter yet.</p>
            )}
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
                gap: '0.75rem',
              }}
            >
              {allBadges.map((badge) => (
                <BadgeCard key={badge.id} badge={badge} earned={earnedIds.has(badge.id)} />
              ))}
            </div>
          </Card.Body>
        </Card>

        {/* Active quests */}
        <div style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: 8 }}>
          <h2 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.15rem', color: 'var(--navy)' }}>
            Active Quests
          </h2>
          <span
            style={{
              background: 'var(--info-bg)', color: 'var(--info)',
              borderRadius: 99, padding: '2px 8px', fontSize: '.7rem', fontWeight: 600,
            }}
          >
            {activeQuests.length}
          </span>
        </div>

        {questsLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading quests...</p>
            </Card.Body>
          </Card>
        )}

        {!questsLoading && activeQuests.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No active quests right now. Check back soon!
              </p>
            </Card.Body>
          </Card>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {activeQuests.map((quest) => (
            <QuestCard key={quest.id} quest={quest} />
          ))}
        </div>
      </main>
    </>
  )
}
