# Deliberix — Deliberative Survey Platform

Deliberative survey platform (wiki survey) where participants rate statements as agree/disagree/abstain, mark them as important, and submit new ones for moderation. Inspired by Pol.is and All Our Ideas.

Website: https://deliberix.com/ | Academix ecosystem: https://www.academix.cz/deliberix

## Tech Stack

- **Frontend**: Ionic 8 / Angular 21 / Capacitor 8 — standalone components, signals, SCSS
- **Backend**: Go 1.26 — chi router, mibk/dali ORM
- **Database**: MariaDB 12 — singular table names
- **i18n**: @ngx-translate (cs/en), JSON files in `client/src/assets/i18n/`
- **Docker**: dev + prod compose files, Adminer on :8081 in dev
- **Auth**: JWT (HMAC-SHA256, 30-day expiry), email/password. Auto-verify for now (no SMTP yet)

## Coding Conventions

- **Go**: K&R braces, backtick SQL (no `\n`), descriptive package names (not generic like "service")
- **Angular**: standalone components, signals (not BehaviorSubjects), separate .html/.scss/.ts files
- **Double quotes** everywhere (TS, Go, JSON) — enforced by Prettier
- **Prettier** on save (VS Code), `npm run format` for CLI. Config in `client/.prettierrc`
- **Minimal targeted changes** over full file rewrites
- **Simple/readable** over DRY complexity

## Project Structure

```
server/
├── main.go              # Router setup, wiring
├── config/config.go     # Env var loading
├── store/               # Database layer (dali ORM)
│   ├── store.go         # DB connection, queryOne[T] generic helper
│   ├── dalix.go         # InTx transaction helper
│   ├── user.go, survey.go, statement.go, response.go, participant.go
├── model/               # Structs with db/json tags
│   ├── user.go, survey.go, statement.go, response.go, participant.go, stats.go
├── handler/             # HTTP handlers (AppHandlerFunc returning error)
│   ├── handler.go       # JSON helpers, ErrorHandler wrapper
│   ├── auth.go, survey.go, statement.go, response.go, moderation.go, results.go, me.go
├── middleware/auth.go   # JWT Bearer extraction → identity context
├── identity/identity.go # ContextWithUser, GetUserFromContext
├── service/auth.go      # bcrypt + JWT (TODO: rename to auth/)
├── slug/slug.go         # Czech-aware slug generation

client/src/app/
├── services/            # Injectable services with signals
│   ├── api, auth, locale, storage, survey, statement, response, moderation, results, theme, toast
├── models/              # TypeScript interfaces
├── pages/               # Ionic pages (separate .html/.scss/.ts)
│   ├── login, register, survey-list, survey-create, survey-detail
│   ├── survey-vote, survey-join, survey-moderation, survey-results
│   ├── settings, not-found
├── components/          # Reusable components
│   ├── seed-statements, submit-statement
├── guards/auth.guard.ts
├── interceptors/auth.interceptor.ts
```

## API Routes (all under /api/v1/)

```
POST   /auth/register, /auth/login
GET    /auth/me                          (protected)

GET    /survey                           (user's surveys)
GET    /survey/public                    (discoverable surveys)
POST   /survey                           (create)
GET    /survey/{slug}                    (detail)
PATCH  /survey/{slug}                    (update)
POST   /survey/{slug}/join
GET    /survey/{slug}/participant/me
GET    /survey/{slug}/statement
POST   /survey/{slug}/statement          (user submitted → pending)
POST   /survey/{slug}/statement/seed     (admin → approved)
GET    /survey/{slug}/statement/next
GET    /survey/{slug}/moderation
GET    /survey/{slug}/results
GET    /survey/{slug}/stats
GET    /survey/{slug}/progress
POST   /statement/{id}/response
PATCH  /statement/{id}/moderate
```

## Database Schema

MariaDB with tables: `organization`, `user`, `survey`, `survey_participant`, `statement`, `response`, `email_verification`, `password_reset`. Full schema in `db/schema.sql`, migrations in `db/migrations/`.

Key fields on `survey`: slug (unique, auto-generated), status (draft/active/closed), visibility, privacy_mode, invitation_mode, result_visibility, statement_order, statement_char_min/max, intake_config (JSON), closes_at.

Survey uses slugs for URLs (e.g., `/survey/budoucnost-hory`). No numeric IDs in routes.

## Key Patterns

- **dali ORM**: `?values` for INSERT, `?set` with `dali.Map` for partial UPDATE, `One()`/`All()` for SELECT
- **queryOne[T]**: generic helper in `store/store.go` for single-row queries (returns nil on ErrNoRows)
- **AppHandlerFunc**: handlers return error, `ErrorHandler` wraps them into http.HandlerFunc
- **getSurveyFromSlug**: shared helper resolves survey from `{slug}` URL param
- **Signals**: Angular services use `signal()`, `computed()` for state. Pages use `ionViewWillEnter` for reload-on-navigate
- **Toast service**: centralized `toast.success()`, `toast.error()`, `toast.apiError()`
- **Theme service**: auto/light/dark with `ion-palette-dark` class, persisted in Capacitor Preferences
- **replaceUrl**: login, register, create, join pages use `replaceUrl: true` to prevent back-button issues

## Development

```bash
docker-compose -f docker-compose-dev.yml up        # start all
docker-compose -f docker-compose-dev.yml up --build # rebuild after dependency changes
```

- Client: http://localhost:4200 (hot reload on src/ changes)
- Server: http://localhost:8080 (air hot reload on .go changes)
- Adminer: http://localhost:8081
- MariaDB healthcheck ensures server waits for DB readiness

## What's Implemented (MVP)

- Auth (register, login, JWT, guards, interceptor)
- Survey CRUD (create, list, detail, update, activate/close with confirmation)
- Statements (seed by admin, submit by participants → pending moderation)
- Voting (one-at-a-time with progress bar, haptic feedback, importance toggle)
- Moderation (approve/reject queue for admins/moderators)
- Results (stats grid, vote distribution bars, sort by votes/agreement/importance, visibility rules)
- Survey discovery (public surveys for non-participants)
- Side menu (split-pane: sidebar on desktop, hamburger on mobile)
- Settings (language switch cs/en, theme auto/light/dark)
- 404 page, friendly slug URLs, skeleton loading, button spinners, error/success toasts
- Prettier + format-on-save, ESLint, organize imports

## What's NOT Implemented Yet

- Survey settings UI (visibility, privacy, result_visibility, statement order, char limits, closes_at — all in DB but no form)
- Participant management (list participants, assign moderator role)
- Email system (verification, password reset, SMTP, multilingual templates)
- User profile editing (name, password change)
- Intake config builder UI
- Organization management
- Invitation system
- Data export / analytics
- Landing page
- Security hardening (rate limiting, CSRF, account lockout)
- Native mobile builds

See `.claude/plans/post-mvp-backlog.md` for the full phased roadmap.
