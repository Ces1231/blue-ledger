import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useCurrentMember } from '../hooks/useCurrentMember'

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
      { label: 'Quests & Badges', path: '/quests',    icon: '⭐' },
      { label: 'Mentorship',    path: '/mentorship',  icon: '🎓' },
      { label: 'Voting',        path: '/votes',       icon: '🗳️' },
      { label: 'Minutes',       path: '/minutes',     icon: '📝' },
      { label: 'Store',         path: '/store',       icon: '🛍️' },
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

export function Sidebar() {
  const { user, role, isSysadmin, logout } = useAuth()
  const { member } = useCurrentMember()
  const navigate = useNavigate()

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
    <nav className="sidebar">
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
                >
                  <span className="nav-icon">{item.icon}</span>
                  {item.label}
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
