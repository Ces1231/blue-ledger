import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getSettings, updateSettings } from '../../api/settings'
import { getEngagementLog } from '../../api/xp'

function Toggle({
  label,
  description,
  checked,
  onChange,
}: {
  label: string
  description?: string
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0.85rem 0',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <div>
        <div style={{ fontSize: '.88rem', fontWeight: 600, color: 'var(--ink)' }}>{label}</div>
        {description && (
          <div style={{ fontSize: '.75rem', color: 'var(--muted)', marginTop: 2 }}>{description}</div>
        )}
      </div>
      <button
        onClick={() => onChange(!checked)}
        style={{
          width: 44, height: 24, borderRadius: 99,
          background: checked ? 'var(--navy)' : 'var(--border)',
          border: 'none', cursor: 'pointer',
          position: 'relative', flexShrink: 0, transition: 'background 0.2s',
        }}
      >
        <span
          style={{
            position: 'absolute', top: 3,
            left: checked ? 23 : 3,
            width: 18, height: 18, borderRadius: '50%',
            background: checked ? 'var(--gold)' : '#fff',
            transition: 'left 0.2s, background 0.2s',
          }}
        />
      </button>
    </div>
  )
}

export function SettingsPage() {
  const { logout } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()

  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })

  const [notifPrefs, setNotifPrefs] = useState<Record<string, boolean>>({
    badges: true, props: true, dues: true, events: true, announcements: true, messages: true,
  })
  const [privacyPrefs, setPrivacyPrefs] = useState<Record<string, boolean>>({
    showEmployer: true, showCity: true, showLinkedIn: true, showXP: true,
  })
  const [darkMode, setDarkMode] = useState(false)

  useEffect(() => {
    if (settings) {
      const prefs = settings.notification_prefs ?? {}
      setNotifPrefs((prev) => ({ ...prev, ...prefs }))
      setDarkMode(settings.dark_mode ?? false)
    }
  }, [settings])

  const saveMutation = useMutation({
    mutationFn: () =>
      updateSettings({
        notification_prefs: { ...notifPrefs, ...privacyPrefs },
        dark_mode: darkMode,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      showToast('Settings saved.', 'success')
    },
    onError: () => showToast('Failed to save settings.', 'error'),
  })

  const handleExport = async () => {
    try {
      const log = await getEngagementLog({ per_page: 1000 })
      const rows = log.data.map((e) =>
        [e.created_at, e.activity, e.xp_awarded, e.source].join(',')
      )
      const csv = ['Date,Activity,XP,Source', ...rows].join('\n')
      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'blue-ledger-my-data.csv'
      a.click()
      URL.revokeObjectURL(url)
      showToast('Data exported.', 'success')
    } catch {
      showToast('Export failed.', 'error')
    }
  }

  return (
    <>
      <Topbar title="Settings" />
      <main className="page-body" style={{ maxWidth: 600 }}>
        {/* Notification preferences */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Notification Preferences" />
          <Card.Body>
            <Toggle
              label="Badges"
              description="Notify me when I earn a new badge"
              checked={notifPrefs.badges ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, badges: v }))}
            />
            <Toggle
              label="Props"
              description="Notify me when I receive props"
              checked={notifPrefs.props ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, props: v }))}
            />
            <Toggle
              label="Dues Reminders"
              description="Reminders for upcoming or past due dues"
              checked={notifPrefs.dues ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, dues: v }))}
            />
            <Toggle
              label="Events"
              description="Upcoming events and RSVP reminders"
              checked={notifPrefs.events ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, events: v }))}
            />
            <Toggle
              label="Announcements"
              description="Chapter announcements and news"
              checked={notifPrefs.announcements ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, announcements: v }))}
            />
            <Toggle
              label="Messages"
              description="New direct messages"
              checked={notifPrefs.messages ?? true}
              onChange={(v) => setNotifPrefs((p) => ({ ...p, messages: v }))}
            />
          </Card.Body>
        </Card>

        {/* Privacy settings */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Privacy" />
          <Card.Body>
            <Toggle
              label="Show Employer"
              description="Display employer in the member directory"
              checked={privacyPrefs.showEmployer ?? true}
              onChange={(v) => setPrivacyPrefs((p) => ({ ...p, showEmployer: v }))}
            />
            <Toggle
              label="Show City"
              description="Display city in the member directory"
              checked={privacyPrefs.showCity ?? true}
              onChange={(v) => setPrivacyPrefs((p) => ({ ...p, showCity: v }))}
            />
            <Toggle
              label="Show LinkedIn"
              description="Display LinkedIn link in your profile"
              checked={privacyPrefs.showLinkedIn ?? true}
              onChange={(v) => setPrivacyPrefs((p) => ({ ...p, showLinkedIn: v }))}
            />
            <Toggle
              label="Show XP on Leaderboard"
              description="Include your XP on the public leaderboard"
              checked={privacyPrefs.showXP ?? true}
              onChange={(v) => setPrivacyPrefs((p) => ({ ...p, showXP: v }))}
            />
          </Card.Body>
        </Card>

        {/* Display settings */}
        <Card style={{ marginBottom: '1.25rem' }}>
          <Card.Header title="Display" />
          <Card.Body>
            <Toggle
              label="Dark Mode"
              description="Enable dark theme (coming in full release)"
              checked={darkMode}
              onChange={setDarkMode}
            />
          </Card.Body>
        </Card>

        {/* Save button */}
        <Button
          variant="gold"
          onClick={() => saveMutation.mutate()}
          loading={saveMutation.isPending}
          style={{ marginBottom: '1.25rem', width: '100%' }}
        >
          Save Settings
        </Button>

        {/* Data & account */}
        <Card>
          <Card.Header title="Data & Account" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <Button variant="outline" onClick={handleExport}>
                  Export My Data (CSV)
                </Button>
                <p style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 6 }}>
                  Downloads your XP activity log.
                </p>
              </div>
              <div
                style={{
                  borderTop: '1px solid var(--border)',
                  paddingTop: 12,
                }}
              >
                <Button
                  variant="danger"
                  onClick={() => {
                    if (confirm('Are you sure you want to sign out?')) logout()
                  }}
                >
                  Sign Out
                </Button>
              </div>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
