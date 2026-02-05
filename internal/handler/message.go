package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/in-jun/github-profile-guestbook/internal/model"
	"github.com/lib/pq"
)

var zalgoPattern = regexp.MustCompile(`[\p{Mn}\p{Me}\p{Mc}]`)

type MessageHandler struct {
	db *sql.DB
}

func NewMessageHandler(db *sql.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

func (h *MessageHandler) Create(c *gin.Context) {
	username := c.Param("username")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authorID := userID.(int64)

	var req struct {
		Content string `json:"content"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content not provided"})
		return
	}

	runes := []rune(req.Content)
	if len(runes) > 200 {
		req.Content = string(runes[:200])
	}

	if zalgoPattern.MatchString(req.Content) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content"})
		return
	}

	// Store raw content in DB, escape only when rendering (SVG, HTML)
	_, err := h.db.Exec(
		`INSERT INTO messages (receiver_id, author_id, content)
		 VALUES ((SELECT id FROM users WHERE github_login = $1), $2, $3)`,
		username, authorID, req.Content,
	)
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already has a message"})
		} else if isNullViolation(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message created"})
}

func (h *MessageHandler) List(c *gin.Context) {
	username := c.Param("username")

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE github_login = $1)", username).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitHub user not found"})
		return
	}

	var currentUserID *int64
	if uid, ok := c.Get("user_id"); ok {
		id := uid.(int64)
		currentUserID = &id
	}

	query := `SELECT
		m.id,
		a.github_login AS author_login,
		a.id           AS author_id,
		m.content,
		m.is_owner_liked,
		COUNT(CASE WHEN r.type = 1 THEN 1 END)                     AS likes,
		COUNT(CASE WHEN r.type = -1 THEN 1 END)                    AS dislikes,
		COALESCE(BOOL_OR(r.user_id = $2 AND r.type = 1), FALSE)    AS is_liked,
		COALESCE(BOOL_OR(r.user_id = $2 AND r.type = -1), FALSE)   AS is_disliked
	FROM messages m
	JOIN users a          ON a.id = m.author_id
	JOIN users recv       ON recv.id = m.receiver_id
	LEFT JOIN reactions r ON r.message_id = m.id
	WHERE recv.github_login = $1
	GROUP BY m.id, a.github_login, a.id, m.content, m.is_owner_liked
	ORDER BY
		CASE WHEN a.id = $2 THEN 0 ELSE 1 END,
		m.is_owner_liked DESC,
		(COUNT(CASE WHEN r.type = 1 THEN 1 END) - COUNT(CASE WHEN r.type = -1 THEN 1 END)) DESC`

	rows, err := h.db.Query(query, username, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer rows.Close()

	messages := make([]model.MessageResponse, 0)
	for rows.Next() {
		var mr model.MessageResponse
		var authorID int64
		if err := rows.Scan(&mr.ID, &mr.Author, &authorID, &mr.Content, &mr.IsOwnerLiked, &mr.Likes, &mr.Dislikes, &mr.IsLiked, &mr.IsDisliked); err != nil {
			continue
		}
		messages = append(messages, mr)
	}

	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) Delete(c *gin.Context) {
	username := c.Param("username")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authorID := userID.(int64)

	var receiverExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE github_login = $1)", username).Scan(&receiverExists)
	if !receiverExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitHub user not found"})
		return
	}

	result, err := h.db.Exec(
		`DELETE FROM messages
		 WHERE receiver_id = (SELECT id FROM users WHERE github_login = $1)
		   AND author_id = $2`,
		username, authorID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func isNullViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23502"
	}
	return false
}
