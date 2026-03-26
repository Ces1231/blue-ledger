import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import { getGoals, createGoal, type Goal, type CreateGoalPayload } from '../../api/goals'

const CATEGORY_COLORS: Record<string, { bg: string; fg: string }> = {
  Academic:    { bg: '#F5E6C8', fg: '#7A4A00' },
  Service:     { bg: '#EAF5EE', fg: '#1A6B3A' },
  Financial:   { bg: '#FCE4EC', fg: '#8B1A1A' },
  Membership:  { bg: '#EAF0FB', fg: '#003087' },
  Leadership:  { bg: '#001A4D', fg: '#C9A84C' },
  Brotherhood: { bg: '#F2EDE4', fg: '#6B6657' },
  Other:       { bg: '#F2EDE4', fg: '#6B6657' },
}

function statusFromGoal(g: Goal): { label: string; bg: string; fg: string } {
  if (!g.is_active) return { label: 'Inactive', bg: 'var(--cream2)', fg: 'var(--muted)' }
  const pct = g.target_value > 0 ? g.current_value / g.target_value : 0
  if (pct >= 1) return { label: 'Complete', bg: 'var(--success-bg)', fg: 'var(--success)' }
  if (g.due_date && new Date(g.due_date) < new Date()) return { label: 'Overdue', bg: 'var(--danger-bg)', fg: 'var(--danger)' }
  return { label: 'In Progress', bg: 'var(--info-bg)', fg: 'var(--info)' }
}

const CATEGORIES = ['Academic', 'Service', 'Financial', 'Membership', 'Leadership', 'Brotherhood', 'Other']

const EMPTY_FORM: CreateGoalPayload = {
  title: '', description: '', category: 'Membership',
  target_value: 0, xp_reward: 0, due_date: '',
}

function GoalCard({ goal }: { goal: Goal }) {
  const colors = CATEGORY_COLORS[goal.category] ?? CATEGORY_COLORS.Other
  const status = statusFromGoal(goal)
  const pct = goal.target_value > 0
    ? Math.min(100, Math.round((goal.current_value / goal.target_value) * 100))
    : 0

  return (
    <div className="card fade-in" style={{ padding: '1.25rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8, marginBottom: '0.5rem' }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap', marginBottom: 4 }}>
            <span style={{ fontWeight: 700, fontSize: '.92rem', color: 'var(--ink)' }}>{goal.title}</span>
            <span
              style={{
                background: colors.bg, color: colors.fg,
                borderRadius: 99, padding: '1px 8px', fontSize: '.65rem', fontWeight: 600,
              }}
            >
              {goal.category}
            </span>
            <span
              style={{
                background: status.bg, color: status.fg,
                borderRadius: 99, padding: '1px 8px', fontSize: '.65rem', fontWeight: 600,
              }}
            >
              {status.label}
            </span>
          </div>
          {goal.description && (
            <p style={{ fontSize: '.78rem', color: 'var(--muted)', lineHeight: 1.4 }}>
              {goal.description}
            </p>
          )}
        </div>
        {goal.xp_reward > 0 && (
          <span
            style={{
              background: 'var(--gold-bg)', color: 'var(--warn)',
              borderRadius: 99, padding: '3px 10px', fontSize: '.7rem', fontWeight: 700, whiteSpace: 'nowrap',
            }}
          >
            +{goal.xp_reward} XP
          </span>
        )}
      </div>

      {/* Progress bar */}
      <div style={{ marginTop: '0.75rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.75rem', marginBottom: 6 }}>
          <span style={{ color: 'var(--muted)' }}>Progress</span>
          <span style={{ fontFamily: 'DM Mono, monospace', color: 'var(--ink2)' }}>
            {goal.current_value.toLocaleString()} / {goal.target_value.toLocaleString()} ({pct}%)
          </span>
        </div>
        <div style={{ height: 8, background: 'var(--cream2)', borderRadius: 99, overflow: 'hidden' }}>
          <div
            style={{
              height: '100%',
              width: `${pct}%`,
              background: pct >= 100 ? 'var(--success)' : 'var(--navy)',
              borderRadius: 99,
              transition: 'width 0.4s ease',
            }}
          />
        </div>
      </div>

      {goal.due_date && (
        <div style={{ marginTop: '0.5rem', fontSize: '.72rem', color: 'var(--faint)' }}>
          Due {new Date(goal.due_date).toLocaleDateString('en', { month: 'long', day: 'numeric', year: 'numeric' })}
        </div>
      )}
    </div>
  )
}

export function GoalsPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState<CreateGoalPayload>(EMPTY_FORM)

  const { data: goals = [], isLoading } = useQuery({
    queryKey: ['goals'],
    queryFn: getGoals,
  })

  const createMutation = useMutation({
    mutationFn: createGoal,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['goals'] })
      showToast('Chapter goal created.', 'success')
      setShowCreate(false)
      setForm(EMPTY_FORM)
    },
    onError: () => showToast('Failed to create goal.', 'error'),
  })

  const active = goals.filter((g) => g.is_active)
  const inactive = goals.filter((g) => !g.is_active)

  return (
    <>
      <Topbar title="Chapter Goals" />
      <main className="page-body">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            {active.length} active goal{active.length !== 1 ? 's' : ''}
          </p>
          {isAdmin && (
            <Button size="sm" variant="gold" onClick={() => setShowCreate(true)}>
              + Add Goal
            </Button>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading goals...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && goals.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No chapter goals set yet.
              </p>
            </Card.Body>
          </Card>
        )}

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
          {active.map((g) => <GoalCard key={g.id} goal={g} />)}
        </div>

        {inactive.length > 0 && (
          <>
            <h2 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.1rem', color: 'var(--ink2)', margin: '1.5rem 0 0.75rem' }}>
              Completed / Inactive
            </h2>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1rem' }}>
              {inactive.map((g) => <GoalCard key={g.id} goal={g} />)}
            </div>
          </>
        )}
      </main>

      <Modal isOpen={showCreate} onClose={() => setShowCreate(false)} title="Add Chapter Goal">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Title</label>
            <input
              className="form-input"
              value={form.title}
              onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
              placeholder="e.g. Achieve 90% dues collection"
            />
          </div>
          <div className="form-group">
            <label className="form-label">Description</label>
            <textarea
              className="form-input"
              rows={2}
              value={form.description ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              style={{ resize: 'vertical' }}
            />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">Category</label>
              <select
                className="form-input"
                value={form.category}
                onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
              >
                {CATEGORIES.map((c) => <option key={c} value={c}>{c}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">Target Value</label>
              <input
                className="form-input"
                type="number"
                value={form.target_value}
                onChange={(e) => setForm((f) => ({ ...f, target_value: Number(e.target.value) }))}
              />
            </div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div className="form-group">
              <label className="form-label">XP Reward</label>
              <input
                className="form-input"
                type="number"
                value={form.xp_reward ?? 0}
                onChange={(e) => setForm((f) => ({ ...f, xp_reward: Number(e.target.value) }))}
              />
            </div>
            <div className="form-group">
              <label className="form-label">Due Date</label>
              <input
                className="form-input"
                type="date"
                value={form.due_date ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, due_date: e.target.value }))}
              />
            </div>
          </div>
          <Button
            variant="gold"
            onClick={() => createMutation.mutate(form)}
            loading={createMutation.isPending}
            disabled={!form.title || form.target_value <= 0}
          >
            Create Goal
          </Button>
        </div>
      </Modal>
    </>
  )
}
