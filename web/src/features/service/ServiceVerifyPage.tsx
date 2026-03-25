import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { getServiceLog, verifyService } from '../../api/service'

export function ServiceVerifyPage() {
  const { isChair } = useAuth()
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['service', 'verify'],
    queryFn: () => getServiceLog({ per_page: 100 }),
    enabled: isChair,
  })

  const verifyMutation = useMutation({
    mutationFn: verifyService,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['service'] }),
  })

  if (!isChair) {
    return (
      <>
        <Topbar title="Verify Service" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Chair or admin access required.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  const pending = (data?.data ?? []).filter((e) => !e.verified)

  return (
    <>
      <Topbar title="Verify Service Hours" />
      <main className="page-body">
        <Card>
          <Card.Header title={`Pending Verifications (${pending.length})`} />
          <Card.Body>
            {isLoading && <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>}
            {!isLoading && pending.length === 0 && (
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '1.5rem 0' }}>
                No pending verifications. All caught up!
              </p>
            )}
            {pending.map((entry) => (
              <div
                key={entry.id}
                style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '12px 0', borderBottom: '1px solid var(--border)',
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: '.9rem' }}>{entry.event_name}</div>
                  <div style={{ fontSize: '.75rem', color: 'var(--muted)' }}>
                    {entry.member_first_name} {entry.member_last_name} •{' '}
                    {entry.hours}h on {new Date(entry.service_date).toLocaleDateString()}
                    {entry.organization && ` • ${entry.organization}`}
                  </div>
                  {entry.notes && (
                    <div style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 2, fontStyle: 'italic' }}>
                      "{entry.notes}"
                    </div>
                  )}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
                  <span style={{ fontFamily: 'DM Mono, monospace', fontSize: '.9rem', fontWeight: 700 }}>
                    {entry.hours}h
                  </span>
                  <Button
                    size="sm"
                    variant="gold"
                    onClick={() => verifyMutation.mutate(entry.id)}
                    loading={verifyMutation.isPending && verifyMutation.variables === entry.id}
                  >
                    Verify
                  </Button>
                </div>
              </div>
            ))}
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
