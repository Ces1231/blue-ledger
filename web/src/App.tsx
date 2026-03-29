import { Routes, Route, Navigate, Outlet, useSearchParams, useNavigate } from 'react-router-dom'
import { Suspense, lazy, useEffect, useState } from 'react'
import { SidebarContext } from './context/SidebarContext'
import { PresenceProvider } from './context/PresenceContext'
import { useMutation } from '@tanstack/react-query'
import { Sidebar } from './components/Sidebar'
import { ProtectedRoute } from './components/ProtectedRoute'
import { consumeMagicLink } from './api/auth'
import { useAuth } from './hooks/useAuth'

// ── Auth ──
import { LoginPage } from './features/auth/LoginPage'

// ── Eager-loaded (above-the-fold) ──
import { DashboardPage } from './features/dashboard/DashboardPage'
import { DirectoryPage } from './features/members/DirectoryPage'
import { MemberProfilePage } from './features/members/MemberProfilePage'
import { LeaderboardPage } from './features/xp-leaderboard/LeaderboardPage'

// ── Lazy-loaded pages ──
const EventsPage            = lazy(() => import('./features/events/EventsPage').then(m => ({ default: m.EventsPage })))
const EventDetailPage       = lazy(() => import('./features/events/EventDetailPage').then(m => ({ default: m.EventDetailPage })))
const ScannerPage           = lazy(() => import('./features/events/ScannerPage').then(m => ({ default: m.ScannerPage })))
const ServicePage           = lazy(() => import('./features/service/ServicePage').then(m => ({ default: m.ServicePage })))
const DuesPage              = lazy(() => import('./features/dues/DuesPage').then(m => ({ default: m.DuesPage })))
const QuestsPage            = lazy(() => import('./features/quests/QuestsPage').then(m => ({ default: m.QuestsPage })))
const QuestDetailPage       = lazy(() => import('./features/quests/QuestDetailPage').then(m => ({ default: m.QuestDetailPage })))
const AnnouncementsPage     = lazy(() => import('./features/announcements/AnnouncementsPage').then(m => ({ default: m.AnnouncementsPage })))
const PropsPage             = lazy(() => import('./features/props/PropsPage').then(m => ({ default: m.PropsPage })))
const NotificationsPage     = lazy(() => import('./features/notifications/NotificationsPage').then(m => ({ default: m.NotificationsPage })))
const DigitalIDPage         = lazy(() => import('./features/digital-id/DigitalIDPage').then(m => ({ default: m.DigitalIDPage })))
const IntakePage            = lazy(() => import('./features/intake/IntakePage').then(m => ({ default: m.IntakePage })))
const VotesPage             = lazy(() => import('./features/votes/VotesPage').then(m => ({ default: m.VotesPage })))
const MentorshipPage        = lazy(() => import('./features/mentorship/MentorshipPage').then(m => ({ default: m.MentorshipPage })))
const MinutesPage           = lazy(() => import('./features/minutes/MinutesPage').then(m => ({ default: m.MinutesPage })))
const ScholarshipsPage      = lazy(() => import('./features/scholarships/ScholarshipsPage').then(m => ({ default: m.ScholarshipsPage })))
const FundraisingPage       = lazy(() => import('./features/fundraising/FundraisingPage').then(m => ({ default: m.FundraisingPage })))
const StorePage             = lazy(() => import('./features/store/StorePage').then(m => ({ default: m.StorePage })))
const ChapterGoalsPage      = lazy(() => import('./features/goals/ChapterGoalsPage').then(m => ({ default: m.ChapterGoalsPage })))
const JobBoardPage          = lazy(() => import('./features/job-board/JobBoardPage').then(m => ({ default: m.JobBoardPage })))
const CommitteesPage        = lazy(() => import('./features/committees/CommitteesPage').then(m => ({ default: m.CommitteesPage })))
const StudyGroupsPage       = lazy(() => import('./features/study-groups/StudyGroupsPage').then(m => ({ default: m.StudyGroupsPage })))
const MessagesPage          = lazy(() => import('./features/messages/MessagesPage').then(m => ({ default: m.MessagesPage })))
const ResourcesPage         = lazy(() => import('./features/resources/ResourcesPage').then(m => ({ default: m.ResourcesPage })))
const MilestonesPage        = lazy(() => import('./features/milestones/MilestonesPage').then(m => ({ default: m.MilestonesPage })))
const AlumniPage            = lazy(() => import('./features/alumni/AlumniPage').then(m => ({ default: m.AlumniPage })))
const ChallengesPage        = lazy(() => import('./features/challenges/ChallengesPage').then(m => ({ default: m.ChallengesPage })))
const AdminPage             = lazy(() => import('./features/admin/AdminPage').then(m => ({ default: m.AdminPage })))
const AdminMembersPage      = lazy(() => import('./features/admin/AdminMembersPage').then(m => ({ default: m.AdminMembersPage })))
const AdminEventsPage       = lazy(() => import('./features/admin/AdminEventsPage').then(m => ({ default: m.AdminEventsPage })))
const AdminDuesPage         = lazy(() => import('./features/admin/AdminDuesPage').then(m => ({ default: m.AdminDuesPage })))
const AdminBadgesPage       = lazy(() => import('./features/admin/AdminBadgesPage').then(m => ({ default: m.AdminBadgesPage })))
const AdminSettingsPage     = lazy(() => import('./features/admin/AdminSettingsPage').then(m => ({ default: m.AdminSettingsPage })))
const AdminBillingPage      = lazy(() => import('./features/admin/AdminBillingPage').then(m => ({ default: m.AdminBillingPage })))
const SysadminConsolePage   = lazy(() => import('./features/sysadmin/SysadminConsolePage').then(m => ({ default: m.SysadminConsolePage })))
const EngagementLogPage     = lazy(() => import('./features/admin/EngagementLogPage').then(m => ({ default: m.EngagementLogPage })))
const PointEconomyPage      = lazy(() => import('./features/admin/PointEconomyPage').then(m => ({ default: m.PointEconomyPage })))
const AIConfigPage          = lazy(() => import('./features/admin/AIConfigPage'))
const AssistantPage         = lazy(() => import('./features/ai/AssistantPage'))

