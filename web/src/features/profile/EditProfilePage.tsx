import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useCurrentMember } from '../../hooks/useCurrentMember'
import { updateMember, type UpdateMemberPayload } from '../../api/members'

const AVATAR_BG_OPTIONS = [
  { bg: '#001A4D', fg: '#C9A84C', label: 'Navy / Gold' },
  { bg: '#1A6B3A', fg: '#FFFFFF', label: 'Green / White' },
  { bg: '#8B1A1A', fg: '#FFFFFF', label: 'Crimson / White' },
  { bg: '#003087', fg: '#FFF3DC', label: 'Royal / Cream' },
  { bg: '#2A2A2A', fg: '#C9A84C', label: 'Onyx / Gold' },
  { bg: '#C9A84C', fg: '#001A4D', label: 'Gold / Navy' },
]

export function EditProfilePage() {
  const { member } = useCurrentMember()
  const qc = useQueryClient()
  const navigate = useNavigate()
  const { showToast } = useToast()

  const [form, setForm] = useState<UpdateMemberPayload>({})

  useEffect(() => {
    if (member) {
      setForm({
        inducted_year: member.inducted_year,
        employer: member.employer ?? '',
        job_title: member.job_title ?? '',
        city: member.city ?? '',
        linkedin_url: member.linkedin_url ?? '',
        avatar_bg: member.avatar_bg ?? '#001A4D',
        avatar_fg: member.avatar_fg ?? '#C9A84C',
      })
    }
  }, [member])

  const mutation = useMutation({
    mutationFn: () => updateMember(member!.id, form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['member', 'current'] })
      showToast('Profile updated successfully.', 'success')
      navigate('/profile')
    },
    onError: () => showToast('Failed to save profile.', 'error'),
  })

  if (!member) return null

  const initials = `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()

  return (
    <>
      <Topbar title="Edit Profile" />
      <main className="page-body" style={{ maxWidth: 560 }}>
        <Card>
          <Card.Header title="Profile Details" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {/* Avatar preview */}
              <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 8 }}>
                <div
                  style={{
                    width: 80, height: 80, borderRadius: '50%',
                    background: form.avatar_bg ?? '#001A4D',
                    color: form.avatar_fg ?? '#C9A84C',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: '1.6rem', fontWeight: 700,
                  }}
                >
                  {initials}
                </div>
              </div>

              {/* Avatar color picker */}
              <div className="form-group">
                <label className="form-label">Avatar Color</label>
                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  {AVATAR_BG_OPTIONS.map((opt) => (
                    <button
                      key={opt.bg}
                      onClick={() => setForm((f) => ({ ...f, avatar_bg: opt.bg, avatar_fg: opt.fg }))}
                      title={opt.label}
                      style={{
                        width: 36, height: 36, borderRadius: '50%',
                        background: opt.bg, color: opt.fg,
                        border: form.avatar_bg === opt.bg ? '3px solid var(--gold)' : '2px solid transparent',
                        cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: '.7rem', fontWeight: 700,
                      }}
                    >
                      {initials}
                    </button>
                  ))}
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Employer</label>
                  <input
                    className="form-input"
                    value={form.employer ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, employer: e.target.value }))}
                    placeholder="Company name"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Job Title</label>
                  <input
                    className="form-input"
                    value={form.job_title ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, job_title: e.target.value }))}
                    placeholder="Your title"
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">City</label>
                  <input
                    className="form-input"
                    value={form.city ?? ''}
                    onChange={(e) => setForm((f) => ({ ...f, city: e.target.value }))}
                    placeholder="Atlanta, GA"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Inducted Year</label>
                  <input
                    className="form-input"
                    type="number"
                    value={form.inducted_year ?? ''}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, inducted_year: e.target.value ? Number(e.target.value) : undefined }))
                    }
                    placeholder="2022"
                  />
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">LinkedIn URL</label>
                <input
                  className="form-input"
                  value={form.linkedin_url ?? ''}
                  onChange={(e) => setForm((f) => ({ ...f, linkedin_url: e.target.value }))}
                  placeholder="https://linkedin.com/in/yourname"
                />
              </div>

              <div style={{ display: 'flex', gap: 10, marginTop: 8 }}>
                <Button
                  variant="gold"
                  onClick={() => mutation.mutate()}
                  loading={mutation.isPending}
                >
                  Save Changes
                </Button>
                <Button variant="outline" onClick={() => navigate('/profile')}>
                  Cancel
                </Button>
              </div>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
