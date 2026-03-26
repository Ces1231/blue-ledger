# Vision Observability Report

Generated: 2026-03-24
Branch: main (no git history; full static codebase scan)
Language: HTML + vanilla JavaScript (single-file SPA)
Logging: None (no structured logging framework detected)
Metrics: None (no metrics framework detected)
Tracing: None (no distributed tracing detected)
Scan mode: Full

---

## Verdict: NOT READY for production as instrumented

### Summary

- Critical findings: 4
- High findings: 6
- Medium findings: 7
- Low findings: 5
- Info: 3

---

## Architectural Context

The Blue Ledger is a single-file static HTML/JavaScript SPA (290 KB, ~4,140 lines).
It runs entirely in the browser. The production architecture is:

  Browser (app/index.html) --> Glide (no-code app builder) --> Google Sheets (database)
                                                             --> Make.com (automation)
                                                             --> Cloudinary (image storage)

This architecture has zero server-side code owned by the team. All backend
services (Glide, Google Sheets, Make.com, Cloudinary) are third-party SaaS
platforms with their own observability dashboards.

This fundamentally shapes every finding below. "Logging" and "metrics" in this
context mean: client-side error capture, external monitoring integrations, and
the audit trail built into the prototype's in-memory data — not server logs.

---

## Logging Scorecard

| File | Functions | Log Stmts | Error Paths Logged | Context Quality | Grade |
|------|-----------|-----------|-------------------|-----------------|-------|
| app/index.html (login) | 3 | 0 | 0/3 | No logging at all | F |
| app/index.html (scanner/XP) | 4 | 0 | 0/4 | No logging at all | F |
| app/index.html (admin ops) | 8 | 0 (toast only) | 0/8 | No logging at all | F |
| app/index.html (sys console) | 6 | In-memory array only | Partial (3/6) | No timestamps | D |
| app/index.html (data mutations) | 12 | 0 | 0/12 | No logging at all | F |

Note: The System Console has a `SYS_AUDIT_LOG` array that records some admin
actions (member add, XP recalc, danger zone). This is the only logging-adjacent
mechanism in the entire codebase. It is in-memory only — it is destroyed on
page reload.

---

## Error Handling Scorecard

| Area | Errors Handled | Errors Swallowed | Silent Failures | Grade |
|------|:--------------:|:----------------:|:---------------:|:-----:|
| Login | 0 wrong / 0 errors | Yes (falls back to MEMBERS[0]) | 1 | D |
| QR Scanner | 0 | Yes (early return, no feedback) | 2 | D |
| Danger zone ops | 0 | Yes (toast shown but no logging) | 3 | D |
| Business mutations (XP, dues) | 0 | Yes (optional chaining hides failures) | 5 | F |
| Form submissions | Partial (required field checks) | Yes (silent on missing IDs) | 3 | C |

---

## Metrics Coverage

| Operation | Instrumented | Notes |
|-----------|:------------:|-------|
| Page navigation / screen views | No | No analytics framework |
| XP award events | No | State changes with no telemetry |
| QR scan completions | No | scanCount is in-memory only |
| Dues payment actions | No | No record beyond member object mutation |
| Login / sign-out | No | No session tracking |
| Error rate | No | No error boundary or reporting |
| Feature usage | No | No click tracking |
| Performance (Time to Interactive) | No | No performance observer |

---

## Health Check Status

| Dependency | Health Check Exists | Notes |
|------------|:-------------------:|-------|
| Glide platform | No | No programmatic check; must monitor Glide status page manually |
| Google Sheets (backend DB) | No | No connectivity verification |
| Make.com (automations) | Simulated | System Console shows mocked "connected" status; no real API call |
| Cloudinary (image hosting) | No | No check |
| Stripe (planned) | Simulated | System Console shows "not_configured"; no real probe |

---

## Timeout Coverage

No external HTTP calls exist in app/index.html. All data is in-memory mock data.
In the production Glide app, HTTP timeouts are managed by Glide's runtime — the
team has no control over timeout configuration. This is an architectural risk, not
a code-level gap, but it must be documented.

---

## Request ID Propagation

Not applicable to a static single-page app with no server backend.
In the Glide + Make.com production system, request tracing is provided by
Make.com's execution history. No propagation is possible from the browser.

