import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  sysListChapters,
  sysListUsers,
  sysGetAuditLog,
  sysRecalculateXP,
  type AuditEntry,
} from '../../api/platform'
import { getEngagementLog } from '../../api/xp'
import { getPointEconomy, updatePointEconomy } from '../../api/settings'

// ── Dark console color tokens ──────────────────────────────────────────────────
const C = {
  bg:      '#0D1117',
  surface: '#161B22',
  border:  '#30363D',
  text:    '#E6EDF3',
  muted:   '#8B949E',
  gold:    '#C9A84C',
  green:   '#3FB950',
  yellow:  '#D29922',
  red:     '#F85149',
  blue:    '#58A6FF',
  navy:    '#001A4D',
}

const TABS = ['Overview', 'Users', 'Integrations', 'Data Tools', 'Audit Log', 'App Config'] as const
type Tab = typeof TABS[number]

function ConsoleCard({ title, children }: { title?: string; children: React.ReactNode }) {
  return (
    <div
      style={{
        background: C.surface, border: `1px solid ${C.border}`,
        borderRadius: 10, overflow: 'hidden', marginBottom: '1rem',
      }}
    >
      {title && (
        <div
          style={{
            padding: '0.75rem 1rem',
            borderBottom: `1px solid ${C.border}`,
            fontWeight: 600, fontSize: '.82rem', color: C.muted,
            fontFamily: 'DM Mono, monospace', letterSpacing: '.08em', textTransform: 'uppercase',
          }}
        >
          {title}
        </div>
      )}
      <div style={{ padding: '1rem' }}>{children}</div>
    </div>
  )
}

function LevelBadge({ level }: { level: string }) {
  const colors: Record<string, { bg: string; fg: string }> = {
    ok:    { bg: '#0D2B1A', fg: C.green },
    warn:  { bg: '#2B1E0A', fg: C.yellow },
    error: { bg: '#2B0D0D', fg: C.red },
    info:  { bg: '#0A1628', fg: C.blue },
  }
  const c = colors[level] ?? colors.info
  return (
    <span
      style={{
        background: c.bg, color: c.fg,
        borderRadius: 4, padding: '2px 6px',
        fontSize: '.65rem', fontWeight: 700, fontFamily: 'DM Mono, monospace',
        textTransform: 'uppercase', letterSpacing: '.08em',
      }}
    >
      {level}
    </span>
  )
}

function ServiceHealthCard({ name, ok }: { name: string; ok: boolean }) {
  return (
    <div
      style={{
        background: C.bg, border: `1px solid ${C.border}`,
        borderRadius: 8, padding: '0.85rem 1rem',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      }}
    >
      <span style={{ fontSize: '.85rem', color: C.text }}>{name}</span>
      <span
        style={{
          color: ok ? C.green : C.red,
          fontSize: '.75rem', fontFamily: 'DM Mono, monospace',
        }}
      >
        {ok ? '● Online' : '● Offline'}
      </span>
    </div>
  )
}

// ── Tab: Overview ─────────────────────────────────────────────────────────────
function OverviewTab() {
  const { data: chapters } = useQuery({ queryKey: ['sys-chapters'], queryFn: () => sysListChapters() })
  const { data: auditData } = useQuery({ queryKey: ['sys-audit-recent'], queryFn: () => sysGetAuditLog({ per_page: 5 }) })

  const chapterCount = chapters?.meta.total ?? 0
  const recentEntries = auditData?.data ?? []

  const SERVICES = ['Postgres', 'Redis', 'Stripe', 'Zeffy', 'Resend', 'R2']

  return (
    <>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2,1fr)', gap: '0.75rem', marginBottom: '1rem' }}>
        <ConsoleCard title="Chapters">
          <div style={{ fontSize: '2rem', fontFamily: 'DM Mono, monospace', color: C.gold }}>{chapterCount}</div>
          <div style={{ fontSize: '.72rem', color: C.muted, marginTop: 4 }}>active deployments</div>
        </ConsoleCard>
        <ConsoleCard title="Platform">
          <div style={{ fontSize: '2rem', fontFamily: 'DM Mono, monospace', color: C.green }}>v1.0</div>
          <div style={{ fontSize: '.72rem', color: C.muted, marginTop: 4 }}>blue-ledger-api</div>
        </ConsoleCard>
      </div>

      <ConsoleCard title="Service Health">
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {SERVICES.map((s) => (
            <ServiceHealthCard key={s} name={s} ok={true} />
          ))}
        </div>
      </ConsoleCard>

      <ConsoleCard title="Recent Audit Log">
        {recentEntries.length === 0 && (
          <p style={{ color: C.muted, fontSize: '.8rem' }}>No recent entries.</p>
        )}
        {recentEntries.map((e) => (
          <div
            key={e.id}
            style={{
              display: 'flex', gap: 12, alignItems: 'center',
              padding: '0.5rem 0', borderBottom: `1px solid ${C.border}`,
              fontSize: '.78rem',
            }}
          >
            <LevelBadge level="info" />
            <span style={{ color: C.text, flex: 1 }}>{e.action}</span>
            <span style={{ color: C.muted, fontFamily: 'DM Mono, monospace', fontSize: '.7rem', whiteSpace: 'nowrap' }}>
              {new Date(e.created_at).toLocaleTimeString()}
            </span>
          </div>
        ))}
      </ConsoleCard>
    </>
  )
}

