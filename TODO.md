## TODO — Priority Order

### P0 — Must fix before deploy

- **Server error messages are raw English strings.** They appear in toasts via `apiError()`. Need a system where server returns error codes and client maps them to i18n keys. Currently ~70 `writeError()` calls with hardcoded English. Temporary fix: at least make them proper sentences with periods.

### P1 — Important UX improvements

- **Intake config builder UX is terrible.** Cards inside cards, tiny inputs, no visual hierarchy. Research how Typeform, Google Forms, SurveyMonkey handle form builders. Consider: slides/stepper pattern, drag-and-drop via CDK, collapsible field editors, inline previews. This needs a full redesign, not incremental fixes.

- **Voting flow could be more engaging.** Current vote page works but is basic. Consider: card swipe animations, keyboard shortcuts (1/2/3 for agree/disagree/skip), bigger touch targets, statement counter in header, smooth transitions between statements.

- **Results page needs better visualization.** The inline results tab is good but basic. Consider: animated bars, hover/tap for details, comparison view (most agreed vs most disagreed), filtering by own votes, export button.

- **Survey list filtering/sorting.** Need tabs or segments: Active (with unvoted statements), Completed, Closed. Separate section for admin-owned vs participant surveys. Sort by recent activity. Currently everything is in one flat list.

- **Mobile tabs should be at the bottom.** The survey detail segments are in the header toolbar which works on desktop but is hard to reach on mobile. Ionic's `ion-tabs` with `ion-tab-bar` at the bottom would be better for mobile. This requires a routing refactor — segments can't be swapped to bottom tabs easily.

- **Result visibility enforcement on client.** The server guards results based on `resultVisibility` setting, but the client always shows the Results tab. When the server returns 403, the tab should be grayed out / show a message. Admins should always see results regardless.

### P2 — Functional gaps

- **Privacy modes** — `anonymous`, `public`, `participant_choice` are stored but never enforced. Results don't show who voted what (effectively always anonymous). When `public` mode is implemented, results should show individual voter names. `participant_choice` would let each participant decide.

- **Intake data editing** — participants should be able to update their intake data after joining (e.g., they made a mistake in their demographic info). Need a PATCH endpoint and UI.

- **Form validation** — login, register, create survey, settings forms have no client-side validation. Need: required field indicators, min/max length enforcement, email format validation, inline error messages, focus first invalid input on submit.

- **Anonymous link-only surveys** (conference mode) — detailed in Czech email from stakeholder. Would allow surveys where participants don't need accounts, just a link + basic intake form. Needs: `auth_mode` enum on survey (`registered` / `anonymous_link`), session-based anonymous participants, duplicate vote prevention via cookies/fingerprint.

### P3 — Polish & infrastructure

- **All server error messages need i18n.** Current approach of raw English strings in `writeError()` doesn't scale. Options: (a) server returns error codes, client maps to i18n keys, (b) server accepts `Accept-Language` header and returns translated messages. Option (a) is cleaner.

- **Login redirect** — when an unauthenticated user visits a deep link (e.g. an unlisted survey), the auth guard redirects to `/login` but loses the original URL. After login, the user should be redirected back to the page they originally requested. The auth guard should store the attempted URL (e.g. in a query param or service) and the login page should redirect there on success.

- **Security audit** — rate limiting on auth endpoints (login, register, magic link), CSRF protection, input sanitization audit, JWT refresh token rotation, account lockout after failed attempts, CORS tightening for production.

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

## More TODO

- Prepared intake form questions
- Admin
- When submitting a new statement, users should see right away the min characters of the statement, not just a red char limit
- when I approve a statement in moderation and then click on particip or overview, I can't see it right away and have to refresh the page first. Participants probably have the same issue. We need fresh data when clicking on the particip/overview or other tabs in the survey detail
- after survey is closed, not even admin can delete participants or change their role
- think about integrating gemmatranslate into the docker image - optional feature
