import { useQuery } from '@tanstack/react-query'
import { getMembers } from '../../api/members'
import { getEvents } from '../../api/events'
import { getLeaderboard } from '../../api/xp'
import { getMyDues } from '../../api/dues'
import { useAuth } from '../../hooks/useAuth'
import { useCurrentMember } from '../../hooks/useCurrentMember'

export function useDashboard() {
  const { isAuthenticated } = useAuth()
  const { member } = useCurrentMember()

  const membersQuery = useQuery({
    queryKey: ['members', 'list'],
    queryFn: () => getMembers({ per_page: 100 }),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 2,
  })

  const eventsQuery = useQuery({
    queryKey: ['events', 'upcoming'],
    queryFn: () => getEvents({ upcoming: true, per_page: 5 }),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 5,
  })

  const leaderboardQuery = useQuery({
    queryKey: ['leaderboard', 'alltime'],
    queryFn: () => getLeaderboard('alltime'),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 2,
  })

  const duesQuery = useQuery({
    queryKey: ['dues', 'me'],
    queryFn: () => getMyDues(),
    enabled: isAuthenticated,
    staleTime: 1000 * 60 * 5,
  })

  const totalMembers = membersQuery.data?.meta.total ?? 0
  const upcomingEvents = eventsQuery.data?.data ?? []
  const leaderboard = leaderboardQuery.data ?? []
  const myDues = duesQuery.data ?? []

  const myRank = member
    ? leaderboard.findIndex((e) => e.member_id === member.id) + 1
    : 0

  const unpaidDues = myDues.filter((d) => d.status !== 'paid' && d.status !== 'waived')

  return {
    member,
    totalMembers,
    upcomingEvents,
    leaderboard: leaderboard.slice(0, 5),
    myRank,
    unpaidDues,
    isLoading:
      membersQuery.isLoading ||
      eventsQuery.isLoading ||
      leaderboardQuery.isLoading,
  }
}
