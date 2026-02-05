package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	db *sql.DB
}

func NewLikeHandler(db *sql.DB) *LikeHandler {
	return &LikeHandler{db: db}
}

func (h *LikeHandler) parseMessageID(c *gin.Context) (int64, bool) {
	messageID := c.Param("messageID")
	id, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Message ID"})
		return 0, false
	}
	return id, true
}

func (h *LikeHandler) requireAuth(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return 0, false
	}
	return userID.(int64), true
}

func (h *LikeHandler) Like(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var authorID int64
	err := h.db.QueryRow("SELECT author_id FROM messages WHERE id = $1", messageID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like message"})
		return
	}

	if authorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't like your own message"})
		return
	}

	var existingType *int16
	err = h.db.QueryRow("SELECT type FROM reactions WHERE message_id = $1 AND user_id = $2", messageID, userID).Scan(&existingType)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like message"})
		return
	}
	if existingType != nil {
		if *existingType == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked this message"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already disliked this message"})
		}
		return
	}

	if _, err := h.db.Exec("INSERT INTO reactions (message_id, user_id, type) VALUES ($1, $2, 1)", messageID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message liked"})
}

func (h *LikeHandler) RemoveLike(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)", messageID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	result, err := h.db.Exec("DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND type = 1", messageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message not liked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Like removed"})
}

func (h *LikeHandler) Dislike(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var authorID int64
	err := h.db.QueryRow("SELECT author_id FROM messages WHERE id = $1", messageID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dislike message"})
		return
	}

	if authorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't dislike your own message"})
		return
	}

	var existingType *int16
	err = h.db.QueryRow("SELECT type FROM reactions WHERE message_id = $1 AND user_id = $2", messageID, userID).Scan(&existingType)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dislike message"})
		return
	}
	if existingType != nil {
		if *existingType == -1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already disliked this message"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked this message"})
		}
		return
	}

	if _, err := h.db.Exec("INSERT INTO reactions (message_id, user_id, type) VALUES ($1, $2, -1)", messageID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dislike message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message disliked"})
}

func (h *LikeHandler) RemoveDislike(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)", messageID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	result, err := h.db.Exec("DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND type = -1", messageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove dislike"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message not disliked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dislike removed"})
}

func (h *LikeHandler) OwnerLike(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	result, err := h.db.Exec(
		"UPDATE messages SET is_owner_liked = TRUE WHERE id = $1 AND receiver_id = $2 AND is_owner_liked = FALSE",
		messageID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like message"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var exists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)", messageID).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
			return
		}

		var isOwner bool
		h.db.QueryRow("SELECT receiver_id = $2 FROM messages WHERE id = $1", messageID, userID).Scan(&isOwner)
		if !isOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only like your own message"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message liked"})
}

func (h *LikeHandler) OwnerRemoveLike(c *gin.Context) {
	messageID, ok := h.parseMessageID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	result, err := h.db.Exec(
		"UPDATE messages SET is_owner_liked = FALSE WHERE id = $1 AND receiver_id = $2 AND is_owner_liked = TRUE",
		messageID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var exists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)", messageID).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
			return
		}

		var isOwner bool
		h.db.QueryRow("SELECT receiver_id = $2 FROM messages WHERE id = $1", messageID, userID).Scan(&isOwner)
		if !isOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only remove like from your own message"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "You have not liked this message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Like removed"})
}
