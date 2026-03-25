import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getMembers, updateMember, deleteMember, type UpdateMemberPayload } from '../../api/members'
import type { Member, MemberRole, MemberStatus, DuesStatus } from '../../types'

const ROLE_OPTIONS: MemberRole[] = ['member', 'chair', 'pia', 'admin']
const STATUS_OPTIONS: MemberStatus[] = ['active', 'inactive', 'alumni', 'suspended', 'pledging']
const DUES_OPTIONS: DuesStatus[] = ['paid', 'unpaid', 'late', 'waived', 'outstanding']

const ROLE_COLORS: Record<string, { bg: string; fg: string }> = {
  member:  { bg: '#EAF0FB', fg: '#003087' },
  chair:   { bg: '#F5E6C8', fg: '#7A4A00' },
  pia:     { bg: '#EAF5EE', fg: '#1A6B3A' },
  admin:   { bg: '#001A4D', fg: '#C9A84C' },
  sysadmin:{ bg: '#FCE4EC', fg: '#8B1A1A' },
}

const STATUS_COLORS: Record<string, { bg: string; fg: string }> = {
  active:    { bg: '#EAF5EE', fg: '#1A6B3A' },
  inactive:  { bg: '#F2EDE4', fg: '#6B6657' },
  alumni:    { bg: '#EAF0FB', fg: '#003087' },
  suspended: { bg: '#FCE4EC', fg: '#8B1A1A' },
  pledging:  { bg: '#FFF3DC', fg: '#7A4A00' },
}

const DUES_COLORS: Record<string, { bg: string; fg: string }> = {
  paid:        { bg: '#EAF5EE', fg: '#1A6B3A' },
  unpaid:      { bg: '#FCE4EC', fg: '#8B1A1A' },
  late:        { bg: '#FFF3DC', fg: '#7A4A00' },
  waived:      { bg: '#EAF0FB', fg: '#003087' },
  outstanding: { bg: '#FCE4EC', fg: '#8B1A1A' },
}

function StatusBadge({ text, colors }: { text: string; colors: { bg: string; fg: string } }) {
  return (
    <span
      style={{
        background: colors.bg, color: colors.fg,
        borderRadius: 99, padding: '2px 8px', fontSize: '.68rem',
        fontWeight: 600, textTransform: 'capitalize', whiteSpace: 'nowrap',
      }}
    >
      {text}
    </span>
  )
}

export function AdminMembersPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [editTarget, setEditTarget] = useState<Member | null>(null)
  const [editForm, setEditForm] = useState<UpdateMemberPayload>({})
  const [search, setSearch] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-members'],
    queryFn: () => getMembers({ per_page: 200 }),
    enabled: isAdmin,
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateMemberPayload }) =>
      updateMember(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-members'] })
      showToast('Member updated.', 'success')
      setEditTarget(null)
    },
    onError: () => showToast('Failed to update member.', 'error'),
  })

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => updateMember(id, { status: 'inactive' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-members'] })
      showToast('Member deactivated.', 'info')
    },
    onError: () => showToast('Failed to deactivate member.', 'error'),
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Members — Admin" />
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

  const members = data?.data ?? []
  const filtered = members.filter(
    (m) =>
      `${m.first_name} ${m.last_name}`.toLowerCase().includes(search.toLowerCase()) ||
      m.email.toLowerCase().includes(search.toLowerCase()) ||
      m.member_display_id.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <>
      <Topbar title="Members — Admin" />
      <main className="page-body">
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap' }}>
          <input
            className="form-input"
            style={{ flex: 1, minWidth: 200, maxWidth: 360 }}
            placeholder="Search by name, email, or ID..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <span style={{ fontSize: '.82rem', color: 'var(--muted)' }}>
            {filtered.length} of {members.length} members
          </span>
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading members...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && filtered.length > 0 && (
          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Member</th>
                  <th>ID</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Dues</th>
                  <th>XP</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((m) => (
                  <tr key={m.id}>
                    <td>
                      <div style={{ fontWeight: 600, fontSize: '.88rem' }}>
                        {m.first_name} {m.last_name}
                      </div>
                      <div style={{ fontSize: '.72rem', color: 'var(--muted)' }}>{m.email}</div>
                    </td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.78rem' }}>
                      {m.member_display_id}
                    </td>
                    <td>
                      <StatusBadge text={m.role} colors={ROLE_COLORS[m.role] ?? ROLE_COLORS.member} />
                    </td>
                    <td>
                      <StatusBadge text={m.status} colors={STATUS_COLORS[m.status] ?? STATUS_COLORS.inactive} />
                    </td>
                    <td>
                      <StatusBadge text={m.dues_status} colors={DUES_COLORS[m.dues_status] ?? DUES_COLORS.unpaid} />
                    </td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem' }}>
                      {m.xp_total.toLocaleString()}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: 6 }}>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            setEditTarget(m)
                            setEditForm({ role: m.role, status: m.status, dues_status: m.dues_status })
                          }}
                        >
                          Edit
                        </Button>
                        {m.status === 'active' && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              if (confirm(`Deactivate ${m.first_name} ${m.last_name}?`)) {
                                deactivateMutation.mutate(m.id)
                              }
                            }}
                          >
                            Deactivate
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>

      {/* Edit modal */}
      <Modal isOpen={!!editTarget} onClose={() => setEditTarget(null)} title="Edit Member" size="sm">
        {editTarget && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <p style={{ fontSize: '.88rem', color: 'var(--ink)', fontWeight: 600 }}>
              {editTarget.first_name} {editTarget.last_name}
            </p>
            <div className="form-group">
              <label className="form-label">Role</label>
              <select
                className="form-input"
                value={editForm.role ?? editTarget.role}
                onChange={(e) => setEditForm((f) => ({ ...f, role: e.target.value }))}
              >
                {ROLE_OPTIONS.map((r) => <option key={r} value={r}>{r}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Status</label>
              <select
                className="form-input"
                value={editForm.status ?? editTarget.status}
                onChange={(e) => setEditForm((f) => ({ ...f, status: e.target.value }))}
              >
                {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Dues Status</label>
              <select
                className="form-input"
                value={editForm.dues_status ?? editTarget.dues_status}
                onChange={(e) => setEditForm((f) => ({ ...f, dues_status: e.target.value }))}
              >
                {DUES_OPTIONS.map((d) => <option key={d} value={d}>{d}</option>)}
              </select>
            </div>
            <Button
              variant="gold"
              onClick={() => updateMutation.mutate({ id: editTarget.id, payload: editForm })}
              loading={updateMutation.isPending}
            >
              Save Changes
            </Button>
          </div>
        )}
      </Modal>
    </>
  )
}
