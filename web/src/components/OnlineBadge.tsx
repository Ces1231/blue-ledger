import { usePresence } from '../context/PresenceContext'

interface OnlineBadgeProps {
  memberID: string
  size?: 'sm' | 'md' | 'lg'
}

/**
 * Renders a green pulsing dot when the member is currently online.
 * Wrap a relative-positioned avatar container with this component as a sibling.
 *
 * Usage:
 *   <div style={{ position: 'relative', display: 'inline-block' }}>
 *     <img src={avatar} />
 *     <OnlineBadge memberID={member.id} />
 *   </div>
 */
export function OnlineBadge({ memberID, size = 'md' }: OnlineBadgeProps) {
  const { isOnline } = usePresence()
  if (!isOnline(memberID)) return null

  const sizeClass = size === 'sm' ? 'online-dot online-dot--sm'
    : size === 'lg' ? 'online-dot online-dot--lg'
    : 'online-dot'

  return <span className={sizeClass} aria-label="Online now" title="Online now" />
}
