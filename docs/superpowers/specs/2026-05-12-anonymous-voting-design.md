# Anonymous Voting — Design Spec

Date: 2026-05-12
Status: Approved, ready for implementation plan
Tracks: Item 9b from `Poznámky k úpravám Deliberix` (post-launch backlog)

## Goal

Let people scan a survey's QR code (or open its share link) and immediately participate — vote, submit statements, see results — without creating an account. The QR/share UI is already shipped (commit `0e8c693`); this spec covers the structural change that makes the link work for not-yet-registered visitors.

## Non-goals (out of scope for v1)

- Anon → user vote claim/migration on signup. Anon votes stay anon.
- CAPTCHA, advanced bot/abuse detection.
- Per-IP rate limiting on join/vote.
- Anon data-subject-rights export. (Anon rows hold only intake answers; removal is via the admin "Remove participant" flow already in place.)
- iOS Universal Links / Android App Links to route QR scans into a native app when installed. URL format is forward-compatible; this is added when the native app ships.
- CSRF token. The anon endpoints mutate via a `SameSite=Lax` cookie and CORS is already restricted; can revisit if SameSite policy is later relaxed.

## Settled decisions (from brainstorm)

1. **Per-survey admin opt-in** via a new `allow_anonymous` flag (default `false`). Survey owners explicitly enable it.
2. **Reuse existing tables** with nullable `user_id` + new `anon_session_id` columns on `survey_participant` and `response`. No parallel anon tables.
3. **Signed-cookie abuse posture only** for v1 — HttpOnly, HMAC-signed UUID cookie. Honest-actor scope (conferences, classrooms). Stronger defenses added later if needed.
4. **Full participant capabilities** for anon visitors once they've joined: vote, mark important, submit statements (subject to moderation), view results (subject to `result_visibility`).
5. **Intake form applies to anon** the same as to authenticated participants. If the survey has `intake_config`, anon visitors complete it before voting.
6. **Anon votes stay anon** — no transfer on signup.
7. **`allow_anonymous` is orthogonal to `visibility`** — admin combines them at their discretion. UI may warn on weird combos (e.g., `private` + `allow_anonymous`) but the backend imposes no hard restriction.
8. **Unified routes with dual-auth middleware**, not parallel public/private route trees. Admin endpoints stay JWT-only.
9. **Single DB migration file** for all schema changes (additive only — no existing data is touched).

## § 1 · Surface (API + client)

### New survey field

- `survey.allow_anonymous TINYINT(1) NOT NULL DEFAULT 0` — editable in draft phase only (locks on activation, mirroring `moderation_enabled`).
- Exposed in **survey-create** Advanced settings and **survey-detail** Settings tab as a toggle "Allow anonymous voting via shared link".

### New endpoints

- `POST /api/v1/survey/{slug}/anon/join` — **public, no auth required**. Issues the `dlbx_anon` cookie, creates a participant row with `user_id = NULL, anon_session_id = <uuid>, intake_data = <body>`. Returns `403 anon_not_allowed` if `survey.allow_anonymous = false`. Returns `409 already_a_participant` if the caller already has a participant row (via cookie or JWT) for this survey.
- `POST /api/v1/anon/logout` — clears the anon cookie. Returns 204. No-op if no cookie present.

### Endpoints upgraded to dual-auth (JWT **OR** anon cookie)

- `GET    /api/v1/survey/{slug}`
- `GET    /api/v1/survey/{slug}/participant/me`
- `GET    /api/v1/survey/{slug}/statement`
- `GET    /api/v1/survey/{slug}/statement/next`
- `POST   /api/v1/survey/{slug}/statement` (submitting a new statement)
- `POST   /api/v1/statement/{id}/response` (casting a vote)
- `GET    /api/v1/survey/{slug}/results` (`result_visibility` gating unchanged)
- `GET    /api/v1/survey/{slug}/stats`
- `GET    /api/v1/survey/{slug}/progress`

### Endpoints that stay JWT-only

