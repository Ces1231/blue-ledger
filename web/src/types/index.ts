// Blue Ledger — TypeScript types matching the API response shapes and DB schema

export interface Chapter {
  id: string
  name: string
  greek_letters: string
  city: string
  state_code: string
  university?: string
  district?: string
  charter_date?: string
  member_id_prefix: string
  semester_start?: string
  stripe_customer_id?: string
  subscription_status: 'trialing' | 'active' | 'past_due' | 'canceled' | 'paused'
  trial_ends_at?: string
  plan_tier: 'starter' | 'growth' | 'chapter_pro'
  settings: ChapterSettings
  created_at: string
  updated_at: string
}

export interface ChapterSettings {
  xp_levels?: XPLevel[]
  zeffy_form_id?: string
  notification_prefs?: Record<string, boolean>
  [key: string]: unknown
}

export interface XPLevel {
  key: string
  label: string
  min: number
}

export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  display_name?: string
  avatar_url?: string
  is_sysadmin: boolean
  created_at: string
}

export interface Member {
  id: string
  chapter_id: string
  user_id: string
  member_display_id: string
  display_id?: string
  first_name: string
  last_name: string
  display_name?: string
  email: string
  phone?: string
  bio?: string
  avatar_url?: string
  role: MemberRole
  status: MemberStatus
  inducted_year?: number
  employer?: string
  job_title?: string
  city?: string
  linkedin_url?: string
  xp_total: number
  xp_semester: number
  level: string
  level_key: string
  dues_status: DuesStatus
  avatar_bg?: string
  avatar_fg?: string
  service_hours_total?: number
  service_hours_semester?: number
  created_at: string
  updated_at: string
}

export type MemberRole = 'member' | 'chair' | 'pia' | 'admin' | 'sysadmin'
export type MemberStatus = 'active' | 'inactive' | 'alumni' | 'suspended' | 'pledging'
export type DuesStatus = 'paid' | 'unpaid' | 'late' | 'waived' | 'outstanding'

export interface Event {
  id: string
  chapter_id: string
  name: string
  description?: string
  event_date: string
  event_time?: string
  location?: string
  event_type: EventType
  xp_attend: number
  xp_rsvp: number
  rsvp_deadline?: string
  capacity?: number
  is_active: boolean
  qr_code_token?: string
  created_by?: string
  created_at: string
  updated_at: string
}

export type EventType = 'Meeting' | 'Service' | 'Social' | 'Conference' | 'Committee' | 'SBC' | 'Other'

export interface RSVP {
  id: string
  chapter_id: string
  event_id: string
  member_id: string
  status: 'yes' | 'no' | 'maybe'
  attended: boolean
  no_show: boolean
  checked_in_at?: string
  checked_in_by?: string
  created_at: string
}

export interface EngagementLog {
  id: string
  chapter_id?: string
  member_id: string
  first_name?: string
  last_name?: string
  activity: string
  xp_awarded: number
  source: XPSource
  reference_id?: string
  reference_type?: string
  semester?: string
  note?: string
  created_at: string
}

export type XPSource = 'checkin' | 'rsvp' | 'admin' | 'system' | 'quiz' | 'service' | 'dues' | 'badge' | 'props' | 'mentorship'

export interface DuesRecord {
  id: string
  chapter_id: string
  member_id: string
  semester: string
  amount_cents: number
  due_date: string
  paid_at?: string
  payment_method?: string
  zeffy_form_id?: string
  zeffy_transaction_id?: string
  status: DuesStatus
  xp_awarded: boolean
  created_at: string
  updated_at: string
}

export interface Badge {
  id: string
  chapter_id: string
  name: string
  icon?: string
  category?: string
  description?: string
  requirement?: string
  xp_reward: number
  rarity: 'common' | 'uncommon' | 'rare' | 'legendary'
  is_active: boolean
  criteria: BadgeCriteria
  created_at: string
}

export interface BadgeCriteria {
  type: string
  value?: number
}

export interface MemberBadge {
  id: string
  chapter_id: string
  member_id: string
  badge_id: string
  awarded_at: string
  awarded_by?: string
  badge?: Badge
}

export interface Announcement {
  id: string
  chapter_id: string
  title: string
  body: string
  category?: string
  posted_by: string
  is_pinned: boolean
  is_active: boolean
  created_at: string
  updated_at: string
  poster?: Pick<Member, 'first_name' | 'last_name'>
  poster_first_name?: string
  poster_last_name?: string
}

export interface Props {
  id: string
  chapter_id: string
  from_id: string
  to_id: string
  category: PropsCategory
  message?: string
  xp_awarded: number
  created_at: string
  from_member?: Pick<Member, 'first_name' | 'last_name' | 'avatar_url' | 'avatar_bg' | 'avatar_fg'>
  to_member?: Pick<Member, 'first_name' | 'last_name' | 'avatar_url' | 'avatar_bg' | 'avatar_fg'>
}

export type PropsCategory = 'Leadership' | 'Brotherhood' | 'Service' | 'Academic' | 'Professionalism' | 'Other'

export interface Notification {
  id: string
  chapter_id: string
  user_id: string
  type: NotificationType
  title: string
  body: string
  link?: string
  is_read: boolean
  created_at: string
}

export type NotificationType = 'badge' | 'prop' | 'xp' | 'dues' | 'event' | 'announcement' | 'level' | 'system'

export interface ServiceLogEntry {
  id: string
  chapter_id: string
  member_id: string
  member_first_name?: string
  member_last_name?: string
  event_name: string
  organization?: string
  service_date: string
  hours: number
  xp_awarded: number
  verified: boolean
  verified_by?: string
  verified_at?: string
  notes?: string
  created_at: string
}

export interface LeaderboardEntry {
  rank: number
  member_id: string
  member_display_id: string
  first_name: string
  last_name: string
  avatar_url?: string
  avatar_bg?: string
  avatar_fg?: string
  xp_total: number
  xp_semester: number
  level: string
  level_key: string
  role: MemberRole
}

export interface Quest {
  id: string
  chapter_id: string
  title: string
  description?: string
  xp_reward: number
  badge_reward_id?: string
  steps: QuestStep[]
  is_active: boolean
  created_at: string
  progress?: QuestProgress
}

export interface QuestStep {
  title: string
  description?: string
  target_count: number
}

export interface QuestProgress {
  [stepIndex: string]: number
}

// ── API Response envelopes ────────────────────────────────────────────────────

export interface APIResponse<T> {
  data: T
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: PaginationMeta
}

export interface PaginationMeta {
  total: number
  page: number
  per_page: number
  pages: number
}

export interface APIError {
  error: {
    code: string
    message: string
    status: number
  }
}

// ── Auth types ────────────────────────────────────────────────────────────────

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserSummary
}

export interface UserSummary {
  id: string
  email: string
  first_name: string
  last_name: string
  chapter_id: string
  member_id: string
  role: MemberRole | 'sysadmin'
  is_sysadmin: boolean
}
