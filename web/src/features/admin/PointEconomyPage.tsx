import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getPointEconomy, updatePointEconomy, type PointEconomy } from '../../api/settings'

const FIELD_LABELS: Record<string, string> = {
  attend_event:    'Attend Event',
  rsvp_event:      'RSVP to Event',
  service_per_hour:'Service (per hour)',
  props:           'Give/Receive Props',
  dues_paid:       'Pay Dues On Time',
  badge_bonus:     'Earn a Badge',
}

export function PointEconomyPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [form, setForm] = useState<PointEconomy | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['point-economy'],
    queryFn: getPointEconomy,
    enabled: isAdmin,
  })

  useEffect(() => {
    if (data) setForm({ ...data })
  }, [data])

  const saveMutation = useMutation({
    mutationFn: () => updatePointEconomy(form!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['point-economy'] })
      showToast('Point economy saved.', 'success')
    },
    onError: () => showToast('Failed to save.', 'error'),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Point Economy" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--danger)' }}>Admin access required.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Point Economy" />
      <main className="page-body" style={{ maxWidth: 540 }}>
        <Card>
          <Card.Header title="XP Values" />
          <Card.Body>
            {isLoading && <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p>}
            {form && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                {(Object.keys(form) as Array<keyof PointEconomy>).map((key) => (
                  <div key={key} className="form-group">
                    <label className="form-label">{FIELD_LABELS[key] ?? key.replace(/_/g, ' ')}</label>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <input
                        className="form-input"
                        type="number"
                        value={form[key]}
                        onChange={(e) => setForm((f) => f ? { ...f, [key]: Number(e.target.value) } : f)}
                        style={{ maxWidth: 120 }}
                        min={0}
                      />
                      <span style={{ fontSize: '.75rem', color: 'var(--muted)' }}>XP</span>
                    </div>
                  </div>
                ))}
                <Button
                  variant="gold"
                  onClick={() => saveMutation.mutate()}
                  loading={saveMutation.isPending}
                >
                  Save Point Economy
                </Button>
              </div>
            )}
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
