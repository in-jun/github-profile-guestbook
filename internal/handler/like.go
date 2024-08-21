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

func (h *LikeHandler) parseCommentID(c *gin.Context) (int64, bool) {
	commentID := c.Param("commentID")
	id, err := strconv.ParseInt(commentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Comment ID"})
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
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var authorID int64
	err := h.db.QueryRow("SELECT author_id FROM comments WHERE id = $1", commentID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like comment"})
		return
	}

	if authorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't like your own comment"})
		return
	}

	var alreadyLiked, alreadyDisliked bool
	h.db.QueryRow(`SELECT
		EXISTS (SELECT 1 FROM likes    WHERE comment_id = $1 AND user_id = $2) AS already_liked,
		EXISTS (SELECT 1 FROM dislikes WHERE comment_id = $1 AND user_id = $2) AS already_disliked`,
		commentID, userID,
	).Scan(&alreadyLiked, &alreadyDisliked)

	if alreadyLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked this comment"})
		return
	}
	if alreadyDisliked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already disliked this comment"})
		return
	}

	if _, err := h.db.Exec("INSERT INTO likes (comment_id, user_id) VALUES ($1, $2)", commentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment liked"})
}

func (h *LikeHandler) RemoveLike(c *gin.Context) {
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	result, err := h.db.Exec("DELETE FROM likes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment not liked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Like removed"})
}

func (h *LikeHandler) Dislike(c *gin.Context) {
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var authorID int64
	err := h.db.QueryRow("SELECT author_id FROM comments WHERE id = $1", commentID).Scan(&authorID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dislike comment"})
		return
	}

	if authorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can't dislike your own comment"})
		return
	}

	var alreadyLiked, alreadyDisliked bool
	h.db.QueryRow(`SELECT
		EXISTS (SELECT 1 FROM likes    WHERE comment_id = $1 AND user_id = $2),
		EXISTS (SELECT 1 FROM dislikes WHERE comment_id = $1 AND user_id = $2)`,
		commentID, userID,
	).Scan(&alreadyLiked, &alreadyDisliked)

	if alreadyDisliked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already disliked this comment"})
		return
	}
	if alreadyLiked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked this comment"})
		return
	}

	if _, err := h.db.Exec("INSERT INTO dislikes (comment_id, user_id) VALUES ($1, $2)", commentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dislike comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment disliked"})
}

func (h *LikeHandler) RemoveDislike(c *gin.Context) {
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	result, err := h.db.Exec("DELETE FROM dislikes WHERE comment_id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove dislike"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment not disliked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dislike removed"})
}

func (h *LikeHandler) OwnerLike(c *gin.Context) {
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	result, err := h.db.Exec(
		"UPDATE comments SET is_owner_liked = TRUE WHERE id = $1 AND receiver_id = $2 AND is_owner_liked = FALSE",
		commentID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like comment"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var exists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}

		var isOwner bool
		h.db.QueryRow("SELECT receiver_id = $2 FROM comments WHERE id = $1", commentID, userID).Scan(&isOwner)
		if !isOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only like your own comment"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already liked comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment liked"})
}

func (h *LikeHandler) OwnerRemoveLike(c *gin.Context) {
	commentID, ok := h.parseCommentID(c)
	if !ok {
		return
	}
	userID, ok := h.requireAuth(c)
	if !ok {
		return
	}

	result, err := h.db.Exec(
		"UPDATE comments SET is_owner_liked = FALSE WHERE id = $1 AND receiver_id = $2 AND is_owner_liked = TRUE",
		commentID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var exists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)", commentID).Scan(&exists)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}

		var isOwner bool
		h.db.QueryRow("SELECT receiver_id = $2 FROM comments WHERE id = $1", commentID, userID).Scan(&isOwner)
		if !isOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only remove like from your own comment"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "You have not liked this comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Like removed"})
}
