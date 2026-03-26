import React from 'react'

type ButtonVariant = 'primary' | 'gold' | 'outline' | 'ghost' | 'danger'
type ButtonSize = 'default' | 'sm'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
  children: React.ReactNode
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: 'btn-primary',
  gold:    'btn-gold',
  outline: 'btn-outline',
  ghost:   'btn-ghost',
  danger:  'btn-danger',
}

const sizeClasses: Record<ButtonSize, string> = {
  default: '',
  sm:      'btn-sm',
}

export function Button({
  variant = 'primary',
  size = 'default',
  loading = false,
  children,
  className = '',
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`btn ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
      )}
      {children}
    </button>
  )
}
