package auth

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(db *sql.DB, secret []byte, atTTL, rtTTL int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if atCookie, err := c.Cookie("access_token"); err == nil {
			if claims, err := Parse(atCookie, secret); err == nil {
				c.Set("user_id", claims.UserID)
				c.Next()
				return
			}
		}

		if rtCookie, err := c.Cookie("refresh_token"); err == nil {
			if userID, ok := rotateRefreshToken(db, rtCookie, secret, atTTL, rtTTL, c); ok {
				c.Set("user_id", userID)
				c.Next()
				return
			}
		}

		c.Next()
	}
}

func rotateRefreshToken(db *sql.DB, rtRaw string, secret []byte, atTTL, rtTTL int, c *gin.Context) (int64, bool) {
	tokenHash := HashToken(rtRaw)

	var userID int64
	var expiresAt time.Time
	err := db.QueryRow("SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1", tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, false
	}

	if expiresAt.Before(time.Now()) {
		db.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", tokenHash)
		return 0, false
	}

	if _, err := db.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", tokenHash); err != nil {
		return 0, false
	}

	newRTRaw, err := GenerateRandomToken()
	if err != nil {
		return 0, false
	}
	newRTHash := HashToken(newRTRaw)
	newRTExpires := time.Now().Add(time.Duration(rtTTL) * time.Second)

	if _, err := db.Exec(
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, newRTHash, newRTExpires,
	); err != nil {
		return 0, false
	}

	newAT := NewAccessToken(userID, secret, atTTL)
	SetTokenCookies(c.Writer, newAT, newRTRaw, atTTL, rtTTL)

	return userID, true
}