---

## Findings (by severity)

### CRITICAL

**C1. No global error boundary — JavaScript exceptions crash silently**
- Location: app/index.html — entire application
- The application has no `window.onerror` handler, no `window.addEventListener('unhandledrejection')`, and no try/catch around any render function.
- Risk: Any runtime JavaScript error (null dereference, missing DOM element, type error) will silently halt the affected function. The user will see a broken screen with no message. The team will have no record of what failed, for whom, or how often.
- Evidence: Every render function (`renderDashboard`, `renderScanner`, `renderAdmin`, etc.) executes raw DOM manipulation with no error handling. `renderView()` at line 989 calls all 37 view functions with zero protection.
- Fix: Add `window.onerror` + `window.addEventListener('unhandledrejection')` with reporting to an error service (Sentry free tier handles this). Wrap `renderView()` in a try/catch that shows a graceful fallback screen.

**C2. Login accepts any email and silently falls back to admin account**
- Location: app/index.html line 867-871
- `doLogin()` does `const m = MEMBERS.find(x=>x.email===email)||MEMBERS[0]`. If the email is not found, it logs in as MEMBERS[0] — which is the Admin/President account.
- Risk: In a production deployment where the prototype is exposed to untrusted users, any failed login silently grants admin access. No error is shown, no log is written, no rate limit exists.
- Fix (production): This login flow must be replaced by Glide's sign-in allowlist before any real deployment. The README acknowledges this but the code does not guard against accidental exposure.

**C3. The in-memory audit log is destroyed on page reload**
- Location: app/index.html — `SYS_AUDIT_LOG` array and `renderSysLogs()`
- The System Console displays an audit log of admin actions (member created, XP recalculated, danger zone executions). This array lives only in JavaScript memory. Any page reload, tab close, or browser crash wipes the entire audit trail.
- Risk: The audit log is the only accountability mechanism for destructive operations (Clear Engagement Log, Reset All Dues, Lock App). If an action is taken and the page is refreshed, there is no record it happened.
- Fix: Persist `SYS_AUDIT_LOG` to `localStorage` at minimum. In production: route audit events to a Google Sheet tab or Make.com webhook so they survive session boundaries.

**C4. Danger zone operations execute without server-side verification**
- Location: app/index.html lines 3589-3601 (`confirmDangerAction`)
- "Clear Engagement Log", "Reset All Dues Statuses", and "Reset Semester Data" are executed entirely in-browser against in-memory data. The confirmation is a client-side CONFIRM text check.
- Risk: In a production Glide + Sheets deployment, the equivalent operations (clearing sheet tabs, mass-updating columns) would be triggered via Make.com webhooks. If this client-side pattern is replicated, a compromised or confused admin could wipe real data with no server-side authorization check, rate limit, or rollback mechanism.
- Fix: Any destructive operation in production must require a Make.com webhook call with server-side role verification, not a browser-side string comparison.

---

### HIGH

**H1. No error reporting service integrated**
- Location: app/index.html — entire codebase
- Zero calls to `console.error`, Sentry, LogRocket, Datadog, or any equivalent.
- Risk: When the app breaks in production for a member — wrong XP, broken scanner, failed form — the team has no visibility. They will learn about bugs from Discord messages, not alerts.
- Fix: Integrate Sentry's free browser SDK (one script tag). It captures unhandled exceptions automatically with user context, stack traces, and reproduction steps.

**H2. XP mutations have no audit trail or idempotency guard**
- Location: `confirmScan()` (line 1516), `verifyService()` (line 1562), `awardXP()` (line 1715), `rsvpEvent()` (line 1838)
- XP is awarded by directly mutating `m.xp += xp`. There is no check for duplicate awards, no log entry written for most operations, and no way to determine post-hoc whether a member's XP total is correct.
- Risk: If a chair double-scans a member's QR code, XP is doubled with no warning, no log, and no way to reverse it except the admin `sysRecalcXP()` function (which re-derives from `engLog` — but `engLog` is also in-memory and not written to for scanner events).
- Evidence: `confirmScan()` at line 1519 does `m.xp += xp` and `updateLevel(m)` but never writes to `engLog`. The engagement log (which is the source of truth for XP recalculation) is never updated by the scanner.
- Fix: Every XP mutation must write a record to `engLog` with `memberId`, `activity`, `xp`, `timestamp`, and `awardedBy`. The scanner flow is the highest-risk gap.

