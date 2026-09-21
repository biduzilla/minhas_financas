package cookies

import (
	"net/http"
	"os"
)

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	CookiePath         = "/"
)

func isProd() bool {
	return os.Getenv("APP_ENV") == "production"
}

func sameSiteMode() http.SameSite {
	if isProd() {
		return http.SameSiteStrictMode
	}

	return http.SameSiteLaxMode
}

func SetAccessTokenCookie(
	w http.ResponseWriter,
	token string,
	expiresInSeconds int64,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    token,
		Path:     CookiePath,
		HttpOnly: true,
		Secure:   isProd(),
		SameSite: sameSiteMode(),
		MaxAge:   int(expiresInSeconds),
	})
}

func SetRefreshTokenCookie(w http.ResponseWriter, token string, expiresInSeconds int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    token,
		Path:     CookiePath,
		HttpOnly: true,
		Secure:   isProd(),
		SameSite: sameSiteMode(),
		MaxAge:   int(expiresInSeconds),
	})
}

func ClearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{AccessTokenCookie, RefreshTokenCookie} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     CookiePath,
			HttpOnly: true,
			Secure:   isProd(),
			SameSite: sameSiteMode(),
			MaxAge:   -1,
		})
	}
}

func TokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(AccessTokenCookie); err == nil && c.Value != "" {
		return c.Value
	}

	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) > len(prefix) && header[:len(prefix)] == prefix {
		return header[len(prefix):]
	}
	return ""
}

func RefreshTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(RefreshTokenCookie); err == nil {
		return c.Value
	}
	return ""
}
