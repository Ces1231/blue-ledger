# Blue Ledger User Guide

**Version:** 1.0.0  
**Last Updated:** March 30, 2026  
**Status:** Production Ready

---

## Table of Contents

- [Getting Started](#getting-started)
- [Logging In](#logging-in)
- [Dashboard](#dashboard)
- [Member Profile](#member-profile)
- [XP & Levels](#xp--levels)
- [Leaderboard](#leaderboard)
- [Events & RSVPs](#events--rsvps)
- [Quests & Badges](#quests--badges)
- [Service Log](#service-log)
- [Admin Features](#admin-features)
- [FAQ](#faq)

---

## Getting Started

### Access the Application

**Local Development:**
```
http://localhost:3001
```

**Production:**
```
https://blue-ledger.com
```

### Supported Browsers

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

### System Requirements

- Minimum 2GB RAM
- Stable internet connection
- Modern web browser

---

## Logging In

### Step 1: Navigate to Login

Go to the application URL and you'll see the login screen.

### Step 2: Enter Credentials

**Demo Account:**
```
Email: admin@tausigmasigma.org
Password: BlueLedger2026!
```

**Or use your chapter account credentials:**
```
Email: your-email@chapter.org
Password: Your password
```

### Step 3: Select Role

If you have multiple roles, select your role from the dropdown:
- **Member** - Standard chapter member access
- **Chair** - Officer with scanning privileges
- **Admin** - Full chapter management
- **Sysadmin** - System-level administration

### Step 4: Click Login

Your browser will store your session. You'll be redirected to the dashboard.

### Logout

Click your profile icon in the top-right corner → **Logout**

---

## Dashboard

### Overview

The dashboard is your personalized chapter hub with:

```
┌─────────────────────────────────────────────┐
│  Welcome, [Your Name]!                      │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Your XP  │  │ Level    │  │ Position │ │
│  │  2,450   │  │  Gold    │  │  #3      │ │
│  │ Legend   │  │ Legend   │  │ on Board │ │
│  └──────────┘  └──────────┘  └──────────┘ │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ Quick Actions                       │   │
│  │ ├─ Attend Meeting                  │   │
│  │ ├─ Log Service Hours               │   │
│  │ ├─ Take Quiz                       │   │
│  │ └─ Give Props                      │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ Recent Activity                     │   │
│  │ • You earned 50 XP for attendance  │   │
│  │ • You unlocked "Dedicated" badge   │   │
│  │ • James gave you 10 XP prop        │   │
│  └─────────────────────────────────────┘   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │ Upcoming Events                     │   │
│  │ • Chapter Meeting - Today 7:00 PM  │   │
│  │ • Service Day - Saturday 9:00 AM   │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### Your Stats

**XP (Experience Points)**
- Earned for attending meetings, completing quests, logging service
- Shows your engagement level
- Click to see detailed breakdown

**Level & Avatar**
- Based on total XP accumulated
- Shows progression path
- Visual indicator of rank

**Position on Leaderboard**
- Your rank among chapter members
- Updates in real-time
- Click to see full leaderboard

### Quick Actions

**Attend Meeting**
- Show QR code to chair or have them scan your ID
- Instantly earns 50 XP (on-time) or 25 XP (late)

**Log Service Hours**
- Enter service activity details
- Chair verifies your hours
- Earn 30 XP per hour (once verified)

**Take Quiz**
- Complete fraternity history or constitution quiz
- Earn 25-75 XP based on score
- Can retake once per semester

**Give Props**
- Recognize a fellow member for outstanding activity
- They earn 10 XP
- Shows on both profiles

---

## Member Profile

### View Your Profile

1. Click the profile icon (top-right)
2. Select "My Profile"

Or navigate to **Members** → Your name

### Profile Sections

```
┌────────────────────────────────────┐
│  [Avatar Image]                    │
│  John Doe                          │
│  Member | Joined March 2024        │
├────────────────────────────────────┤
│                                    │
│  Professional Information          │
│  ├─ Email: john@chapter.org        │
│  ├─ City: San Francisco            │
│  ├─ Job Title: Software Engineer   │
│  └─ LinkedIn: [Link]               │
│                                    │
│  Engagement Stats                  │
│  ├─ XP: 2,450                      │
│  ├─ Level: Gold Legend             │
│  ├─ Badges: 8/10 earned           │
│  ├─ Service Hours: 45              │
│  └─ Meetings Attended: 24/25       │
│                                    │
│  Social Links                      │
│  ├─ Twitter: @johndoe              │
│  ├─ Instagram: @johndoe            │
│  └─ Website: johndoe.com           │
│                                    │
│  [Edit Profile] [Share Profile]    │
└────────────────────────────────────┘
```

### Edit Profile

1. Click **Edit Profile** button
2. Update your information:
   - City
   - Job Title
   - Social media handles
   - Professional photo (optional)
3. Click **Save Changes**

### Share Profile

1. Click **Share Profile**
2. Choose:
   - Copy link to clipboard
   - Share on social media
   - Generate QR code

---

## XP & Levels

### XP System

XP (Experience Points) tracks your chapter engagement.

### XP Sources

| Activity | XP | Notes |
|---|---|---|
| Chapter Meeting (on-time) | 50 | Via chair QR scan |
| Chapter Meeting (late) | 25 | Via chair QR scan |
| Service Hour | 30 | Chair verified |
| Quiz Pass (≥80%) | 25-75 | Per correct answer |
| Prop Received | 10 | From fellow member |
| Study Group Join | 20 | Per group |
| Committee Chair | 100 | Per semester |
| E-Board Officer | 200 | Per term |

### Levels

Your XP determines your level and avatar:

```
🔵 Neophyte
├─ XP: 0 - 499
├─ Avatar: Base (Blue Circle)
└─ Status: New member

🥉 Bronze Varsity
├─ XP: 500 - 999
├─ Avatar: Blue Polo + Bronze Pin
└─ Status: Consistent engagement

🥈 Silver Elite
├─ XP: 1,000 - 1,499
├─ Avatar: Cap + Silver Pin
└─ Status: Strong engagement

🥇 Gold Legend
├─ XP: 1,500 - 2,499
├─ Avatar: Blazer + Gold Pin
└─ Status: Exceptional dedication

👑 Chapter Icon
├─ XP: 2,500+
├─ Avatar: Crown + Legend Frame
└─ Status: Elite chapter member
```

### View XP History

1. Click your avatar/name → **My Profile**
2. Click **XP History**
3. Filter by:
   - Date range
   - Activity type
   - Verified/pending

---

## Leaderboard

### Access Leaderboard

1. Main menu → **Leaderboard**
2. See all chapter members ranked by XP

### Features

**Ranking Display**
```
#1  🏆 Marcus Williams     2,850 XP  👑 Icon
#2  🥈 DeShawn Carter      2,640 XP  🥇 Gold
#3  🥉 Elijah Brooks       2,510 XP  🥇 Gold
#4     Jordan Hayes        1,890 XP  🥇 Gold
#5     Nicholas Patel      1,650 XP  🥈 Silver
...
```

**Filter Options**
- **Semester** - This semester or all time
- **Role** - Show all or specific roles
- **Level** - Filter by achievement level

**Search**
- Find a specific member by name
- See where they rank

**Timeline**
- Hover over names to see recent activity
- Shows engagement trend

---

## Events & RSVPs

### View Events

1. Main menu → **Events**
2. See list of upcoming chapter events

### Event Details

```
┌──────────────────────────────────┐
│ Chapter Meeting                  │
├──────────────────────────────────┤
│                                  │
│ Date: Friday, March 30, 2026     │
│ Time: 7:00 PM - 9:00 PM          │
│ Location: Student Center Room 201│
│ Attendees: 24/30                 │
│                                  │
│ Description:                     │
│ Regular chapter meeting with     │
│ announcements and elections.     │
│                                  │
│ [RSVP: Yes] [Maybe] [No]        │
│                                  │
│ Attendee List:                   │
│ ├─ Marcus Williams (confirmed)   │
│ ├─ DeShawn Carter (confirmed)    │
│ └─ [8 more...]                   │
│                                  │
└──────────────────────────────────┘
```

### RSVP to Event

1. Find the event
2. Click **RSVP: Yes**
3. Check mark appears next to your name
4. Chair can use for attendance tracking

### Modify RSVP

1. Click event again
2. Select **Change Response**
3. Choose new status:
   - **Yes** - I'll attend
   - **Maybe** - Still deciding
   - **No** - Can't make it

---

## Quests & Badges

### Badges

Badges represent achievements and milestones.

### Available Badges

```
📌 Founder
└─ For charter members

📚 Scholar
└─ Pass all quizzes

🤝 Community
└─ 50+ service hours

🎯 Focused
└─ 10+ consecutive meetings

💪 Iron Brother
└─ 2,500+ XP

🎖️ Officer
└─ Serve on E-Board

🌟 Mentor
└─ Mentor 3+ members

🏆 All-Star
└─ Top 3 leaderboard for 2 semesters

👑 Chapter Icon
└─ 2,500+ XP (special status)

🎓 Alumnus
└─ Graduated from chapter
```

### View Your Badges

1. Main menu → **Quests & Badges**
2. See earned badges (gold) and locked badges (gray)
3. Click badge to see requirements

### Unlock Badges

Badges unlock automatically when you meet requirements:

```
📍 Earned Badges
├─ Founder (Unlocked on joining)
├─ Scholar (Unlocked after quiz 5/5)
└─ Community (Unlocked at 50 hrs)

🔒 Locked Badges
├─ Iron Brother (Need 2,500 XP - at 1,850)
└─ All-Star (Top 3 for 2 semesters)
```

### Side Quests

One-time challenges with XP rewards:

```
⚡ Quest: Complete Profile
├─ Reward: 100 XP
├─ Status: In Progress
└─ Progress: 7/10 fields filled

⚡ Quest: Attend 5 Meetings
├─ Reward: 150 XP
├─ Status: In Progress
└─ Progress: 3/5 meetings attended

⚡ Quest: Give 3 Props
├─ Reward: 75 XP
├─ Status: Complete ✓
└─ Claimed: Yes
```

---

## Service Log

### Log Service Hours

1. Main menu → **Service Log**
2. Click **+ Log Service**
3. Fill in form:
   - **Activity**: Select from dropdown
   - **Hours**: Number of hours served
   - **Description**: Details of service
   - **Date**: When service occurred
4. Click **Submit for Verification**

### Service Activity Types

- Community Clean-up
- Food Bank
- Mentoring
- Tutoring
- Campus Mentorship
- Alumni Outreach
- Fundraising
- Other (describe)

### Service Status

```
Pending Review (⏳)
├─ Submitted but not yet verified
├─ Chair or admin will review
└─ You'll get notification when approved

Verified (✓)
├─ Approved by chair/admin
├─ XP credited to your account
└─ Counts toward badges

Rejected (✗)
├─ Didn't meet requirements
├─ See rejection reason
└─ Can resubmit with corrections
```

### View History

1. Click **Service History**
2. See all submitted and verified hours
3. Filter by:
   - Activity type
   - Date range
   - Status

### Service Hours Goals

Track progress toward semester goals:

```
Goal: 50 Hours This Semester

Progress: ████████░░ 40/50 hours

Breakdown:
├─ Community Service: 18 hours
├─ Mentoring: 15 hours
├─ Tutoring: 7 hours
└─ Other: 0 hours

Status: On Track! (10 hours to goal)
```

---

## Admin Features

### Admin Dashboard

Available if you have **Admin** or **Chair** role.

### Member Management

**View All Members**
1. Main menu → **Members**
2. See all chapter members
3. Sort by: XP, name, join date, status

**Edit Member**
1. Click member name
2. Update profile information
3. Click **Save**

**Deactivate Member**
1. Click member name
2. Click **Deactivate Member**
3. Member remains in history but inactive

**Reactivate Member**
1. Click member name
2. Click **Reactivate Member**
3. Member profile active again

### Check-In Verification

**For Chair:**
1. Main menu → **QR Scanner**
2. Point device camera at member's QR code
3. Automatically logs attendance
4. Member earns XP instantly

**For Admin:**
1. Main menu → **Check-In History**
2. View all attendance records
3. Manually adjust if needed

### Reports

1. Main menu → **Reports**
2. Select report type:
   - **Attendance Report** - Meeting attendance by member
   - **Service Hours** - Total hours logged
   - **XP Distribution** - XP earned by activity
   - **Badge Progress** - Badge unlock rates

---

## FAQ

### General Questions

**Q: What is XP?**  
A: XP (Experience Points) measure your engagement in the chapter. Earn XP for attending meetings, service, quizzes, and more. Higher XP = higher level and avatar.

**Q: How do I earn XP?**  
A: Attend meetings (50 XP), log service hours (30/hour), pass quizzes (25-75 XP), receive props (10 XP), and more.

**Q: What's the difference between levels?**  
A: Levels (Neophyte → Icon) are visual representations of your XP. Each level unlocks new badges and status.

**Q: Can I lose XP?**  
A: No, XP only increases. However, members can request corrections for logging errors.

### Profile Questions

**Q: How do I update my profile?**  
A: Click **My Profile** → **Edit Profile** → Update information → **Save**.

**Q: Can I change my avatar?**  
A: Avatars are automatically based on your XP level. You can't customize them, but your level advances as you earn XP.

**Q: How do I reset my password?**  
A: Click **Forgot Password** on login screen, enter your email, and follow the reset link.

### Events & Attendance

**Q: How do I RSVP to an event?**  
A: Go to **Events**, find the event, click **RSVP: Yes/Maybe/No**.

**Q: What happens if I don't RSVP?**  
A: You can still attend and get credit, but the chair knows your intent if you RSVP.

**Q: How does attendance get tracked?**  
A: Chair scans your QR code (in Digital ID) or manually marks attendance.

### Service & Badges

**Q: How do I log service hours?**  
A: Go to **Service Log** → **+ Log Service** → Fill form → Submit for verification.

**Q: Who verifies service hours?**  
A: Chair or admin reviews and approves service hours.

**Q: When do I unlock badges?**  
A: Badges unlock automatically when you meet the requirements.

**Q: Can I track progress toward badges?**  
A: Yes, go to **Quests & Badges** and see locked badges with progress bars.

### Technical Issues

**Q: The app is loading slowly. What do I do?**  
A: Clear your browser cache (Ctrl+Shift+Del) and refresh. Check your internet connection.

**Q: My login isn't working.**  
A: Verify your email and password are correct. Try **Forgot Password** if unsure.

**Q: I'm seeing a 500 error.**  
A: This is a server issue. Try refreshing. If it persists, contact your admin.

**Q: Can I access on mobile?**  
A: Yes! The app is fully responsive. Open in any mobile browser.

---

## Getting Help

### Contact Your Admin

For chapter-specific questions, contact:
- **Chapter President**: [Email]
- **Tech Lead**: [Email]
- **Service Chair**: [Email]

### Report Issues

Found a bug? Have a feature request?

1. Note what happened
2. Take a screenshot
3. Open an issue: https://github.com/Ces1231/blue-ledger/issues

### Documentation

- **Full API Docs**: See docs/API_DOCUMENTATION.md
- **Architecture**: See docs/ARCHITECTURE.md
- **Deployment**: See docs/DEPLOYMENT_GUIDE.md

---

## Support

For help:
- **Email**: admin@tausigmasigma.org
- **GitHub Issues**: https://github.com/Ces1231/blue-ledger/issues
- **Documentation**: https://github.com/Ces1231/blue-ledger#readme
