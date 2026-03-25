import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { getDues, markDuesPaid } from '../../api/dues'

export function AdminDuesPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['dues', 'all'],
    queryFn: () => getDues({ per_page: 200 }),
    enabled: isAdmin,
  })

  const markPaidMutation = useMutation({
    mutationFn: (id: string) => markDuesPaid(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dues'] }),
  })

  if (!isAdmin) return null

  const dues = (data?.data ?? []).filter((d) => {
    const name = `${d.semester}`.toLowerCase()
    return name.includes(search.toLowerCase())
  })

  const unpaidCount = dues.filter((d) => d.status !== 'paid' && d.status !== 'waived').length

  return (
    <>
      <Topbar title="Admin — Dues Management" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <input
            className="form-input"
            style={{ maxWidth: 280 }}
            placeholder="Search by semester..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <span style={{ fontSize: '.83rem', color: 'var(--muted)' }}>
            {unpaidCount} outstanding
          </span>
        </div>

        <Card>
          <Card.Header title="All Dues Records" />
          <Card.Body>
            {isLoading && <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>}
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '.83rem' }}>
                <thead>
                  <tr style={{ borderBottom: '2px solid var(--border)' }}>
                    <th style={{ padding: '8px 4px', textAlign: 'left', color: 'var(--muted)', fontWeight: 600 }}>Member</th>
                    <th style={{ padding: '8px 4px', textAlign: 'left', color: 'var(--muted)', fontWeight: 600 }}>Semester</th>
                    <th style={{ padding: '8px 4px', textAlign: 'right', color: 'var(--muted)', fontWeight: 600 }}>Amount</th>
                    <th style={{ padding: '8px 4px', textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>Status</th>
                    <th style={{ padding: '8px 4px', textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {dues.map((d) => (
                    <tr key={d.id} style={{ borderBottom: '1px solid var(--border)' }}>
                      <td style={{ padding: '8px 4px' }}>
                        <span style={{ fontFamily: 'DM Mono, monospace', fontSize: '.75rem', color: 'var(--muted)' }}>
                          {d.member_id.slice(0, 8)}
                        </span>
                      </td>
                      <td style={{ padding: '8px 4px', fontWeight: 500 }}>{d.semester}</td>
                      <td style={{ padding: '8px 4px', textAlign: 'right', fontFamily: 'DM Mono, monospace' }}>
                        ${(d.amount_cents / 100).toFixed(2)}
                      </td>
                      <td style={{ padding: '8px 4px', textAlign: 'center' }}>
                        <span style={{
                          background: d.status === 'paid' ? 'var(--success-bg)' : 'var(--danger-bg)',
                          color: d.status === 'paid' ? 'var(--success)' : 'var(--danger)',
                          borderRadius: 99, padding: '2px 8px', fontSize: '.7rem', fontWeight: 600,
                        }}>
                          {d.status}
                        </span>
                      </td>
                      <td style={{ padding: '8px 4px', textAlign: 'center' }}>
                        {d.status !== 'paid' && d.status !== 'waived' && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => markPaidMutation.mutate(d.id)}
                            loading={markPaidMutation.isPending && markPaidMutation.variables === d.id}
                          >
                            Mark Paid
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
