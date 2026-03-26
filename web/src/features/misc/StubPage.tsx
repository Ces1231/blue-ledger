import { useNavigate } from 'react-router-dom'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'

interface StubPageProps {
  title: string
  icon: string
  description?: string
}

export function StubPage({ title, icon, description }: StubPageProps) {
  const navigate = useNavigate()

  return (
    <>
      <Topbar title={title} />
      <main className="page-body" style={{ maxWidth: 500 }}>
        <Card>
          <Card.Body>
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem' }}>
              <div style={{ fontSize: '3rem', marginBottom: '0.75rem' }}>{icon}</div>
              <h2 style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.4rem', color: 'var(--navy)', marginBottom: '0.5rem' }}>
                {title}
              </h2>
              <p style={{ fontSize: '.88rem', color: 'var(--muted)', lineHeight: 1.6, marginBottom: '1.5rem' }}>
                {description ?? 'Coming soon in the full release.'}
              </p>
              <Button variant="outline" size="sm" onClick={() => navigate(-1)}>
                Go Back
              </Button>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
