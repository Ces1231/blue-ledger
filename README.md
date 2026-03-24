# The Blue Ledger

**Chapter Engagement & Gamification Platform for Phi Beta Sigma**
Built on Glide + Google Sheets + Make.com · Fully documented · Ready to deploy

[![Screens](https://img.shields.io/badge/Screens-47-001A4D)](app/index.html)
[![Cost](https://img.shields.io/badge/Year_1_Cost-$0-1A6B3A)](#cost)
[![JS](https://img.shields.io/badge/JavaScript-Clean-C9A84C)](#)                            
[![License](https://img.shields.io/badge/License-MIT-blue)](#license)

---

## Table of Contents

- [What This Is](#what-this-is)
- [Live Demo](#live-demo)
- [Repository Structure](#repository-structure)
- [Quick Start](#quick-start)
- [Feature Overview](#feature-overview)
- [Tech Stack & Cost](#tech-stack--cost)
- [Master Sheet — 15 Tabs](#master-sheet--15-tabs)
- [12-Week Build Guide](#12-week-build-guide)
- [Year 1 Enhancements](#year-1-enhancements)
- [User Roles](#user-roles)
- [Point Economy](#point-economy)
- [Avatar Levels](#avatar-levels)
- [CES1231 System Admin](#ces1231-system-admin)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

---

## What This Is

The Blue Ledger is a mobile-first engagement platform that transforms chapter participation into a gamified "Quest." Every meeting attendance, service hour, dues payment, quiz pass, and peer recognition earns XP and levels up a brother's avatar — from Neophyte to Chapter Icon.

**Core loop:** Brother shows up → Chair scans QR code → XP awarded → Leaderboard updates → Avatar levels up → Chapter engaged.

---

## Live Demo

Open `app/index.html` in any browser. No server required. Sign in with any demo account:

| Account | Username | Role | Access |
|---|---|---|---|
| Marcus J. Williams | m.williams@chapter.org | Admin / President | Full access — all 47 screens |
| DeShawn A. Carter | d.carter@chapter.org | Chair | Scanner + member features |
| Elijah T. Brooks | e.brooks@chapter.org | PIA Analyst | Reports + analytics |
| Jordan M. Hayes | j.hayes@chapter.org | Member | Standard member experience |
| **CES1231** | ces1231@blueledger.sys | **System Admin** | **System Console — all 6 tabs** |

> **CES1231** has a separate dark-mode System Console with User Management, Integrations monitor, Data Tools, Audit Log, and App Config.

---

## Repository Structure

```
blue-ledger/
│
├── app/
│   └── index.html                        ← Full working app (47 screens, 290KB)
│
├── assets/
│   └── avatars/
│       ├── avatar-base.png               ← Neophyte    (0–499 XP)
│       ├── avatar-bronze.png             ← Bronze Varsity (500–999 XP)
│       ├── avatar-silver.png             ← Silver Elite (1,000–1,499 XP)
│       ├── avatar-gold.png               ← Gold Legend (1,500–2,499 XP)
│       └── avatar-icon.png               ← Chapter Icon (2,500+ XP)
│
├── data/
│   ├── MasterSheet.xlsx                  ← Google Sheets backend (15 tabs, ready to upload)
│   └── QuizBank.xlsx                     ← 40 quiz questions (History + Constitution)
│
├── docs/
│   └── MasterPlan.html                   ← Full 12-week execution plan
│
├── guides/
│   ├── Glide_Config_Guide.html           ← All 12 screens, computed columns, role matrix
│   ├── Week3_Onboarding.html             ← Splash → sign-in → profile → Digital ID → home
│   ├── Week4_Scanner_Roles.html          ← QR scanner, role-gating, 5-step action chain
│   ├── Week5_Financials_Resources.html   ← Dues view, admin panel, expense form
│   ├── Week6_Alpha_Milestone.html        ← Live event test, 7 alpha checks
│   ├── Week7_Quest_Engine.html           ← Side quest quizzes, scoring automation
│   ├── Week8_Avatar_Drip.html            ← Cloudinary upload, If-Then-Else, 4 screens
│   ├── Week9_Leaderboard_Badges.html     ← Leaderboard upgrade, all 10 badge automations
│   ├── Week10_PIA_Analytics.html         ← PIA screen, PDF export, Gmail automation
│   ├── Week11_Beta_Testing.html          ← E-Board stress test, bug fix reference
│   ├── Week12_Launch.html                ← Announcement template, demo runbook, adoption
│   └── Year1_Enhancements.html          ← 6 post-launch enhancements (RSVP, Stripe, Props…)
│
├── setup-github.sh                       ← Helper script to push to GitHub
├── README.md                             ← This file
└── .gitignore
```

---

## Quick Start

### Option A — Demo Right Now (30 seconds)
```bash
# Just open the file
open app/index.html        # macOS
start app/index.html       # Windows
xdg-open app/index.html    # Linux
```

### Option B — Deploy Online (2 minutes, free)
1. Go to [app.netlify.com/drop](https://app.netlify.com/drop)
2. Drag `app/index.html` onto the page
3. Share the live URL with your chapter

### Option C — Build the Real Glide App
1. Upload `data/MasterSheet.xlsx` to Google Drive
2. Create a free account at [glideapps.com](https://glideapps.com)
3. New App → connect your Google Sheet
4. Follow `guides/Glide_Config_Guide.html` — every screen is documented

---

## Feature Overview

### Member Features (all roles)
| Feature | Description |
|---|---|
| 🏠 Home Dashboard | Live XP, avatar, progress bar, pinned announcements, upcoming events |
| 🪪 Digital ID | Scannable QR code for event check-in |
| 🏆 Leaderboard | Chapter-wide ranking with avatar images, podium, semester filter |
| 🎖️ Quests & Badges | 10 badges with earned/locked states and progress tracking |
| 📚 Side Quest Quizzes | Fraternity History + Constitution quizzes (3 retakes/semester) |
| 🙌 Service Log | Log hours, view history, running totals |
| 👥 Chapter Directory | Searchable by name or employer, LinkedIn links |
| 👤 My Profile | Edit professional info, social links, photo |
| 📅 Events & RSVP | Browse events, RSVP, see headcount |
| 👏 Give Props | Peer recognition, +10 XP per prop received |
| 📢 Announcements | Chapter feed with pinned posts and categories |
| 💼 Job Board | Post and browse opportunities shared by brothers |
| 📖 Committees | Browse all committees, members, next meetings |
| ⭐ Milestones | Celebrate birthdays, new jobs, graduations |
| 🗓️ Chapter Calendar | Monthly grid with event dots, deadline markers |
| 🔔 Notifications | Unread inbox — badges, props, dues, events |
| 📋 Meeting Minutes | Searchable records with full modal view |
| 🎓 Alumni Network | Life Members and graduate brothers |
| 🗺️ Brother Network | City-grouped view of where brothers live |
| 🏛️ Chapter History | Visual timeline from founding |
| 🤝 Mentorship | Available mentors, request matching |
| 🛒 XP Store | Redeem XP for apparel, accessories, digital items |
| 📚 Study Groups | Create and join brother-led study sessions |
| 📁 Resources | Searchable document library |
| 🎂 Brother of the Month | Nominations and anonymous voting |
| 🎯 Chapter Goals | Semester OKRs with XP rewards |
| 💬 Direct Messages | Threaded brother-to-brother messaging |
| 📊 Activity Feed | Real-time chapter-wide timeline |
| ⚙️ Settings | Notification toggles, privacy, display preferences |

### Operational Features (Chair + Admin)
| Feature | Access |
|---|---|
| 📷 QR Scanner | Chair + Admin |
| 🙌 Service Verification | Chair + Admin |
| 📈 PIA Builder + Export | PIA + Admin |
| 💰 Financials (all members) | Admin only |
| 🎓 Scholarship Manager | Admin + PIA |
| 🔄 Intake Pipeline | Admin only |
| ⚙️ Admin Panel | Admin only |
| 📊 Chapter Health Dashboard | Admin + PIA |
| 📋 Check-In History | Chair + Admin |
| 🎤 Fundraising Tracker | Admin only |

### System Console (CES1231 only)
| Tab | Capabilities |
|---|---|
| Overview | Service health, Make.com ops budget, audit activity |
| User Management | Full CRUD — edit, deactivate, add members |
| Integrations | Glide, Sheets, Make.com, Cloudinary, Stripe status |
| Data Tools | Import/export, force sync, XP recalc, danger zone |
| Audit Log | Filterable system log, export |
| App Config | Chapter settings, point economy editor, webhooks |

---

## Tech Stack & Cost

| Tool | Purpose | Plan | Monthly Cost |
|---|---|---|---|
| [Glide](https://glideapps.com) | Mobile app UI, QR scanner, roles | Free | $0 |
| [Google Sheets](https://sheets.google.com) | Database backend (15 tabs) | Free | $0 |
| [Make.com](https://make.com) | Automations (XP, badges, email) | Free (1k ops/mo) | $0 |
| [Cloudinary](https://cloudinary.com) | Avatar image hosting | Free (25GB) | $0 |
| Zelle / PayPal | Dues payments | Existing | $0 |
| **Total Year 1** | | | **$0** |

**Upgrade triggers:**
- Glide Maker ($49/mo) — when roster exceeds 500 rows or need custom domain
- Make.com Core ($9/mo) — when automations exceed 1,000 ops/month
- Stripe (2.9% + 30¢/txn) — when adding in-app dues payment (Enhancement E3)

---

## Master Sheet — 15 Tabs

Upload `data/MasterSheet.xlsx` to Google Drive. It becomes a 15-tab Google Sheet.

| Tab | Purpose |
|---|---|
| 👥 Users | Members, XP, roles, dues, levels |
| ⚡ Point Economy | XP values for all activities |
| 📋 Engagement Log | Every XP event |
| 🏆 Leaderboard | Ranked member view |
| 🙌 Service Log | Service hours with verification |
| 🎖️ Quests & Badges | 10 badges with requirements |
| 📊 PIA Builder | Program assessment data |
| 🗺️ 12-Week Roadmap | Build timeline |
| 💸 Expense Reports | Google Form submissions |
| 📝 Quiz Sessions | Quiz attempt history |
| ⚙️ Config | App settings (semester, dues, district email) |
| 📅 Events | Event management + RSVP |
| 📅 RSVP Responses | Member RSVP history |
| 👏 Props | Peer recognition log |
| 📢 Announcements | Chapter feed posts |

---

## 12-Week Build Guide

| Week | Phase | Guide | Est. Time |
|---|---|---|---|
| 1–2 | Architecture & Data | MasterSheet.xlsx + QuizBank.xlsx | Done ✅ |
| 3 | Onboarding & Profile | Week3_Onboarding.html | 3–4 hrs |
| 4 | QR Scanner & Roles | Week4_Scanner_Roles.html | 4 hrs |
| 5 | Financials & Resources | Week5_Financials_Resources.html | 90 min |
| 6 | Alpha Milestone | Week6_Alpha_Milestone.html | Live test |
| 7 | Quest Engine | Week7_Quest_Engine.html | 3 hrs |
| 8 | Avatar & Drip | Week8_Avatar_Drip.html | 2 hrs |
| 9 | Leaderboard & Badges | Week9_Leaderboard_Badges.html | 2.5 hrs |
| 10 | PIA Builder | Week10_PIA_Analytics.html | 2.5 hrs |
| 11 | Beta Testing | Week11_Beta_Testing.html | Event |
| 12 | Full Launch | Week12_Launch.html | Launch day |

Open any guide in a browser — each has phone mockups, step-by-step Glide build instructions, Make.com automation flows, and a completion checklist.

---

## Year 1 Enhancements

Documented in `guides/Year1_Enhancements.html`:

| Enhancement | Month | Build Time |
|---|---|---|
| E2 — Dues Reminder Automation | Month 1 | 45 min |
| E5 — Chapter Announcements | Month 1 | 1 hr |
| E4 — Give Props | Month 2 | 1.5 hrs |
| E1 — Event RSVP System | Month 2 | 2.5 hrs |
| E3 — Stripe In-App Payment | Month 3 | 1.5 hrs |
| E6 — Multi-Chapter Expansion | Month 4–6 | 4–6 hrs |

---

## User Roles

| Role | Key Permissions |
|---|---|
| **Member** | Personal profile, Digital ID, leaderboard, quests, service log, directory, messaging |
| **Chair** | All member features + QR scanner, service verification |
| **PIA Analyst** | All member features + PIA Builder, scholarship manager, chapter health |
| **Admin** | Everything + financials, admin panel, role management, intake pipeline |
| **CES1231 (Sysadmin)** | System Console only — user CRUD, integrations, data ops, audit log, config |

---

## Point Economy

| Activity | XP | Notes |
|---|---|---|
| Chapter Meeting (on time) | 50 | Via QR scan |
| Chapter Meeting (late) | 25 | Via QR scan |
| Community Service (per hour) | 30 | Chair verified |
| Dues Paid (on time) | 50 | Auto via Make.com |
| Complete 100% Profile | 100 | One-time |
| Quiz Pass (≥80%) | 25–75 | Per correct answer |
| Committee Chair Role | 100 | Per semester |
| E-Board Officer | 200 | Per term |
| Props Received | 10 | Per prop |
| Study Group Join | 20 | Per group |
| Fundraising Contribution | 50 | Per donation |
| Mentor Match | 50 | One-time |

---

## Avatar Levels

| Level | XP Range | Unlocks |
|---|---|---|
| 🔵 Neophyte | 0–499 | Base avatar |
| 🥉 Bronze Varsity | 500–999 | Blue polo + bronze pin |
| 🥈 Silver Elite | 1,000–1,499 | Cap + silver pin |
| 🥇 Gold Legend | 1,500–2,499 | Blazer + gold pin |
| 👑 Chapter Icon | 2,500+ | Crown + legend frame |

Avatar PNGs are in `assets/avatars/`. Host on Cloudinary (free) and paste URLs into Glide's If-Then-Else computed column.

---

## CES1231 System Admin

A separate system-level account with a dark-mode console interface.

**Login:** Click the `CES1231 — System Administrator` button on the login screen.

**Capabilities:**
- **User Management** — add, edit, deactivate any member account
- **Integrations** — monitor Glide, Google Sheets, Make.com, Cloudinary, Stripe
- **Data Tools** — bulk import/export, force sync, XP recalculation, danger zone operations
- **Audit Log** — filterable system log; all CES1231 actions auto-logged
- **App Config** — chapter settings, point economy XP values, Make.com webhook URL

> In production: restrict `ces1231@blueledger.sys` to the Glide admin allowlist only. Never share this credential with chapter officers.

---

## Deployment

### GitHub → Netlify (Recommended for demo)
```bash
# 1. Push to GitHub
git remote add origin https://github.com/YOUR_USERNAME/blue-ledger.git
git branch -M main
git push -u origin main

# 2. Connect to Netlify
# netlify.com → New site → Import from GitHub → select repo
# Build command: (leave empty)
# Publish directory: app
# Done — live URL in 60 seconds
```

### Glide Production Build
```bash
# 1. Upload MasterSheet.xlsx to Google Drive
# 2. glideapps.com → New App → connect sheet
# 3. Follow guides/Glide_Config_Guide.html (all 12 screens)
# 4. Set theme: Primary #001A4D / Accent #C9A84C
# 5. Add member emails to sign-in allowlist
# 6. Publish → share URL with chapter
```

### Weekly Sync Workflow
```bash
# After each build week or feature addition:
git add .
git commit -m "Week X — [what was built]"
git push
```

---

## Contributing

This repo is maintained by the chapter's build lead. If you are building this for another chapter:

1. Fork the repo
2. Update `data/MasterSheet.xlsx` with your chapter's roster
3. Update chapter name, ID, and district email in the ⚙️ Config tab
4. Follow the 12-week guides
5. Open a pull request with your chapter-specific improvements

---

## License

MIT License — free to use, fork, and deploy for any Phi Beta Sigma chapter.

Attribution appreciated but not required.

---

*The Blue Ledger · Phi Beta Sigma Fraternity, Inc. · Tau Tau Sigma Chapter*
*Built with Claude · Version 1.0.0 · 47 Screens · 15 Sheet Tabs*
