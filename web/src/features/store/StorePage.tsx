import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useCurrentMember } from '../../hooks/useCurrentMember'
import {
  getStoreItems,
  purchaseItem,
  getOrders,
  type StoreItem,
} from '../../api/store'

const ITEM_ICONS = ['🎽', '📚', '🏆', '🎖️', '🎓', '🎯', '💼', '⭐']

function ItemCard({ item, onBuy }: { item: StoreItem; onBuy: () => void }) {
  const icon = ITEM_ICONS[Math.abs(item.name.charCodeAt(0)) % ITEM_ICONS.length]
  const outOfStock = item.quantity !== undefined && item.quantity !== null && item.quantity <= 0

  return (
    <div
      className="card fade-in"
      style={{
        padding: '1.25rem',
        display: 'flex', flexDirection: 'column', gap: '0.75rem',
        opacity: outOfStock ? 0.6 : 1,
      }}
    >
      <div style={{ fontSize: '2.5rem', textAlign: 'center' }}>{icon}</div>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontWeight: 700, fontSize: '.92rem', color: 'var(--ink)' }}>{item.name}</div>
        {item.description && (
          <p style={{ fontSize: '.78rem', color: 'var(--muted)', marginTop: 4, lineHeight: 1.4 }}>
            {item.description}
          </p>
        )}
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span
          style={{
            background: 'var(--gold-bg)', color: 'var(--warn)',
            borderRadius: 99, padding: '3px 10px', fontSize: '.75rem', fontWeight: 700,
          }}
        >
          {item.xp_cost.toLocaleString()} XP
        </span>
        {item.quantity !== undefined && item.quantity !== null && (
          <span style={{ fontSize: '.72rem', color: outOfStock ? 'var(--danger)' : 'var(--faint)' }}>
            {outOfStock ? 'Out of stock' : `${item.quantity} left`}
          </span>
        )}
      </div>
      <Button
        variant={outOfStock ? 'ghost' : 'gold'}
        size="sm"
        disabled={outOfStock}
        onClick={onBuy}
      >
        {outOfStock ? 'Unavailable' : 'Purchase'}
      </Button>
    </div>
  )
}

