import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { createAnnouncement, type CreateAnnouncementPayload } from '../../api/announcements'

const CATEGORIES = ['General', 'Events', 'Service', 'Dues', 'Academic', 'Brotherhood', 'Alumni', 'Other']

export function CreateAnnouncementPage() {
  const { isAdmin } = useAuth()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { register, handleSubmit, formState: { errors } } = useForm<CreateAnnouncementPayload>({
    defaultValues: { is_pinned: false },
  })

  const mutation = useMutation({
    mutationFn: createAnnouncement,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['announcements'] })
      navigate('/announcements')
    },
  })

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Post Announcement" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--danger)' }}>Admin access required.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Post Announcement" />
      <main className="page-body" style={{ maxWidth: 640 }}>
        <Card>
          <Card.Header title="New Announcement" />
          <Card.Body>
            <form
              onSubmit={handleSubmit((d) => mutation.mutate(d))}
              style={{ display: 'flex', flexDirection: 'column', gap: 16 }}
            >
              <div className="form-group">
                <label className="form-label">Title</label>
                <input
                  className="form-input"
                  {...register('title', { required: 'Title is required' })}
                  placeholder="Important announcement..."
                />
                {errors.title && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.title.message}</p>}
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="form-group">
                  <label className="form-label">Category</label>
                  <select className="form-input" {...register('category')}>
                    <option value="">Select category</option>
                    {CATEGORIES.map((c) => <option key={c}>{c}</option>)}
                  </select>
                </div>
                <div className="form-group" style={{ display: 'flex', alignItems: 'flex-end', paddingBottom: 2 }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', fontSize: '.88rem' }}>
                    <input type="checkbox" {...register('is_pinned')} />
                    Pin this announcement
                  </label>
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">Message</label>
                <textarea
                  className="form-input"
                  rows={6}
                  {...register('body', { required: 'Message is required' })}
                  placeholder="Write your announcement..."
                />
                {errors.body && <p style={{ color: 'var(--danger)', fontSize: '.8rem' }}>{errors.body.message}</p>}
              </div>

              {mutation.isError && (
                <p style={{ color: 'var(--danger)', fontSize: '.85rem' }}>Failed to post. Try again.</p>
              )}

              <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                <Button variant="ghost" type="button" onClick={() => navigate('/announcements')}>
                  Cancel
                </Button>
                <Button variant="gold" type="submit" loading={mutation.isPending}>
                  Post Announcement
                </Button>
              </div>
            </form>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