// ── Tab: Users ────────────────────────────────────────────────────────────────
function UsersTab() {
  const { data } = useQuery({ queryKey: ['sys-users'], queryFn: () => sysListUsers({ per_page: 50 }) })
  const users = data?.data ?? []

  return (
    <ConsoleCard title={`Users (${data?.meta.total ?? 0})`}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '.8rem' }}>
        <thead>
          <tr style={{ color: C.muted }}>
            <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', borderBottom: `1px solid ${C.border}` }}>Email</th>
            <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', borderBottom: `1px solid ${C.border}` }}>Name</th>
            <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', borderBottom: `1px solid ${C.border}` }}>Role</th>
            <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', borderBottom: `1px solid ${C.border}` }}>Chapters</th>
            <th style={{ textAlign: 'left', padding: '0.4rem 0.5rem', borderBottom: `1px solid ${C.border}` }}>Joined</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id}>
              <td style={{ padding: '0.5rem', color: C.text, fontFamily: 'DM Mono, monospace', fontSize: '.75rem' }}>
                {u.email}
              </td>
              <td style={{ padding: '0.5rem', color: C.text }}>
                {u.first_name} {u.last_name}
              </td>
              <td style={{ padding: '0.5rem' }}>
                {u.is_sysadmin ? (
                  <span style={{ color: C.red, fontSize: '.68rem', fontWeight: 700 }}>SYSADMIN</span>
                ) : (
                  <span style={{ color: C.muted }}>member</span>
                )}
              </td>
              <td style={{ padding: '0.5rem', color: C.muted, fontFamily: 'DM Mono, monospace' }}>
                {u.chapter_count}
              </td>
              <td style={{ padding: '0.5rem', color: C.muted, fontSize: '.72rem', whiteSpace: 'nowrap' }}>
                {new Date(u.created_at).toLocaleDateString('en', { month: 'short', day: 'numeric', year: '2-digit' })}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </ConsoleCard>
  )
}

// ── Tab: Integrations ─────────────────────────────────────────────────────────
function IntegrationsTab() {
  const { showToast } = useToast()

  const INTEGRATIONS = [
    { name: 'PostgreSQL', description: 'Primary database (Fly.io Postgres)', endpoint: 'internal' },
    { name: 'Redis', description: 'Session store and rate limiter (Fly.io)', endpoint: 'internal' },
    { name: 'Stripe', description: 'Chapter subscriptions and dues payments', endpoint: 'stripe.com' },
    { name: 'Zeffy', description: 'Dues collection form embed', endpoint: 'zeffy.com' },
    { name: 'Resend', description: 'Transactional email (magic links, reminders)', endpoint: 'resend.com' },
    { name: 'Cloudflare R2', description: 'Avatar and resource file storage', endpoint: 'r2.cloudflarestorage.com' },
    { name: 'Sentry', description: 'Error monitoring (API + React)', endpoint: 'sentry.io' },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
      {INTEGRATIONS.map((integ) => (
        <div
          key={integ.name}
          style={{
            background: C.bg, border: `1px solid ${C.border}`,
            borderRadius: 8, padding: '1rem',
            display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16,
          }}
        >
          <div>
            <div style={{ fontWeight: 600, color: C.text, fontSize: '.88rem' }}>{integ.name}</div>
            <div style={{ fontSize: '.75rem', color: C.muted, marginTop: 2 }}>{integ.description}</div>
            <div style={{ fontSize: '.68rem', color: C.muted, fontFamily: 'DM Mono, monospace', marginTop: 4 }}>
              {integ.endpoint}
            </div>
          </div>
          <button
            onClick={() => showToast(`Test connection to ${integ.name} is available in the full release.`, 'info')}
            style={{
              background: C.surface, border: `1px solid ${C.border}`,
              borderRadius: 6, padding: '5px 12px', color: C.text,
              cursor: 'pointer', fontSize: '.75rem', fontWeight: 600, whiteSpace: 'nowrap',
            }}
          >
            Test Connection
          </button>
        </div>
      ))}
    </div>
  )
}