**H3. Role-based access control is client-side only**
- Location: app/index.html lines 952-963 (`renderSidebar`), `renderView()` switch
- Nav items are filtered by role in the sidebar (`if(item.roles&&!item.roles.includes(u.role)) return`), but `navigate('financials')` can be called directly from the browser console by any user, bypassing all role checks.
- Risk: A member who opens browser devtools can call `navigate('admin')` or `navigate('intake')` and access all admin views and data.
- Fix: `renderView()` must enforce role checks before rendering each protected view, not just in the nav. This is also an inherent limitation of client-side-only auth that Glide's server-side role enforcement must address in production.

**H4. System Console credential (CES1231) has no session expiry or lockout**
- Location: app/index.html line 863 (`loginAs()`) and `doLogin()`
- The sysadmin account has no timeout, no failed-attempt lockout, and no MFA check.
- Risk: If the browser tab is left open, any person at the keyboard has full system access. The README warns about this but the code has no mitigation.
- Fix: For the demo prototype, add a session timeout (auto sign-out after 30 minutes of inactivity). In production, the CES1231 account should not exist (README acknowledges this).

**H5. "Test Connection" buttons in System Console are pure UI stubs**
- Location: app/index.html line 3553 (`renderSysIntegrations`)
- Every integration card has a "Test Connection" button that calls `showToast('Testing ${s.name} connection…')`. No actual connectivity check occurs.
- Risk: An admin who uses System Console to verify that Glide, Make.com, or Cloudinary are healthy will receive a false positive. If any integration is broken, the console will still show "connected" status.
- Fix: In production, "Test Connection" must make a real API call (or trigger a Make.com scenario) and display a genuine status response. The current stub creates a false sense of operational confidence.

**H6. Config save buttons are pure UI stubs**
- Location: app/index.html lines 3641, 3662 (`renderSysConfig`)
- "Save Changes" in Chapter Settings and "Save XP Values" in Point Economy both call `showToast('Config saved!','success')` with no actual data persistence.
- Risk: An admin believes chapter settings and XP values have been saved. They have not. The next page reload reverts all changes. This is prototype behavior but will cause data loss if not clearly communicated before any live use.

---

### MEDIUM

**M1. No performance monitoring**
- No `PerformanceObserver`, no `window.performance.mark`, no timing around DOM rendering.
- With 47 views and 4,140 lines of JavaScript executing in a single file, render performance degradation will be invisible until users complain. The leaderboard sort (`[...members].sort((a,b)=>b.xp-a.xp)`) runs on every dashboard render with no caching.

**M2. The engagement log is not updated by all XP-awarding operations**
- `confirmScan()` (check-in via QR) awards XP but never writes to `engLog`.
- `rsvpEvent()` awards +10 XP but never writes to `engLog`.
- `submitProps()` awards +10 XP but writes to `propsData` only, not `engLog`.
- This means `sysRecalcXP()` (which derives from `engLog`) will produce incorrect results for members who gained XP through these paths.
- Impact: XP recalculation, the primary recovery mechanism for data corruption, will silently undercount XP.

**M3. Form submissions use optional chaining to swallow missing element errors**
- Pattern throughout: `document.getElementById('ev-name')?.value`, `document.getElementById('xp-amount').value||0`
- When an element is not found (wrong view, stale DOM), the form silently submits with empty/default values. No error is shown; no log is written.
- Example: `awardXP()` at line 1715 — if `xp-member` element is not found, `m` is undefined. `showToast('+0 XP awarded to undefined', 'success')` would appear.

**M4. No usage analytics for production adoption tracking**
- The README references a 12-week launch guide and adoption goals, but the app has no mechanism to track which features members use, which views they visit most, or where they drop off.
- Without this, the build lead cannot make data-driven decisions about which Year 1 enhancements to prioritize.
- Fix: Add a lightweight analytics call (Google Analytics 4 is free) on `navigate()` to track screen views.

