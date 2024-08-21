package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/in-jun/github-profile-comments/internal/auth"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
)

type AuthHandler struct {
	db              *sql.DB
	oauthCfg        *oauth2.Config
	oauthState      string
	jwtSecret       []byte
	accessTokenTTL  int
	refreshTokenTTL int
	originURL       string
}

type AuthHandlerConfig struct {
	OriginURL       string
	ClientID        string
	ClientSecret    string
	OAuthState      string
	JWTSecret       []byte
	AccessTokenTTL  int
	RefreshTokenTTL int
}

func NewAuthHandler(db *sql.DB, cfg *AuthHandlerConfig) *AuthHandler {
	return &AuthHandler{
		db: db,
		oauthCfg: &oauth2.Config{
			RedirectURL:  cfg.OriginURL + "/api/auth/callback",
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     ghOAuth.Endpoint,
		},
		oauthState:      cfg.OAuthState,
		jwtSecret:       cfg.JWTSecret,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		originURL:       cfg.OriginURL,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	redirectPath := c.Query("current")
	cfg := *h.oauthCfg
	if redirectPath != "" {
		cfg.RedirectURL += "?current=" + redirectPath
	}
	c.Redirect(http.StatusTemporaryRedirect, cfg.AuthCodeURL(h.oauthState))
}

func (h *AuthHandler) Callback(c *gin.Context) {
	if c.Query("state") != h.oauthState {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := h.oauthCfg.Exchange(c, c.Query("code"))
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := h.oauthCfg.Client(c, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	var user map[string]interface{}
	if err := dec.Decode(&user); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	login, ok1 := user["login"].(string)
	idNum, ok2 := user["id"].(json.Number)
	if !ok1 || !ok2 {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	githubID, err := idNum.Int64()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var internalID int64
	err = h.db.QueryRow(
		`INSERT INTO users (github_id, github_login)
		 VALUES ($1, $2)
		 ON CONFLICT (github_id) DO UPDATE SET github_login = EXCLUDED.github_login
		 RETURNING id`,
		githubID, login,
	).Scan(&internalID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	accessToken := auth.NewAccessToken(internalID, h.jwtSecret, h.accessTokenTTL)
	rtRaw, err := auth.GenerateRandomToken()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	rtHash := auth.HashToken(rtRaw)
	rtExpires := time.Now().Add(time.Duration(h.refreshTokenTTL) * time.Second)

	if _, err := h.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		internalID, rtHash, rtExpires,
	); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	auth.SetTokenCookies(c.Writer, accessToken, rtRaw, h.accessTokenTTL, h.refreshTokenTTL)

	if redirectPath := c.Query("current"); redirectPath != "" {
		c.Redirect(http.StatusFound, h.originURL+redirectPath)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if rtCookie, err := c.Cookie("refresh_token"); err == nil {
		rtHash := auth.HashToken(rtCookie)
		h.db.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", rtHash)
	}
	auth.ClearTokenCookies(c.Writer)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