// ── Page loading fallback ──
function PageLoader() {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '60vh', color: 'var(--muted)', fontSize: '.85rem' }}>
      Loading...
    </div>
  )
}

// ── Authenticated app shell (sidebar + outlet) ──
function AppShell() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const closeSidebar = () => setSidebarOpen(false)

  return (
    <SidebarContext.Provider value={{ toggleSidebar: () => setSidebarOpen(v => !v) }}>
      <PresenceProvider>
        <div className="app-game-theme" style={{ display: 'flex', minHeight: '100vh' }}>
          {/* Overlay — tapping closes sidebar on mobile */}
          <div
            className={`sidebar-overlay${sidebarOpen ? ' visible' : ''}`}
            onClick={closeSidebar}
            aria-hidden="true"
          />
          <Sidebar isOpen={sidebarOpen} onClose={closeSidebar} />
          <div className="main-content">
            <Suspense fallback={<PageLoader />}>
              <Outlet />
            </Suspense>
          </div>
        </div>
      </PresenceProvider>
    </SidebarContext.Provider>
  )
}

export function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/magic" element={<MagicLinkPage />} />

      {/* Protected app shell */}
      <Route
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/dashboard" replace />} />

        {/* ── Core ── */}
        <Route path="/dashboard"   element={<DashboardPage />} />
        <Route path="/profile"     element={<MemberProfileSelf />} />
        <Route path="/digital-id"  element={<DigitalIDPage />} />
        <Route path="/notifications" element={<NotificationsPage />} />

        {/* ── Members ── */}
        <Route path="/members"     element={<DirectoryPage />} />
        <Route path="/members/:id" element={<MemberProfilePage />} />

        {/* ── Events ── */}
        <Route path="/events"          element={<EventsPage />} />
        <Route path="/events/:id"      element={<EventDetailPage />} />
        <Route
          path="/scanner"
          element={
            <ProtectedRoute requiredRole="chair">
              <ScannerPage />
            </ProtectedRoute>
          }
        />

        {/* ── XP ── */}
        <Route path="/leaderboard"     element={<LeaderboardPage />} />

        {/* ── Engagement ── */}
        <Route path="/service"         element={<ServicePage />} />
        <Route path="/dues"            element={<DuesPage />} />
        <Route path="/quests"          element={<QuestsPage />} />
        <Route path="/quests/:id"      element={<QuestDetailPage />} />
        <Route path="/announcements"   element={<AnnouncementsPage />} />
        <Route path="/props"           element={<PropsPage />} />

        {/* ── Chapter ── */}
        <Route path="/votes"           element={<VotesPage />} />
        <Route path="/mentorship"      element={<MentorshipPage />} />
        <Route path="/minutes"         element={<MinutesPage />} />
        <Route path="/scholarships"    element={<ScholarshipsPage />} />
        <Route path="/fundraising"     element={<FundraisingPage />} />
        <Route path="/store"           element={<StorePage />} />
        <Route path="/goals"           element={<ChapterGoalsPage />} />
        <Route path="/job-board"       element={<JobBoardPage />} />
        <Route path="/committees"      element={<CommitteesPage />} />
        <Route path="/study-groups"    element={<StudyGroupsPage />} />
        <Route path="/messages"        element={<MessagesPage />} />
        <Route path="/resources"       element={<ResourcesPage />} />
        <Route path="/milestones"      element={<MilestonesPage />} />
        <Route path="/alumni"          element={<AlumniPage />} />
        <Route path="/assistant"       element={<AssistantPage />} />
        <Route path="/challenges"      element={<ChallengesPage />} />

        {/* ── Chair routes ── */}
        <Route
          path="/intake"
          element={
            <ProtectedRoute requiredRole="chair">
              <IntakePage />
            </ProtectedRoute>
          }
        />

        {/* ── Admin routes ── */}
        <Route
          path="/admin"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/members"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminMembersPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/events"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminEventsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/dues"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminDuesPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/badges"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminBadgesPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/engagement-log"
          element={
            <ProtectedRoute requiredRole="admin">
              <EngagementLogPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/point-economy"
          element={
            <ProtectedRoute requiredRole="admin">
              <PointEconomyPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/settings"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminSettingsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/ai-config"
          element={
            <ProtectedRoute requiredRole="admin">
              <AIConfigPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/admin/billing"
          element={
            <ProtectedRoute requiredRole="admin">
              <AdminBillingPage />
            </ProtectedRoute>
          }
        />

        {/* ── Sysadmin ── */}
        <Route
          path="/sys"
          element={
            <ProtectedRoute requiredRole="sysadmin">
              <SysadminConsolePage />
            </ProtectedRoute>
          }
        />
      </Route>

      {/* Catch-all */}
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

// ── Inline mini-pages for small utility routes ──

// Self-profile redirect using memberID from auth
function MemberProfileSelf() {
  const { memberID } = useAuth()
  if (!memberID) return <Navigate to="/dashboard" replace />
  return <Navigate to={`/members/${memberID}`} replace />
}

// Magic link consumption page
function MagicLinkPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const { login } = useAuth()

  const mutation = useMutation({
    mutationFn: () => consumeMagicLink(params.get('token') ?? ''),
    onSuccess: (data: { user: any; access_token: string; refresh_token: string }) => {
      login(data.user, data.access_token, data.refresh_token)
      navigate('/dashboard', { replace: true })
    },
    onError: () => {
      navigate('/login?error=magic_link_invalid', { replace: true })
    },
  })

  useEffect(() => {
    const token = params.get('token')
    if (token) {
      mutation.mutate()
    } else {
      navigate('/login', { replace: true })
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'var(--navy)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <div style={{ textAlign: 'center', color: '#fff' }}>
        <div style={{ fontFamily: 'DM Serif Display, serif', fontSize: '1.4rem', color: 'var(--gold)', marginBottom: '0.5rem' }}>
          Signing you in...
        </div>
        <div style={{ fontSize: '.78rem', color: 'rgba(255,255,255,.4)', fontFamily: 'DM Mono, monospace' }}>
          Verifying your magic link
        </div>
      </div>
    </div>
  )
}
