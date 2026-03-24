import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

interface ProtectedRouteProps {
  children: React.ReactNode
  requiredRole?: string[]
}

/**
 * Wraps routes that require authentication.
 * Redirects to /login if not authenticated.
 * Optionally enforces a minimum role.
 */
export function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { isAuthenticated, role, isSysadmin } = useAuth()
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (requiredRole && requiredRole.length > 0) {
    const hasRole = isSysadmin || requiredRole.includes(role)
    if (!hasRole) {
      return <Navigate to="/dashboard" replace />
    }
  }

  return <>{children}</>
}
