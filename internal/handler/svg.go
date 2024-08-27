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
	// Determine colors
	borderColor := "#e0e0e0"
	grayColor := "#666666"
	if boxColor == "black" {
		borderColor = "#333333"
		grayColor = "#999999"
	} else if boxColor == "transparent" {
		borderColor = "#e0e0e0"
		grayColor = "#666666"
	}

	return generateHTMLContent(userName, comments, textColor, boxColor, borderColor, grayColor)
}

func generateHTMLContent(userName string, comments []model.SvgCommentModel, textColor, boxColor, borderColor, grayColor string) string {
	var parts []string

	// SVG wrapper without height - will auto-size to content
	parts = append(parts, `<svg xmlns="http://www.w3.org/2000/svg" width="800">`)
	parts = append(parts, `<foreignObject x="0" y="0" width="100%" height="100%">`)
	parts = append(parts, `<div xmlns="http://www.w3.org/1999/xhtml">`)

	// Embedded styles
	parts = append(parts, `<style>
		@import url('https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable.min.css');
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { font-family: "Pretendard Variable", Pretendard, -apple-system, sans-serif; }
	</style>`)

	// Main container
	parts = append(parts, fmt.Sprintf(`<div style="background: %s; color: %s; padding: 24px; min-height: 100%%;">`, boxColor, textColor))

	// Header
	parts = append(parts, fmt.Sprintf(`<div style="font-size: 24px; font-weight: 700; margin-bottom: 16px;">%s</div>`, template.HTMLEscapeString(userName)))
	parts = append(parts, fmt.Sprintf(`<div style="border-bottom: 1px solid %s; margin-bottom: 24px;"></div>`, borderColor))

	// Section title
	parts = append(parts, fmt.Sprintf(`<div style="font-size: 14px; font-weight: 700; color: %s; letter-spacing: 0.5px; margin-bottom: 24px;">COMMENTS</div>`, grayColor))

	// Comments
	if len(comments) == 0 {
		// Empty state
		parts = append(parts, fmt.Sprintf(`<div style="border: 1px solid %s; padding: 30px; text-align: center; color: %s;">`, borderColor, grayColor))
		parts = append(parts, `<div style="font-size: 24px; font-weight: 700; margin-bottom: 12px;">—</div>`)
		parts = append(parts, `<div style="font-size: 14px;">No comments yet</div>`)
		parts = append(parts, `</div>`)
	} else {
		for _, comment := range comments {
			parts = append(parts, fmt.Sprintf(`<div style="border: 1px solid %s; padding: 12px 16px; margin-bottom: 16px; display: flex; justify-content: space-between; align-items: flex-start;">`, borderColor))

			// Left: text content
			parts = append(parts, `<div style="flex: 1; margin-right: 16px; min-width: 0;">`)
			parts = append(parts, fmt.Sprintf(`<div style="font-weight: 700; margin-bottom: 4px;">%s</div>`, template.HTMLEscapeString(comment.Author)))
			parts = append(parts, fmt.Sprintf(`<div style="word-wrap: break-word; overflow-wrap: break-word; line-height: 1.5;">%s</div>`, template.HTMLEscapeString(comment.Content)))
			parts = append(parts, `</div>`)

			// Right: buttons
			parts = append(parts, `<div style="display: flex; gap: 8px; flex-shrink: 0;">`)
			parts = append(parts, fmt.Sprintf(`<div style="border: 1px solid %s; padding: 4px 12px; font-size: 12px; font-weight: 500;">+ %d</div>`, borderColor, comment.Likes))
			parts = append(parts, fmt.Sprintf(`<div style="border: 1px solid %s; padding: 4px 12px; font-size: 12px; font-weight: 500;">- %d</div>`, borderColor, comment.Dislikes))
			if comment.IsOwnerLiked {
				parts = append(parts, fmt.Sprintf(`<div style="border: 1px solid %s; padding: 4px 8px; font-size: 12px;">★</div>`, borderColor))
			}
			parts = append(parts, `</div>`)

			parts = append(parts, `</div>`) // End comment
		}
	}

	parts = append(parts, `</div>`) // End container
	parts = append(parts, `</div>`) // End xmlns div
	parts = append(parts, `</foreignObject>`)
	parts = append(parts, `</svg>`)

	return strings.Join(parts, "\n")
}
