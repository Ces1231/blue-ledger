import { useCurrentMember } from '../hooks/useCurrentMember'

interface TopbarProps {
  title: string
}

export function Topbar({ title }: TopbarProps) {
  const { member } = useCurrentMember()

  return (
    <header className="topbar">
      <h1 className="topbar-title">{title}</h1>
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
