package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"user_id": "Not logged in", "logged_in": false})
		return
	}

	var login string
	err := h.db.QueryRow("SELECT github_login FROM users WHERE id = $1", userID.(int64)).Scan(&login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": login, "logged_in": true})
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	rows, err := h.db.Query("SELECT id, github_id, github_login FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}
	defer rows.Close()

	type userRow struct {
		ID          int64  `json:"id"`
		GitHubID    int64  `json:"github_id"`
		GitHubLogin string `json:"github_login"`
	}

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.GitHubID, &u.GitHubLogin); err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}
