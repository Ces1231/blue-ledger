import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card, StatCard } from '../../components/Card'
import { Button } from '../../components/Button'
import { getDues } from '../../api/dues'
import { getSettings } from '../../api/settings'

function statusBadge(status: string) {
  const map: Record<string, { bg: string; fg: string }> = {
    paid:        { bg: 'var(--success-bg)', fg: 'var(--success)' },
    unpaid:      { bg: 'var(--danger-bg)',  fg: 'var(--danger)' },
    late:        { bg: '#FFF3DC',           fg: '#7A4A00' },
    outstanding: { bg: 'var(--danger-bg)',  fg: 'var(--danger)' },
    waived:      { bg: 'var(--info-bg)',    fg: 'var(--info)' },
  }
  const s = map[status] ?? map.unpaid
  return (
    <span style={{
      background: s.bg, color: s.fg,
      borderRadius: 99, padding: '2px 10px', fontSize: '.72rem', fontWeight: 600,
    }}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  )
}

export function DuesPage() {
  const { data: duesData, isLoading } = useQuery({
    queryKey: ['dues', 'my'],
    queryFn: () => getDues(),
  })

  const { data: settingsResp } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })

  const dues = duesData?.data ?? []
  const zeffyFormId = settingsResp?.zeffy_form_id

  const unpaid = dues.filter((d) => d.status !== 'paid' && d.status !== 'waived')
  const totalOwed = unpaid.reduce((sum, d) => sum + d.amount_cents, 0)

  return (
    <>
      <Topbar title="Dues" />
      <main className="page-body">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 10, marginBottom: '1.25rem' }}>
          <StatCard
            label="Status"
            value={unpaid.length === 0 ? 'Paid' : 'Outstanding'}
            sub={unpaid.length === 0 ? 'All clear' : `${unpaid.length} unpaid`}
          />
          <StatCard
            label="Amount Owed"
            value={unpaid.length === 0 ? '$0' : `$${(totalOwed / 100).toFixed(2)}`}
            sub="current balance"
          />
          <StatCard
            label="Total Records"
            value={dues.length}
            sub="all semesters"
          />
        </div>

        {unpaid.length > 0 && zeffyFormId && (
          <Card style={{ marginBottom: '1.25rem', background: 'var(--warn-bg)', border: '1px solid var(--warn-border)' }}>
            <Card.Body>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ fontWeight: 700, color: 'var(--warn)' }}>Dues Payment Due</div>
                  <div style={{ fontSize: '.83rem', color: 'var(--muted)', marginTop: 2 }}>
                    Pay securely via Zeffy — 0% platform fees
                  </div>
                </div>
                <Button
                  variant="gold"
                  onClick={() => {
                    window.open(
                      `https://www.zeffy.com/embed/donation-form/${zeffyFormId}`,
                      'zeffy-payment',
                      'width=640,height=700,scrollbars=yes'
                    )
                  }}
                >
                  Pay Dues via Zeffy
                </Button>
              </div>
            </Card.Body>
          </Card>
        )}

        <Card>
          <Card.Header title="Dues History" />
          <Card.Body>
            {isLoading && <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>}
            {!isLoading && dues.length === 0 && (
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '1rem 0' }}>
                No dues records found.
              </p>
            )}
            {dues.map((d) => (
              <div
                key={d.id}
                style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '10px 0', borderBottom: '1px solid var(--border)',
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: '.88rem' }}>{d.semester}</div>
                  <div style={{ fontSize: '.73rem', color: 'var(--muted)' }}>
                    Due {new Date(d.due_date).toLocaleDateString()}
                    {d.paid_at && ` • Paid ${new Date(d.paid_at).toLocaleDateString()}`}
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <span style={{ fontFamily: 'DM Mono, monospace', fontSize: '.88rem', fontWeight: 600 }}>
                    ${(d.amount_cents / 100).toFixed(2)}
                  </span>
                  {statusBadge(d.status)}
                </div>
              </div>
            ))}
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
