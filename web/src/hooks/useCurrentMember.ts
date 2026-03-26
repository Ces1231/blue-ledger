import { useQuery } from '@tanstack/react-query'
import { getMember } from '../api/members'
import { useAuth } from './useAuth'
import type { Member } from '../types'

/**
 * Fetches and caches the current authenticated user's member record.
 * Returns null while loading or if no user is authenticated.
 */
export function useCurrentMember() {
  const { memberID, isAuthenticated } = useAuth()

  const query = useQuery<Member, Error>({
    queryKey: ['member', 'current', memberID],
    queryFn: () => getMember(memberID),
    enabled: isAuthenticated && !!memberID,
    staleTime: 1000 * 60 * 5, // 5 minutes — member profiles don't change often
    retry: 1,
  })

  return {
    member: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  }
}
