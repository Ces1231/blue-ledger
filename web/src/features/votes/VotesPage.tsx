import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  getVotes,
  respondToVote,
  getVoteResults,
  type Vote,
  type VoteResult,
} from '../../api/votes'

function VoteCard({ vote }: { vote: Vote }) {
  const qc = useQueryClient()
  const { showToast } = useToast()
  const { memberID } = useAuth()
  const hasVoted = !!vote.my_response

  const [selected, setSelected] = useState<string>(vote.my_response ?? '')

  const { data: results } = useQuery({
    queryKey: ['vote-results', vote.id],
    queryFn: () => getVoteResults(vote.id),
    enabled: hasVoted || !vote.is_open,
  })

  const voteMutation = useMutation({
    mutationFn: () => respondToVote(vote.id, selected),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['votes'] })
      qc.invalidateQueries({ queryKey: ['vote-results', vote.id] })
      showToast('Your vote has been recorded.', 'success')
    },
    onError: () => showToast('Failed to submit vote.', 'error'),
  })

  const closesAt = vote.closes_at ? new Date(vote.closes_at) : null
  const isClosed = !vote.is_open || (closesAt != null && closesAt < new Date())

  return (
    <div className="card fade-in" style={{ padding: '1.25rem', marginBottom: '1rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8, marginBottom: '0.5rem' }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: '.95rem', color: 'var(--ink)' }}>{vote.title}</div>
          {vote.description && (
            <p style={{ fontSize: '.8rem', color: 'var(--muted)', marginTop: 4, lineHeight: 1.5 }}>
              {vote.description}
            </p>
          )}
        </div>
        <span
          style={{
            background: isClosed ? 'var(--cream2)' : 'var(--success-bg)',
            color: isClosed ? 'var(--muted)' : 'var(--success)',
            borderRadius: 99, padding: '2px 10px', fontSize: '.68rem', fontWeight: 600, whiteSpace: 'nowrap',
          }}
        >
          {isClosed ? 'Closed' : 'Open'}
        </span>
      </div>

      {closesAt && !isClosed && (
        <div style={{ fontSize: '.72rem', color: 'var(--faint)', marginBottom: '0.75rem' }}>
          Closes {closesAt.toLocaleDateString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
        </div>
      )}

      {/* Results view */}
      {(hasVoted || isClosed) && results ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', marginTop: '0.75rem' }}>
          {vote.options.map((option) => {
            const count = results.options[option] ?? 0
            const pct = results.total > 0 ? Math.round((count / results.total) * 100) : 0
            const isMyVote = option === vote.my_response

            return (
              <div key={option}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '.8rem', marginBottom: 4 }}>
                  <span style={{ fontWeight: isMyVote ? 700 : 400, color: isMyVote ? 'var(--navy)' : 'var(--ink2)' }}>
                    {isMyVote ? '✓ ' : ''}{option}
                  </span>
                  <span style={{ color: 'var(--muted)', fontFamily: 'DM Mono, monospace', fontSize: '.72rem' }}>
                    {count} vote{count !== 1 ? 's' : ''} — {pct}%
                  </span>
                </div>
                <div style={{ height: 8, background: 'var(--cream2)', borderRadius: 99, overflow: 'hidden' }}>
                  <div
                    style={{
                      height: '100%',
                      width: `${pct}%`,
                      background: isMyVote ? 'var(--navy)' : 'var(--gold)',
                      borderRadius: 99,
                      transition: 'width 0.4s ease',
                    }}
                  />
                </div>
              </div>
            )
          })}
          <div style={{ fontSize: '.72rem', color: 'var(--faint)', marginTop: 4 }}>
            {results.total} total vote{results.total !== 1 ? 's' : ''}
          </div>
        </div>
      ) : (
        /* Voting form */
        !isClosed && (
          <div style={{ marginTop: '0.75rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '0.75rem' }}>
              {vote.options.map((option) => (
                <label
                  key={option}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 10,
                    padding: '10px 12px',
                    borderRadius: 8,
                    border: `1.5px solid ${selected === option ? 'var(--navy)' : 'var(--border)'}`,
                    background: selected === option ? 'var(--info-bg)' : 'var(--white)',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                  }}
                >
                  <input
                    type="radio"
                    name={`vote-${vote.id}`}
                    value={option}
                    checked={selected === option}
                    onChange={() => setSelected(option)}
                    style={{ accentColor: 'var(--navy)' }}
                  />
                  <span style={{ fontSize: '.85rem', color: 'var(--ink)', fontWeight: selected === option ? 600 : 400 }}>
                    {option}
                  </span>
                </label>
              ))}
            </div>
            <Button
              variant="gold"
              size="sm"
              onClick={() => voteMutation.mutate()}
              loading={voteMutation.isPending}
              disabled={!selected}
            >
              Submit Vote
            </Button>
          </div>
        )
      )}
    </div>
  )
}

export function VotesPage() {
  const { data: votes = [], isLoading } = useQuery({
    queryKey: ['votes'],
    queryFn: getVotes,
  })

  const openVotes = votes.filter((v) => v.is_open)
  const closedVotes = votes.filter((v) => !v.is_open)

  return (
    <>
      <Topbar title="Chapter Votes" />
      <main className="page-body" style={{ maxWidth: 680 }}>
        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading votes...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && votes.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No active votes right now.
              </p>
            </Card.Body>
          </Card>
        )}

        {openVotes.length > 0 && (
          <>
            <div style={{ marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: 8 }}>
              <h2 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.1rem', color: 'var(--navy)' }}>
                Active Votes
              </h2>
              <span
                style={{
                  background: 'var(--success-bg)', color: 'var(--success)',
                  borderRadius: 99, padding: '2px 8px', fontSize: '.7rem', fontWeight: 600,
                }}
              >
                {openVotes.length}
              </span>
            </div>
            {openVotes.map((v) => <VoteCard key={v.id} vote={v} />)}
          </>
        )}

        {closedVotes.length > 0 && (
          <>
            <div style={{ margin: '1.5rem 0 0.75rem', display: 'flex', alignItems: 'center', gap: 8 }}>
              <h2 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.1rem', color: 'var(--ink2)' }}>
                Past Votes
              </h2>
            </div>
            {closedVotes.map((v) => <VoteCard key={v.id} vote={v} />)}
          </>
        )}
      </main>
    </>
  )
}
