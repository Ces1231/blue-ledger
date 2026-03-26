import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { Button } from '../../components/Button'
import { useToast } from '../../components/Toast'
import apiClient from '../../api/client'
import type { Notification, NotificationType, APIResponse } from '../../types'

async function getNotifications(): Promise<Notification[]> {
  const { data } = await apiClient.get<APIResponse<Notification[]>>('/notifications')
  return data.data
}

async function markRead(id: string): Promise<void> {
  await apiClient.put(`/notifications/${id}/read`, {})
}

async function markAllRead(): Promise<void> {
  await apiClient.put('/notifications/read-all', {})
}

const TYPE_ICONS: Record<NotificationType, string> = {
  badge:        '🎖️',
  prop:         '👏',
  xp:           '⭐',
  dues:         '💰',
  event:        '📅',
  announcement: '📢',
  level:        '🏆',
  system:       '⚙️',
}

const TYPE_COLORS: Record<NotificationType, { bg: string; fg: string }> = {
  badge:        { bg: '#F5E6C8', fg: '#7A4A00' },
  prop:         { bg: '#EAF5EE', fg: '#1A6B3A' },
  xp:           { bg: '#FFF3DC', fg: '#7A4A00' },
  dues:         { bg: '#FCE4EC', fg: '#8B1A1A' },
  event:        { bg: '#EAF0FB', fg: '#003087' },
  announcement: { bg: '#F2EDE4', fg: '#6B6657' },
  level:        { bg: '#001A4D', fg: '#C9A84C' },
  system:       { bg: '#F2EDE4', fg: '#6B6657' },
}

function NotificationRow({
  notif,
  onRead,
}: {
  notif: Notification
  onRead: (id: string) => void
}) {
  const icon = TYPE_ICONS[notif.type] ?? '🔔'
  const colors = TYPE_COLORS[notif.type] ?? TYPE_COLORS.system

  return (
    <div
      onClick={() => { if (!notif.is_read) onRead(notif.id) }}
      style={{
        display: 'flex',
        gap: '0.75rem',
        alignItems: 'flex-start',
        padding: '0.9rem 1.25rem',
        borderBottom: '1px solid var(--border)',
        cursor: notif.is_read ? 'default' : 'pointer',
        background: notif.is_read ? 'transparent' : 'var(--gold-pale)',
        transition: 'background 0.15s',
      }}
    >
      <div
        style={{
          width: 38, height: 38, flexShrink: 0,
          background: colors.bg, color: colors.fg,
          borderRadius: '50%', display: 'flex', alignItems: 'center',
          justifyContent: 'center', fontSize: '1.1rem',
        }}
      >
        {icon}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
          <span style={{ fontWeight: notif.is_read ? 400 : 700, fontSize: '.88rem', color: 'var(--ink)' }}>
            {notif.title}
          </span>
          {!notif.is_read && (
            <span
              style={{
                width: 8, height: 8, borderRadius: '50%',
                background: 'var(--gold)', flexShrink: 0, marginTop: 4,
              }}
            />
          )}
        </div>
        <div style={{ fontSize: '.8rem', color: 'var(--muted)', marginTop: 2 }}>{notif.body}</div>
        <div style={{ fontSize: '.72rem', color: 'var(--faint)', marginTop: 4 }}>
          {new Date(notif.created_at).toLocaleDateString('en', {
            month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
          })}
        </div>
      </div>
    </div>
  )
}

export function NotificationsPage() {
  const qc = useQueryClient()
  const { showToast } = useToast()

  const { data: notifications = [], isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: getNotifications,
  })

  const readMutation = useMutation({
    mutationFn: markRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  })

  const readAllMutation = useMutation({
    mutationFn: markAllRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
      showToast('All notifications marked as read.', 'success')
    },
  })

  const unread = notifications.filter((n) => !n.is_read)
  const read = notifications.filter((n) => n.is_read)

  return (
    <>
      <Topbar title="Notifications" />
      <main className="page-body" style={{ maxWidth: 680 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div style={{ fontSize: '.85rem', color: 'var(--muted)' }}>
            {unread.length > 0 ? `${unread.length} unread` : 'All caught up'}
          </div>
          {unread.length > 0 && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => readAllMutation.mutate()}
              loading={readAllMutation.isPending}
            >
              Mark all read
            </Button>
          )}
        </div>

        {isLoading && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading notifications...</p>
            </Card.Body>
          </Card>
        )}

        {!isLoading && notifications.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                No notifications yet. We'll let you know when something happens!
              </p>
            </Card.Body>
          </Card>
        )}

        {unread.length > 0 && (
          <Card style={{ marginBottom: '1rem' }}>
            <Card.Header title={`Unread (${unread.length})`} />
            {unread.map((n) => (
              <NotificationRow
                key={n.id}
                notif={n}
                onRead={(id) => readMutation.mutate(id)}
              />
            ))}
          </Card>
        )}

        {read.length > 0 && (
          <Card>
            <Card.Header title="Read" />
            {read.map((n) => (
              <NotificationRow
                key={n.id}
                notif={n}
                onRead={(id) => readMutation.mutate(id)}
              />
            ))}
          </Card>
        )}
      </main>
    </>
  )
}
