import { useMutation } from '@tanstack/react-query'
import { login as loginAPI, requestMagicLink } from '../../api/auth'
import { useAuth } from '../../hooks/useAuth'
import { useToast } from '../../components/Toast'
import { useNavigate } from 'react-router-dom'

export function useLogin() {
  const { login } = useAuth()
  const { showToast } = useToast()
  const navigate = useNavigate()

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginAPI({ email, password }),
    onSuccess: (data) => {
      login(data.user, data.access_token, data.refresh_token)
      navigate('/dashboard')
    },
    onError: () => {
      showToast('Invalid email or password', 'error')
    },
  })

  const magicLinkMutation = useMutation({
    mutationFn: (email: string) => requestMagicLink(email),
    onSuccess: () => {
      showToast('Magic link sent! Check your email.', 'success')
    },
    onError: () => {
      // Still show success to prevent email enumeration
      showToast('If that email is registered, a magic link has been sent.', 'info')
    },
  })

  return {
    login: loginMutation.mutate,
    isLoggingIn: loginMutation.isPending,
    sendMagicLink: magicLinkMutation.mutate,
    isSendingMagicLink: magicLinkMutation.isPending,
  }
}
