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
