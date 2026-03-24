import React from 'react'

interface CardProps {
  children: React.ReactNode
  className?: string
}

interface CardHeaderProps {
  title: string
  action?: React.ReactNode
}

export function Card({ children, className = '' }: CardProps) {
  return (
    <div className={`card ${className}`}>
      {children}
    </div>
  )
}

Card.Header = function CardHeader({ title, action }: CardHeaderProps) {
  return (
    <div className="card-header">
      <span className="card-title">{title}</span>
      {action && <div>{action}</div>}
    </div>
  )
}

Card.Body = function CardBody({ children, className = '' }: CardProps) {
  return (
    <div className={`card-body ${className}`}>
      {children}
    </div>
  )
}

interface StatCardProps {
  label: string
  value: React.ReactNode
  sub?: string
}

export function StatCard({ label, value, sub }: StatCardProps) {
  return (
    <div className="stat-card">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      {sub && <div className="stat-sub">{sub}</div>}
    </div>
  )
}
