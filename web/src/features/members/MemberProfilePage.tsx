import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { useState } from 'react'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { useToast } from '../../components/Toast'
import { getMember, getMemberXPHistory, updateMember } from '../../api/members'
import { useAuth } from '../../hooks/useAuth'
import type { Member } from '../../types'

interface EditFormData {
  first_name: string
  last_name: string
  phone: string
  employer: string
  job_title: string
  city: string
  linkedin_url: string
  bio: string
}

function XPHistoryItem({ entry }: { entry: { activity: string; xp_awarded: number; created_at: string; source: string } }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 0', borderBottom: '1px solid var(--border)' }}>
      <div style={{ flex: 1 }}>
        <div style={{ fontSize: '.85rem', color: 'var(--ink)', fontWeight: 500 }}>{entry.activity}</div>
        <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '2px' }}>
          {new Date(entry.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
          {' · '}{entry.source}
        </div>
      </div>
      <div style={{ fontFamily: 'DM Mono, monospace', fontSize: '.83rem', fontWeight: 700, color: entry.xp_awarded >= 0 ? 'var(--navy)' : 'var(--danger)' }}>
        {entry.xp_awarded >= 0 ? '+' : ''}{entry.xp_awarded} XP
      </div>
    </div>
  )
}

function EditModal({ member, onClose }: { member: Member; onClose: () => void }) {
  const { showToast } = useToast()
  const qc = useQueryClient()
  const { register, handleSubmit } = useForm<EditFormData>({
    defaultValues: {
      first_name: member.first_name,
      last_name: member.last_name,
      phone: member.phone ?? '',
      employer: member.employer ?? '',
      job_title: member.job_title ?? '',
      city: member.city ?? '',
      linkedin_url: member.linkedin_url ?? '',
      bio: member.bio ?? '',
    },
  })

  const mutation = useMutation({
    mutationFn: (data: EditFormData) => updateMember(member.id, data),
    onSuccess: () => {
      showToast('Profile updated', 'success')
      qc.invalidateQueries({ queryKey: ['member', member.id] })
      onClose()
    },
    onError: () => showToast('Update failed', 'error'),
  })

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <span className="modal-title">Edit Profile</span>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit((d) => mutation.mutate(d))}>
          <div className="modal-body">
            <div className="form-row">
              <div className="field">
                <label className="input-label">First Name</label>
                <input className="input" {...register('first_name', { required: true })} />
              </div>
              <div className="field">
                <label className="input-label">Last Name</label>
                <input className="input" {...register('last_name', { required: true })} />
              </div>
            </div>
            <div className="field">
              <label className="input-label">Phone</label>
              <input className="input" {...register('phone')} />
            </div>
            <div className="form-row">
              <div className="field">
                <label className="input-label">Employer</label>
                <input className="input" {...register('employer')} />
              </div>
              <div className="field">
                <label className="input-label">Job Title</label>
                <input className="input" {...register('job_title')} />
              </div>
            </div>
            <div className="field">
              <label className="input-label">City</label>
              <input className="input" {...register('city')} />
            </div>
            <div className="field">
              <label className="input-label">LinkedIn URL</label>
              <input className="input" {...register('linkedin_url')} placeholder="https://linkedin.com/in/..." />
            </div>
            <div className="field">
              <label className="input-label">Bio</label>
              <textarea
                className="input"
                rows={3}
                style={{ resize: 'vertical' }}
                {...register('bio')}
              />
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-outline btn-sm" onClick={onClose}>Cancel</button>
            <button type="submit" className="btn btn-primary btn-sm" disabled={mutation.isPending}>
              {mutation.isPending ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export function MemberProfilePage() {
  const { id } = useParams<{ id: string }>()
  const { memberID, isAdmin, isSysadmin } = useAuth()
  const [editOpen, setEditOpen] = useState(false)

  const { data: member, isLoading } = useQuery({
    queryKey: ['member', id],
    queryFn: () => getMember(id!),
    enabled: !!id,
  })

  const { data: xpHistory = [] } = useQuery({
    queryKey: ['member', id, 'xp-history'],
    queryFn: () => getMemberXPHistory(id!),
    enabled: !!id,
  })

  const canEdit = member && (isAdmin || isSysadmin || memberID === member.id)

  if (isLoading) {
    return (
      <>
        <Topbar title="Member Profile" />
        <main className="page-body">
          <div style={{ textAlign: 'center', paddingTop: '3rem', color: 'var(--muted)' }}>Loading...</div>
        </main>
      </>
    )
  }

  if (!member) {
    return (
      <>
        <Topbar title="Member Profile" />
        <main className="page-body">
          <div className="empty-state">
            <div className="empty-state-icon">👤</div>
            <div className="empty-state-text">Member not found</div>
            <Link to="/members" className="btn btn-outline btn-sm" style={{ marginTop: '1rem', display: 'inline-flex' }}>
              Back to Directory
            </Link>
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Member Profile" />
      <main className="page-body">
        {editOpen && <EditModal member={member} onClose={() => setEditOpen(false)} />}

        {/* Profile hero */}
        <div className="profile-hero fade-in">
          <div
            className="profile-avatar-lg"
            style={{ background: member.avatar_bg ?? '#001A4D', color: member.avatar_fg ?? '#C9A84C' }}
          >
            {member.first_name[0]}{member.last_name[0]}
          </div>
          <div style={{ flex: 1 }}>
            <div className="profile-name-big">
              Bro. {member.first_name} {member.last_name}
            </div>
            <div className="profile-meta">
              {member.display_id}
              {member.inducted_year ? ` · Inducted ${member.inducted_year}` : ''}
            </div>
            <div className="profile-badges-row">
              <span className={`badge badge-${member.role}`}>{member.role}</span>
              <span className={`badge badge-${member.level_key}`}>{member.level}</span>
              {member.dues_status === 'paid' && (
                <span className="badge badge-success">Dues Paid</span>
              )}
              {member.dues_status !== 'paid' && member.dues_status !== 'waived' && (
                <span className="badge badge-danger">Dues {member.dues_status}</span>
              )}
            </div>
          </div>
          {canEdit && (
            <button className="btn btn-gold btn-sm" onClick={() => setEditOpen(true)}>
              Edit Profile
            </button>
          )}
        </div>

        <div className="info-grid">
          {/* Left column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {/* XP Overview */}
            <Card>
              <Card.Header title="XP Overview" />
              <Card.Body>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                  <div className="stat-card">
                    <div className="stat-label">Total XP</div>
                    <div className="stat-value">{member.xp_total.toLocaleString()}</div>
                  </div>
                  <div className="stat-card">
                    <div className="stat-label">Semester XP</div>
                    <div className="stat-value">{member.xp_semester.toLocaleString()}</div>
                  </div>
                </div>
              </Card.Body>
            </Card>

            {/* Contact / Professional */}
            <Card>
              <Card.Header title="Professional" />
              <Card.Body>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                  {[
                    { label: 'Employer', value: member.employer },
                    { label: 'Job Title', value: member.job_title },
                    { label: 'City', value: member.city },
                    { label: 'Phone', value: member.phone },
                    {
                      label: 'LinkedIn',
                      value: member.linkedin_url ? (
                        <a href={member.linkedin_url} target="_blank" rel="noopener noreferrer">
                          {member.linkedin_url.replace(/^https?:\/\//i, '')}
                        </a>
                      ) : null,
                    },
                  ].map(({ label, value }) =>
                    value ? (
                      <div key={label} className="info-item">
                        <div className="info-key">{label}</div>
                        <div className="info-val">{value}</div>
                      </div>
                    ) : null
                  )}
                  {member.bio && (
                    <div className="info-item">
                      <div className="info-key">Bio</div>
                      <div className="info-val" style={{ lineHeight: 1.5 }}>{member.bio}</div>
                    </div>
                  )}
                </div>
              </Card.Body>
            </Card>
          </div>

          {/* Right column */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            {/* Service hours */}
            <Card>
              <Card.Header title="Service" />
              <Card.Body>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
                  <div className="stat-card">
                    <div className="stat-label">Total Hours</div>
                    <div className="stat-value">{member.service_hours_total ?? 0}</div>
                  </div>
                  <div className="stat-card">
                    <div className="stat-label">Semester Hrs</div>
                    <div className="stat-value">{member.service_hours_semester ?? 0}</div>
                  </div>
                </div>
              </Card.Body>
            </Card>

            {/* XP History */}
            <Card>
              <Card.Header title="XP History" />
              <Card.Body>
                {xpHistory.length === 0 && (
                  <div style={{ color: 'var(--muted)', fontSize: '.83rem' }}>No XP history yet</div>
                )}
                {xpHistory.slice(0, 10).map((entry, i) => (
                  <XPHistoryItem key={i} entry={entry} />
                ))}
              </Card.Body>
            </Card>
          </div>
        </div>
      </main>
    </>
  )
}
