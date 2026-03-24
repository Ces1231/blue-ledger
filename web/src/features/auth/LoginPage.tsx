import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useLogin } from './useLogin'

interface LoginFormData {
  email: string
  password: string
}

export function LoginPage() {
  const [showMagicLink, setShowMagicLink] = useState(false)
  const [magicEmail, setMagicEmail] = useState('')
  const { login, isLoggingIn, sendMagicLink, isSendingMagicLink } = useLogin()

  const { register, handleSubmit, formState: { errors } } = useForm<LoginFormData>()

  const onSubmit = (data: LoginFormData) => {
    login({ email: data.email, password: data.password })
  }

  const onMagicLink = (e: React.FormEvent) => {
    e.preventDefault()
    sendMagicLink(magicEmail)
  }

  return (
    <div id="login-screen">
      <div className="login-card">
        {/* Shield logo */}
        <div className="login-shield">
          <svg width="100" height="110" viewBox="0 0 100 110">
            <defs>
              <clipPath id="shield">
                <polygon points="50,0 100,15 100,60 50,110 0,60 0,15" />
              </clipPath>
            </defs>
            <polygon points="50,0 100,15 100,60 50,110 0,60 0,15" fill="#C9A84C" opacity=".15" />
            <polygon points="50,0 100,15 100,60 50,110 0,60 0,15" fill="none" stroke="#C9A84C" strokeWidth="2" />
            <text
              x="50" y="68"
              textAnchor="middle"
              fontFamily="DM Serif Display, serif"
              fontSize="22"
              fill="#C9A84C"
            >ΦΒΣ</text>
          </svg>
        </div>

        <div className="login-title">The Blue Ledger</div>
        <div className="login-sub">Chapter Engagement Platform</div>

        {!showMagicLink ? (
          <form onSubmit={handleSubmit(onSubmit)}>
            <div>
              <label className="login-label">Email</label>
              <input
                className="login-input"
                type="email"
                placeholder="brother@chapter.org"
                autoComplete="email"
                {...register('email', { required: true })}
              />
              {errors.email && (
                <p style={{ color: '#F4C0D1', fontSize: '.72rem', marginTop: '-8px', marginBottom: '8px' }}>
                  Email is required
                </p>
              )}
            </div>

            <div>
              <label className="login-label">Password</label>
              <input
                className="login-input"
                type="password"
                placeholder="••••••••"
                autoComplete="current-password"
                {...register('password', { required: true })}
              />
              {errors.password && (
                <p style={{ color: '#F4C0D1', fontSize: '.72rem', marginTop: '-8px', marginBottom: '8px' }}>
                  Password is required
                </p>
              )}
            </div>

            <button
              type="submit"
              className="login-btn"
              disabled={isLoggingIn}
            >
              {isLoggingIn ? 'Signing in...' : 'Sign In'}
            </button>

            <div className="login-divider">OR</div>

            <button
              type="button"
              className="login-btn"
              style={{ background: 'rgba(255,255,255,.06)', color: 'rgba(255,255,255,.7)', border: '1px solid rgba(255,255,255,.12)' }}
              onClick={() => setShowMagicLink(true)}
            >
              Sign in with Magic Link
            </button>
          </form>
        ) : (
          <form onSubmit={onMagicLink}>
            <div>
              <label className="login-label">Your Email Address</label>
              <input
                className="login-input"
                type="email"
                placeholder="brother@chapter.org"
                value={magicEmail}
                onChange={(e) => setMagicEmail(e.target.value)}
                autoFocus
              />
            </div>

            <button
              type="submit"
              className="login-btn"
              disabled={isSendingMagicLink || !magicEmail}
            >
              {isSendingMagicLink ? 'Sending...' : 'Send Magic Link'}
            </button>

            <div className="login-divider">OR</div>

            <button
              type="button"
              className="login-btn"
              style={{ background: 'rgba(255,255,255,.06)', color: 'rgba(255,255,255,.7)', border: '1px solid rgba(255,255,255,.12)' }}
              onClick={() => setShowMagicLink(false)}
            >
              Back to Password Login
            </button>
          </form>
        )}

        <p style={{ textAlign: 'center', marginTop: '1.5rem', fontSize: '.68rem', color: 'rgba(255,255,255,.2)' }}>
          Phi Beta Sigma Fraternity, Inc. • Chapter Management Platform
        </p>
      </div>
    </div>
  )
}
