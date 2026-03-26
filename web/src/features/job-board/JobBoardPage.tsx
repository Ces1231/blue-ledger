import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getJobs,
  createJob,
  deleteJob,
  type JobPosting,
} from '../../api/job-board'

interface CreateJobForm {
  title: string
  company: string
  location: string
  url: string
  description: string
}

export function JobBoardPage() {
  const { isAdmin, isChair } = useAuth()
  const canPost = isAdmin || isChair
  const [modalOpen, setModalOpen] = useState(false)
  const [search, setSearch] = useState('')
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['job-board'],
    queryFn: getJobs,
  })

  const filtered = items.filter((j) => {
    const q = search.toLowerCase()
    return (
      !q ||
      j.title.toLowerCase().includes(q) ||
      j.company.toLowerCase().includes(q) ||
      (j.location ?? '').toLowerCase().includes(q)
    )
  })

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } =
    useForm<CreateJobForm>()

  const createMut = useMutation({
    mutationFn: (vals: CreateJobForm) =>
      createJob({
        title: vals.title,
        company: vals.company,
        location: vals.location || undefined,
        url: vals.url || undefined,
        description: vals.description,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['job-board'] })
      setModalOpen(false)
      reset()
    },
  })

  const deleteMut = useMutation({
    mutationFn: deleteJob,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['job-board'] }),
  })

  const closeModal = () => { setModalOpen(false); reset() }

  return (
    <>
      <Topbar title="Job Board" />
      <main className="page-body">

        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', gap: 8, flexWrap: 'wrap' }}>
          <input
            className="input"
            placeholder="🔍 Search jobs, companies, locations…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{ maxWidth: 320, fontSize: '.85rem' }}
          />
          {canPost && (
            <Button size="sm" variant="gold" onClick={() => setModalOpen(true)}>
              + Post Job
            </Button>
          )}
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading…</p></Card.Body></Card>
        )}

        {!isLoading && filtered.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                💼 No job postings found{search ? ` for "${search}"` : ''}.
                {canPost && !search && ' Be the first to post an opportunity!'}
              </p>
            </Card.Body>
          </Card>
        )}

        {filtered.map((j: JobPosting) => (
          <div key={j.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
            <div style={{ padding: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div style={{ flex: 1, minWidth: 0 }}>
                {/* Title + company row */}
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, flexWrap: 'wrap' }}>
                  <span style={{ fontWeight: 700, fontSize: '.96rem', color: 'var(--ink)' }}>{j.title}</span>
                  <span style={{ fontSize: '.82rem', color: 'var(--ink2)' }}>@ {j.company}</span>
                  {!j.active && (
                    <span style={{ fontSize: '.68rem', background: '#ffeaea', color: '#b00', borderRadius: 99, padding: '1px 8px', fontWeight: 600 }}>Closed</span>
                  )}
                </div>

                {/* Meta row */}
                <div style={{ display: 'flex', gap: 12, marginTop: 4, flexWrap: 'wrap', fontSize: '.75rem', color: 'var(--muted)' }}>
                  {j.location && <span>📍 {j.location}</span>}
                  <span>📅 {new Date(j.created_at).toLocaleDateString()}</span>
                </div>

                {/* Description */}
                <p style={{ fontSize: '.84rem', color: 'var(--ink2)', marginTop: 6, lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
                  {j.description}
                </p>

                {/* Apply link */}
                {j.url && (
                  <div style={{ marginTop: 8 }}>
                    <a
                      href={j.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{
                        display: 'inline-block',
                        fontSize: '.78rem',
                        fontWeight: 600,
                        color: 'var(--gold-dark, #7a5a00)',
                        background: 'var(--gold-pale)',
                        border: '1.5px solid var(--gold)',
                        borderRadius: 8,
                        padding: '4px 14px',
                        textDecoration: 'none',
                      }}
                    >
                      Apply →
                    </a>
                  </div>
                )}
              </div>

              {isAdmin && (
                <button
                  onClick={() => deleteMut.mutate(j.id)}
                  disabled={deleteMut.isPending}
                  title="Delete posting"
                  style={{ marginLeft: 12, background: 'none', border: 'none', cursor: 'pointer', color: 'var(--muted)', fontSize: '1rem', flexShrink: 0, padding: '2px 4px', borderRadius: 4 }}
                  onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--danger, #c0392b)')}
                  onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--muted)')}
                >
                  🗑️
                </button>
              )}
            </div>
          </div>
        ))}
      </main>

      {/* Post Job Modal */}
      <Modal isOpen={modalOpen} onClose={closeModal} title="Post a Job" size="md">
        <form
          onSubmit={handleSubmit((vals) => createMut.mutate(vals))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}
        >
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
              Job Title <span style={{ color: 'var(--danger, #c0392b)' }}>*</span>
            </label>
            <input
              className="input"
              placeholder="e.g. Software Engineering Intern"
              {...register('title', { required: 'Job title is required' })}
            />
            {errors.title && <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.title.message}</p>}
          </div>

          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
              Company <span style={{ color: 'var(--danger, #c0392b)' }}>*</span>
            </label>
            <input
              className="input"
              placeholder="e.g. Acme Corp"
              {...register('company', { required: 'Company is required' })}
            />
            {errors.company && <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.company.message}</p>}
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
            <div>
              <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Location</label>
              <input
                className="input"
                placeholder="e.g. Remote / Atlanta, GA"
                {...register('location')}
              />
            </div>
            <div>
              <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Apply URL</label>
              <input
                className="input"
                type="url"
                placeholder="https://..."
                {...register('url')}
              />
            </div>
          </div>

          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
              Description <span style={{ color: 'var(--danger, #c0392b)' }}>*</span>
            </label>
            <textarea
              className="input"
              rows={4}
              placeholder="Role overview, requirements, how to apply…"
              style={{ resize: 'vertical' }}
              {...register('description', { required: 'Description is required' })}
            />
            {errors.description && <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.description.message}</p>}
          </div>

          {createMut.isError && (
            <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.78rem' }}>Failed to post job. Please try again.</p>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" size="sm" onClick={closeModal}>Cancel</Button>
            <Button type="submit" variant="gold" size="sm" disabled={isSubmitting || createMut.isPending}>
              {createMut.isPending ? 'Posting…' : 'Post Job'}
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
