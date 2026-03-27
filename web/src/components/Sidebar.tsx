import { NavLink, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '../hooks/useAuth'
import { useCurrentMember } from '../hooks/useCurrentMember'
import { useWebSocket } from '../hooks/useWebSocket'
import { listChallenges } from '../api/challenges'

interface NavItem {
  label: string
  path: string
  icon: string
  roles?: string[] // if undefined, all roles can see it
}

const NAV_SECTIONS = [
  {
    label: 'Main',
    items: [
      { label: 'Dashboard',    path: '/dashboard',   icon: '🏠' },
      { label: 'My Profile',   path: '/profile',     icon: '👤' },
      { label: 'Digital ID',   path: '/digital-id',  icon: '🪪' },
      { label: 'Leaderboard',  path: '/leaderboard', icon: '🏆' },
      { label: 'Notifications',path: '/notifications',icon: '🔔' },
    ],
  },
  {
    label: 'Chapter',
    items: [
      { label: 'Members',      path: '/members',      icon: '👥' },
      { label: 'Events',       path: '/events',       icon: '📅' },
      { label: 'Service',      path: '/service',      icon: '🤝' },
      { label: 'Dues',         path: '/dues',         icon: '💳' },
      { label: 'Announcements',path: '/announcements',icon: '📢' },
      { label: 'Props',        path: '/props',        icon: '👏' },
    ],
  },
  {
    label: 'Engage',
    items: [
      { label: 'Quests & Badges', path: '/quests',      icon: '⭐' },
      { label: 'Challenges',      path: '/challenges',  icon: '⚔️' },
      { label: 'Mentorship',      path: '/mentorship',  icon: '🎓' },
      { label: 'Voting',          path: '/votes',       icon: '🗳️' },
      { label: 'Minutes',         path: '/minutes',     icon: '📝' },
      { label: 'Store',           path: '/store',       icon: '🛍️' },
    ],
  },
  {
    label: 'Programs',
    items: [
      { label: 'Scholarships',  path: '/scholarships',icon: '🎓', roles: ['pia', 'admin', 'sysadmin'] },
      { label: 'Intake',        path: '/intake',      icon: '🚪', roles: ['chair', 'admin', 'sysadmin'] },
      { label: 'Fundraising',   path: '/fundraising', icon: '💰' },
      { label: 'Job Board',     path: '/job-board',   icon: '💼' },
      { label: 'Goals',         path: '/goals',       icon: '🎯' },
    ],
  },
  {
    label: 'Info',
    items: [
      { label: 'Alumni',        path: '/alumni',      icon: '🎩' },
      { label: 'History',       path: '/history',     icon: '📖' },
      { label: 'Resources',     path: '/resources',   icon: '📂' },
    ],
  },
  {
    label: 'AI',
    items: [
      { label: 'Assistant',     path: '/assistant',   icon: '✨' },
    ],
  },
  {
    label: 'Administration',
    items: [
      { label: 'Admin Panel',   path: '/admin',       icon: '⚙️',  roles: ['admin', 'sysadmin'] },
      { label: 'AI Config',     path: '/admin/ai-config', icon: '🤖', roles: ['admin', 'sysadmin'] },
      { label: 'PIA Reports',   path: '/pia',         icon: '📊',  roles: ['pia', 'admin', 'sysadmin'] },
      { label: 'System Console',path: '/sys',         icon: '🖥️',  roles: ['sysadmin'] },
    ],
  },
]

interface SidebarProps {
  isOpen: boolean
  onClose: () => void
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { user, role, isSysadmin, logout, memberID } = useAuth()
  const { member } = useCurrentMember()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { on } = useWebSocket()

  // Pending incoming challenge count for badge
  const { data: challenges = [] } = useQuery({
    queryKey: ['challenges'],
    queryFn: listChallenges,
    staleTime: 30_000,
    refetchInterval: 30_000,
  })
  const pendingCount = challenges.filter(
    (c) => c.challenged_id === memberID && c.status === 'pending'
  ).length

  // Real-time update when a new challenge invite arrives
  useEffect(() => {
    const unsub = on('CHALLENGE_INVITE', () => {
      qc.invalidateQueries({ queryKey: ['challenges'] })
    })
    return unsub
  }, [on, qc])

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  const canSee = (item: NavItem) => {
    if (!item.roles) return true
    if (isSysadmin) return true
    return item.roles.includes(role)
  }

  const initials = user
    ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    : '??'

  return (
    <nav className={`sidebar${isOpen ? ' open' : ''}`}>
      {/* Logo */}
      <div className="sidebar-logo">
        <div className="sidebar-logo-row">
          <div className="sidebar-shield">
            <span>ΦΒΣ</span>
          </div>
          <div>
            <div className="sidebar-brand">The Blue Ledger</div>
            <div className="sidebar-tagline">Chapter OS</div>
          </div>
        </div>
      </div>

      {/* User info */}
      {user && (
        <div className="sidebar-user">
          <div
            className="sidebar-user-avatar"
            style={{
              background: member?.avatar_bg ?? '#C9A84C',
              color: member?.avatar_fg ?? '#001A4D',
            }}
          >
            {initials}
          </div>
          <div>
            <div className="sidebar-user-name">
              {user.first_name} {user.last_name}
            </div>
            <div className="sidebar-user-role">
              {member?.member_display_id ?? role}
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <div className="sidebar-nav">
        {NAV_SECTIONS.map((section) => {
          const visibleItems = section.items.filter(canSee)
          if (visibleItems.length === 0) return null

          return (
            <div key={section.label}>
              <div className="nav-section-label">{section.label}</div>
              {visibleItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  className={({ isActive }) => `nav-item${isActive ? ' active' : ''}`}
                  onClick={onClose}
                >
                  <span className="nav-icon">{item.icon}</span>
                  {item.label}
                  {item.path === '/challenges' && pendingCount > 0 && (
                    <span
                      style={{
                        marginLeft: 'auto',
                        background: 'var(--danger, #e53935)',
                        color: '#fff',
                        borderRadius: '50%',
                        minWidth: '18px',
                        height: '18px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '.62rem',
                        fontWeight: 700,
                        padding: '0 3px',
                        lineHeight: 1,
                      }}
                    >
                      {pendingCount > 9 ? '9+' : pendingCount}
                    </span>
                  )}
                </NavLink>
              ))}
            </div>
          )
        })}
      </div>

      {/* Footer */}
      <div className="sidebar-footer">
        {member && (
          <div style={{ marginBottom: '8px' }}>
            <div className="xp-pill">{member.xp_total.toLocaleString()} XP</div>
          </div>
        )}
        <button className="signout-btn" onClick={handleLogout}>
          <span>⬅</span> Sign Out
        </button>
      </div>
    </nav>
  )
}
