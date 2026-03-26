import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useAuth } from '../../hooks/useAuth'
import {
  getResources,
  createResource,
  deleteResource,
  type Resource,
} from '../../api/resources'

const CATEGORY_EMOJI: Record<string, string> = {
  document:    '📄',
  study:       '📚',
  leadership:  '👑',
  finance:     '💰',
  archive:     '🗄️',
  other:       '📎',
}

function categoryEmoji(cat?: string | null): string {
  if (!cat) return '📎'
  return CATEGORY_EMOJI[cat.toLowerCase()] ?? '📎'
}

interface CreateResourceForm {
  title: string
  url: string
  category: string
  description: string
}

export function ResourcesPage() {
  const { isAdmin, isChair } = useAuth()
  const canManage = isAdmin || isChair
  const [modalOpen, setModalOpen] = useState(false)
  const [activeCategory, setActiveCategory] = useState<string>('all')
  const qc = useQueryClient()

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['resources'],
    queryFn: getResources,
  })

  const categories = useMemo(() => {
    const cats = Array.from(
      new Set(items.map((r) => r.category ?? 'other').filter(Boolean))
    ).sort()
    return cats
  }, [items])

  const filtered = useMemo(() => {
    if (activeCategory === 'all') return items
    return items.filter((r) => (r.category ?? 'other') === activeCategory)
  }, [items, activeCategory])

  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } =
    useForm<CreateResourceForm>({ defaultValues: { category: 'other' } })

  const createMut = useMutation({
    mutationFn: (vals: CreateResourceForm) =>
      createResource({
        title: vals.title,
        url: vals.url,
        category: vals.category || undefined,
        description: vals.description || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['resources'] })
      setModalOpen(false)
      reset()
    },
  })

  const deleteMut = useMutation({
    mutationFn: deleteResource,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['resources'] }),
  })

  const closeModal = () => {
    setModalOpen(false)
    reset()
  }

  return (
    <>
      <Topbar title="Resources" />
      <main className="page-body">

        {/* Header row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: 8 }}>
          {/* Category filter pills */}
          <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
            {['all', ...categories].map((cat) => (
              <button
                key={cat}
                onClick={() => setActiveCategory(cat)}
                style={{
                  fontSize: '.72rem',
                  fontWeight: activeCategory === cat ? 700 : 400,
                  padding: '3px 12px',
                  borderRadius: 99,
                  border: `1.5px solid ${activeCategory === cat ? 'var(--gold)' : 'var(--cream3)'}`,
                  background: activeCategory === cat ? 'var(--gold-pale)' : 'var(--cream2)',
                  color: activeCategory === cat ? 'var(--gold-dark)' : 'var(--ink2)',
                  cursor: 'pointer',
                  transition: 'all .15s',
                  textTransform: 'capitalize',
                }}
              >
                {cat === 'all' ? 'All' : cat}
              </button>
            ))}
          </div>

          {canManage && (
            <Button size="sm" variant="gold" onClick={() => setModalOpen(true)}>
              + Add Resource
            </Button>
          )}
        </div>

        {/* Loading */}
        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading...</p></Card.Body></Card>
        )}

        {/* Empty state */}
        {!isLoading && filtered.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                📚 No resources found{activeCategory !== 'all' ? ` in "${activeCategory}"` : ''}.
                {canManage && ' Add the first one!'}
              </p>
            </Card.Body>
          </Card>
        )}

        {/* Resource list */}
        {filtered.map((r: Resource) => (
          <div key={r.id} className="card fade-in" style={{ marginBottom: '0.75rem' }}>
            <div style={{ padding: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div style={{ display: 'flex', gap: 14, alignItems: 'flex-start', flex: 1, minWidth: 0 }}>
                <span style={{ fontSize: '1.6rem', lineHeight: 1, flexShrink: 0 }}>
                  {categoryEmoji(r.category)}
                </span>
                <div style={{ minWidth: 0, flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                    <a
                      href={r.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ fontWeight: 600, fontSize: '.93rem', color: 'var(--ink)', textDecoration: 'none' }}
                      onMouseEnter={(e) => (e.currentTarget.style.textDecoration = 'underline')}
                      onMouseLeave={(e) => (e.currentTarget.style.textDecoration = 'none')}
                    >
                      {r.title}
                    </a>
                    {r.category && (
                      <span style={{ fontSize: '.68rem', background: 'var(--cream2)', borderRadius: 99, padding: '1px 8px', color: 'var(--ink2)', textTransform: 'capitalize', flexShrink: 0 }}>
                        {r.category}
                      </span>
                    )}
                  </div>
                  {r.description && (
                    <p style={{ fontSize: '.82rem', color: 'var(--ink2)', marginTop: 3, lineHeight: 1.55 }}>
                      {r.description}
                    </p>
                  )}
                  <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: 4, display: 'flex', gap: 10, flexWrap: 'wrap' }}>
                    <span>🔗 <a href={r.url} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--muted)' }}>{r.url.length > 50 ? r.url.slice(0, 50) + '…' : r.url}</a></span>
                    <span>📅 {new Date(r.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </div>

              {isAdmin && (
                <button
                  onClick={() => deleteMut.mutate(r.id)}
                  disabled={deleteMut.isPending}
                  title="Delete resource"
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

      {/* Add Resource Modal */}
      <Modal isOpen={modalOpen} onClose={closeModal} title="Add Resource" size="md">
        <form
          onSubmit={handleSubmit((vals) => createMut.mutate(vals))}
          style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '1rem' }}
        >
          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
              Title <span style={{ color: 'var(--danger, #c0392b)' }}>*</span>
            </label>
            <input
              className="input"
              placeholder="e.g. Parliamentary Procedure Guide"
              {...register('title', { required: 'Title is required', minLength: { value: 2, message: 'Min 2 characters' } })}
            />
            {errors.title && <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.title.message}</p>}
          </div>

          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>
              URL <span style={{ color: 'var(--danger, #c0392b)' }}>*</span>
            </label>
            <input
              className="input"
              type="url"
              placeholder="https://..."
              {...register('url', { required: 'URL is required' })}
            />
            {errors.url && <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.75rem', marginTop: 2 }}>{errors.url.message}</p>}
          </div>

          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Category</label>
            <select className="input" {...register('category')}>
              <option value="other">Other</option>
              <option value="document">Document</option>
              <option value="study">Study</option>
              <option value="leadership">Leadership</option>
              <option value="finance">Finance</option>
              <option value="archive">Archive</option>
            </select>
          </div>

          <div>
            <label style={{ fontSize: '.8rem', fontWeight: 600, display: 'block', marginBottom: 4 }}>Description</label>
            <textarea
              className="input"
              rows={3}
              placeholder="Brief description of this resource…"
              style={{ resize: 'vertical' }}
              {...register('description')}
            />
          </div>

          {createMut.isError && (
            <p style={{ color: 'var(--danger, #c0392b)', fontSize: '.78rem' }}>
              Failed to add resource. Please try again.
            </p>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <Button type="button" variant="ghost" size="sm" onClick={closeModal}>Cancel</Button>
            <Button type="submit" variant="gold" size="sm" disabled={isSubmitting || createMut.isPending}>
              {createMut.isPending ? 'Saving…' : 'Add Resource'}
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}
