import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'
import {
  getMentors,
  becomeMentor,
  requestMentorshipMatch,
  getMyMatch,
  type Mentor,
} from '../../api/mentorship'

function MentorCard({ mentor, onRequest }: { mentor: Mentor; onRequest: () => void }) {
  const initials = [mentor.first_name?.[0], mentor.last_name?.[0]].filter(Boolean).join('').toUpperCase()

  return (
    <div className="card fade-in" style={{ padding: '1.25rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
      <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
        <div
          style={{
            width: 48, height: 48, borderRadius: '50%', flexShrink: 0,
            background: 'var(--navy)', color: 'var(--gold)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: '1rem', fontWeight: 700,
          }}
        >
          {mentor.avatar_url ? (
            <img src={mentor.avatar_url} alt={initials} style={{ width: '100%', height: '100%', borderRadius: '50%', objectFit: 'cover' }} />
          ) : initials}
        </div>
        <div>
          <div style={{ fontWeight: 700, fontSize: '.92rem', color: 'var(--ink)' }}>
            Bro. {mentor.first_name} {mentor.last_name}
          </div>
          {mentor.role && (
            <div style={{ fontSize: '.75rem', color: 'var(--muted)', textTransform: 'capitalize' }}>
              {mentor.role}
            </div>
          )}
        </div>
      </div>

      {mentor.specialties.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
          {mentor.specialties.map((spec) => (
            <span
              key={spec}
              style={{
                background: 'var(--info-bg)', color: 'var(--info)',
                borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
              }}
            >
              {spec}
            </span>
          ))}
        </div>
      )}

      {mentor.bio && (
        <p style={{ fontSize: '.8rem', color: 'var(--muted)', lineHeight: 1.5 }}>
          {mentor.bio.length > 140 ? mentor.bio.slice(0, 140) + '...' : mentor.bio}
        </p>
      )}

      <Button size="sm" variant="gold" onClick={onRequest}>
        Request Mentor
      </Button>
    </div>
  )
}

export function MentorshipPage() {
  const { memberID } = useAuth()
  const qc = useQueryClient()
  const { showToast } = useToast()
  const [showBecomeModal, setShowBecomeModal] = useState(false)
  const [bio, setBio] = useState('')
  const [specialtyInput, setSpecialtyInput] = useState('')
  const [specialties, setSpecialties] = useState<string[]>([])

  const { data: mentors = [], isLoading } = useQuery({
    queryKey: ['mentors'],
    queryFn: getMentors,
  })

  const { data: myMatch } = useQuery({
    queryKey: ['my-match'],
    queryFn: getMyMatch,
  })

  const requestMutation = useMutation({
    mutationFn: (mentorId: string) => requestMentorshipMatch(mentorId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['my-match'] })
      showToast('Mentorship request sent!', 'success')
    },
    onError: () => showToast('Failed to send request.', 'error'),
  })

  const becomeMutation = useMutation({
    mutationFn: () => becomeMentor({ bio, specialties }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mentors'] })
      showToast('You are now registered as a mentor!', 'success')
      setShowBecomeModal(false)
    },
    onError: () => showToast('Failed to register as mentor.', 'error'),
  })

  const addSpecialty = () => {
    const trimmed = specialtyInput.trim()
    if (trimmed && !specialties.includes(trimmed)) {
      setSpecialties((s) => [...s, trimmed])
      setSpecialtyInput('')
    }
  }

  const isMentor = mentors.some((m) => m.member_id === memberID)
  const otherMentors = mentors.filter((m) => m.member_id !== memberID)

  return (
    <>
      <Topbar title="Mentorship" />
      <main className="page-body">
        {/* My match section */}
        {myMatch && (
          <Card style={{ marginBottom: '1.5rem', borderLeft: '4px solid var(--gold)' }}>
            <Card.Header title="My Mentorship Match" />
            <Card.Body>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: '.92rem' }}>
                      Mentor: Bro. {myMatch.mentor_first_name} {myMatch.mentor_last_name}
                    </div>
                    <div style={{ fontWeight: 700, fontSize: '.88rem', color: 'var(--muted)' }}>
                      Mentee: Bro. {myMatch.mentee_first_name} {myMatch.mentee_last_name}
                    </div>
                  </div>
                  <span
                    style={{
                      background: myMatch.status === 'active' ? 'var(--success-bg)' : 'var(--cream2)',
                      color: myMatch.status === 'active' ? 'var(--success)' : 'var(--muted)',
                      borderRadius: 99, padding: '2px 10px', fontSize: '.7rem', fontWeight: 600,
                      textTransform: 'capitalize',
                    }}
                  >
                    {myMatch.status}
                  </span>
                </div>
                {myMatch.started_at && (
                  <div style={{ fontSize: '.75rem', color: 'var(--faint)' }}>
                    Started {new Date(myMatch.started_at).toLocaleDateString('en', { month: 'long', year: 'numeric' })}
                  </div>
                )}
              </div>
            </Card.Body>
          </Card>
        )}

        {/* Header row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <p style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            {otherMentors.length} mentor{otherMentors.length !== 1 ? 's' : ''} available
          </p>
          {!isMentor && (
            <Button variant="outline" size="sm" onClick={() => setShowBecomeModal(true)}>
              Become a Mentor
            </Button>
          )}
          {isMentor && (
            <span
              style={{
                background: 'var(--success-bg)', color: 'var(--success)',
                borderRadius: 99, padding: '4px 12px', fontSize: '.75rem', fontWeight: 600,
              }}
            >
              You are a mentor
            </span>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading mentors...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && otherMentors.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No mentors available yet. Be the first to sign up!
              </p>
            </Card.Body>
          </Card>
        )}

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: '1rem',
          }}
        >
          {otherMentors.map((mentor) => (
            <MentorCard
              key={mentor.id}
              mentor={mentor}
              onRequest={() => requestMutation.mutate(mentor.id)}
            />
          ))}
        </div>
      </main>

      <Modal isOpen={showBecomeModal} onClose={() => setShowBecomeModal(false)} title="Become a Mentor">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div className="form-group">
            <label className="form-label">Bio</label>
            <textarea
              className="form-input"
              rows={4}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="Tell potential mentees about your background, expertise, and what you can offer..."
              style={{ resize: 'vertical' }}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Focus Areas</label>
            <div style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
              <input
                className="form-input"
                value={specialtyInput}
                onChange={(e) => setSpecialtyInput(e.target.value)}
                placeholder="e.g. Finance, Public Speaking"
                style={{ flex: 1 }}
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addSpecialty() } }}
              />
              <Button size="sm" variant="outline" onClick={addSpecialty}>Add</Button>
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
              {specialties.map((s) => (
                <span
                  key={s}
                  style={{
                    background: 'var(--info-bg)', color: 'var(--info)',
                    borderRadius: 99, padding: '2px 8px', fontSize: '.75rem',
                    display: 'flex', alignItems: 'center', gap: 4,
                  }}
                >
                  {s}
                  <button
                    onClick={() => setSpecialties((sp) => sp.filter((x) => x !== s))}
                    style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit', padding: 0 }}
                  >
                    ✕
                  </button>
                </span>
              ))}
            </div>
          </div>

          <Button
            variant="gold"
            onClick={() => becomeMutation.mutate()}
            loading={becomeMutation.isPending}
          >
            Register as Mentor
          </Button>
        </div>
      </Modal>
    </>
  )
}
