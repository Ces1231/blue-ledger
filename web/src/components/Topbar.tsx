import { useCurrentMember } from '../hooks/useCurrentMember'
import { useSidebarToggle } from '../context/SidebarContext'

interface TopbarProps {
  title: string
}

export function Topbar({ title }: TopbarProps) {
  const { member } = useCurrentMember()
  const { toggleSidebar } = useSidebarToggle()

  return (
    <header className="topbar">
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
        <button
          className="mobile-menu-btn"
          onClick={toggleSidebar}
          aria-label="Open navigation menu"
        >
          ☰
        </button>
        <h1 className="topbar-title">{title}</h1>
      </div>
      <div className="topbar-right">
        {member && (
          <>
            <span className="xp-pill">{member.xp_total.toLocaleString()} XP</span>
            <span className={`level-pill badge-${member.level_key}`}>{member.level}</span>
          </>
        )}
      </div>
    </header>
  )
}
