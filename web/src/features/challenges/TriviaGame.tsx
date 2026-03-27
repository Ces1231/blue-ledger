import { useState, useEffect, useCallback } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '../../components/Toast'
import { submitAnswers, type Challenge } from '../../api/challenges'

// ── Types ──────────────────────────────────────────────────────────────────────

interface TriviaQuestion {
  question: string
  options: string[]
  answer: number
}

interface TriviaGameProps {
  challenge: Challenge
  myID: string
  onClose: () => void
}

// ── Helpers ────────────────────────────────────────────────────────────────────

const ANSWER_LABELS = ['A', 'B', 'C', 'D']
const TOTAL_TIME = 120 // seconds

function pad(n: number) {
  return String(n).padStart(2, '0')
}

// ── TriviaGame component ───────────────────────────────────────────────────────

export function TriviaGame({ challenge, myID, onClose }: TriviaGameProps) {
  const { showToast } = useToast()
  const qc = useQueryClient()

  // Extract questions from game_data
  const questions: TriviaQuestion[] = (
    (challenge.game_data?.questions as TriviaQuestion[]) ?? []
  )

  const [currentIdx, setCurrentIdx] = useState(0)
  const [selected, setSelected] = useState<(number | null)[]>(
    Array(questions.length).fill(null)
  )
  const [timeLeft, setTimeLeft] = useState(TOTAL_TIME)
  const [submitted, setSubmitted] = useState(false)
  const [result, setResult] = useState<{
    score: number
    total: number
    won: boolean
    xpChange: number
  } | null>(null)

  const currentQ = questions[currentIdx]
  const isLast = currentIdx === questions.length - 1
  const answered = selected[currentIdx] !== null

  // ── Countdown timer ──
  useEffect(() => {
    if (submitted) return
    if (timeLeft <= 0) {
      handleSubmit()
      return
    }
    const t = setTimeout(() => setTimeLeft(s => s - 1), 1000)
    return () => clearTimeout(t)
  }, [timeLeft, submitted])

  // ── Submit mutation ──
  const mutation = useMutation({
    mutationFn: (answers: number[]) => submitAnswers(challenge.id, answers),
    onSuccess: (updated) => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
      const isWinner = updated.winner_id === myID
      const correctCount = selected.filter((ans, i) => {
        const q = questions[i]
        return q && ans === q.answer
      }).length
      setResult({
        score: correctCount,
        total: questions.length,
        won: isWinner,
        xpChange: isWinner ? challenge.xp_stake : -challenge.xp_stake,
      })
    },
    onError: () => {
      showToast('Failed to submit answers.', 'error')
    },
  })

  const handleSubmit = useCallback(() => {
    if (submitted) return
    setSubmitted(true)
    // Fill any un-answered questions with -1 (wrong)
    const answers = selected.map(s => (s === null ? -1 : s))
    mutation.mutate(answers)
  }, [submitted, selected, mutation])

  const handleSelect = (optionIdx: number) => {
    if (submitted) return
    setSelected(prev => {
      const next = [...prev]
      next[currentIdx] = optionIdx
      return next
    })
  }

  const handleNext = () => {
    if (currentIdx < questions.length - 1) {
      setCurrentIdx(i => i + 1)
    }
  }

  const handlePrev = () => {
    if (currentIdx > 0) {
      setCurrentIdx(i => i - 1)
    }
  }

  // ── Empty state ──
  if (questions.length === 0) {
    return (
      <div className="modal-overlay" style={{ zIndex: 10000 }}>
        <div className="modal-box" style={{ maxWidth: '400px', textAlign: 'center' }}>
          <div style={{ fontSize: '2rem', marginBottom: '1rem' }}>⚠️</div>
          <div style={{ color: 'var(--muted)', marginBottom: '1.5rem' }}>
            No trivia questions found for this challenge.
          </div>
          <button className="btn btn-primary btn-sm" onClick={onClose}>Close</button>
        </div>
      </div>
    )
  }

  // ── Result screen ──
  if (result) {
    return (
      <div
        className="modal-overlay"
        style={{ zIndex: 10000, background: 'rgba(6,13,26,.97)' }}
      >
        <div
          style={{
            maxWidth: '420px',
            width: '95%',
            padding: '2.5rem 2rem',
            background: 'var(--surface)',
            borderRadius: '16px',
            border: `2px solid ${result.won ? '#f5c842' : 'var(--danger)'}`,
            textAlign: 'center',
            animation: 'fadeIn .3s ease',
          }}
        >
          <div style={{ fontSize: '3.5rem', marginBottom: '.75rem' }}>
            {result.won ? '🏆' : '💀'}
          </div>
          <div
            style={{
              fontSize: '1.6rem',
              fontWeight: 900,
              color: result.won ? 'var(--gold)' : 'var(--danger)',
              marginBottom: '.4rem',
              fontFamily: 'Space Grotesk, sans-serif',
            }}
          >
            {result.won ? 'VICTORY!' : 'DEFEAT'}
          </div>
          <div style={{ fontSize: '.9rem', color: 'var(--muted)', marginBottom: '1.5rem' }}>
            You answered <strong style={{ color: 'var(--ink)' }}>{result.score}/{result.total}</strong> correctly
          </div>

          {/* XP result */}
          <div
            style={{
              padding: '1rem',
              borderRadius: '10px',
              background: result.won ? 'rgba(245,200,66,.1)' : 'rgba(229,57,53,.1)',
              border: `1px solid ${result.won ? 'rgba(245,200,66,.3)' : 'rgba(229,57,53,.3)'}`,
              marginBottom: '1.5rem',
            }}
          >
            <div
              style={{
                fontFamily: 'DM Mono, monospace',
                fontSize: '1.5rem',
                fontWeight: 700,
                color: result.won ? 'var(--gold)' : 'var(--danger)',
              }}
            >
              {result.won ? '+' : ''}{result.xpChange} XP
            </div>
            <div style={{ fontSize: '.72rem', color: 'var(--muted)', marginTop: '4px' }}>
              {result.won
                ? 'XP stake awarded — glory is yours!'
                : 'XP deducted by chapter admin after review'}
            </div>
          </div>

          {/* Score breakdown */}
          <div style={{ marginBottom: '1.5rem' }}>
            {questions.map((q, i) => {
              const userAns = selected[i]
              const correct = userAns === q.answer
              return (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    padding: '6px 0',
                    borderBottom: '1px solid var(--border)',
                    textAlign: 'left',
                  }}
                >
                  <span style={{ fontSize: '1rem', flexShrink: 0 }}>{correct ? '✅' : '❌'}</span>
                  <span style={{ fontSize: '.76rem', color: 'var(--muted)', flex: 1, lineHeight: 1.3 }}>
                    {q.question}
                  </span>
                  {!correct && (
                    <span style={{ fontSize: '.7rem', color: 'var(--gold)', flexShrink: 0 }}>
                      {ANSWER_LABELS[q.answer]}
                    </span>
                  )}
                </div>
              )
            })}
          </div>

          <button className="btn btn-primary" onClick={onClose} style={{ width: '100%' }}>
            Done
          </button>
        </div>
      </div>
    )
  }

  // ── Loading (waiting for second player) ──
  if (mutation.isPending) {
    return (
      <div className="modal-overlay" style={{ zIndex: 10000, background: 'rgba(6,13,26,.95)' }}>
        <div style={{ textAlign: 'center', color: 'var(--ink)' }}>
          <div style={{ fontSize: '2rem', marginBottom: '1rem' }}>⚙️</div>
          <div style={{ fontSize: '.9rem', color: 'var(--muted)' }}>Submitting answers…</div>
        </div>
      </div>
    )
  }

  // ── Game screen ──
  const mins = Math.floor(timeLeft / 60)
  const secs = timeLeft % 60
  const timerDanger = timeLeft < 30
  const progress = ((currentIdx + 1) / questions.length) * 100

  return (
    <div
      className="modal-overlay"
      style={{ zIndex: 10000, background: 'rgba(6,13,26,.97)' }}
      onClick={(e) => e.stopPropagation()}
    >
      <div
        style={{
          width: '100%',
          maxWidth: '560px',
          margin: '0 auto',
          padding: '0 1rem',
          display: 'flex',
          flexDirection: 'column',
          height: '100%',
          maxHeight: '100dvh',
          justifyContent: 'center',
        }}
      >
        {/* Header bar */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '1rem',
            padding: '0 .25rem',
          }}
        >
          {/* Question counter */}
          <div
            style={{
              fontFamily: 'DM Mono, monospace',
              fontSize: '.8rem',
              color: 'var(--muted)',
            }}
          >
            {currentIdx + 1} / {questions.length}
          </div>

          {/* Title */}
          <div style={{ fontWeight: 700, fontSize: '.9rem', color: 'var(--gold)' }}>
            🧠 Trivia Battle
          </div>

          {/* Countdown */}
          <div
            style={{
              fontFamily: 'DM Mono, monospace',
              fontSize: '1rem',
              fontWeight: 700,
              color: timerDanger ? 'var(--danger)' : 'var(--gold)',
              animation: timerDanger ? 'pulse-green 1s infinite' : undefined,
            }}
          >
            {pad(mins)}:{pad(secs)}
          </div>
        </div>

        {/* Progress bar */}
        <div
          style={{
            height: '4px',
            background: 'rgba(255,255,255,.1)',
            borderRadius: '2px',
            marginBottom: '1.25rem',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              height: '100%',
              width: `${progress}%`,
              background: 'var(--gold)',
              borderRadius: '2px',
              transition: 'width .3s ease',
            }}
          />
        </div>

        {/* Question card */}
        <div
          style={{
            background: 'var(--surface)',
            borderRadius: '14px',
            border: '1px solid var(--border)',
            padding: '1.5rem 1.75rem',
            marginBottom: '1rem',
          }}
        >
          <div
            style={{
              fontSize: '1rem',
              fontWeight: 600,
              color: 'var(--ink)',
              lineHeight: 1.5,
              marginBottom: '1.25rem',
            }}
          >
            {currentQ.question}
          </div>

          {/* Answer options */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            {currentQ.options.map((opt, i) => {
              const isSelected = selected[currentIdx] === i
              return (
                <button
                  key={i}
                  onClick={() => handleSelect(i)}
                  disabled={submitted}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '12px 16px',
                    borderRadius: '10px',
                    border: `2px solid ${isSelected ? 'var(--gold)' : 'var(--border)'}`,
                    background: isSelected ? 'rgba(201,168,76,.15)' : 'rgba(255,255,255,.03)',
                    cursor: 'pointer',
                    textAlign: 'left',
                    transition: 'all .15s',
                  }}
                >
                  <span
                    style={{
                      width: '28px',
                      height: '28px',
                      borderRadius: '50%',
                      background: isSelected ? 'var(--gold)' : 'var(--border)',
                      color: isSelected ? '#060D1A' : 'var(--muted)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '.75rem',
                      fontWeight: 700,
                      flexShrink: 0,
                      transition: 'all .15s',
                    }}
                  >
                    {ANSWER_LABELS[i]}
                  </span>
                  <span style={{ fontSize: '.86rem', color: 'var(--ink)', lineHeight: 1.4 }}>
                    {opt}
                  </span>
                </button>
              )
            })}
          </div>
        </div>

        {/* Navigation */}
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'space-between' }}>
          <button
            className="btn btn-outline btn-sm"
            onClick={handlePrev}
            disabled={currentIdx === 0}
          >
            ← Prev
          </button>

          <div style={{ display: 'flex', gap: '6px' }}>
            {questions.map((_, i) => (
              <button
                key={i}
                onClick={() => setCurrentIdx(i)}
                style={{
                  width: '28px',
                  height: '28px',
                  borderRadius: '50%',
                  border: `2px solid ${i === currentIdx ? 'var(--gold)' : selected[i] !== null ? 'var(--neon-blue)' : 'var(--border)'}`,
                  background: i === currentIdx ? 'rgba(201,168,76,.2)' : selected[i] !== null ? 'rgba(74,158,255,.2)' : 'transparent',
                  cursor: 'pointer',
                  fontSize: '.7rem',
                  fontWeight: 700,
                  color: i === currentIdx ? 'var(--gold)' : selected[i] !== null ? 'var(--neon-blue)' : 'var(--muted)',
                  transition: 'all .15s',
                }}
              >
                {i + 1}
              </button>
            ))}
          </div>

          {isLast ? (
            <button
              className="btn btn-primary btn-sm"
              onClick={handleSubmit}
              disabled={submitted || mutation.isPending}
            >
              Submit ✓
            </button>
          ) : (
            <button
              className="btn btn-outline btn-sm"
              onClick={handleNext}
            >
              Next →
            </button>
          )}
        </div>

        {/* Submit early prompt */}
        {!isLast && selected.filter(s => s !== null).length === questions.length && (
          <div style={{ textAlign: 'center', marginTop: '1rem' }}>
            <button
              className="btn btn-primary btn-sm"
              onClick={handleSubmit}
              disabled={submitted}
            >
              All answered — Submit Now ✓
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
