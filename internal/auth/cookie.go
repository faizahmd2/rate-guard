package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
)

const adminCookieName = "rateguard_admin"

type CookieManager struct {
	secret []byte
}

func NewCookieManager(secret string) *CookieManager {
	return &CookieManager{
		secret: []byte(secret),
	}
}

func (c *CookieManager) Set(
	w http.ResponseWriter,
	username string,
) {
	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(username),
	)

	signature := c.sign(payload)

	value := payload + "." + signature

	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   60 * 60 * 24 * 30,
	})
}

func (c *CookieManager) Username(
	r *http.Request,
) (string, bool) {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil {
		return "", false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", false
	}

	payload := parts[0]
	providedSignature := parts[1]

	expectedSignature := c.sign(payload)

	if !hmac.Equal(
		[]byte(providedSignature),
		[]byte(expectedSignature),
	) {
		return "", false
	}

	usernameBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", false
	}

	return string(usernameBytes), true
}

func (c *CookieManager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (c *CookieManager) sign(payload string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(payload))

	return base64.RawURLEncoding.EncodeToString(
		mac.Sum(nil),
	)
}
