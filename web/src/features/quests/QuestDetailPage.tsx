import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { getQuest, getQuestProgress, type Quest } from '../../api/quests'

export function QuestDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: quest, isLoading: questLoading } = useQuery({
    queryKey: ['quest', id],
    queryFn: () => getQuest(id!),
    enabled: !!id,
  })

  const { data: progress, isLoading: progressLoading } = useQuery({
    queryKey: ['quest-progress', id],
    queryFn: () => getQuestProgress(id!),
    enabled: !!id,
  })

  if (questLoading || progressLoading) {
    return (
      <>
        <Topbar title="Quest" />
        <main className="page-body">
          <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>
        </main>
      </>
    )
  }

  if (!quest) {
    return (
      <>
        <Topbar title="Quest" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                Quest not found.
              </p>
              <Button variant="outline" onClick={() => navigate('/quests')}>
                ← Back to Quests
              </Button>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  const steps = quest.steps ?? []
  const progressData = progress?.progress ?? {}

  return (
    <>
      <Topbar title={quest.title} />
      <main className="page-body" style={{ maxWidth: 720 }}>
        <div style={{ marginBottom: '1rem' }}>
          <Button size="sm" variant="outline" onClick={() => navigate('/quests')}>
            ← Back to Quests
          </Button>
        </div>

        <Card style={{ marginBottom: '1.5rem' }}>
          <Card.Header title={quest.title} />
          <Card.Body>
            {quest.description && (
              <p style={{ fontSize: '.88rem', color: 'var(--ink)', lineHeight: 1.6, marginBottom: '1rem' }}>
                {quest.description}
              </p>
            )}

            <div
              style={{
                display: 'flex',
                gap: '1rem',
                padding: '1rem',
                background: 'var(--info-bg)',
                borderRadius: '8px',
                marginBottom: '1rem',
                flexWrap: 'wrap',
              }}
            >
              <div style={{ flex: 1, minWidth: 150 }}>
                <div style={{ fontSize: '.7rem', color: 'var(--muted)', textTransform: 'uppercase', fontWeight: 600 }}>
                  XP Reward
                </div>
                <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--gold)' }}>
                  +{quest.xp_reward} XP
                </div>
              </div>
              <div style={{ flex: 1, minWidth: 150 }}>
                <div style={{ fontSize: '.7rem', color: 'var(--muted)', textTransform: 'uppercase', fontWeight: 600 }}>
                  Status
                </div>
                <div
                  style={{
                    fontSize: '.88rem',
                    fontWeight: 600,
                    color: quest.is_active ? 'var(--success)' : 'var(--muted)',
                    marginTop: '4px',
                  }}
                >
                  {quest.is_active ? '✅ Active' : '⏸ Inactive'}
                </div>
              </div>
            </div>

            <div style={{ marginTop: '1.5rem' }}>
              <h3
                style={{
                  fontFamily: 'DM Serif Display, serif',
                  fontSize: '1rem',
                  color: 'var(--navy)',
                  marginBottom: '1rem',
                }}
              >
                Quest Steps
              </h3>

              {steps.length === 0 ? (
                <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>No steps defined.</p>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {steps.map((step: any, idx: number) => {
                    const current = progressData[String(idx)] ?? 0
                    const pct = Math.min(100, Math.round((current / step.target_count) * 100))
                    const done = current >= step.target_count

                    return (
                      <div key={idx}>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            marginBottom: '8px',
                          }}
                        >
                          <div>
                            <span
                              style={{
                                fontSize: '.9rem',
                                fontWeight: done ? 600 : 400,
                                color: done ? 'var(--success)' : 'var(--ink)',
                              }}
                            >
                              {done ? '✅ ' : '○ '}
                              {step.title}
                            </span>
                          </div>
                          <span
                            style={{
                              fontSize: '.75rem',
                              color: 'var(--muted)',
                              fontFamily: 'DM Mono, monospace',
                              fontWeight: 600,
                            }}
                          >
                            {current}/{step.target_count}
                          </span>
                        </div>

                        <div
                          style={{
                            height: 8,
                            background: 'var(--cream2)',
                            borderRadius: 99,
                            overflow: 'hidden',
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

                        {step.description && (
                          <p
                            style={{
                              fontSize: '.75rem',
                              color: 'var(--muted)',
                              marginTop: '6px',
                              fontStyle: 'italic',
                            }}
                          >
                            {step.description}
                          </p>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