// ── Tab: Data Tools ───────────────────────────────────────────────────────────
function DataToolsTab() {
  const { showToast } = useToast()
  const [confirmText, setConfirmText] = useState('')

  const recalcMutation = useMutation({
    mutationFn: sysRecalculateXP,
    onSuccess: (r) => showToast(`Recalculated XP for ${r.members_updated} members.`, 'success'),
    onError: () => showToast('Recalculation failed.', 'error'),
  })

  const handleExport = async () => {
    try {
      const log = await getEngagementLog({ per_page: 5000 })
      const rows = log.data.map((e) =>
        [e.created_at, e.member_id, e.activity, e.xp_awarded, e.source].join(',')
      )
      const csv = ['Date,MemberID,Activity,XP,Source', ...rows].join('\n')
      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `blue-ledger-platform-export-${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
      showToast('Platform data exported.', 'success')
    } catch {
      showToast('Export failed.', 'error')
    }
  }

  const TOOLS = [
    {
      label: 'Import Members (CSV)',
      description: 'Bulk import member records from CSV file.',
      action: () => showToast('Bulk import is available in the full release.', 'info'),
      variant: 'outline' as const,
    },
    {
      label: 'Export Platform Data',
      description: 'Download full engagement log as CSV.',
      action: handleExport,
      variant: 'outline' as const,
    },
    {
      label: 'Force Sync Integrations',
      description: 'Trigger manual sync with all connected services.',
      action: () => showToast('Force sync is available in the full release.', 'info'),
      variant: 'outline' as const,
    },
    {
      label: 'Recalculate All XP',
      description: 'Re-derive xp_total for all members from the engagement log.',
      action: () => recalcMutation.mutate(),
      variant: 'primary' as const,
      loading: recalcMutation.isPending,
    },
  ]

  return (
    <>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginBottom: '1.5rem' }}>
        {TOOLS.map((tool) => (
          <div
            key={tool.label}
            style={{
              background: C.bg, border: `1px solid ${C.border}`,
              borderRadius: 8, padding: '1rem',
              display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16,
            }}
          >
            <div>
              <div style={{ fontWeight: 600, color: C.text, fontSize: '.88rem' }}>{tool.label}</div>
              <div style={{ fontSize: '.75rem', color: C.muted, marginTop: 2 }}>{tool.description}</div>
            </div>
            <button
              onClick={tool.action}
              disabled={tool.loading}
              style={{
                background: tool.variant === 'primary' ? C.navy : C.surface,
                border: `1px solid ${C.border}`,
                borderRadius: 6, padding: '5px 14px',
                color: C.text, cursor: 'pointer',
                fontSize: '.78rem', fontWeight: 600, whiteSpace: 'nowrap',
                opacity: tool.loading ? 0.6 : 1,
              }}
            >
              {tool.loading ? '...' : tool.label.split(' ')[0]}
            </button>
          </div>
        ))}
      </div>

      {/* Danger Zone */}
      <div
        style={{
          background: '#1A0A0A', border: `1px solid ${C.red}40`,
          borderRadius: 10, padding: '1.25rem',
        }}
      >
        <div style={{ color: C.red, fontWeight: 700, fontSize: '.88rem', marginBottom: '0.5rem' }}>
          Danger Zone
        </div>
        <p style={{ fontSize: '.78rem', color: C.muted, marginBottom: '0.75rem', lineHeight: 1.5 }}>
          These actions are destructive and irreversible. Type CONFIRM to proceed.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <input
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            placeholder="Type CONFIRM..."
            style={{
              flex: 1, background: C.bg, border: `1px solid ${C.red}60`,
              borderRadius: 6, padding: '6px 10px',
              color: C.text, fontSize: '.82rem', fontFamily: 'DM Mono, monospace',
            }}
          />
          <button
            disabled={confirmText !== 'CONFIRM'}
            onClick={() => {
              showToast('Destructive action executed (stub — no actual deletion in dev mode).', 'warning')
              setConfirmText('')
            }}
            style={{
              background: confirmText === 'CONFIRM' ? C.red : '#2B0D0D',
              border: `1px solid ${C.red}`,
              borderRadius: 6, padding: '6px 14px',
              color: '#fff', cursor: confirmText === 'CONFIRM' ? 'pointer' : 'not-allowed',
              fontSize: '.78rem', fontWeight: 700, transition: 'background 0.15s',
            }}
          >
            Execute
          </button>
        </div>
      </div>
    </>
  )
}

// ── Tab: Audit Log ────────────────────────────────────────────────────────────
function AuditLogTab() {
  const [filter, setFilter] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['sys-audit-full'],
    queryFn: () => sysGetAuditLog({ per_page: 100 }),
  })

  const entries = (data?.data ?? []).filter(
    (e) =>
      !filter ||
      e.action.toLowerCase().includes(filter.toLowerCase()) ||
      (e.actor_email?.toLowerCase().includes(filter.toLowerCase()) ?? false)
  )

  const inferLevel = (e: AuditEntry): string => {
    const a = e.action.toLowerCase()
    if (a.includes('error') || a.includes('fail') || a.includes('delete')) return 'error'
    if (a.includes('warn') || a.includes('overrid')) return 'warn'
    if (a.includes('creat') || a.includes('updat') || a.includes('login')) return 'ok'
    return 'info'
  }

  return (
    <>
      <div style={{ marginBottom: '0.75rem' }}>
        <input
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Filter by action or email..."
          style={{
            width: '100%', background: C.surface, border: `1px solid ${C.border}`,
            borderRadius: 6, padding: '7px 10px',
            color: C.text, fontSize: '.82rem',
          }}
        />
      </div>

      <ConsoleCard>
        {isLoading && <p style={{ color: C.muted, fontSize: '.8rem' }}>Loading...</p>}
        {!isLoading && entries.length === 0 && (
          <p style={{ color: C.muted, fontSize: '.8rem' }}>No audit entries found.</p>
        )}
        {entries.map((e) => (
          <div
            key={e.id}
            style={{
              display: 'flex', gap: 10, alignItems: 'center',
              padding: '0.55rem 0', borderBottom: `1px solid ${C.border}`,
              flexWrap: 'wrap',
            }}
          >
            <LevelBadge level={inferLevel(e)} />
            <span style={{ color: C.text, fontSize: '.8rem', flex: 1 }}>{e.action}</span>
            {e.actor_email && (
              <span style={{ color: C.muted, fontSize: '.72rem', fontFamily: 'DM Mono, monospace' }}>
                {e.actor_email}
              </span>
            )}
            <span style={{ color: C.muted, fontSize: '.68rem', fontFamily: 'DM Mono, monospace', whiteSpace: 'nowrap' }}>
              {new Date(e.created_at).toLocaleString('en', {
                month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
              })}
            </span>
          </div>
        ))}
      </ConsoleCard>
    </>
  )
}

// ── Tab: App Config ───────────────────────────────────────────────────────────
function AppConfigTab() {
  const { showToast } = useToast()
  const qc = useQueryClient()

  const { data: pointEconomy } = useQuery({
    queryKey: ['point-economy'],
    queryFn: getPointEconomy,
  })

  const [econForm, setEconForm] = useState<Record<string, number>>({})

  const saveMutation = useMutation({
    mutationFn: () => updatePointEconomy(econForm as any),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['point-economy'] })
      showToast('Point economy saved.', 'success')
    },
    onError: () => showToast('Failed to save.', 'error'),
  })

  const economy = pointEconomy ?? {}

  return (
    <>
      <ConsoleCard title="XP Point Economy">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {Object.entries(economy).map(([key, val]) => (
            <div key={key} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <label
                style={{
                  fontSize: '.8rem', color: C.muted,
                  fontFamily: 'DM Mono, monospace',
                }}
              >
                {key.replace(/_/g, ' ')}
              </label>
              <input
                type="number"
                defaultValue={val as number}
                onChange={(e) => setEconForm((f) => ({ ...f, [key]: Number(e.target.value) }))}
                style={{
                  width: 80, background: C.bg, border: `1px solid ${C.border}`,
                  borderRadius: 6, padding: '4px 8px',
                  color: C.gold, fontSize: '.82rem', fontFamily: 'DM Mono, monospace',
                  textAlign: 'right',
                }}
              />
            </div>
          ))}
          <button
            onClick={() => saveMutation.mutate()}
            style={{
              marginTop: 8, background: C.navy, border: `1px solid ${C.gold}40`,
              borderRadius: 6, padding: '7px 14px',
              color: C.gold, cursor: 'pointer', fontSize: '.8rem', fontWeight: 600,
            }}
          >
            {saveMutation.isPending ? 'Saving...' : 'Save Point Economy'}
          </button>
        </div>
      </ConsoleCard>

      <ConsoleCard title="Webhook URLs">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {['Make.com Webhook', 'Stripe Webhook', 'Zeffy Callback'].map((name) => (
            <div key={name}>
              <label style={{ fontSize: '.72rem', color: C.muted, display: 'block', marginBottom: 4 }}>
                {name}
              </label>
              <input
                style={{
                  width: '100%', background: C.bg, border: `1px solid ${C.border}`,
                  borderRadius: 6, padding: '6px 10px',
                  color: C.text, fontSize: '.78rem', fontFamily: 'DM Mono, monospace',
                }}
                placeholder={`https://...`}
              />
            </div>
          ))}
        </div>
      </ConsoleCard>
    </>
  )
}

