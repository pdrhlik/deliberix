## TODO — Priority Order

### P0 — Must fix before deploy

- **Server error messages are raw English strings.** They appear in toasts via `apiError()`. Need a system where server returns error codes and client maps them to i18n keys. Currently ~70 `writeError()` calls with hardcoded English. Temporary fix: at least make them proper sentences with periods.

- ~~**Spacing/alignment inconsistencies across pages.**~~ DONE — Replaced `ion-card` and `ion-list > ion-item` with custom `.content-card` / `.readonly-field` / `.survey-card` divs everywhere. All inputs use `fill="outline"`, all containers use 600px max-width. Consistent card pattern: `background: var(--ion-color-light)`, `border-radius: 8px`, `padding: 0.75-1rem`.

- ~~**Survey detail page admin buttons**~~ DONE — Moved activate/close buttons to the settings tab.

- **Double join step for public surveys.** Survey detail shows a "Join" button that navigates to `/survey/:slug/join` which then shows another join button. For surveys without an intake form, joining should be a single click. Only show the join page if there's an intake form to fill out.

- **Dates not reflecting current language.** The `DatePipe` with `locale` param was added to some pages but not all. The `LocaleService` registers locale data but Angular's `LOCALE_ID` is set at bootstrap and doesn't change dynamically. Every `| date` pipe needs the `:locale` suffix, or use a shared helper/pipe.

### P1 — Important UX improvements

- **Intake config builder UX is terrible.** Cards inside cards, tiny inputs, no visual hierarchy. Research how Typeform, Google Forms, SurveyMonkey handle form builders. Consider: slides/stepper pattern, drag-and-drop via CDK, collapsible field editors, inline previews. This needs a full redesign, not incremental fixes.

- **Voting flow could be more engaging.** Current vote page works but is basic. Consider: card swipe animations, keyboard shortcuts (1/2/3 for agree/disagree/skip), bigger touch targets, statement counter in header, smooth transitions between statements.

- **Results page needs better visualization.** The inline results tab is good but basic. Consider: animated bars, hover/tap for details, comparison view (most agreed vs most disagreed), filtering by own votes, export button.

- **Survey list filtering/sorting.** Need tabs or segments: Active (with unvoted statements), Completed, Closed. Separate section for admin-owned vs participant surveys. Sort by recent activity. Currently everything is in one flat list.

- **Mobile tabs should be at the bottom.** The survey detail segments are in the header toolbar which works on desktop but is hard to reach on mobile. Ionic's `ion-tabs` with `ion-tab-bar` at the bottom would be better for mobile. This requires a routing refactor — segments can't be swapped to bottom tabs easily.

- **Result visibility enforcement on client.** The server guards results based on `resultVisibility` setting, but the client always shows the Results tab. When the server returns 403, the tab should be grayed out / show a message. Admins should always see results regardless.

### P2 — Functional gaps

- **Unlisted visibility** — what does it mean? Currently `ListPublicSurveys` includes both `public` and `unlisted`. The intended behavior: unlisted surveys are accessible by direct link but don't appear in the public browse list. Need to exclude `unlisted` from the public list query.

- **Privacy modes** — `anonymous`, `public`, `participant_choice` are stored but never enforced. Results don't show who voted what (effectively always anonymous). When `public` mode is implemented, results should show individual voter names. `participant_choice` would let each participant decide.

- **Intake data editing** — participants should be able to update their intake data after joining (e.g., they made a mistake in their demographic info). Need a PATCH endpoint and UI.

- **Form validation** — login, register, create survey, settings forms have no client-side validation. Need: required field indicators, min/max length enforcement, email format validation, inline error messages, focus first invalid input on submit.

- **Anonymous link-only surveys** (conference mode) — detailed in Czech email from stakeholder. Would allow surveys where participants don't need accounts, just a link + basic intake form. Needs: `auth_mode` enum on survey (`registered` / `anonymous_link`), session-based anonymous participants, duplicate vote prevention via cookies/fingerprint.

### P3 — Polish & infrastructure

- **All server error messages need i18n.** Current approach of raw English strings in `writeError()` doesn't scale. Options: (a) server returns error codes, client maps to i18n keys, (b) server accepts `Accept-Language` header and returns translated messages. Option (a) is cleaner.

- **Security audit** — rate limiting on auth endpoints (login, register, magic link), CSRF protection, input sanitization audit, JWT refresh token rotation, account lockout after failed attempts, CORS tightening for production.

- ~~**Production deployment**~~ DONE — `DEPLOY.md` has full instructions. Schema auto-applied on first start, no manual migrations needed.

- **Survey categories/tags** for better browsing.

- **Data export** — CSV/JSON export of survey results for admins.

- **Organization management** — multi-tenant support.

---

## NOTES/QUESTIONS for team

- Do we want to mark in which order the statements were displayed to the user?
- Admin is automatically a participant and can vote. Do we want that or does he have to join as well?
- Should the intake config builder support reordering options within a field via drag-and-drop?
- For the anonymous survey mode: what's the minimum acceptable duplicate prevention? Cookie-only? Email uniqueness check?
- Name suggestion - Deliberix
