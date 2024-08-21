package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/in-jun/github-profile-comments/internal/model"
)

type SVGHandler struct {
	db *sql.DB
}

func NewSVGHandler(db *sql.DB) *SVGHandler {
	return &SVGHandler{db: db}
}

func (h *SVGHandler) GetSVG(c *gin.Context) {
	username := c.Param("username")

	var exists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE github_login = $1)", username).Scan(&exists)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitHub user not found"})
		return
	}

	rows, err := h.db.Query(`SELECT
		c.id,
		a.github_login,
		c.content,
		c.is_owner_liked,
		COUNT(DISTINCT l.id)  AS likes,
		COUNT(DISTINCT d.id)  AS dislikes
	FROM comments c
	JOIN users a         ON a.id = c.author_id
	JOIN users r         ON r.id = c.receiver_id
	LEFT JOIN likes l    ON l.comment_id = c.id
	LEFT JOIN dislikes d ON d.comment_id = c.id
	WHERE r.github_login = $1
	GROUP BY c.id, a.github_login, c.content, c.is_owner_liked
	ORDER BY
		c.is_owner_liked DESC,
		(COUNT(DISTINCT l.id) - COUNT(DISTINCT d.id)) DESC`, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}
	defer rows.Close()

	comments := make([]model.SvgCommentModel, 0)
	for rows.Next() {
		var cm model.SvgCommentModel
		if err := rows.Scan(&cm.ID, &cm.Author, &cm.Content, &cm.IsOwnerLiked, &cm.Likes, &cm.Dislikes); err != nil {
			continue
		}
		comments = append(comments, cm)
	}

	theme := c.Query("theme")
	var bgColor, textColor string
	switch theme {
	case "black":
		bgColor, textColor = "black", "white"
	case "white":
		bgColor, textColor = "white", "black"
	case "transparent":
		bgColor, textColor = "transparent", "gray"
	default:
		bgColor, textColor = "white", "black"
	}

	svgContent := generateCommentBox(username, comments, textColor, bgColor)

	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.String(http.StatusOK, svgContent)
}

func generateCommentBox(userName string, comments []model.SvgCommentModel, textColor, boxColor string) string {
	const (
		additionalHeightPerComment = 35
		commentBoxMargin           = 5
	)

	numComments := len(comments)
	commentsHeight := numComments * additionalHeightPerComment
	inputBoxY := 60 + commentsHeight
	height := inputBoxY + additionalHeightPerComment

	svgHeader := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, 540, height)
	commentBox := fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s" stroke="%s" rx="5" ry="5"/>`, 540, height, boxColor, textColor)
	userNameText := fmt.Sprintf(`<text x="%d" y="20" font-family="Arial" font-size="16" fill="%s">%s</text>`, commentBoxMargin, textColor, userName)

	var commentBoxes []string
	for i, comment := range comments {
		commentY := 40 + i*additionalHeightPerComment
		box := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="30" fill="%s" stroke="%s" rx="5" ry="5"/>`, commentBoxMargin, commentY, 540-2*commentBoxMargin, boxColor, textColor)
		text := fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial" font-size="14" fill="%s">%s: %s</text>`, commentBoxMargin*2, commentY+20, textColor, template.HTMLEscapeString(comment.Author), comment.Content)
		commentBoxes = append(commentBoxes, box, text)
	}

	inputBox := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="30" fill="%s" stroke="%s" rx="5" ry="5"/>`, commentBoxMargin, inputBoxY, 540-2*commentBoxMargin, boxColor, textColor)
	inputText := fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial" font-size="14" fill="gray">Enter your comment...</text>`, commentBoxMargin*2, inputBoxY+20)

	parts := append([]string{svgHeader, commentBox, userNameText}, commentBoxes...)
	parts = append(parts, inputBox, inputText, "</svg>")
	return strings.Join(parts, "\n")
}
