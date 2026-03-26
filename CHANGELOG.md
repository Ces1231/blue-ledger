# Changelog

All notable changes to The Blue Ledger are documented here.

## [1.0.0] — 2025-03-20 — Initial Release

### Screens Added (47 total)
- Home Dashboard, Digital ID, Leaderboard, Quests & Badges, Side Quest Quizzes
- Service Log, Chapter Directory, My Profile
- Events & RSVP, Give Props, Announcements, Job Board, Committees, Milestones
- Chapter Calendar, Notifications, Meeting Minutes, Alumni Network, Brother Network
- Chapter History, Mentorship, XP Store, Study Groups, Resources
- Fundraising Tracker, Brother of the Month, Chapter Goals
- Direct Messages, Activity Feed, Settings
- QR Scanner (Chair+), Financials (Admin), PIA Builder (PIA+)
- Scholarship Manager (Admin+PIA), Intake Pipeline (Admin), Admin Panel
- Chapter Health Dashboard (Admin+PIA), Check-In History (Chair+)
- System Console — CES1231 (Sysadmin)

### Data Layers
- 7 member profiles with full XP, badges, service hours
- 19 activities in Point Economy
- 10 badges with requirements and XP bonuses
- 8 engagement log entries
- 4 events with RSVP responses
- 4 props, 4 announcements, 3 job postings
- 4 committees, 5 milestones, 4 alumni
- 10 social feed items, 4 nominations, 6 chapter goals
- 4 scholarship applications, 4 intake prospects
- 8 notification items, 2 meeting minutes records
- 3 message threads, 4 study groups, 8 resources
- 3 fundraising campaigns

### Tech
- Stack: Glide + Google Sheets + Make.com + Cloudinary
- Master Sheet: 15 tabs
- Quiz Bank: 40 questions (20 History + 20 Constitution)
- Avatar PNGs: 5 levels (400×400px, transparent background)
- Build guides: Weeks 3–12 + Year 1 enhancements
- System Admin: CES1231 with dark-mode console

### Known Limitations (Prototype)
- Data is in-memory — resets on page refresh
- QR scanner simulated (use Glide for real camera access)
- Make.com automations require setup (documented in guides)
- Stripe payment requires configuration (Enhancement E3)
