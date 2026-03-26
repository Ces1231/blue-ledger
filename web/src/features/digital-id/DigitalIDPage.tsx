import { useRef } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useCurrentMember } from '../../hooks/useCurrentMember'
import { useAuth } from '../../hooks/useAuth'

const XP_LEVELS = [
  { key: 'neo',    min: 0,    max: 249,   label: 'Neophyte' },
  { key: 'bronze', min: 250,  max: 749,   label: 'Bronze Varsity' },
  { key: 'silver', min: 750,  max: 1499,  label: 'Silver Elite' },
  { key: 'gold',   min: 1500, max: 2499,  label: 'Gold Legend' },
  { key: 'icon',   min: 2500, max: 99999, label: 'Chapter Icon' },
]

function getLevel(xp: number) {
  return XP_LEVELS.find((l) => xp >= l.min && xp <= l.max) ?? XP_LEVELS[0]
}

export function DigitalIDPage() {
  const { user } = useAuth()
  const { member, isLoading } = useCurrentMember()
  const { showToast } = useToast()
  const cardRef = useRef<HTMLDivElement>(null)

  if (isLoading) {
    return (
      <>
        <Topbar title="Digital ID" />
        <main className="page-body">
          <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading your ID...</p>
        </main>
      </>
    )
  }

  if (!member) return null

  const level = getLevel(member.xp_total)
  const qrValue = `BL:${member.member_display_id}:${member.first_name} ${member.last_name}`
  const inductedYear = member.inducted_year ?? new Date().getFullYear()
  const initials = `${member.first_name[0]}${member.last_name[0]}`.toUpperCase()

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'Blue Ledger Digital ID',
          text: `Bro. ${member.first_name} ${member.last_name} — ${member.member_display_id}`,
        })
      } catch {
        showToast('Share cancelled.', 'info')
      }
    } else {
      await navigator.clipboard.writeText(qrValue)
      showToast('ID data copied to clipboard.', 'success')
    }
  }

  const handleDownload = () => {
    const svg = cardRef.current?.querySelector('svg')
    if (!svg) return
    const blob = new Blob([svg.outerHTML], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `blue-ledger-id-${member.member_display_id}.svg`
    a.click()
    URL.revokeObjectURL(url)
    showToast('QR code downloaded.', 'success')
  }

  return (
    <>
      <Topbar title="Digital ID" />
      <main className="page-body" style={{ maxWidth: 420, margin: '0 auto' }}>
        {/* ID Card */}
        <div
          ref={cardRef}
          style={{
            background: 'var(--navy)',
            borderRadius: 20,
            overflow: 'hidden',
            boxShadow: '0 8px 32px rgba(0,26,77,0.35)',
            marginBottom: '1.5rem',
            position: 'relative',
          }}
        >
          {/* Card header */}
          <div
            style={{
              padding: '1.25rem 1.5rem 1rem',
              borderBottom: '1px solid rgba(201,168,76,0.25)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <div>
              <div
                style={{
                  fontFamily: 'DM Serif Display, serif',
                  fontSize: '1.1rem',
                  color: 'var(--gold)',
                  letterSpacing: '.04em',
                }}
              >
                The Blue Ledger
              </div>
              <div style={{ fontSize: '.68rem', color: 'rgba(255,255,255,.5)', letterSpacing: '.1em', textTransform: 'uppercase', marginTop: 2 }}>
                Member Identification
              </div>
            </div>
            <div
              style={{
                width: 32, height: 36,
                background: 'var(--gold)',
                clipPath: 'polygon(50% 0%, 100% 15%, 100% 60%, 50% 100%, 0% 60%, 0% 15%)',
              }}
            />
          </div>

          {/* Card body */}
          <div style={{ padding: '1.25rem 1.5rem', display: 'flex', gap: '1.25rem', alignItems: 'flex-start' }}>
            {/* Avatar */}
            <div
              style={{
                width: 64, height: 64, borderRadius: '50%', flexShrink: 0,
                background: member.avatar_bg ?? 'var(--gold)',
                color: member.avatar_fg ?? 'var(--navy)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '1.5rem', fontWeight: 700,
                border: '2px solid rgba(201,168,76,0.4)',
              }}
            >
              {initials}
            </div>

            {/* Member info */}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div
                style={{
                  fontFamily: 'DM Serif Display, serif',
                  fontSize: '1.15rem',
                  color: '#fff',
                  lineHeight: 1.2,
                }}
              >
                Bro. {member.first_name} {member.last_name}
              </div>
              <div
                style={{
                  fontFamily: 'DM Mono, monospace',
                  fontSize: '.75rem',
                  color: 'var(--gold)',
                  marginTop: 4,
                  letterSpacing: '.06em',
                }}
              >
                {member.member_display_id}
              </div>
              <div
                style={{
                  background: 'rgba(201,168,76,0.15)',
                  border: '1px solid rgba(201,168,76,0.3)',
                  borderRadius: 99,
                  padding: '2px 8px',
                  fontSize: '.65rem',
                  color: 'var(--gold)',
                  fontWeight: 600,
                  marginTop: 6,
                  display: 'inline-block',
                }}
              >
                {level.label}
              </div>
            </div>
          </div>

          {/* Stats row */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(4, 1fr)',
              borderTop: '1px solid rgba(201,168,76,0.2)',
              borderBottom: '1px solid rgba(201,168,76,0.2)',
            }}
          >
            {[
              { label: 'XP', value: member.xp_total.toLocaleString() },
              { label: 'Inducted', value: String(inductedYear) },
              { label: 'Level', value: level.key.toUpperCase() },
              { label: 'Role', value: member.role.toUpperCase() },
            ].map((stat, i) => (
              <div
                key={stat.label}
                style={{
                  padding: '0.7rem 0.5rem',
                  textAlign: 'center',
                  borderRight: i < 3 ? '1px solid rgba(201,168,76,0.15)' : 'none',
                }}
              >
                <div style={{ fontFamily: 'DM Mono, monospace', fontSize: '.8rem', color: '#fff', fontWeight: 600 }}>
                  {stat.value}
                </div>
                <div style={{ fontSize: '.6rem', color: 'rgba(255,255,255,.45)', textTransform: 'uppercase', letterSpacing: '.08em', marginTop: 2 }}>
                  {stat.label}
                </div>
              </div>
            ))}
          </div>

          {/* QR Code */}
          <div style={{ padding: '1.25rem', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
            <div
              style={{
                background: '#fff',
                borderRadius: 12,
                padding: 12,
                display: 'inline-flex',
              }}
            >
              <QRCodeSVG
                value={qrValue}
                size={140}
                fgColor="#001A4D"
                bgColor="#ffffff"
                level="M"
              />
            </div>
          </div>

          {/* Card footer */}
          <div style={{ padding: '0.75rem 1.5rem', borderTop: '1px solid rgba(201,168,76,0.15)', textAlign: 'center' }}>
            <div style={{ fontSize: '.65rem', color: 'rgba(255,255,255,.35)', fontFamily: 'DM Mono, monospace', letterSpacing: '.08em' }}>
              SCAN TO VERIFY MEMBERSHIP
            </div>
          </div>
        </div>

        {/* Action buttons */}
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="primary" style={{ flex: 1 }} onClick={handleDownload}>
            Download QR
          </Button>
          <Button variant="outline" style={{ flex: 1 }} onClick={handleShare}>
            Share
          </Button>
        </div>
      </main>
    </>
  )
}
