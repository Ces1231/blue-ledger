import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import { useAuth } from '../../hooks/useAuth'

export function AdminBillingPage() {
  const { isAdmin } = useAuth()
  const { showToast } = useToast()

  if (!isAdmin) {
    return (
      <>
        <Topbar title="Billing" />
        <main className="page-body">
          <Card><Card.Body><p style={{ color: 'var(--danger)' }}>Admin access required.</p></Card.Body></Card>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar title="Billing" />
      <main className="page-body" style={{ maxWidth: 560 }}>
        <Card>
          <Card.Header title="Chapter Subscription" />
          <Card.Body>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <p style={{ fontSize: '.85rem', color: 'var(--muted)', lineHeight: 1.6 }}>
                Manage your chapter's Blue Ledger subscription, update payment methods, and view billing history through the Stripe customer portal.
              </p>
              <Button
                variant="gold"
                onClick={() => showToast('Stripe billing portal is available in the full release.', 'info')}
              >
                Open Billing Portal
              </Button>
            </div>
          </Card.Body>
        </Card>
      </main>
    </>
  )
}