// ── Main page ──────────────────────────────────────────────────────────────────
export function SysadminConsolePage() {
  const { isSysadmin } = useAuth()
  const [activeTab, setActiveTab] = useState<Tab>('Overview')

  if (!isSysadmin) {
    return (
      <div style={{ background: C.bg, minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ color: C.red, fontFamily: 'DM Mono, monospace', fontSize: '.9rem' }}>
          ACCESS DENIED — sysadmin role required
        </div>
      </div>
    )
  }

  return (
    <div style={{ background: C.bg, minHeight: '100vh', color: C.text }}>
      {/* Console header */}
      <div
        style={{
          padding: '1rem 1.5rem',
          borderBottom: `1px solid ${C.border}`,
          display: 'flex', alignItems: 'center', gap: 14,
        }}
      >
        <div
          style={{
            width: 28, height: 32,
            background: C.gold,
            clipPath: 'polygon(50% 0%, 100% 15%, 100% 60%, 50% 100%, 0% 60%, 0% 15%)',
            flexShrink: 0,
          }}
        />
        <div>
          <div style={{ fontFamily: 'DM Mono, monospace', fontWeight: 700, fontSize: '.92rem', color: C.gold }}>
            Blue Ledger Platform Console
          </div>
          <div style={{ fontSize: '.68rem', color: C.muted, marginTop: 1, letterSpacing: '.08em' }}>
            SYSADMIN — ALL CHAPTERS VISIBLE
          </div>
        </div>
      </div>

      {/* Tab bar */}
      <div
        style={{
          display: 'flex', gap: 0,
          borderBottom: `1px solid ${C.border}`,
          overflowX: 'auto',
          padding: '0 1.5rem',
        }}
      >
        {TABS.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            style={{
              background: 'none', border: 'none', cursor: 'pointer',
              padding: '0.75rem 1.25rem',
              fontSize: '.8rem', fontFamily: 'DM Mono, monospace',
              color: activeTab === tab ? C.gold : C.muted,
              borderBottom: activeTab === tab ? `2px solid ${C.gold}` : '2px solid transparent',
              transition: 'color 0.15s, border-color 0.15s',
              whiteSpace: 'nowrap',
            }}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div style={{ padding: '1.5rem', maxWidth: 900 }}>
        {activeTab === 'Overview' && <OverviewTab />}
        {activeTab === 'Users' && <UsersTab />}
        {activeTab === 'Integrations' && <IntegrationsTab />}
        {activeTab === 'Data Tools' && <DataToolsTab />}
        {activeTab === 'Audit Log' && <AuditLogTab />}
        {activeTab === 'App Config' && <AppConfigTab />}
      </div>
    </div>
  )
}
