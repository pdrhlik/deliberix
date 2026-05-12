# Anonymous Voting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let people scan a survey's QR code or open its share link and immediately participate (vote, submit statements, see results) without creating an account — gated by a per-survey admin opt-in.

**Architecture:** Reuse existing `survey_participant` and `response` tables with nullable `user_id` and a new `anon_session_id` column. A signed HttpOnly cookie (`dlbx_anon`) carries the session id. A unified `identity.Actor` abstraction lets every shared handler serve both JWT users and anon callers via a new `AuthOptional` middleware; admin endpoints stay JWT-only via the renamed `AuthJWT`.

**Tech Stack:** Go 1.26 (chi router, dali ORM), MariaDB 12, Angular 21 / Ionic 8, signed-cookie auth via HMAC-SHA256 with the existing `JWT_SECRET`.

**Spec:** `docs/superpowers/specs/2026-05-12-anonymous-voting-design.md`

---

## Task 1 — DB migration + schema

**Files:**
- Create: `db/migrations/005_add_anonymous_voting.sql`
- Modify: `db/schema.sql`

- [ ] **Step 1: Create the migration**

Write `db/migrations/005_add_anonymous_voting.sql`:

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

- [ ] **Step 2: Update `db/schema.sql` to match (fresh installs)**

In the `survey` table block, after the `moderation_enabled` column:

```sql
moderation_enabled TINYINT(1) NOT NULL DEFAULT 1,
allow_anonymous TINYINT(1) NOT NULL DEFAULT 0,
```

In the `survey_participant` table block, change `user_id` to nullable and add `anon_session_id`:

