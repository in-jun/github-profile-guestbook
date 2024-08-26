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
		width            = 800
		padding          = 24
		headerTextSize   = 24
		headerBaseline   = 28  // Distance from top to text baseline
		headerBottom     = 16  // Space from text to border line
		sectionGap       = 32  // Gap between header line and section title
		sectionTitleSize = 14
		titleToComments  = 16  // Gap from title to first comment
		commentPadding   = 16  // Padding inside comment box
		commentHeight    = 56  // Height of each comment box
		commentGap       = 16  // Gap between comment boxes
		bottomPadding    = 24
	)

	// Calculate layout positions
	y := padding

	// Header section
	headerTextY := y + headerBaseline
	headerLineY := headerTextY + headerBottom
	y = headerLineY

	// Section title
	y += sectionGap
	sectionTitleY := y + sectionTitleSize
	y = sectionTitleY

	// Comments section
	y += titleToComments
	commentsStartY := y

	// Calculate total height
	numComments := len(comments)
	if numComments == 0 {
		totalHeight := commentsStartY + commentHeight + bottomPadding
		return generateSVGContent(userName, comments, textColor, boxColor, width, totalHeight, headerTextY, headerLineY, sectionTitleY, commentsStartY)
	}

	totalCommentsHeight := numComments*commentHeight + (numComments-1)*commentGap
	totalHeight := commentsStartY + totalCommentsHeight + bottomPadding

	return generateSVGContent(userName, comments, textColor, boxColor, width, totalHeight, headerTextY, headerLineY, sectionTitleY, commentsStartY)
}

func generateSVGContent(userName string, comments []model.SvgCommentModel, textColor, boxColor string, width, totalHeight int, headerTextY, headerLineY, sectionTitleY, commentsStartY int) string {
	const (
		padding        = 24
		commentHeight  = 56
		commentGap     = 16
		commentPadding = 16
	)

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
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="24" font-weight="700" fill="%s">%s</text>`,
		padding, headerTextY, textColor, template.HTMLEscapeString(userName)))
	parts = append(parts, fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
		padding, headerLineY, width-padding, headerLineY, borderColor))

	// Comments section title
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="%s" letter-spacing="0.5">COMMENTS</text>`,
		padding, sectionTitleY, grayColor))

	// Comments
	if len(comments) == 0 {
		// Empty state
		emptyY := commentsStartY
		parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
			padding, emptyY, width-padding*2, commentHeight, borderColor))

		// Icon (centered vertically and horizontally)
		iconY := emptyY + commentHeight/2 - 10
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="32" font-weight="700" fill="%s" text-anchor="middle">—</text>`,
			width/2, iconY+24, grayColor))

		// Text (below icon)
		textY := iconY + 40
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="%s" text-anchor="middle">No comments yet</text>`,
			width/2, textY, grayColor))
	} else {
		for i, comment := range comments {
			commentY := commentsStartY + i*(commentHeight+commentGap)

			// Comment box
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				padding, commentY, width-padding*2, commentHeight, borderColor))

			// Text content (left side)
			textX := padding + commentPadding

			// Author (top)
			authorY := commentY + commentPadding + 14 // 16 padding + 14 baseline
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="%s">%s</text>`,
				textX, authorY, textColor, template.HTMLEscapeString(comment.Author)))

			// Content (below author, 8px gap)
			contentY := authorY + 20
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="%s">%s</text>`,
				textX, contentY, textColor, template.HTMLEscapeString(comment.Content)))

			// Buttons (right side, centered vertically)
			buttonY := commentY + commentHeight/2
			buttonStartX := width - padding - 16

			// Calculate button positions from right to left
			buttonWidth := 50
			buttonHeight := 24
			buttonGap := 8

			currentX := buttonStartX

			// Owner like (star) - rightmost if exists
			if comment.IsOwnerLiked {
				starWidth := 32
				currentX -= starWidth
				parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
					currentX, buttonY-buttonHeight/2, starWidth, buttonHeight, borderColor))
				parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="%s" text-anchor="middle" dominant-baseline="middle">★</text>`,
					currentX+starWidth/2, buttonY, textColor))
				currentX -= buttonGap
			}

			// Dislike button
			currentX -= buttonWidth
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, buttonWidth, buttonHeight, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s" text-anchor="middle" dominant-baseline="middle">- %d</text>`,
				currentX+buttonWidth/2, buttonY, textColor, comment.Dislikes))
			currentX -= buttonGap

			// Like button
			currentX -= buttonWidth
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, buttonWidth, buttonHeight, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s" text-anchor="middle" dominant-baseline="middle">+ %d</text>`,
				currentX+buttonWidth/2, buttonY, textColor, comment.Likes))
		}
	}

	parts = append(parts, "</svg>")
	return strings.Join(parts, "\n")
}
