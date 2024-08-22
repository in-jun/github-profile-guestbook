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
		width         = 800
		padding       = 24
		headerHeight  = 70
		sectionMargin = 32
		commentHeight = 60
		commentGap    = 16
	)

	// Calculate total height
	numComments := len(comments)
	if numComments == 0 {
		numComments = 1 // For empty state
	}
	commentsHeight := numComments*commentHeight + (numComments-1)*commentGap
	totalHeight := headerHeight + sectionMargin + 20 + commentsHeight + padding*2

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

	var parts []string

	// SVG header with Pretendard font
	parts = append(parts, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, width, totalHeight))
	parts = append(parts, `<style>
		@import url('https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable.min.css');
		text { font-family: "Pretendard Variable", Pretendard, -apple-system, sans-serif; }
	</style>`)

	// Background
	parts = append(parts, fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, width, totalHeight, boxColor))

	// Header section
	headerY := padding
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="24" font-weight="700" fill="%s">%s</text>`,
		padding, headerY+28, textColor, template.HTMLEscapeString(userName)))
	parts = append(parts, fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
		padding, headerY+headerHeight-16, width-padding, headerY+headerHeight-16, borderColor))

	// Comments section title
	sectionY := headerY + headerHeight + sectionMargin
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="%s" letter-spacing="0.5">COMMENTS</text>`,
		padding, sectionY, grayColor))

	// Comments
	commentStartY := sectionY + 20
	if len(comments) == 0 {
		// Empty state
		emptyY := commentStartY
		parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
			padding, emptyY, width-padding*2, commentHeight, borderColor))
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="28" font-weight="700" fill="%s" text-anchor="middle">—</text>`,
			width/2, emptyY+30, grayColor))
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="%s" text-anchor="middle">No comments yet</text>`,
			width/2, emptyY+48, grayColor))
	} else {
		for i, comment := range comments {
			commentY := commentStartY + i*(commentHeight+commentGap)

			// Comment box
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				padding, commentY, width-padding*2, commentHeight, borderColor))

			// Author
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="%s">%s</text>`,
				padding+16, commentY+22, textColor, template.HTMLEscapeString(comment.Author)))

			// Content
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="%s">%s</text>`,
				padding+16, commentY+42, textColor, template.HTMLEscapeString(comment.Content)))

			// Buttons (right side)
			buttonX := width - padding - 200
			buttonY := commentY + 30

			// Like button
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="60" height="28" fill="none" stroke="%s" stroke-width="1"/>`,
				buttonX, buttonY-18, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s">+ %d</text>`,
				buttonX+12, buttonY-2, textColor, comment.Likes))

			// Dislike button
			buttonX += 68
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="60" height="28" fill="none" stroke="%s" stroke-width="1"/>`,
				buttonX, buttonY-18, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s">- %d</text>`,
				buttonX+12, buttonY-2, textColor, comment.Dislikes))

			// Owner like (star)
			if comment.IsOwnerLiked {
				buttonX += 68
				parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="36" height="28" fill="none" stroke="%s" stroke-width="1"/>`,
					buttonX, buttonY-18, borderColor))
				parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="%s">★</text>`,
					buttonX+10, buttonY-2, textColor))
			}
		}
	}

	parts = append(parts, "</svg>")
	return strings.Join(parts, "\n")
}
