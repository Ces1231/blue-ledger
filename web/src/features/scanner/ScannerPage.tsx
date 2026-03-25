import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useAuth } from '../../hooks/useAuth'
import { checkinQR, getQRToken, type CheckInResult } from '../../api/events'

export function ScannerPage() {
  const { isAdmin, isChair } = useAuth()
  const [eventId, setEventId] = useState('')
  const [manualId, setManualId] = useState('')
  const [qrToken, setQRToken] = useState<string | null>(null)
  const [lastResult, setLastResult] = useState<CheckInResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const tokenMutation = useMutation({
    mutationFn: () => getQRToken(eventId),
    onSuccess: (token) => setQRToken(token),
  })

  const checkinMutation = useMutation({
    mutationFn: (memberId: string) => checkinQR(eventId, {
      member_display_id: memberId,
      scanner_token: qrToken ?? '',
    }),
    onSuccess: (result) => {
      setLastResult(result)
      setManualId('')
      setError(null)
    },
    onError: () => setError('Check-in failed. Verify the member ID and try again.'),
  })

  if (!isAdmin && !isChair) {
    return (
      <>
        <Topbar title="Scanner" />
        <main className="page-body">
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--danger)' }}>Chair or admin access required to use the scanner.</p>
            </Card.Body>
          </Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="QR Check-in Scanner" />
      <main className="page-body" style={{ maxWidth: 540 }}>
        <Card>
          <Card.Header title="Event Check-in" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div className="form-group">
                <label className="form-label">Event ID</label>
                <div style={{ display: 'flex', gap: 8 }}>
                  <input
                    className="form-input"
                    value={eventId}
                    onChange={(e) => setEventId(e.target.value)}
                    placeholder="Paste event UUID"
                    style={{ flex: 1 }}
                  />
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => tokenMutation.mutate()}
                    loading={tokenMutation.isPending}
                    disabled={!eventId}
                  >
                    Load
                  </Button>
                </div>
              </div>

              {qrToken && (
                <>
                  <div
                    style={{
                      background: 'var(--success-bg)', border: '1px solid var(--success-border)',
                      borderRadius: 8, padding: '8px 12px', fontSize: '.8rem', color: 'var(--success)',
                    }}
                  >
                    Scanner active — token loaded
                  </div>

                  {/* Manual ID fallback */}
                  <div className="form-group">
                    <label className="form-label">Member Display ID (manual fallback)</label>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <input
                        className="form-input"
                        value={manualId}
                        onChange={(e) => setManualId(e.target.value.toUpperCase())}
                        placeholder="MBR-001"
                        style={{ flex: 1 }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' && manualId) checkinMutation.mutate(manualId)
                        }}
                      />
                      <Button
                        variant="gold"
                        onClick={() => checkinMutation.mutate(manualId)}
                        loading={checkinMutation.isPending}
                        disabled={!manualId}
                      >
                        Check In
                      </Button>
                    </div>
                    <p style={{ fontSize: '.73rem', color: 'var(--muted)', marginTop: 4 }}>
                      Scan a QR code or type the member's display ID and press Enter.
                    </p>
                  </div>

                  <p style={{ fontSize: '.78rem', color: 'var(--muted)' }}>
                    Note: Camera-based QR scanning requires the <code>react-qr-reader</code> package.
                    Using manual ID entry mode.
                  </p>
                </>
              )}

              {error && (
                <div style={{
                  background: 'var(--danger-bg)', border: '1px solid var(--danger-border)',
                  borderRadius: 8, padding: '8px 12px', color: 'var(--danger)', fontSize: '.85rem',
                }}>
                  {error}
                </div>
              )}
            </div>
          </Card.Body>
        </Card>

        {lastResult && (
          <Card style={{ marginTop: '1rem' }}>
            <Card.Header title="Check-in Successful" />
            <Card.Body>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <div style={{ fontSize: '1.1rem', fontWeight: 700, fontFamily: 'DM Serif Display, serif' }}>
                  {lastResult.member_name}
                </div>
                <div style={{ color: 'var(--success)', fontWeight: 600 }}>
                  +{lastResult.xp_awarded} XP awarded
                </div>
                <div style={{ fontSize: '.83rem', color: 'var(--muted)' }}>
                  Level: {lastResult.level}
                </div>
                {lastResult.new_badges.length > 0 && (
                  <div style={{ color: 'var(--gold)', fontSize: '.85rem', fontWeight: 600 }}>
                    New badge{lastResult.new_badges.length > 1 ? 's' : ''}: {lastResult.new_badges.join(', ')}
                  </div>
                )}
              </div>
            </Card.Body>
          </Card>
        )}
      </main>
    </>
  )
}
