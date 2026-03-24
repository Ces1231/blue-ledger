import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { getMembers } from '../../api/members'
import { useAuth } from '../../hooks/useAuth'
import type { Member } from '../../types'

const ROLE_ORDER: Record<string, number> = {
  admin: 0,
  chair: 1,
  pia: 2,
  member: 3,
  sysadmin: 99,
}

function MemberRow({ member }: { member: Member }) {
  return (
    <Link
      to={`/members/${member.id}`}
      style={{ textDecoration: 'none', color: 'inherit' }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '12px',
          padding: '12px 1.25rem',
          borderBottom: '1px solid var(--border)',
          transition: 'background .12s',
          cursor: 'pointer',
        }}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--cream)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
      >
        <div
          className="avatar-circle"
          style={{
            width: 40,
            height: 40,
            background: member.avatar_bg ?? '#001A4D',
            color: member.avatar_fg ?? '#C9A84C',
            fontSize: '.8rem',
            flexShrink: 0,
          }}
        >
          {member.first_name[0]}{member.last_name[0]}
        </div>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: '.88rem', fontWeight: 600, color: 'var(--ink)', display: 'flex', alignItems: 'center', gap: '8px' }}>
            {member.first_name} {member.last_name}
            {member.role !== 'member' && (
              <span className={`badge badge-${member.role}`} style={{ flexShrink: 0 }}>
                {member.role}
              </span>
            )}
          </div>
          <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '2px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {member.display_id}
            {member.employer ? ` · ${member.employer}` : ''}
          </div>
        </div>

        <div style={{ textAlign: 'right', flexShrink: 0 }}>
          <div
            style={{
              fontFamily: 'DM Mono, monospace',
              fontSize: '.78rem',
              fontWeight: 700,
              color: 'var(--navy)',
            }}
          >
            {member.xp_total.toLocaleString()} XP
          </div>
          <div style={{ marginTop: '3px' }}>
            <span className={`badge badge-${member.level_key}`}>{member.level}</span>
          </div>
        </div>
      </div>
    </Link>
  )
}

export function DirectoryPage() {
  const { isAuthenticated } = useAuth()
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('all')
  const [statusFilter, setStatusFilter] = useState('active')

  const { data, isLoading } = useQuery({
    queryKey: ['members', 'list', { per_page: 200 }],
    queryFn: () => getMembers({ per_page: 200 }),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 2,
  })

  const members = data?.data ?? []

  const filtered = members
    .filter((m) => {
      if (statusFilter === 'active' && m.status !== 'active') return false
      if (statusFilter === 'inactive' && m.status !== 'inactive') return false
      if (roleFilter !== 'all' && m.role !== roleFilter) return false
      if (search) {
        const q = search.toLowerCase()
        const name = `${m.first_name} ${m.last_name}`.toLowerCase()
        return name.includes(q) || m.display_id.toLowerCase().includes(q) || (m.employer ?? '').toLowerCase().includes(q)
      }
      return true
    })
    .sort((a, b) => {
      const ro = (ROLE_ORDER[a.role] ?? 4) - (ROLE_ORDER[b.role] ?? 4)
      if (ro !== 0) return ro
      return b.xp_total - a.xp_total
    })

  return (
    <>
      <Topbar title="Member Directory" />
      <main className="page-body">
        {/* Stats row */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: '10px', marginBottom: '1.25rem' }}>
          <div className="stat-card">
            <div className="stat-label">Total Members</div>
            <div className="stat-value">{data?.meta.total ?? '—'}</div>
            <div className="stat-sub">in chapter</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Active</div>
            <div className="stat-value">
              {members.filter((m) => m.status === 'active').length}
            </div>
            <div className="stat-sub">in good standing</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Avg XP</div>
            <div className="stat-value">
              {members.length > 0
                ? Math.round(members.reduce((s, m) => s + m.xp_total, 0) / members.length).toLocaleString()
                : '—'}
            </div>
            <div className="stat-sub">per member</div>
          </div>
        </div>

        {/* Filter bar */}
        <Card>
          <Card.Body>
            <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
              <input
                className="input"
                type="text"
                placeholder="Search by name, ID, employer..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                style={{ flex: 1, minWidth: '200px' }}
              />
              <select
                className="input select"
                value={roleFilter}
                onChange={(e) => setRoleFilter(e.target.value)}
                style={{ width: 'auto' }}
              >
                <option value="all">All Roles</option>
                <option value="admin">Admin</option>
                <option value="chair">Chair</option>
                <option value="pia">PIA</option>
                <option value="member">Member</option>
              </select>
              <select
                className="input select"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                style={{ width: 'auto' }}
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="all">All Statuses</option>
              </select>
            </div>
          </Card.Body>
        </Card>

        {/* Results */}
        <div style={{ marginTop: '1.25rem' }}>
          <Card>
            <Card.Header
              title={`${filtered.length} member${filtered.length !== 1 ? 's' : ''}`}
            />
            <div>
              {isLoading && (
                <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--muted)', fontSize: '.85rem' }}>
                  Loading members...
                </div>
              )}
              {!isLoading && filtered.length === 0 && (
                <div className="empty-state">
                  <div className="empty-state-icon">👥</div>
                  <div className="empty-state-text">No members match your filters</div>
                </div>
              )}
              {filtered.map((m) => (
                <MemberRow key={m.id} member={m} />
              ))}
            </div>
          </Card>
        </div>
      </main>
    </>
  )
}