**M5. Duplicate check-in not prevented**
- `confirmScan()` has no guard against scanning the same member twice in one session.
- `scanLog` records every scan but `confirmScan()` never checks it before awarding XP.
- A chair who accidentally taps "Confirm Check-In" twice will award double XP with no warning.

**M6. Session state not persisted across page reloads**
- All application state (members, XP, scan logs, RSVPs, engagement log) is reset on every page reload.
- Any data entered during a session (new events, service log entries, admin changes) is lost when the browser tab is closed.
- This is inherent to a static HTML demo, but there is no warning shown to users before they enter real data.

**M7. No network connectivity detection**
- The app has no handler for `window.addEventListener('offline')`.
- In the production Glide app, offline Glide usage is limited. But the prototype gives no indication to users when the system is offline, and makes no attempt to queue or retry operations.

---

### LOW

**L1. showToast is the only error feedback mechanism — no severity distinction in UX**
- `showToast('msg', 'error')` and `showToast('msg', 'success')` look nearly identical (both navy with different left-border colors). Error states are visually understated.
- An admin who marks dues paid for the wrong member, or a chair who double-scans, receives the same visual weight of feedback as a successful action.

**L2. The `SYS_AUDIT_LOG` timestamps use "Just now" for all entries**
- Every audit log entry gets `time:'Just now'` regardless of when it was created.
- After a session with multiple admin actions, the entire audit log reads "Just now" for every entry, making it impossible to reconstruct event ordering.
- Fix: Use `new Date().toLocaleTimeString()` or an ISO timestamp.

**L3. No .env.example or environment variable documentation**
- The Make.com webhook URL is hardcoded as a placeholder in the System Console config.
- There is no `.env.example` or equivalent to document what environment-specific values must be set before deployment.

**L4. Dead placeholder actions emit misleading success toasts**
- "Save to Photos", "Share", "Export My Data", "Contact Support", "Opening import wizard", "Force Sheet Sync" all call `showToast('…', 'success')` for operations that do nothing.
- If these are not clearly distinguished in a deployment guide, a real user will believe these operations succeeded.

**L5. XP awarded via `requestMentor()` is not logged to `engLog`**
- `requestMentor()` at line 2809 awards +50 XP directly (`u.xp += 50`) with no engagement log entry and no audit record.

---

### INFO

**I1. Consider a client-side structured logging wrapper**
- Even without a backend, a logging wrapper (`logger.info(event, context)`) that batches events and sends them to a free endpoint (PostHog, Mixpanel free tier, or a Make.com webhook) would give the chapter visibility into real usage patterns.

**I2. The System Console health overview could become a genuine liveness dashboard**
- The Overview tab in the System Console has the right shape for an operational dashboard (service statuses, ops budget, last sync). Wiring it to real Make.com and Glide APIs would make it genuinely useful for the chapter admin.

**I3. Glide + Make.com have native observability that should be documented**
- Make.com's execution history is the primary audit trail for the production system. The team should document: where to find Make.com execution logs, how to diagnose a failed XP automation, and how to check Google Sheets sync status in Glide. This is the real ops runbook for this architecture.

---

## Operational Debt

- Placeholder stub functions with misleading success feedback: 14 buttons
- In-memory audit log (no persistence): 1 critical path
- TODO / FIXME items in codebase: 0 (none found)
- Dead code (unused exports): Not applicable (single-file JS, no module system)
- Undocumented environment values: Make.com webhook URL, Glide app URL, Cloudinary URLs

---

## Production Readiness Checklist (for Glide deployment)

| Item | Status |
|------|--------|
| Remove / replace prototype login with Glide allowlist | Required |
| Disable or remove CES1231 sysadmin account from Glide allowlist | Required |
| Wire "Test Connection" to real API calls | Required |
| Persist audit log to Google Sheet or Make.com | Required |
| Add error reporting (Sentry or equivalent) | Strongly recommended |
| Add screen-view analytics | Recommended |
| Document Make.com execution log as primary ops runbook | Required |
| XP mutation idempotency (prevent double-scan) | Required |
| Role enforcement in `renderView()` not just sidebar | Required |
| Add real timestamps to `SYS_AUDIT_LOG` | Required |

---

*Report written to: /Users/carnell.smithgotyto.com/Documents/blue-ledger/.claude/vision/observability-report.md*