Everything under `/api/v1/auth/*`, survey CRUD (`POST/PATCH/DELETE /api/v1/survey[/{slug}]`), participants admin (`/participants`, `/role`, removal), `POST /api/v1/survey` (create), `POST /api/v1/survey/{slug}/statement/seed`, moderation (`/moderation`, `/moderate`), `GET /api/v1/survey` (the user's own surveys), `GET /api/v1/survey/public` (discovery).

### Client route guard

`authGuard` on `/survey/:slug` is replaced with `surveyAccessGuard`:

- JWT valid → allow.
- No JWT → fetch `GET /api/v1/survey/{slug}` (the now dual-auth endpoint, sent with credentials so the anon cookie is included if present). If it returns 200, allow. If 404/403, redirect to `/login?redirect=/survey/{slug}`.

All other authed routes keep using `authGuard`.

## § 2 · Database

### Migration `db/migrations/005_add_anonymous_voting.sql`

```sql
ALTER TABLE survey
  ADD COLUMN allow_anonymous TINYINT(1) NOT NULL DEFAULT 0 AFTER moderation_enabled;

ALTER TABLE survey_participant
  MODIFY user_id INT UNSIGNED DEFAULT NULL,
  ADD COLUMN anon_session_id CHAR(36) DEFAULT NULL AFTER user_id,
  ADD CONSTRAINT chk_participant_identity
    CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
  ADD UNIQUE KEY uq_survey_anon (survey_id, anon_session_id);

ALTER TABLE response
  MODIFY user_id INT UNSIGNED DEFAULT NULL,
  ADD COLUMN anon_session_id CHAR(36) DEFAULT NULL AFTER user_id,
  ADD CONSTRAINT chk_response_identity
    CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
  ADD UNIQUE KEY uq_statement_anon (statement_id, anon_session_id);
```

`db/schema.sql` gets the equivalent column additions so fresh installs match the migrated state.

### Notes

- Existing `uq_survey_user (survey_id, user_id)` and `uq_statement_user (statement_id, user_id)` keys stay untouched. MariaDB treats NULLs as distinct in unique indexes, so anon rows (where `user_id IS NULL`) don't trip these keys. The new `uq_survey_anon` / `uq_statement_anon` cover the anon side. The XOR CHECK keeps the two key sets from overlapping.
- `survey_participant.user_id` FK to `user(id) ON DELETE CASCADE` stays — MariaDB allows nullable FK columns.
- Every existing `survey_participant` and `response` row has a non-null `user_id`; the XOR check passes trivially. No data migration needed.
- `statement.author_id` is already nullable (`ON DELETE SET NULL`), so anon-submitted statements drop in as `author_id = NULL` — no schema change.

## § 3 · Auth / middleware

### `identity.Actor`

A unified actor abstraction replaces direct `User` lookups for any handler that may now serve either kind of caller.

```go
// server/identity/actor.go
type Actor struct {
    UserID        *uint   // set if JWT
    AnonSessionID *string // set if anon cookie
    User          *User   // full user record when UserID set; nil for anon
}

func (a *Actor) IsAnon() bool         { return a != nil && a.AnonSessionID != nil }
func (a *Actor) IsAuthenticated() bool { return a != nil && a.UserID != nil }
```

Helpers:

```go
func ContextWithActor(ctx context.Context, a *Actor) context.Context
func GetActorFromContext(ctx context.Context) *Actor
```

The existing `User`-from-context API stays around for JWT-only routes that need the full user record. New helpers complement it; they don't replace.

### Middleware

- **`middleware.AuthJWT`** (existing `Auth` renamed; behavior unchanged) — 401 if no valid JWT. Attaches `*User` and `*Actor{UserID, User}` to context.
- **`middleware.AuthOptional`** (new) — tries Bearer JWT first; if absent, tries `dlbx_anon` cookie; if a valid cookie is found, attaches `*Actor{AnonSessionID}` to context. If neither, no actor is set and the request continues — handlers decide what to do with no actor.

The "email verification required" gate stays on `AuthJWT` only: anon callers have no email and bypass that block.

### Route wiring in `main.go`

```go
// Public — no auth at all
r.Post("/api/v1/auth/*", ...)              // existing
r.Post("/api/v1/survey/{slug}/anon/join", ...) // NEW
r.Post("/api/v1/anon/logout", ...)             // NEW

// JWT-only (admin / user-management)
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthJWT(...))
    r.Get("/api/v1/auth/me", ...)
    r.Post("/api/v1/survey", ...)
    r.Patch("/api/v1/survey/{slug}", ...)
    r.Delete("/api/v1/survey/{slug}", ...)
    r.Get("/api/v1/survey", ...)             // list user's own surveys
    r.Get("/api/v1/survey/public", ...)      // discovery
    r.Get("/api/v1/survey/{slug}/participants", ...)
    r.Patch("/api/v1/survey/{slug}/participant/{userId}/role", ...)
    r.Delete("/api/v1/survey/{slug}/participant/{userId}", ...)
    r.Post("/api/v1/survey/{slug}/statement/seed", ...)
    r.Get("/api/v1/survey/{slug}/moderation", ...)
    r.Patch("/api/v1/statement/{id}/moderate", ...)
})

// Dual-auth (anon cookie OR JWT)
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthOptional(...))
    r.Get("/api/v1/survey/{slug}", ...)
    r.Post("/api/v1/survey/{slug}/join", ...)         // existing JWT join is kept for backward compat
    r.Get("/api/v1/survey/{slug}/participant/me", ...)
    r.Get("/api/v1/survey/{slug}/statement", ...)
    r.Get("/api/v1/survey/{slug}/statement/next", ...)
    r.Post("/api/v1/survey/{slug}/statement", ...)
    r.Post("/api/v1/statement/{id}/response", ...)
    r.Get("/api/v1/survey/{slug}/results", ...)
    r.Get("/api/v1/survey/{slug}/stats", ...)
    r.Get("/api/v1/survey/{slug}/progress", ...)
})
```

### Cookie spec

- Name: `dlbx_anon`
- Value: `{anon_session_id}.{hmac_base64url}` where `hmac = HMAC-SHA256(JWT_SECRET, anon_session_id)`
- Flags: `HttpOnly`, `SameSite=Lax`, `Secure` in production (off when serving over `http://localhost`), `Path=/`, `Max-Age=31536000` (365 days)
- On `POST /anon/join`: server generates a new UUID v4, signs it, sets the cookie, persists the participant row
- On `POST /anon/logout`: server sends `Set-Cookie: dlbx_anon=; Max-Age=0; Path=/`
- The cookie carries the only authoritative source of an anon session — no DB session table

### Handler pattern

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
}
participant, err := h.Store.GetParticipantByActor(r.Context(), survey.ID, actor)
// ... rest of the handler is identity-agnostic
```

### Store changes

The store grows one shape of helper:

- `GetParticipantByActor(ctx, surveyID, *Actor) (*Participant, error)` — branches internally on `actor.UserID` vs `actor.AnonSessionID`.
- Similar `IsParticipantByActor`, `GetUserVotesByActor`, `GetVoteProgressByActor`, etc., or simply pass the Actor through to existing methods that internally branch on which field is non-nil.

Existing methods (`GetParticipant(surveyID, userID)`) stay for the JWT-only paths to avoid churning every callsite.

## § 4 · Client UX

### Anon visitor lands on `/survey/{slug}`

1. `surveyAccessGuard` allows the route through (per § 1).
2. `survey-detail` page loads survey + participation.
3. Participant fetch returns 404 (anon not yet joined). The page enters its "not yet a participant" state and renders a join CTA — *but* the CTA text/behavior depends on auth state:
   - **Logged-in user, allow_anonymous doesn't apply**: existing "Join survey" button (unchanged behavior).
   - **Not logged in, `survey.allow_anonymous = true`**: primary CTA "Join anonymously"; if `intake_config` exists, this opens the existing intake form; on submit, POST `/anon/join` with intake data; on success the cookie is set, the page reloads survey + participant, voting available. Secondary link "Sign in instead" → `/login?redirect=/survey/{slug}`.
   - **Not logged in, `survey.allow_anonymous = false`**: shouldn't reach this state often (route guard would have redirected to login), but as a safety net the page shows "Sign in to participate" linking to `/login`.

### Already joined anonymously

Identical UI to a logged-in participant. The submit-statement component, voting page, results component all read participation state from the shared `participant/me` response — they don't care whether identity is user or anon.

### Side menu

`<ion-app>` only renders the side menu when `auth.isAuthenticated()`. Anon visitors get the bare `<ion-router-outlet>` — no menu, no nav, no profile. Focused single-page experience for "scan-and-vote".

### "End anonymous session" affordance

A small footer link on the survey page for anon participants: "End anonymous session". Calls `POST /api/v1/anon/logout` and reloads. Useful on shared devices.

### Auth interceptor

The HTTP interceptor that adds `Authorization: Bearer ...` already does so only when there's a JWT in storage. For anon callers there's no JWT, so no header — the cookie travels automatically (assuming `withCredentials: true` is set or Angular's `HttpClient` is configured to include credentials for same-origin). In production this is same-origin already; in dev (Angular on `:4200`, API on `:8180`) we'll need `withCredentials: true` on the `HttpClient` config plus a matching `AllowCredentials` on the server's CORS handler.

### i18n keys (new)

- `survey.join-anon` — "Join anonymously" / "Připojit se anonymně"
- `survey.join-anon-sign-in-instead` — "Sign in instead" / "Nebo se přihlásit"
- `survey.anon-end-session` — "End anonymous session" / "Ukončit anonymní hlasování"
- `errors.anon_not_allowed` — "This survey doesn't accept anonymous voting." / "Tento průzkum neumožňuje anonymní hlasování."
- `survey.allow-anonymous` — settings toggle label
- `survey.allow-anonymous-help` — toggle help text

## § 5 · Admin views

### Participants list

Currently joins `survey_participant ↔ user` and renders `name <email>`. Updated to a LEFT JOIN and to handle anon rows:

- Authenticated participant: `John Doe <john@example.com>` (unchanged)
- Anonymous participant: localized label "Anonymous (joined <date>)" + intake summary if present
- Removal works as today — the existing DELETE endpoint, given the participant row id, deletes it via the existing cascade (votes go too).

### Moderation queue

`statement.author_id IS NULL` already happens when an author user was deleted. The moderation list will be tweaked to render "Anonymous author" in that case — a single i18n branch.

### Stats

`total_participants` updates from `COUNT(*) FROM survey_participant WHERE survey_id = ?` (which already implicitly counts anon participants once the rows exist). `total_votes` similarly unchanged — `response` rows are counted regardless of identity. No new aggregation columns.

### Survey list / discovery

Both filter by `user_id` against the current JWT user — anon participation doesn't surface here. Unchanged.

## § 6 · Backward compatibility

This is the explicit compatibility contract for the change set.

- **Existing surveys**: `allow_anonymous = 0` after the migration. The new join endpoint returns 403 for them. No surveys change behavior.
- **Existing participants / responses**: every row has a non-null `user_id` and null `anon_session_id` — passes the XOR check. The new unique keys are additive. No data is rewritten.
- **Existing handlers** that read `user_id` directly continue to do so on the JWT path; the anon path is purely additive.
- **JWT users on the dual-auth routes**: get the same `*Actor` filled with `UserID + User`. Handler logic that previously read `User` from context continues to compile (we keep that helper) and continues to return the same results.
- **API clients** (the Angular app): the response shapes don't change. The only client-visible difference is that the new endpoints exist. Existing code paths are untouched on the auth/JWT side.

## Open questions

None remaining at end of brainstorm. All design decisions captured above.