export function StorePage() {
  const qc = useQueryClient()
  const { showToast } = useToast()
  const { member } = useCurrentMember()
  const [tab, setTab] = useState<'store' | 'orders'>('store')
  const [confirmItem, setConfirmItem] = useState<StoreItem | null>(null)

  const { data: items = [], isLoading: itemsLoading } = useQuery({
    queryKey: ['store-items'],
    queryFn: getStoreItems,
  })

  const { data: ordersData, isLoading: ordersLoading } = useQuery({
    queryKey: ['store-orders'],
    queryFn: () => getOrders({ per_page: 50 }),
    enabled: tab === 'orders',
  })

  const purchaseMutation = useMutation({
    mutationFn: (itemId: string) => purchaseItem(itemId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['store-items'] })
      qc.invalidateQueries({ queryKey: ['store-orders'] })
      qc.invalidateQueries({ queryKey: ['member', 'current'] })
      showToast('Purchase successful! XP deducted.', 'success')
      setConfirmItem(null)
    },
    onError: () => showToast('Purchase failed. Check your XP balance.', 'error'),
  })

  const orders = ordersData?.data ?? []
  const activeItems = items.filter((i) => i.is_active)

  return (
    <>
      <Topbar title="XP Store" />
      <main className="page-body">
        {/* Tab + balance row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              size="sm"
              variant={tab === 'store' ? 'primary' : 'outline'}
              onClick={() => setTab('store')}
            >
              Browse Items
            </Button>
            <Button
              size="sm"
              variant={tab === 'orders' ? 'primary' : 'outline'}
              onClick={() => setTab('orders')}
            >
              My Orders
            </Button>
          </div>
          {member && (
            <div
              style={{
                background: 'var(--gold-bg)', borderRadius: 99,
                padding: '6px 14px', fontSize: '.82rem', fontWeight: 700, color: 'var(--warn)',
              }}
            >
              {member.xp_total.toLocaleString()} XP available
            </div>
          )}
        </div>

        {/* Store tab */}
        {tab === 'store' && (
          <>
            {itemsLoading && (
              <Card>
                <Card.Body>
                  <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading store...</p>
                </Card.Body>
              </Card>
            )}
            {!itemsLoading && activeItems.length === 0 && (
              <Card>
                <Card.Body>
                  <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                    No items in the store yet. Check back soon!
                  </p>
                </Card.Body>
              </Card>
            )}
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                gap: '1rem',
              }}
            >
              {activeItems.map((item) => (
                <ItemCard key={item.id} item={item} onBuy={() => setConfirmItem(item)} />
              ))}
            </div>
          </>
        )}

        {/* Orders tab */}
        {tab === 'orders' && (
          <>
            {ordersLoading && (
              <Card>
                <Card.Body>
                  <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading orders...</p>
                </Card.Body>
              </Card>
            )}
            {!ordersLoading && orders.length === 0 && (
              <Card>
                <Card.Body>
                  <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                    No orders yet. Browse the store and redeem your XP!
                  </p>
                </Card.Body>
              </Card>
            )}
            {orders.length > 0 && (
              <div className="table-wrap">
                <table className="table">
                  <thead>
                    <tr>
                      <th>Item</th>
                      <th>XP Spent</th>
                      <th>Status</th>
                      <th>Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {orders.map((o) => (
                      <tr key={o.id}>
                        <td style={{ fontWeight: 600 }}>{o.item_name ?? '—'}</td>
                        <td style={{ fontFamily: 'DM Mono, monospace', fontSize: '.82rem' }}>
                          {o.xp_spent.toLocaleString()} XP
                        </td>
                        <td>
                          <span
                            style={{
                              background: o.status === 'fulfilled' ? 'var(--success-bg)' : 'var(--warn-bg)',
                              color: o.status === 'fulfilled' ? 'var(--success)' : 'var(--warn)',
                              borderRadius: 99, padding: '2px 8px', fontSize: '.68rem', fontWeight: 600,
                              textTransform: 'capitalize',
                            }}
                          >
                            {o.status}
                          </span>
                        </td>
                        <td style={{ fontSize: '.78rem', color: 'var(--muted)', whiteSpace: 'nowrap' }}>
                          {new Date(o.created_at).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
      </main>

      {/* Confirm purchase modal */}
      <Modal isOpen={!!confirmItem} onClose={() => setConfirmItem(null)} title="Confirm Purchase" size="sm">
        {confirmItem && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <p style={{ fontSize: '.88rem', color: 'var(--ink)' }}>
              Purchase <strong>{confirmItem.name}</strong> for{' '}
              <strong style={{ color: 'var(--warn)' }}>{confirmItem.xp_cost.toLocaleString()} XP</strong>?
            </p>
            {member && member.xp_total < confirmItem.xp_cost && (
              <div
                style={{
                  background: 'var(--danger-bg)', border: '1px solid var(--danger-border)',
                  borderRadius: 8, padding: '10px 12px', fontSize: '.82rem', color: 'var(--danger)',
                }}
              >
                Insufficient XP. You have {member.xp_total.toLocaleString()} XP but need {confirmItem.xp_cost.toLocaleString()} XP.
              </div>
            )}
            <div style={{ display: 'flex', gap: 10 }}>
              <Button
                variant="gold"
                onClick={() => purchaseMutation.mutate(confirmItem.id)}
                loading={purchaseMutation.isPending}
                disabled={!!member && member.xp_total < confirmItem.xp_cost}
              >
                Confirm Purchase
              </Button>
              <Button variant="outline" onClick={() => setConfirmItem(null)}>
                Cancel
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  )
}
