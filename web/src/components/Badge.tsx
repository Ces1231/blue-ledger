import React from 'react'

type BadgeVariant =
  | 'admin' | 'chair' | 'pia' | 'member'
  | 'gold' | 'silver' | 'bronze' | 'neo'
  | 'success' | 'warn' | 'danger'
  | string

interface BadgeProps {
  variant?: BadgeVariant
  children: React.ReactNode
  className?: string
}

const variantMap: Record<string, string> = {
  admin:   'badge-admin',
  chair:   'badge-chair',
  pia:     'badge-pia',
  member:  'badge-member',
  gold:    'badge-gold',
  silver:  'badge-silver',
  bronze:  'badge-bronze',
  neo:     'badge-neo',
  success: 'badge-success',
  warn:    'badge-warn',
  danger:  'badge-danger',
  icon:    'badge-gold',
}

export function Badge({ variant = 'member', children, className = '' }: BadgeProps) {
  const cls = variantMap[variant] ?? 'badge-member'
  return (
    <span className={`badge ${cls} ${className}`}>
      {children}
    </span>
  )
}

// Convenience component for XP level badges
export function LevelBadge({ levelKey, label }: { levelKey: string; label: string }) {
  return <Badge variant={levelKey}>{label}</Badge>
}

// Role badge with automatic variant mapping
export function RoleBadge({ role }: { role: string }) {
  return <Badge variant={role}>{role.charAt(0).toUpperCase() + role.slice(1)}</Badge>
}
