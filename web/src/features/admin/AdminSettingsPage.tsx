import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getSettings, updateSettings, type UpdateSettingsPayload } from '../../api/settings'

export function AdminSettingsPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })

  const [form, setForm] = useState<UpdateSettingsPayload & {
    zeffy_form_id?: string
    stripe_payment_link?: string
    make_webhook_url?: string
    semester_label?: string
    dues_amount?: number
    dues_deadline?: string
    grace_period_days?: number
    quiz_retake_limit?: number
    quiz_pass_score?: number
    district_email?: string
  }>({})

  useEffect(() => {
    if (settings) {
      setForm({
        semester_label: settings.semester_label ?? '',
        zeffy_form_id: settings.zeffy_form_id ?? '',
        allow_public_directory: settings.allow_public_directory ?? true,
      })
    }
  }, [settings])

  const saveMutation = useMutation({
    mutationFn: () =>
      updateSettings({
        allow_public_directory: form.allow_public_directory,
        semester_label: form.semester_label,
        notification_prefs: {},
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      showToast('Settings saved successfully.', 'success')
    },
    onError: () => showToast('Failed to save settings.', 'error'),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Chapter Settings" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Admin access required.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  if (isLoading) {
    return (
      <>
        <Topbar title="Chapter Settings" />
        <main className="page-body">
          <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading settings...</p>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Chapter Settings" />
      <main className="page-body" style={{ maxWidth: 640 }}>
        {/* Chapter configuration */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Chapter Configuration" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div className="form-group">
                <label className="form-label">Semester Label</label>
                <input
                  className="form-input"
                  value={form.semester_label ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, semester_label: e.target.value }))}
                  placeholder="e.g. Spring 2026"
                />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Dues Amount ($)</label>
                  <input
                    className="form-input"
                    type="number"
                    value={form.dues_amount ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, dues_amount: Number(e.target.value) }))}
                    placeholder="125.00"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Dues Deadline</label>
                  <input
                    className="form-input"
                    type="date"
                    value={form.dues_deadline ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, dues_deadline: e.target.value }))}
                  />
                </div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Grace Period (days)</label>
                  <input
                    className="form-input"
                    type="number"
                    value={form.grace_period_days ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, grace_period_days: Number(e.target.value) }))}
                    placeholder="14"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">District Email</label>
                  <input
                    className="form-input"
                    type="email"
                    value={form.district_email ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, district_email: e.target.value }))}
                    placeholder="district@example.org"
                  />
                </div>
              </div>
            </div>
          </Card.Body>
        </Card>

        {/* Quiz settings */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Quiz & Assessment" />
          <Card.Body>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <div className="form-group">
                <label className="form-label">Retake Limit</label>
                <input
                  className="form-input"
                  type="number"
                  value={form.quiz_retake_limit ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, quiz_retake_limit: Number(e.target.value) }))}
                  placeholder="3"
                />
              </div>
              <div className="form-group">
                <label className="form-label">Pass Score (%)</label>
                <input
                  className="form-input"
                  type="number"
                  value={form.quiz_pass_score ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, quiz_pass_score: Number(e.target.value) }))}
                  placeholder="80"
                  min={0} max={100}
                />
              </div>
            </div>
          </Card.Body>
        </Card>

        {/* Integrations */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Integrations" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div className="form-group">
                <label className="form-label">Zeffy Form ID</label>
                <input
                  className="form-input"
                  value={form.zeffy_form_id ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, zeffy_form_id: e.target.value }))}
                  placeholder="zeffy-form-uuid"
                />
                <p style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 4 }}>
                  Used for dues collection via Zeffy.
                </p>
              </div>
              <div className="form-group">
                <label className="form-label">Stripe Payment Link</label>
                <input
                  className="form-input"
                  value={form.stripe_payment_link ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, stripe_payment_link: e.target.value }))}
                  placeholder="https://buy.stripe.com/..."
                />
              </div>
              <div className="form-group">
                <label className="form-label">Make.com Webhook URL</label>
                <input
                  className="form-input"
                  value={form.make_webhook_url ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, make_webhook_url: e.target.value }))}
                  placeholder="https://hook.make.com/..."
                />
              </div>
            </div>
          </Card.Body>
        </Card>

        {/* Directory visibility */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Privacy & Visibility" />
          <Card.Body>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0.5rem 0' }}>
              <div>
                <div style={{ fontWeight: 600, fontSize: '.88rem' }}>Public Member Directory</div>
                <div style={{ fontSize: '.75rem', color: 'var(--muted)', marginTop: 2 }}>
                  Allow members to see the full directory
                </div>
              </div>
              <input
                type="checkbox"
                checked={form.allow_public_directory ?? true}
                onChange={(e) => setForm((f) => ({ ...f, allow_public_directory: e.target.checked }))}
                style={{ width: 18, height: 18, accentColor: 'var(--navy)', cursor: 'pointer' }}
              />
            </div>
          </Card.Body>
        </Card>

        <Button
          variant="gold"
          style={{ width: '100%' }}
          onClick={() => saveMutation.mutate()}
          loading={saveMutation.isPending}
        >
          Save Settings
        </Button>
      </main>
    </>
  )
}