```sql
CREATE TABLE IF NOT EXISTS survey_participant (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    survey_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED DEFAULT NULL,
    anon_session_id CHAR(36) DEFAULT NULL,
    role ENUM('participant', 'admin', 'moderator') NOT NULL DEFAULT 'participant',
    intake_data JSON DEFAULT NULL,
    privacy_choice ENUM('anonymous', 'public') DEFAULT NULL,
    invited_by INT UNSIGNED DEFAULT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL DEFAULT NULL,
    UNIQUE KEY uq_survey_user (survey_id, user_id),
    UNIQUE KEY uq_survey_anon (survey_id, anon_session_id),
    CONSTRAINT chk_participant_identity CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
    FOREIGN KEY (survey_id) REFERENCES survey(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES user(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

Similarly for `response`:

```sql
CREATE TABLE IF NOT EXISTS response (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    statement_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED DEFAULT NULL,
    anon_session_id CHAR(36) DEFAULT NULL,
    vote ENUM('agree', 'disagree', 'abstain') NOT NULL,
    is_important TINYINT(1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_statement_user (statement_id, user_id),
    UNIQUE KEY uq_statement_anon (statement_id, anon_session_id),
    CONSTRAINT chk_response_identity CHECK ((user_id IS NULL) <> (anon_session_id IS NULL)),
    FOREIGN KEY (statement_id) REFERENCES statement(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

- [ ] **Step 3: Apply the migration locally**

Apply via Adminer at `http://localhost:8081` (login with the credentials from `.env`, select the `deliberix` database, paste the migration into "SQL command" and execute) OR via mysql client:

```bash
docker exec -i deliberix-mariadb mysql -udeliberix -p${MARIADB_PASSWORD} deliberix < db/migrations/005_add_anonymous_voting.sql
```

Expected: three `ALTER TABLE` statements complete with no errors.

- [ ] **Step 4: Verify the schema**

```bash
docker exec deliberix-mariadb mysql -udeliberix -p${MARIADB_PASSWORD} -e "DESCRIBE deliberix.survey_participant; DESCRIBE deliberix.response; DESCRIBE deliberix.survey;" | grep -E "allow_anonymous|anon_session_id|user_id"
```

Expected: `user_id` is now `YES` for Null on both participant and response; `anon_session_id` appears on both; `allow_anonymous` appears on `survey`.

- [ ] **Step 5: Commit**

```bash
git add db/migrations/005_add_anonymous_voting.sql db/schema.sql
git commit -m "$(cat <<'EOF'
Migration 005: schema for anonymous voting

Adds survey.allow_anonymous flag, makes user_id nullable on
survey_participant and response, adds anon_session_id (UUID) on both,
plus XOR check constraints and unique keys for anon identity.
EOF
)"
```

---

## Task 2 — `identity.Actor` and anon-cookie helpers

**Files:**
- Create: `server/identity/actor.go`
- Create: `server/identity/anon_cookie.go`
- Modify: `server/identity/identity.go`

- [ ] **Step 1: Read the current identity package**

```bash
cat server/identity/identity.go
```

Expected: defines `User`, `CtxUserKey`, `ContextWithUser`, `GetUserFromContext`.

- [ ] **Step 2: Create `server/identity/actor.go`**

```go
package identity

import "context"

type Actor struct {
	UserID        *uint
	AnonSessionID *string
	User          *User
}

func (a *Actor) IsAnon() bool          { return a != nil && a.AnonSessionID != nil }
func (a *Actor) IsAuthenticated() bool { return a != nil && a.UserID != nil }

type actorKey struct{}

func ContextWithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

func GetActorFromContext(ctx context.Context) *Actor {
	a, _ := ctx.Value(actorKey{}).(*Actor)
	return a
}
```

- [ ] **Step 3: Create `server/identity/anon_cookie.go`**

```go
package identity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const AnonCookieName = "dlbx_anon"

func NewAnonSessionID() string {
	return uuid.NewString()
}

func SignAnonSessionID(id, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(id))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return id + "." + sig
}

func VerifyAnonSessionID(signed, secret string) (string, error) {
	dot := strings.LastIndex(signed, ".")
	if dot <= 0 || dot == len(signed)-1 {
		return "", errors.New("malformed cookie value")
	}
	id, sig := signed[:dot], signed[dot+1:]
	expected := base64.RawURLEncoding.EncodeToString(hmacOf(id, secret))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", errors.New("bad signature")
	}
	return id, nil
}

func hmacOf(id, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(id))
	return mac.Sum(nil)
}

// SetAnonCookie writes the signed cookie to the response.
// secureCookie should be true in production (HTTPS).
func SetAnonCookie(w http.ResponseWriter, signedValue string, secureCookie bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     AnonCookieName,
		Value:    signedValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 365,
	})
}

func ClearAnonCookie(w http.ResponseWriter, secureCookie bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     AnonCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// ReadAnonCookie returns the verified anon session ID from the request cookie,
// or empty string if missing/invalid.
func ReadAnonCookie(r *http.Request, secret string) string {
	c, err := r.Cookie(AnonCookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	id, err := VerifyAnonSessionID(c.Value, secret)
	if err != nil {
		return ""
	}
	return id
}
```

- [ ] **Step 4: Add uuid dependency**

```bash
cd server && go get github.com/google/uuid
```

Expected output ends with `go: added github.com/google/uuid vX.Y.Z`.

- [ ] **Step 5: Compile to verify**

```bash
cd server && go build ./...
```

Expected: no errors. Compiles clean.

- [ ] **Step 6: Commit**

```bash
git add server/identity/actor.go server/identity/anon_cookie.go server/go.mod server/go.sum
git commit -m "$(cat <<'EOF'
Add identity.Actor abstraction and anon-cookie helpers

Actor wraps either a UserID (JWT) or an AnonSessionID (cookie).
Cookie helpers sign and verify a UUID with HMAC-SHA256 keyed off
the existing JWT_SECRET. No DB session store: the cookie carries
the only authoritative reference.
EOF
)"
```

---

## Task 3 — Rename `Auth` → `AuthJWT`, add `AuthOptional`

**Files:**
- Modify: `server/middleware/auth.go`
- Modify: `server/main.go`

- [ ] **Step 1: Rename `Auth` to `AuthJWT` in `server/middleware/auth.go`**

Change the function signature from `func Auth(...)` to `func AuthJWT(...)`. Keep everything else identical (the body, `writeAuthError` helper, the verification logic, the email-verified gate).

Top of file should now read:

```go
func AuthJWT(secret string, s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ... unchanged body
		})
	}
}
```

Inside the body, where the actor needs to be attached, find:

```go
ctx := context.WithValue(r.Context(), identity.CtxUserKey, ident)
```

Add the Actor attachment alongside it:

```go
ctx := context.WithValue(r.Context(), identity.CtxUserKey, ident)
actor := &identity.Actor{UserID: &u.ID, User: u}
ctx = identity.ContextWithActor(ctx, actor)
```

(The `Actor.User` is the full DB user record, useful in handlers that needed `u` from `User`.)

- [ ] **Step 2: Add `AuthOptional` to the same file**

After the `AuthJWT` function, append:

```go
// AuthOptional tries Bearer JWT first; if absent or invalid, tries the
// anon cookie. If a valid identity is found it's attached to the request
// context; if neither is present the request continues with no Actor —
// handlers decide what to do.
func AuthOptional(secret string, s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try JWT first
			auth := r.Header.Get("Authorization")
			_, token, _ := strings.Cut(auth, "Bearer ")
			if token != "" {
				if userID, err := service.ValidateToken(token, secret); err == nil {
					if u, err := s.GetUserByID(r.Context(), userID); err == nil && u != nil {
						ident := &identity.User{
							ID:            u.ID,
							Role:          u.Role,
							EmailVerified: u.EmailVerifiedAt != nil,
						}
						ctx := context.WithValue(r.Context(), identity.CtxUserKey, ident)
						actor := &identity.Actor{UserID: &u.ID, User: u}
						ctx = identity.ContextWithActor(ctx, actor)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// Try anon cookie
			if anonID := identity.ReadAnonCookie(r, secret); anonID != "" {
				actor := &identity.Actor{AnonSessionID: &anonID}
				ctx := identity.ContextWithActor(r.Context(), actor)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No identity — pass through with no actor
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 3: Update the lone call site in `main.go`**

In `server/main.go`, change:

```go
r.Use(middleware.Auth(cfg.JWTSecret, s))
```

to:

```go
r.Use(middleware.AuthJWT(cfg.JWTSecret, s))
```

(Don't add `AuthOptional` routes yet — wiring comes in Task 8.)

- [ ] **Step 4: Build and verify**

```bash
cd server && go build ./...
```

Expected: clean compile.

- [ ] **Step 5: Commit**

```bash
git add server/middleware/auth.go server/main.go
git commit -m "$(cat <<'EOF'
Rename Auth middleware to AuthJWT and add AuthOptional

AuthJWT keeps the existing JWT-required semantics and now attaches a
*Actor (UserID + User) alongside the legacy User key. AuthOptional
accepts either a Bearer JWT or the dlbx_anon cookie; either fills in
the Actor, and a request with no credentials passes through with no
Actor for handlers to decide on.
EOF
)"
```

---

## Task 4 — Survey model + create/update for `allow_anonymous`

**Files:**
- Modify: `server/model/survey.go`
- Modify: `server/handler/survey.go`

- [ ] **Step 1: Add `AllowAnonymous` to the model**

In `server/model/survey.go`, in the `Survey` struct, after `ModerationEnabled`:

```go
AllowAnonymous bool `db:"allow_anonymous" json:"allowAnonymous"`
```

In `CreateSurveyRequest` and `UpdateSurveyRequest`, after their `ModerationEnabled` fields:

```go
AllowAnonymous *bool `json:"allowAnonymous,omitempty"`
```

- [ ] **Step 2: Wire defaults + persistence in `CreateSurvey` handler**

In `server/handler/survey.go` `CreateSurvey()`, locate the `moderationEnabled := true` block and add right after it:

```go
allowAnonymous := false
if in.AllowAnonymous != nil {
    allowAnonymous = *in.AllowAnonymous
}
```

In the survey struct literal a few lines down, after `ModerationEnabled: moderationEnabled,`:

```go
AllowAnonymous: allowAnonymous,
```

- [ ] **Step 3: Wire updates in `UpdateSurvey` handler**

Find the existing `in.ModerationEnabled` block:

```go
if in.ModerationEnabled != nil {
    if survey.Status != "draft" {
        return writeError(w, http.StatusBadRequest, "moderation_lock", "moderation_enabled can only be changed while survey is in draft")
    }
    fields["moderation_enabled"] = *in.ModerationEnabled
}
```

Add directly after it:

```go
if in.AllowAnonymous != nil {
    if survey.Status != "draft" {
        return writeError(w, http.StatusBadRequest, "anon_lock", "allow_anonymous can only be changed while survey is in draft")
    }
    fields["allow_anonymous"] = *in.AllowAnonymous
}
```

- [ ] **Step 4: Build and verify**

```bash
cd server && go build ./...
```

Expected: clean compile.

- [ ] **Step 5: Commit**

```bash
git add server/model/survey.go server/handler/survey.go
git commit -m "$(cat <<'EOF'
Add allow_anonymous to survey model and create/update handlers

Default false on creation. Update is rejected with anon_lock when
the survey has left draft, matching how moderation_enabled is locked
after activation.
EOF
)"
```

---

## Task 5 — Store: `*ByActor` helpers

**Files:**
- Modify: `server/store/survey.go`
- Modify: `server/store/participant.go`
- Modify: `server/store/response.go`
- Modify: `server/store/statement.go`

- [ ] **Step 1: Read existing helpers**

```bash
grep -n "GetParticipant\|IsParticipant\|GetUserVotesForSurvey\|GetVoteProgress\|CreateResponse\|CreateStatement" server/store/*.go
```

- [ ] **Step 2: Add Actor-aware lookups to `server/store/survey.go`**

Append after the existing `GetParticipant`:

```go
func (s *Store) GetParticipantByActor(ctx context.Context, surveyID uint, a *identity.Actor) (*model.SurveyParticipant, error) {
	if a == nil {
		return nil, nil
	}
	if a.UserID != nil {
		return s.GetParticipant(ctx, surveyID, *a.UserID)
	}
	if a.AnonSessionID != nil {
		return queryOne[model.SurveyParticipant](s.DB.Query(
			`SELECT * FROM survey_participant WHERE survey_id = ? AND anon_session_id = ?`,
			surveyID, *a.AnonSessionID,
		))
	}
	return nil, nil
}

func (s *Store) IsParticipantByActor(ctx context.Context, surveyID uint, a *identity.Actor) (bool, error) {
	p, err := s.GetParticipantByActor(ctx, surveyID, a)
	if err != nil {
		return false, err
	}
	return p != nil, nil
}
```

Add the import `"github.com/pdrhlik/deliberix/server/identity"` at the top of the file.

- [ ] **Step 3: Add anon participant creation**

In the same file, append:

```go
func (s *Store) CreateAnonParticipant(ctx context.Context, surveyID uint, anonSessionID string, intakeData *json.RawMessage) error {
	p := &model.SurveyParticipant{
		SurveyID:   surveyID,
		AnonSessionID: &anonSessionID,
		Role:       "participant",
		IntakeData: intakeData,
	}
	q := s.DB.Query(`INSERT INTO survey_participant ?values`, p)
	_, err := q.Exec()
	return err
}
```

Add the `encoding/json` import if missing.

Also adjust `model.SurveyParticipant` in `server/model/participant.go`:

- `UserID *uint` (was `uint`)
- Add `AnonSessionID *string` with db tag `anon_session_id` and json tag `anonSessionId,omitempty`

Open `server/model/participant.go` and update the struct accordingly. Save the file.

- [ ] **Step 4: Verify other handlers using `SurveyParticipant.UserID` directly**

```bash
grep -rn "\.UserID" server/handler/ server/store/ | grep -i participant | head -10
```

Any spot dereferencing `participant.UserID` as a plain `uint` needs to handle the pointer. Update to `*participant.UserID` after a `if participant.UserID != nil {` check. Be precise — only inside paths that know the participant is a user.

- [ ] **Step 5: Add `Response`/`Statement` ByActor variants in `server/store/response.go` and `server/store/statement.go`**

In `server/store/response.go`, append:

```go
func (s *Store) CreateResponseByActor(ctx context.Context, statementID uint, a *identity.Actor, vote string, isImportant bool) error {
	resp := &model.Response{
		StatementID: statementID,
		Vote:        vote,
		IsImportant: isImportant,
	}
	if a.UserID != nil {
		resp.UserID = a.UserID
	}
	if a.AnonSessionID != nil {
		resp.AnonSessionID = a.AnonSessionID
	}
	q := s.DB.Query(`INSERT INTO response ?values`, resp)
	_, err := q.Exec()
	return err
}

func (s *Store) GetUserVotesForSurveyByActor(ctx context.Context, surveyID uint, a *identity.Actor) (map[uint]model.UserVote, error) {
	q := s.DB.Query(`
		SELECT r.statement_id, r.vote, r.is_important
		FROM response r
		JOIN statement st ON st.id = r.statement_id
		WHERE st.survey_id = ?
		  AND ((r.user_id IS NOT NULL AND r.user_id = ?) OR
		       (r.anon_session_id IS NOT NULL AND r.anon_session_id = ?))`,
		surveyID, nullableUint(a.UserID), nullableString(a.AnonSessionID),
	)
	var rows []model.UserVote
	if err := q.All(&rows); err != nil {
		return nil, err
	}
	out := make(map[uint]model.UserVote, len(rows))
	for _, v := range rows {
		out[v.StatementID] = v
	}
	return out, nil
}

func (s *Store) GetVoteProgressByActor(ctx context.Context, surveyID uint, a *identity.Actor) (model.VoteProgress, error) {
	var p model.VoteProgress
	q := s.DB.Query(`
		SELECT
		  (SELECT COUNT(*) FROM statement WHERE survey_id = ? AND status = 'approved') AS total,
		  (SELECT COUNT(*) FROM response r
		     JOIN statement st ON st.id = r.statement_id
		     WHERE st.survey_id = ?
		       AND ((r.user_id IS NOT NULL AND r.user_id = ?) OR
		            (r.anon_session_id IS NOT NULL AND r.anon_session_id = ?))) AS voted`,
		surveyID, surveyID, nullableUint(a.UserID), nullableString(a.AnonSessionID),
	)
	if err := q.ScanRow(&p.Total, &p.Voted); err != nil {
		return p, err
	}
	return p, nil
}
```

Helper functions in the same file (or in `server/store/store.go`):

```go
func nullableUint(p *uint) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullableString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}
```

Also update `model.Response` and `model.UserVote` to support nullable user_id + anon_session_id:

In `server/model/response.go`, change `UserID uint` to `UserID *uint` and add `AnonSessionID *string`.

- [ ] **Step 6: Statement.AuthorID — already nullable**

`statement.author_id` is already nullable in the schema. The Go model `Statement.AuthorID *uint` already exists. For statement creation by anon, the handler will set `AuthorID = nil`. No store change needed for statements; just confirm by `grep -n "AuthorID" server/model/statement.go server/store/statement.go`.

- [ ] **Step 7: Build and fix all knock-on issues**

```bash
cd server && go build ./... 2>&1 | head -40
```

Any callsite that previously read `participant.UserID` or `response.UserID` as a plain `uint` will now error. Fix each by dereferencing inside a nil check.

Common touchpoints to check:
- `server/handler/results.go` (`progress.Voted`, `GetUserVotesForSurvey`)
- `server/handler/survey.go` (`GetMyParticipation`)
- `server/handler/moderation.go`
- `server/handler/participant.go` (listing)
- `server/store/survey.go` `ListSurveysByUser` (joins on `user_id`)

Update the SELECT and the unmarshal targets accordingly. Re-run the build until clean.

- [ ] **Step 8: Commit**

```bash
git add server/store/ server/model/
git commit -m "$(cat <<'EOF'
Store + model: actor-aware helpers and nullable identity

SurveyParticipant.UserID and Response.UserID become *uint;
AnonSessionID *string added on both. New *ByActor helpers branch on
which identity is non-nil, including CreateAnonParticipant and
CreateResponseByActor. nullableUint/nullableString helpers used in
queries that need to OR-match either identity.
EOF
)"
```

---

## Task 6 — Anon join + logout handlers

**Files:**
- Create: `server/handler/anon.go`
- Modify: `server/config/config.go` (add ProdMode/Secure flag if missing)

- [ ] **Step 1: Confirm a secure-cookie signal exists in config**

```bash
grep -n "BaseURL\|HTTPS\|Secure" server/config/config.go
```

If there's no straightforward "are we serving over HTTPS" flag, derive it from `BaseURL` in the handler — `strings.HasPrefix(h.Config.BaseURL, "https://")`.

- [ ] **Step 2: Create `server/handler/anon.go`**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pdrhlik/deliberix/server/identity"
)

func (h *Handler) secureCookies() bool {
	return strings.HasPrefix(h.Config.BaseURL, "https://")
}

func (h *Handler) AnonJoin() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		survey, err := h.getSurveyFromSlug(w, r)
		if err != nil {
			return err
		}
		if survey == nil {
			return nil
		}

		if !survey.AllowAnonymous {
			return writeError(w, http.StatusForbidden, "anon_not_allowed", "this survey does not accept anonymous voting")
		}
		if survey.Status != "active" {
			return writeError(w, http.StatusBadRequest, "survey_not_active", "survey is not active")
		}
		if isSurveyClosed(survey) {
			return writeError(w, http.StatusForbidden, "survey_closed", "survey has closed")
		}

		// If the caller already has a valid anon cookie for *this* survey, treat as conflict.
		if existing := identity.ReadAnonCookie(r, h.Config.JWTSecret); existing != "" {
			p, err := h.Store.GetParticipantByActor(r.Context(), survey.ID, &identity.Actor{AnonSessionID: &existing})
			if err != nil {
				return err
			}
			if p != nil {
				return writeError(w, http.StatusConflict, "already_a_participant", "already a participant")
			}
		}

		var body struct {
			IntakeData *json.RawMessage `json:"intakeData,omitempty"`
		}
		_ = parseJSON(r, &body) // body is optional

		sessionID := identity.NewAnonSessionID()
		if err := h.Store.CreateAnonParticipant(r.Context(), survey.ID, sessionID, body.IntakeData); err != nil {
			return err
		}

		identity.SetAnonCookie(w, identity.SignAnonSessionID(sessionID, h.Config.JWTSecret), h.secureCookies())

		return writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

func (h *Handler) AnonLogout() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		identity.ClearAnonCookie(w, h.secureCookies())
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
```

- [ ] **Step 3: Build and verify**

```bash
cd server && go build ./...
```

Expected: clean compile.

- [ ] **Step 4: Commit**

```bash
git add server/handler/anon.go
git commit -m "$(cat <<'EOF'
Handlers: POST /anon/join and /anon/logout

Join: server gates on survey.allow_anonymous, returns 403/409 in
the expected error cases, creates an anon participant row with
optional intake data, sets the signed cookie. Logout clears the
cookie.
EOF
)"
```

---

## Task 7 — Switch existing handlers to Actor-based identity

**Files:**
- Modify: `server/handler/survey.go` (GetSurvey, GetMyParticipation, JoinSurvey)
- Modify: `server/handler/statement.go` (ListStatements, GetNextStatement, SubmitStatement, AddSeedStatement stays JWT)
- Modify: `server/handler/response.go` (SubmitResponse, GetVoteProgress)
- Modify: `server/handler/results.go` (GetResults, GetSurveyStats)

For each handler that's been promoted to dual-auth (per § 1 of the spec), replace the JWT-only identity lookup with the unified Actor one.

- [ ] **Step 1: GetSurvey — accept anon**

In `server/handler/survey.go` `GetSurvey()`, the current code does:

```go
if survey.Visibility == "private" {
    user := identity.GetUserFromContext(r.Context())
    isParticipant, err := h.Store.IsParticipant(r.Context(), survey.ID, user.ID)
    ...
}
```

Replace with:

```go
if survey.Visibility == "private" {
    actor := identity.GetActorFromContext(r.Context())
    if actor == nil {
        return writeError(w, http.StatusNotFound, "survey_not_found", "survey not found")
    }
    isParticipant, err := h.Store.IsParticipantByActor(r.Context(), survey.ID, actor)
    if err != nil {
        return err
    }
    if !isParticipant {
        return writeError(w, http.StatusNotFound, "survey_not_found", "survey not found")
    }
}
```

- [ ] **Step 2: GetMyParticipation — accept anon**

Replace:

```go
user := identity.GetUserFromContext(r.Context())
p, err := h.Store.GetParticipant(r.Context(), survey.ID, user.ID)
```

with:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusNotFound, "not_a_participant", "not a participant")
}
p, err := h.Store.GetParticipantByActor(r.Context(), survey.ID, actor)
```

The rest of the handler is unchanged (404 if p == nil, writeJSON otherwise).

- [ ] **Step 3: JoinSurvey — JWT-only kept, but now coexists with /anon/join**

Leave the existing `JoinSurvey()` exactly as-is. Auth users still join via `/api/v1/survey/{slug}/join`; anon users use `/anon/join`. No code change required for `JoinSurvey`.

- [ ] **Step 4: SubmitStatement — accept anon**

In `server/handler/statement.go` `SubmitStatement()`, change the identity check:

```go
user := identity.GetUserFromContext(r.Context())
isParticipant, err := h.Store.IsParticipant(r.Context(), survey.ID, user.ID)
```

to:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant to submit statements")
}
isParticipant, err := h.Store.IsParticipantByActor(r.Context(), survey.ID, actor)
```

In the statement create struct, switch the author assignment:

```go
st := &model.Statement{
    SurveyID: survey.ID,
    Text:     in.Text,
    Type:     "user_submitted",
    Status:   status,
    AuthorID: actor.UserID, // nil for anon submitters
}
```

Anon-submitted statements end up with `author_id = NULL`. The DB schema already allows this.

- [ ] **Step 5: GetNextStatement — already participant-agnostic, switch the underlying lookup**

In `server/handler/statement.go` `GetNextStatement()`, the current call:

```go
user := identity.GetUserFromContext(r.Context())
st, err := h.Store.GetNextStatement(r.Context(), survey.ID, user.ID, survey.StatementOrder)
```

The `GetNextStatement` store method takes a `userID`. We need an Actor-aware variant or update the store query to OR-match on identity.

In `server/store/statement.go`, add:

```go
func (s *Store) GetNextStatementByActor(ctx context.Context, surveyID uint, a *identity.Actor, order string) (*model.Statement, error) {
	uid := nullableUint(a.UserID)
	aid := nullableString(a.AnonSessionID)
	// Find an approved statement the actor hasn't voted on yet, ordered by survey.statement_order.
	orderBy := "ORDER BY RAND()"
	switch order {
	case "sequential":
		orderBy = "ORDER BY id ASC"
	case "least_voted":
		orderBy = "ORDER BY (SELECT COUNT(*) FROM response r WHERE r.statement_id = statement.id) ASC, RAND()"
	}
	q := s.DB.Query(`
		SELECT *
		FROM statement
		WHERE survey_id = ? AND status = 'approved'
		  AND id NOT IN (
		    SELECT statement_id FROM response
		    WHERE (user_id IS NOT NULL AND user_id = ?)
		       OR (anon_session_id IS NOT NULL AND anon_session_id = ?)
		  )
		`+orderBy+`
		LIMIT 1`,
		surveyID, uid, aid,
	)
	return queryOne[model.Statement](q)
}
```

And in the handler, call:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
}
st, err := h.Store.GetNextStatementByActor(r.Context(), survey.ID, actor, survey.StatementOrder)
```

- [ ] **Step 6: ListStatements — already actor-agnostic**

`ListStatements` in `statement.go` doesn't filter by user identity. It returns all approved statements for the survey. No change.

- [ ] **Step 7: SubmitResponse — accept anon**

In `server/handler/response.go` `SubmitResponse()`, change:

```go
user := identity.GetUserFromContext(r.Context())
...
isParticipant, err := h.Store.IsParticipant(r.Context(), surveyID, user.ID)
```

to:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusForbidden, "must_be_participant", "must be a participant")
}
isParticipant, err := h.Store.IsParticipantByActor(r.Context(), surveyID, actor)
```

Replace the response creation:

```go
resp := &model.Response{
    StatementID: statementID,
    UserID:      &user.ID,
    Vote:        in.Vote,
    IsImportant: in.IsImportant,
}
if err := h.Store.CreateResponse(r.Context(), resp); err != nil {
    return writeError(w, http.StatusConflict, "already_voted", "already voted on this statement")
}
```

with:

```go
if err := h.Store.CreateResponseByActor(r.Context(), statementID, actor, in.Vote, in.IsImportant); err != nil {
    return writeError(w, http.StatusConflict, "already_voted", "already voted on this statement")
}
```

Return value is no longer `resp` since the helper doesn't return it. Change `writeJSON(w, http.StatusCreated, resp)` to `writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})`. Verify the Angular client doesn't depend on the response body (it doesn't — `ResponseService.submitResponse` ignores it).

- [ ] **Step 8: GetVoteProgress — accept anon**

In `server/handler/response.go` `GetVoteProgress()`:

```go
user := identity.GetUserFromContext(r.Context())
progress, err := h.Store.GetVoteProgress(r.Context(), survey.ID, user.ID)
```

becomes:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
}
progress, err := h.Store.GetVoteProgressByActor(r.Context(), survey.ID, actor)
```

- [ ] **Step 9: GetResults — accept anon**

In `server/handler/results.go` `GetResults()`, switch the same pattern:

```go
user := identity.GetUserFromContext(r.Context())
...
progress, err := h.Store.GetVoteProgress(r.Context(), survey.ID, user.ID)
...
myVotes, err := h.Store.GetUserVotesForSurvey(r.Context(), survey.ID, user.ID)
```

becomes:

```go
actor := identity.GetActorFromContext(r.Context())
if actor == nil {
    return writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
}
...
progress, err := h.Store.GetVoteProgressByActor(r.Context(), survey.ID, actor)
...
myVotes, err := h.Store.GetUserVotesForSurveyByActor(r.Context(), survey.ID, actor)
```

- [ ] **Step 10: GetSurveyStats — no per-user filter, no change**

`GetSurveyStats` returns aggregate stats; no per-user filter. No change needed.

- [ ] **Step 11: Build, fix, build until clean**

```bash
cd server && go build ./...
```

Expected: clean compile.

- [ ] **Step 12: Commit**

```bash
git add server/handler/
git commit -m "$(cat <<'EOF'
Handlers: dual-auth Actor-based identity for shared endpoints

GetSurvey, GetMyParticipation, GetNextStatement, SubmitStatement,
SubmitResponse, GetVoteProgress, and GetResults now read the unified
Actor from context and call *ByActor store helpers. Anon-submitted
statements get author_id = NULL. Admin endpoints (CreateSurvey,
moderation, participants admin, etc.) keep their JWT-only flow.
EOF
)"
```

---

## Task 8 — Wire route groups in `main.go` + CORS

**Files:**
- Modify: `server/main.go`

- [ ] **Step 1: Inspect current route layout**

```bash
sed -n '50,120p' server/main.go
```

- [ ] **Step 2: Update CORS for credentials**

Replace the existing `cors.Options` block:

```go
r.Use(cors.Handler(cors.Options{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
}))
```

with:

```go
allowedOrigins := []string{cfg.BaseURL}
if cfg.BaseURL == "" {
    // Dev default — allow the Angular dev server origin
    allowedOrigins = []string{"http://localhost:4200"}
}
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   allowedOrigins,
    AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
}))
```

The CORS spec forbids `AllowedOrigins: ["*"]` together with `AllowCredentials: true`, so the wildcard must drop to an explicit origin.

- [ ] **Step 3: Re-organize the routes**

Find the existing route block and restructure as follows:

```go
// Public — no auth required
r.Post("/api/v1/auth/register", handler.ErrorHandler(h.Register()))
r.Post("/api/v1/auth/login", handler.ErrorHandler(h.Login()))
r.Post("/api/v1/auth/verify-email", handler.ErrorHandler(h.VerifyEmail()))
r.Post("/api/v1/auth/magic-link", handler.ErrorHandler(h.RequestMagicLink()))
r.Post("/api/v1/auth/magic-link/verify", handler.ErrorHandler(h.VerifyMagicLink()))
r.Post("/api/v1/auth/forgot-password", handler.ErrorHandler(h.ForgotPassword()))
r.Post("/api/v1/auth/reset-password", handler.ErrorHandler(h.ResetPassword()))
r.Post("/api/v1/survey/{slug}/anon/join", handler.ErrorHandler(h.AnonJoin()))
r.Post("/api/v1/anon/logout", handler.ErrorHandler(h.AnonLogout()))

// JWT-only (admin + user-management)
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthJWT(cfg.JWTSecret, s))

    r.Get("/api/v1/auth/me", handler.ErrorHandler(h.Me()))
    r.Patch("/api/v1/auth/me", handler.ErrorHandler(h.UpdateProfile()))
    r.Post("/api/v1/auth/change-password", handler.ErrorHandler(h.ChangePassword()))
    r.Post("/api/v1/auth/resend-verification", handler.ErrorHandler(h.ResendVerification()))

    r.Get("/api/v1/survey", handler.ErrorHandler(h.ListSurveys()))
    r.Post("/api/v1/survey", handler.ErrorHandler(h.CreateSurvey()))
    r.Get("/api/v1/survey/public", handler.ErrorHandler(h.ListPublicSurveys()))
    r.Patch("/api/v1/survey/{slug}", handler.ErrorHandler(h.UpdateSurvey()))
    r.Delete("/api/v1/survey/{slug}", handler.ErrorHandler(h.DeleteSurvey()))
    r.Post("/api/v1/survey/{slug}/join", handler.ErrorHandler(h.JoinSurvey()))

    r.Get("/api/v1/survey/{slug}/participants", handler.ErrorHandler(h.ListParticipants()))
    r.Patch("/api/v1/survey/{slug}/participant/{userId}/role", handler.ErrorHandler(h.UpdateParticipantRole()))
    r.Delete("/api/v1/survey/{slug}/participant/{userId}", handler.ErrorHandler(h.RemoveParticipant()))

    r.Post("/api/v1/survey/{slug}/statement/seed", handler.ErrorHandler(h.AddSeedStatement()))

    r.Get("/api/v1/survey/{slug}/moderation", handler.ErrorHandler(h.GetModerationQueue()))
    r.Patch("/api/v1/statement/{id}/moderate", handler.ErrorHandler(h.ModerateStatement()))
})

// Dual-auth (JWT OR anon cookie)
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthOptional(cfg.JWTSecret, s))

    r.Get("/api/v1/survey/{slug}", handler.ErrorHandler(h.GetSurvey()))
    r.Get("/api/v1/survey/{slug}/participant/me", handler.ErrorHandler(h.GetMyParticipation()))
    r.Get("/api/v1/survey/{slug}/statement", handler.ErrorHandler(h.ListStatements()))
    r.Get("/api/v1/survey/{slug}/statement/next", handler.ErrorHandler(h.GetNextStatement()))
    r.Post("/api/v1/survey/{slug}/statement", handler.ErrorHandler(h.SubmitStatement()))
    r.Post("/api/v1/statement/{id}/response", handler.ErrorHandler(h.SubmitResponse()))
    r.Get("/api/v1/survey/{slug}/results", handler.ErrorHandler(h.GetResults()))
    r.Get("/api/v1/survey/{slug}/stats", handler.ErrorHandler(h.GetSurveyStats()))
    r.Get("/api/v1/survey/{slug}/progress", handler.ErrorHandler(h.GetVoteProgress()))
})
```

- [ ] **Step 4: Build and verify**

```bash
cd server && go build ./...
```

Expected: clean compile.

- [ ] **Step 5: Smoke test the build**

```bash
docker-compose -f docker-compose-dev.yml restart server
docker-compose -f docker-compose-dev.yml logs --tail=30 server
```

Expected: log line `server listening on :8080`, no panics.

- [ ] **Step 6: Commit**

```bash
git add server/main.go
git commit -m "$(cat <<'EOF'
Wire route groups for dual-auth + public anon endpoints

Public: /auth/*, /survey/{slug}/anon/join, /anon/logout.
JWT-only: survey CRUD, participants admin, seed statements,
moderation, auth/me. Dual-auth: GetSurvey, participant/me,
statements list/next/submit, responses, results, stats, progress.

CORS: drop AllowedOrigins=* in favor of an explicit origin
(cfg.BaseURL or http://localhost:4200 in dev) and enable
AllowCredentials so the anon cookie can travel cross-origin in
development.
EOF
)"
```

---

## Task 9 — Client foundation: withCredentials + Survey model + surveyAccessGuard

**Files:**
- Modify: `client/src/app/services/api.service.ts`
- Modify: `client/src/app/models/survey.model.ts`
- Create: `client/src/app/guards/survey-access.guard.ts`
- Modify: `client/src/app/app.routes.ts`

- [ ] **Step 1: Send credentials with all API calls**

In `client/src/app/services/api.service.ts`, change every HTTP call to pass `{ withCredentials: true }`:

```typescript
import { HttpClient } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { environment } from "../../environments/environment";

@Injectable({
  providedIn: "root",
})
export class ApiService {
  private http = inject(HttpClient);
  private baseUrl = environment.apiUrl + "/api/v1";
  private opts = { withCredentials: true } as const;

  get<T>(path: string): Observable<T> {
    return this.http.get<T>(`${this.baseUrl}${path}`, this.opts);
  }

  post<T>(path: string, body: any): Observable<T> {
    return this.http.post<T>(`${this.baseUrl}${path}`, body, this.opts);
  }

  patch<T>(path: string, body: any): Observable<T> {
    return this.http.patch<T>(`${this.baseUrl}${path}`, body, this.opts);
  }

  delete<T>(path: string): Observable<T> {
    return this.http.delete<T>(`${this.baseUrl}${path}`, this.opts);
  }
}
```

- [ ] **Step 2: Add `allowAnonymous` to the Survey model**

In `client/src/app/models/survey.model.ts`, in `Survey`, after `moderationEnabled`:

```typescript
allowAnonymous: boolean;
```

In `CreateSurveyRequest` and `UpdateSurveyRequest`, after `moderationEnabled?`:

```typescript
allowAnonymous?: boolean;
```

- [ ] **Step 3: Create the survey-access guard**

`client/src/app/guards/survey-access.guard.ts`:

```typescript
import { inject } from "@angular/core";
import { CanActivateFn, Router } from "@angular/router";
import { firstValueFrom } from "rxjs";
import { ApiService } from "../services/api.service";
import { AuthService } from "../services/auth.service";

export const surveyAccessGuard: CanActivateFn = async (route) => {
  const auth = inject(AuthService);
  const api = inject(ApiService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  const slug = route.paramMap.get("slug");
  if (!slug) {
    return router.parseUrl("/login");
  }

  try {
    // Dual-auth endpoint; succeeds if the survey is reachable anonymously
    // (allow_anonymous or visibility permits) or if the anon cookie is valid.
    await firstValueFrom(api.get(`/survey/${slug}`));
    return true;
  } catch {
    return router.parseUrl(`/login?redirect=${encodeURIComponent("/survey/" + slug)}`);
  }
};
```

- [ ] **Step 4: Update `app.routes.ts` to use the new guard on `/survey/:slug`**

Find the existing `{ path: "survey/:slug", canActivate: [authGuard], ... }` entry. Replace `authGuard` with `surveyAccessGuard`:

```typescript
{
  path: "survey/:slug",
  canActivate: [surveyAccessGuard],
  loadComponent: () =>
    import("./pages/survey-detail/survey-detail.page").then((m) => m.SurveyDetailPage),
},
```

Import the new guard at the top of the file:

```typescript
import { surveyAccessGuard } from "./guards/survey-access.guard";
```

Keep the existing `authGuard` import — it stays on every other survey route (the vote page, join page, etc., which require an active session of some kind; the participant/me check inside handles anon).

For the vote, join, results, moderation routes that begin with `survey/:slug/...`, they still need *some* form of identity — JWT or anon cookie. The dual-auth endpoints will return errors if neither. So we replace `authGuard` with `surveyAccessGuard` on those too, with the same logic.

Update every route under `survey/:slug` that currently has `canActivate: [authGuard]`:
- `survey/:slug/results`
- `survey/:slug/moderation`
- `survey/:slug/vote`
- `survey/:slug/join`

Leave `authGuard` on the truly user-only routes: `surveys`, `profile`, `survey/create`.

- [ ] **Step 5: Build and verify**

```bash
cd client && npx ng build --configuration=development
```

Expected: clean build (only pre-existing warnings).

- [ ] **Step 6: Commit**

```bash
git add client/src/app/services/api.service.ts client/src/app/models/survey.model.ts client/src/app/guards/survey-access.guard.ts client/src/app/app.routes.ts
git commit -m "$(cat <<'EOF'
Client foundation for anon voting

ApiService sends withCredentials on every request so the anon cookie
travels cross-origin in dev. Survey model gains allowAnonymous.
surveyAccessGuard replaces authGuard on every survey/:slug* route:
authenticated users pass through, anonymous ones are allowed only
if GET /survey/{slug} succeeds (dual-auth endpoint), otherwise sent
to /login with a redirect.
EOF
)"
```

---

## Task 10 — Client: Anon join UX in survey-detail + "End anonymous session"

**Files:**
- Modify: `client/src/app/services/survey.service.ts`
- Modify: `client/src/app/pages/survey-detail/survey-detail.page.ts`
- Modify: `client/src/app/pages/survey-detail/survey-detail.page.html`
- Modify: `client/src/app/services/auth.service.ts`

- [ ] **Step 1: Add anon join + logout to SurveyService**

In `client/src/app/services/survey.service.ts`, add:

```typescript
async joinAnonymously(slug: string, intakeData?: unknown): Promise<void> {
  await firstValueFrom(
    this.api.post(`/survey/${slug}/anon/join`, { intakeData: intakeData ?? null }),
  );
}

async anonLogout(): Promise<void> {
  await firstValueFrom(this.api.post(`/anon/logout`, {}));
}
```

- [ ] **Step 2: Survey-detail TS additions**

In `client/src/app/pages/survey-detail/survey-detail.page.ts`:

Add a `joiningAnon = signal(false)` next to the existing `joining = signal(false)`.

Add the `joinAnon()` method (handles intake routing exactly like the existing `joinSurvey()` does for auth users):

```typescript
async joinAnon() {
  const s = this.survey();
  if (!s) return;

  if (s.intakeConfig?.fields?.length) {
    // Reuse the existing intake flow but route through anon-join.
    // Navigate to a dedicated join page in anon mode.
    this.router.navigateByUrl(`/survey/${s.slug}/join?anon=1`);
    return;
  }

  this.joiningAnon.set(true);
  try {
    await this.surveyService.joinAnonymously(s.slug);
    await this.loadSurvey(s.slug);
  } catch (e) {
    this.toast.apiError(e);
  } finally {
    this.joiningAnon.set(false);
  }
}

async endAnonSession() {
  await this.surveyService.anonLogout();
  window.location.reload();
}
```

- [ ] **Step 3: Survey-detail HTML — anon join CTA**

In `client/src/app/pages/survey-detail/survey-detail.page.html`, find the existing `<!-- Join button for non-participants -->` block:

```html
@if (!isParticipantOrAdmin && s.status === "active") {
<div class="content-card">
  <ion-button expand="block" (click)="joinSurvey()" [disabled]="joining()">
    ...
  </ion-button>
</div>
}
```

Replace with:

```html
@if (!isParticipantOrAdmin && s.status === "active") {
<div class="content-card">
  @if (auth.isAuthenticated()) {
    <ion-button expand="block" (click)="joinSurvey()" [disabled]="joining()">
      @if (joining()) {
        <ion-spinner name="crescent"></ion-spinner>
      } @else {
        {{ "survey.join" | translate }}
      }
    </ion-button>
  } @else if (s.allowAnonymous) {
    <ion-button expand="block" (click)="joinAnon()" [disabled]="joiningAnon()">
      @if (joiningAnon()) {
        <ion-spinner name="crescent"></ion-spinner>
      } @else {
        {{ "survey.join-anon" | translate }}
      }
    </ion-button>
    <p class="auth-link">
      <a [routerLink]="['/login']" [queryParams]="{ redirect: '/survey/' + s.slug }">
        {{ "survey.join-anon-sign-in-instead" | translate }}
      </a>
    </p>
  } @else {
    <p class="auth-link">
      <a [routerLink]="['/login']" [queryParams]="{ redirect: '/survey/' + s.slug }">
        {{ "survey.join-sign-in-required" | translate }}
      </a>
    </p>
  }
</div>
}
```

(Make sure `RouterLink` is in the component imports — it likely is already; if not, add `RouterLink` to the imports array.)

- [ ] **Step 4: End-anon-session affordance for joined anon participants**

In the same HTML, locate where participant-only content appears in the overview. After the vote-progress block, add:

```html
@if (isAnonParticipant()) {
  <p class="anon-end-session">
    <a (click)="endAnonSession()">{{ "survey.anon-end-session" | translate }}</a>
  </p>
}
```

In the TS, add the computed signal:

```typescript
isAnonParticipant(): boolean {
  // anon participants have no user_id; the participant record has anonSessionId set
  const p = this.participant();
  return !!p && !!p.anonSessionId && !this.auth.isAuthenticated();
}
```

And update `SurveyParticipant` model at `client/src/app/models/participant.model.ts` to include the optional anonymous field (the server sends `userId` nullable and `anonSessionId` when anon):

```typescript
export interface SurveyParticipant {
  id: number;
  surveyId: number;
  userId?: number | null;
  anonSessionId?: string | null;
  role: "participant" | "admin" | "moderator";
  intakeData?: any;
  joinedAt: string;
}
```

- [ ] **Step 5: Update the survey-join page to handle `?anon=1`**

The current `survey-join.page` POSTs to `/survey/{slug}/join` (the JWT-only endpoint). When `anon=1` is set, it should POST to `/survey/{slug}/anon/join` instead. In `client/src/app/pages/survey-join/survey-join.page.ts`:

Read the query param in the page constructor / `ngOnInit`:

```typescript
isAnonJoin = signal(false);

ngOnInit() {
  this.isAnonJoin.set(this.route.snapshot.queryParamMap.get("anon") === "1");
  // ... existing setup
}
```

And in the submit handler, branch:

```typescript
async submit() {
  // ... existing validation
  try {
    if (this.isAnonJoin()) {
      await this.surveyService.joinAnonymously(this.slug, this.intakeData);
    } else {
      await firstValueFrom(this.api.post(`/survey/${this.slug}/join`, { intakeData: this.intakeData }));
    }
    this.router.navigateByUrl(`/survey/${this.slug}`, { replaceUrl: true });
  } catch (e) {
    this.toast.apiError(e);
  }
}
```

(Inspect the current `survey-join.page.ts` before editing — keep field names intact.)

- [ ] **Step 6: Build, verify, and smoke-test in browser**

```bash
cd client && npx ng build --configuration=development
```

Manual smoke test:
1. Log out completely (or use an incognito window).
2. Create a draft survey from a logged-in account, enable "Allow anonymous voting" (Task 12 covers the UI; skip if you're testing this in order — for now, set the flag via Adminer: `UPDATE survey SET allow_anonymous = 1 WHERE slug = '<your-test-survey>';`).
3. Activate the survey.
4. In a logged-out browser, open `/survey/<slug>`. You should see the "Join anonymously" button.
5. Click it. A toast should NOT appear; the page should reload showing the vote progress.
6. Vote on a statement. Verify the vote registers.
7. Refresh the page — you should still be a participant (cookie persists).
8. Click "End anonymous session" — page reloads to the not-yet-joined state.

- [ ] **Step 7: Commit**

```bash
git add client/src/app/services/survey.service.ts client/src/app/pages/survey-detail client/src/app/pages/survey-join client/src/app/models/participant.model.ts
git commit -m "$(cat <<'EOF'
Client UX: anonymous join flow + end-session affordance

Survey detail renders three CTAs based on auth state and the
survey's allow_anonymous flag: logged-in users see the existing
join, anon-allowed surveys offer an anonymous join (with intake if
configured) plus a "sign in instead" link, and anon-disallowed
surveys offer "sign in to participate". Joined anon participants
see an "End anonymous session" link that clears the cookie. The
join page accepts ?anon=1 to route the submit through the anon
endpoint.
EOF
)"
```

---

## Task 11 — Client: `allow_anonymous` toggle in survey-create and survey-detail Settings

**Files:**
- Modify: `client/src/app/pages/survey-create/survey-create.page.ts`
- Modify: `client/src/app/pages/survey-create/survey-create.page.html`
- Modify: `client/src/app/pages/survey-detail/survey-detail.page.ts`
- Modify: `client/src/app/pages/survey-detail/survey-detail.page.html`

- [ ] **Step 1: Survey-create page state**

In `survey-create.page.ts`, after `moderationEnabled = true;`:

```typescript
allowAnonymous = false;
```

Include it in `onSubmit` body:

```typescript
moderationEnabled: this.moderationEnabled,
allowAnonymous: this.allowAnonymous,
```

- [ ] **Step 2: Survey-create page HTML toggle**

In `survey-create.page.html`, find the existing `moderationEnabled` `<ion-toggle>` block and add directly after it:

```html
<ion-toggle
  [(ngModel)]="allowAnonymous"
  name="allowAnonymous"
  labelPlacement="start"
  justify="space-between"
>
  <div>
    <div>{{ "survey.allow-anonymous" | translate }}</div>
    <ion-note color="medium" class="toggle-hint">
      {{ "survey.allow-anonymous-help" | translate }}
    </ion-note>
  </div>
</ion-toggle>
```

- [ ] **Step 3: Survey-detail Settings tab — state**

In `survey-detail.page.ts`, near `editModerationEnabled = signal(true);`:

```typescript
editAllowAnonymous = signal(false);
```

In `loadSurvey`, after setting `editModerationEnabled`:

```typescript
this.editAllowAnonymous.set(survey.allowAnonymous);
```

In `saveSettings`, after `moderationEnabled: this.editModerationEnabled(),`:

```typescript
allowAnonymous: this.editAllowAnonymous(),
```

- [ ] **Step 4: Survey-detail Settings tab — HTML toggle (draft form + readonly)**

In `survey-detail.page.html`, in the draft-form block find the existing moderation `<ion-toggle>` and add right after it:

```html
<ion-toggle
  [checked]="editAllowAnonymous()"
  (ionChange)="editAllowAnonymous.set($event.detail.checked)"
  labelPlacement="start"
  justify="space-between"
>
  <div>
    <div>{{ "survey.allow-anonymous" | translate }}</div>
    <ion-note color="medium" class="toggle-hint">
      {{ "survey.allow-anonymous-help" | translate }}
    </ion-note>
  </div>
</ion-toggle>
```

In the read-only settings block, after the moderation readonly field add:

```html
<div class="readonly-field">
  <span class="readonly-label">{{ "survey.allow-anonymous" | translate }}</span>
  <span class="readonly-value">
    @if (s.allowAnonymous) {
      {{ "survey.allow-anonymous-on" | translate }}
    } @else {
      {{ "survey.allow-anonymous-off" | translate }}
    }
  </span>
</div>
```

- [ ] **Step 5: Build, verify, smoke-test**

```bash
cd client && npx ng build --configuration=development
```

Smoke test: Create a new survey with "Allow anonymous voting" enabled. Verify the toggle persists after save and shows on the Settings readonly view post-activation.

- [ ] **Step 6: Commit**

```bash
git add client/src/app/pages/survey-create client/src/app/pages/survey-detail
git commit -m "$(cat <<'EOF'
UI: allow_anonymous toggle in survey-create and detail Settings

Toggle exposed in the create page's Advanced settings accordion and
in the detail page's Settings tab (draft only, locked after
activation). Readonly section shows the current state when the
survey has left draft.
EOF
)"
```

---

## Task 12 — Admin views: participants list + moderation queue handle anon

**Files:**
- Modify: `server/store/participant.go` (list query joins user)
- Modify: `server/handler/participant.go` (response shape includes anon flag)
- Modify: `client/src/app/components/participants/participants.component.ts` and `.html`
- Modify: `client/src/app/pages/survey-detail/survey-detail.page.html` (moderation queue rendering)

- [ ] **Step 1: Inspect current participant listing**

```bash
grep -A 12 "ListParticipants" server/store/participant.go server/handler/participant.go
grep -A 10 "participant" client/src/app/components/participants/participants.component.html
```

- [ ] **Step 2: Backend — list participants includes anon**

The existing `ListParticipants` likely does:

```sql
SELECT sp.*, u.name, u.email
FROM survey_participant sp
JOIN user u ON u.id = sp.user_id
WHERE sp.survey_id = ?
```

Change `JOIN` to `LEFT JOIN` and accept null name/email:

```sql
SELECT sp.id, sp.survey_id, sp.user_id, sp.anon_session_id, sp.role, sp.joined_at,
       COALESCE(u.name, '') AS name, COALESCE(u.email, '') AS email
FROM survey_participant sp
LEFT JOIN user u ON u.id = sp.user_id
WHERE sp.survey_id = ?
ORDER BY sp.joined_at ASC
```

Adjust the row scan target accordingly.

In `server/model/participant.go`, the `ParticipantWithUser` (or whatever the listing response shape is named) gets an `AnonSessionID *string` field too.

- [ ] **Step 3: Client — participants component renders anon row**

In `participants.component.html`, where the list renders each participant:

```html
@for (p of participants(); track p.id) {
  <ion-item>
    <ion-label>
      <h3>
        @if (p.userId) {
          {{ p.name }}
        } @else {
          {{ "participants.anon-label" | translate }}
        }
      </h3>
      @if (p.userId) {
        <p>{{ p.email }}</p>
      } @else {
        <p>{{ "participants.joined-at" | translate }}: {{ p.joinedAt | date:"medium" }}</p>
      }
    </ion-label>
    <ion-badge slot="end">{{ "participants.role-" + p.role | translate }}</ion-badge>
  </ion-item>
}
```

Inspect the actual current template before editing — the goal is conditional rendering keyed on whether `userId` is present.

- [ ] **Step 4: Moderation queue — anon-authored statements**

The moderation queue in `survey-detail.page.html` renders pending statements. Since `statement.author_id` can already be NULL (when a user is deleted), the existing UI may already handle nulls. If it doesn't show an explicit "Anonymous author" label, add one.

Inspect `client/src/app/services/moderation.service.ts` and the `Statement` model to confirm `authorId?: number | null` is already present; if the server now also returns `authorId: null` for anon submissions, the front-end behavior is unchanged unless we want to surface "Anonymous" explicitly. For v1, no extra work needed beyond confirming nulls render cleanly.

- [ ] **Step 5: Build, verify, smoke-test**

```bash
cd server && go build ./...
cd ../client && npx ng build --configuration=development
```

Smoke test:
1. Create a survey with anon voting enabled.
2. Have an anon visitor join + vote.
3. As admin, open Participants tab → verify the anon row shows as "Anonymous (joined …)".
4. Have the anon visitor submit a statement.
5. As admin, open Moderation Queue tab → verify the statement appears without an author name.

- [ ] **Step 6: Commit**

```bash
git add server/store/participant.go server/handler/participant.go server/model/participant.go client/src/app/components/participants
git commit -m "$(cat <<'EOF'
Admin views adapt to anonymous participants

ListParticipants uses LEFT JOIN and returns anon_session_id +
optional name/email. The participants UI renders an 'Anonymous'
label and the join timestamp for anon rows. Moderation queue
already tolerated null author_id; no change needed there beyond
confirming the existing render handles it.
EOF
)"
```

---

## Task 13 — i18n keys (cs + en)

**Files:**
- Modify: `client/src/assets/i18n/cs.json`
- Modify: `client/src/assets/i18n/en.json`

- [ ] **Step 1: Add Czech keys**

In `cs.json` under the `survey.*` block, add:

```json
"join-anon": "Připojit se anonymně",
"join-anon-sign-in-instead": "Nebo se přihlásit",
"join-sign-in-required": "Pro připojení se přihlaste",
"anon-end-session": "Ukončit anonymní hlasování",
"allow-anonymous": "Povolit anonymní hlasování přes odkaz",
"allow-anonymous-help": "Účastníci mohou hlasovat bez registrace, pokud znají odkaz nebo naskenují QR kód. Nastavení lze měnit pouze před aktivací průzkumu.",
"allow-anonymous-on": "Povoleno — anonymní hlasování přes odkaz",
"allow-anonymous-off": "Zakázáno — pouze přihlášení uživatelé"
```

Under `errors.*`:

```json
"anon_not_allowed": "Tento průzkum neumožňuje anonymní hlasování.",
"anon_lock": "Anonymní hlasování lze měnit pouze před aktivací průzkumu."
```

Under `participants.*` (create the block if it doesn't have these):

```json
"anon-label": "Anonymní účastník",
"joined-at": "Připojen(a)"
```

- [ ] **Step 2: Add English keys**

Same keys in `en.json`:

```json
"join-anon": "Join anonymously",
"join-anon-sign-in-instead": "Or sign in",
"join-sign-in-required": "Sign in to participate",
"anon-end-session": "End anonymous session",
"allow-anonymous": "Allow anonymous voting via link",
"allow-anonymous-help": "Participants can vote without signing up if they have the link or scan the QR code. Can only be changed before the survey is activated.",
"allow-anonymous-on": "Enabled — anonymous voting via link",
"allow-anonymous-off": "Disabled — signed-in users only"
```

Errors:

```json
"anon_not_allowed": "This survey does not accept anonymous voting.",
"anon_lock": "Anonymous voting can only be changed before the survey is activated."
```

Participants:

```json
"anon-label": "Anonymous participant",
"joined-at": "Joined"
```

- [ ] **Step 3: Validate JSON and commit**

```bash
python3 -c "import json; json.load(open('client/src/assets/i18n/cs.json')); json.load(open('client/src/assets/i18n/en.json')); print('OK')"
```

Expected: `OK`.

```bash
git add client/src/assets/i18n/cs.json client/src/assets/i18n/en.json
git commit -m "$(cat <<'EOF'
i18n: anonymous voting strings (cs + en)

survey.join-anon, allow-anonymous (+ help + readonly variants),
anon-end-session, join-sign-in-required, participants.anon-label,
errors.anon_not_allowed, errors.anon_lock.
EOF
)"
```

---

## Task 14 — Manual end-to-end smoke test

This task has no code changes; verify the feature end-to-end in the browser.

- [ ] **Step 1: Restart the stack with the new code**

```bash
docker-compose -f docker-compose-dev.yml up -d --build server client
```

Watch logs for clean startup.

- [ ] **Step 2: Authenticated admin path**

1. Log in as an existing user.
2. Create a new survey. In Advanced settings, enable "Allow anonymous voting".
3. Add ≥ 3 seed statements.
4. Activate the survey.
5. Open the Share modal (icon top-right of the survey-detail page). Copy the link.

- [ ] **Step 3: Anonymous voter path (incognito window)**

1. Open the copied link in an incognito window.
2. Page should load showing the survey title and a "Join anonymously" button.
3. Click "Join anonymously". If the survey has intake, fill in.
4. Should land back on the survey-detail page with vote progress visible.
5. Open the vote page and cast votes on each statement. Verify each vote sticks (refresh proves cookie persistence).
6. Verify that the "Show results" tab respects `result_visibility` exactly the same way as for an authenticated user.

- [ ] **Step 4: End anonymous session**

1. On the survey-detail page (in the incognito session), click "End anonymous session".
2. The page reloads to the "not yet a participant" state.
3. Click "Join anonymously" again. The user gets a fresh session — votes from the previous session are still in the DB but invisible to this new session.

- [ ] **Step 5: Anon submits a statement (with moderation enabled)**

1. From the anon session, submit a new statement (overview tab after voting all, or wherever the submit UI surfaces).
2. As admin, open Moderation Queue. The new statement appears with no author name.
3. Approve it. Verify it shows up in the next-vote queue for future participants.

- [ ] **Step 6: Anon participant in admin list**

1. As admin, open Participants tab.
2. Verify the anon participant appears as "Anonymous participant" with a join timestamp.
3. Remove them. Verify the cascade clears their votes.

- [ ] **Step 7: Backward compatibility — surveys without allow_anonymous**

1. Open the share link of a survey that was created before this feature (allow_anonymous = 0).
2. In an incognito window: the survey-detail page should redirect to `/login` (or show the "sign in to participate" prompt depending on visibility).
3. Logged-in users continue to join as before — no behavior change.

- [ ] **Step 8: Note any defects in the plan checkboxes**

If something doesn't work as expected, document it as a bullet under "Known issues" in this plan and fix in the next iteration.

---

## Task 15 — Production deploy

This is the operational checklist for rolling out to deliberix.com.

- [ ] **Step 1: Commit and push your branch**

Verify all 12 implementation commits are on the branch:

```bash
git log --oneline master | head -15
```

Push to remote (rebase first if upstream advanced):

```bash
git pull --rebase origin master
git push origin master
```

- [ ] **Step 2: SSH to the prod host and pull**

```bash
cd ~/projects/deliberix
git pull origin master
```

- [ ] **Step 3: Apply the DB migration**

Production runs MariaDB with persistent volume — `schema.sql` is NOT re-applied. Apply the migration explicitly:

```bash
docker exec -i deliberix-mariadb mysql -udeliberix -p"$(grep MARIADB_PASSWORD .env | cut -d= -f2)" deliberix < db/migrations/005_add_anonymous_voting.sql
```

Verify:

```bash
docker exec deliberix-mariadb mysql -udeliberix -p"$(grep MARIADB_PASSWORD .env | cut -d= -f2)" -e "
  DESCRIBE deliberix.survey;
  DESCRIBE deliberix.survey_participant;
  DESCRIBE deliberix.response;
" | grep -E "allow_anonymous|anon_session_id|user_id"
```

Expected: `allow_anonymous`, `anon_session_id` columns visible; `user_id` shows `YES` for Null on both participant and response.

- [ ] **Step 4: Rebuild + restart server and client**

```bash
docker compose up -d --build server client
```

Watch the logs until each container is healthy:

```bash
docker compose logs -f server client
```

Look for `server listening on :8080` and the Angular Caddy/nginx serving the SPA.

- [ ] **Step 5: Health checks**

```bash
curl -s https://deliberix.com/api/v1/survey/public | jq .
# Should return a (possibly empty) list, not an error.
```

Verify the CORS change didn't break authenticated callers — log into the SPA and make sure the survey list still loads.

- [ ] **Step 6: Smoke test in production**

Repeat the same end-to-end smoke test from Task 14 on the prod URL:

1. Create a test survey with anon enabled.
2. Open the share link in an incognito window.
3. Join anonymously, vote, view results.
4. Submit a statement, moderate it as admin.
5. Verify the anon participant in the admin list.

- [ ] **Step 7: Rollback plan (if needed)**

If anything goes wrong:

```bash
# Revert the code
cd ~/projects/deliberix
git revert <range>

# Rebuild without dropping the migration
docker compose up -d --build server client
```

The schema change is additive only — it does not need to be rolled back. Existing surveys keep `allow_anonymous = 0`, anon rows are absent. The old binary will simply ignore the new columns.

If the migration itself needs to be rolled back (extremely unlikely — only if a CHECK constraint or unique key blocks legitimate writes you can't identify), the inverse is:

```sql
ALTER TABLE response
  DROP CHECK chk_response_identity,
  DROP INDEX uq_statement_anon,
  DROP COLUMN anon_session_id,
  MODIFY user_id INT UNSIGNED NOT NULL;

ALTER TABLE survey_participant
  DROP CHECK chk_participant_identity,
  DROP INDEX uq_survey_anon,
  DROP COLUMN anon_session_id,
  MODIFY user_id INT UNSIGNED NOT NULL;

ALTER TABLE survey
  DROP COLUMN allow_anonymous;
```

(But this only works if no anon rows exist. With anon rows present, `MODIFY user_id NOT NULL` would fail — clean anon rows first: `DELETE FROM survey_participant WHERE user_id IS NULL;` cascades to anon responses.)

- [ ] **Step 8: Update DEPLOY.md**

Add a "Migration on update" subsection so future deploys know to run pending migrations:

```markdown
## Migrations on update

`db/schema.sql` is only applied on a fresh MariaDB volume. When the
schema changes between releases, apply pending migrations explicitly
after `git pull`:

\`\`\`bash
docker exec -i deliberix-mariadb mysql -udeliberix -p"$MARIADB_PASSWORD" deliberix < db/migrations/<NNN>_<name>.sql
\`\`\`

Then run `docker compose up -d --build` to deploy the new code.
```

Commit the doc update:

```bash
git add DEPLOY.md
git commit -m "DEPLOY: document the manual migration step for updates"
git push origin master
```

---

## Self-review notes

- **Spec coverage check**: § 1 surface (Tasks 4, 6, 7, 8) ✓ · § 2 DB (Task 1) ✓ · § 3 auth/middleware (Tasks 2, 3) ✓ · § 4 client UX (Tasks 9, 10, 11, 13) ✓ · § 5 admin views (Task 12) ✓ · § 6 backward compat (verified in Task 14, Step 7) ✓.
- **Placeholder scan**: every code-changing step has full code, no TBDs.
- **Type consistency**: `Actor` used identically across server tasks; `allowAnonymous` consistent in client and server; the `*ByActor` helpers named consistently.
- **Cookie/CORS dependency**: Task 9 Step 1 sets `withCredentials: true` on the client, Task 8 Step 2 enables `AllowCredentials` and drops the wildcard origin on the server — they MUST land together (or both will be broken until both are deployed). Tasks 8 and 9 are sequenced so this is unambiguous.

## Known issues / future work

(Empty at plan-write time. Populated during Task 14 if defects surface.)
