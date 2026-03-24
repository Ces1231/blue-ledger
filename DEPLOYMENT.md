# Deployment Guide

## Option 1 — Netlify Drop (Demo, 2 minutes)

1. Go to [app.netlify.com/drop](https://app.netlify.com/drop)
2. Drag `app/index.html` onto the page
3. Copy the live URL and share with your chapter

No account needed. Free forever for static files.

---

## Option 2 — GitHub + Netlify (Production demo)

```bash
# Clone or fork this repo
git clone https://github.com/YOUR_USERNAME/blue-ledger.git
cd blue-ledger

# Push to GitHub
git remote add origin https://github.com/YOUR_USERNAME/blue-ledger.git
git push -u origin main
```

Then:
1. [netlify.com](https://netlify.com) → New Site → Import from GitHub
2. Select `blue-ledger` repo
3. Build command: *(leave empty)*
4. Publish directory: `app`
5. Deploy — live in 60 seconds with auto-deploy on every git push

---

## Option 3 — Glide (Production app)

### Prerequisites
- Google account (free)
- Glide account (free tier at [glideapps.com](https://glideapps.com))
- Make.com account (free tier at [make.com](https://make.com))

### Step 1 — Upload Master Sheet
1. Go to [drive.google.com](https://drive.google.com)
2. Upload `data/MasterSheet.xlsx`
3. Open it — Google Drive converts it to a Google Sheet automatically
4. Confirm all 15 tabs are present

### Step 2 — Create Glide App
1. [glideapps.com](https://glideapps.com) → New App → App (not Pages)
2. Connect your Google Sheet
3. Name: `The Blue Ledger`
4. Primary color: `#001A4D` · Accent: `#C9A84C`

### Step 3 — Build Screens
Follow the guides in order:
1. `guides/Glide_Config_Guide.html` — setup and all 12 screens
2. `guides/Week3_Onboarding.html` — sign-in and profile flow
3. `guides/Week4_Scanner_Roles.html` — QR scanner and roles
4. Continue Week 5–12 at your own pace

### Step 4 — Configure Automations
1. [make.com](https://make.com) → New Scenario
2. Follow automation specs in each weekly guide
3. Connect to your Google Sheet via Make.com's Google Sheets module

### Step 5 — Upload Avatars
1. [cloudinary.com](https://cloudinary.com) → Sign Up (free)
2. Upload all 5 PNGs from `assets/avatars/`
3. Copy each Secure URL
4. In Glide Data Editor → Users table → Add If-Then-Else column `Avatar Image URL`
5. Paste the 5 URLs at the corresponding XP thresholds

### Step 6 — Launch
1. Add all member emails to Glide's sign-in allowlist
2. Glide → Publish → copy share URL
3. Follow `guides/Week12_Launch.html` for the chapter launch runbook

---

## Environment Notes

| File | Purpose | Where it runs |
|---|---|---|
| `app/index.html` | Full demo prototype | Any browser, Netlify |
| `data/MasterSheet.xlsx` | Database template | Google Drive → Sheets |
| `data/QuizBank.xlsx` | Quiz questions | Copy tabs into Master Sheet |
| `assets/avatars/*.png` | Avatar images | Cloudinary (free hosting) |
| `guides/*.html` | Build documentation | Any browser |

---

## CES1231 System Admin in Production

The `ces1231@blueledger.sys` account is a system-level operator. In the Glide production app:

1. Do **not** add `ces1231@blueledger.sys` to the member allowlist
2. This account is for the prototype demo only
3. In production: the system admin functions are handled by the chapter President (Admin role) through the Admin Panel
4. The System Console concept can be rebuilt as a separate Glide app with admin-only access if needed

---

## Updating the Repo

```bash
# After making changes locally:
git add .
git commit -m "describe what changed"
git push

# Netlify auto-deploys on every push to main
```
