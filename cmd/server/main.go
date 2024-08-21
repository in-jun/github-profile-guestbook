package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/in-jun/github-profile-comments/internal/auth"
	"github.com/in-jun/github-profile-comments/internal/config"
	"github.com/in-jun/github-profile-comments/internal/db"
	"github.com/in-jun/github-profile-comments/internal/handler"
	"github.com/in-jun/github-profile-comments/internal/middleware"
	"github.com/in-jun/github-profile-comments/web"
)

func main() {
	cfg := config.Load()

	database, err := db.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := database.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	if err := db.RunMigrations(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	stateBytes := make([]byte, 32)
	rand.Read(stateBytes)
	oauthState := base64.URLEncoding.EncodeToString(stateBytes)

	jwtSecret := []byte(cfg.JWTSecret)

	postLimiter := middleware.NewRateLimiter(30, 60*time.Second)
	getLimiter := middleware.NewRateLimiter(60, 60*time.Second)
	authLimiter := middleware.NewRateLimiter(10, 60*time.Second)

	authHandler := handler.NewAuthHandler(database, &handler.AuthHandlerConfig{
		OriginURL:       cfg.OriginURL,
		ClientID:        cfg.GitHubClientID,
		ClientSecret:    cfg.GitHubClientSecret,
		OAuthState:      oauthState,
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	})
	userHandler := handler.NewUserHandler(database)
	commentHandler := handler.NewCommentHandler(database)
	likeHandler := handler.NewLikeHandler(database)
	svgHandler := handler.NewSVGHandler(database)

	router := gin.Default()
	router.Use(auth.AuthMiddleware(database, jwtSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL))

	api := router.Group("/api")
	{
		api.GET("/", userHandler.GetMe)
		api.GET("/users", getLimiter.Handler(), userHandler.GetUsers)

		user := api.Group("/user")
		{
			user.POST("/:username/comments", postLimiter.Handler(), commentHandler.Create)
			user.GET("/:username/comments", getLimiter.Handler(), commentHandler.List)
			user.DELETE("/:username/comments", postLimiter.Handler(), commentHandler.Delete)
			user.GET("/:username/svg", svgHandler.GetSVG)
		}

		authGroup := api.Group("/auth")
		{
			authGroup.GET("/login", authLimiter.Handler(), authHandler.Login)
			authGroup.GET("/callback", authLimiter.Handler(), authHandler.Callback)
			authGroup.GET("/logout", authHandler.Logout)
		}

		like := api.Group("/like")
		{
			like.POST("/like/:commentID", postLimiter.Handler(), likeHandler.Like)
			like.POST("/remove-like/:commentID", postLimiter.Handler(), likeHandler.RemoveLike)
			like.POST("/dislike/:commentID", postLimiter.Handler(), likeHandler.Dislike)
			like.POST("/remove-dislike/:commentID", postLimiter.Handler(), likeHandler.RemoveDislike)
			like.POST("/owner-like/:commentID", postLimiter.Handler(), likeHandler.OwnerLike)
			like.POST("/owner-remove-like/:commentID", postLimiter.Handler(), likeHandler.OwnerRemoveLike)
		}
	}

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", web.FaviconICO)
	})

	router.GET("/:username", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", web.IndexHTML)
	})

	router.Run(":" + cfg.Port)
}
