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
		width  = 800
		padding = 24

		// Header section
		headerFontSize   = 24
		headerBaseline   = 18 // 0.75 * fontSize
		headerBottomGap  = 16

		// Section title
		sectionTitleFontSize   = 14
		sectionTitleBaseline   = 11 // 0.78 * fontSize
		sectionTitleTopGap     = 24
		sectionTitleBottomGap  = 24
		titleDescent           = 3

		// Comment box
		commentBoxHeight  = 68 // 12 + 21 + 4 + 21 + 10 = 68 (padding + author + gap + content + padding)
		commentBoxGap     = 16
		commentBoxPadding = 12

		// Buttons
		buttonYOffset = 18 // from top of comment box (upper aligned)
		buttonHeight  = 24
		buttonGap     = 8
		likeWidth     = 50
		dislikeWidth  = 50
		starWidth     = 32

		// Empty state
		emptyBoxHeight = 80

		// Bottom padding
		bottomPadding = 24
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

	// Calculate layout positions
	headerTextY := padding + headerBaseline                                    // 24 + 18 = 42
	headerLineY := headerTextY + headerBottomGap                               // 42 + 16 = 58
	sectionTitleY := headerLineY + sectionTitleTopGap + sectionTitleBaseline   // 58 + 24 + 11 = 93
	commentsStartY := sectionTitleY + titleDescent + sectionTitleBottomGap     // 93 + 3 + 24 = 120

	// Calculate total height
	var totalHeight int
	if len(comments) == 0 {
		totalHeight = commentsStartY + emptyBoxHeight + bottomPadding // 120 + 80 + 24 = 224
	} else {
		commentsHeight := len(comments)*commentBoxHeight + (len(comments)-1)*commentBoxGap
		totalHeight = commentsStartY + commentsHeight + bottomPadding // 120 + n*68 + (n-1)*16 + 24 = 144 + n*84
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

	// Header text
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" font-weight="700" fill="%s">%s</text>`,
		padding, headerTextY, headerFontSize, textColor, template.HTMLEscapeString(userName)))

	// Header line
	parts = append(parts, fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
		padding, headerLineY, width-padding, headerLineY, borderColor))

	// Section title
	parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" font-weight="700" fill="%s" letter-spacing="0.5">COMMENTS</text>`,
		padding, sectionTitleY, sectionTitleFontSize, grayColor))

	// Comments
	if len(comments) == 0 {
		// Empty state box
		emptyY := commentsStartY
		parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
			padding, emptyY, width-padding*2, emptyBoxHeight, borderColor))

		// Empty icon and text (centered)
		iconY := emptyY + 40 - 9      // center - half of (24 + 12 + 14) / 2
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="24" font-weight="700" fill="%s" text-anchor="middle">—</text>`,
			width/2, iconY+18, grayColor))

		textY := iconY + 24 + 12 + 11 // icon + gap + text baseline
		parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="%s" text-anchor="middle">No comments yet</text>`,
			width/2, textY, grayColor))
	} else {
		for i, comment := range comments {
			commentY := commentsStartY + i*(commentBoxHeight+commentBoxGap)

			// Comment box border
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				padding, commentY, width-padding*2, commentBoxHeight, borderColor))

			// Author - pure SVG text
			textX := padding + 16
			textWidth := width - padding*2 - 32 - 158 // 158 = button area (50+8+50+8+32+10)
			authorY := commentY + commentBoxPadding + 11 // baseline offset
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="%s">%s</text>`,
				textX, authorY, textColor, template.HTMLEscapeString(comment.Author)))

			// Content - foreignObject for ellipsis only
			contentY := commentY + commentBoxPadding + 21 + 4 // author line-height + gap
			contentHeight := 21 // 1 line
			parts = append(parts, fmt.Sprintf(`<foreignObject x="%d" y="%d" width="%d" height="%d">`,
				textX, contentY, textWidth, contentHeight))
			parts = append(parts, `<div xmlns="http://www.w3.org/1999/xhtml" style="font-family: 'Pretendard Variable', Pretendard, sans-serif; height: 100%;">`)
			parts = append(parts, fmt.Sprintf(`<div style="font-size: 14px; color: %s; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden;">%s</div>`,
				textColor, template.HTMLEscapeString(comment.Content)))
			parts = append(parts, `</div>`)
			parts = append(parts, `</foreignObject>`)

			// Buttons (right side, from right to left)
			buttonY := commentY + buttonYOffset
			currentX := width - padding - 16

			// Owner like (star) - rightmost if exists
			if comment.IsOwnerLiked {
				currentX -= starWidth
				parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
					currentX, buttonY-buttonHeight/2, starWidth, buttonHeight, borderColor))
				parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="%s" text-anchor="middle" dominant-baseline="middle">★</text>`,
					currentX+starWidth/2, buttonY, textColor))
				currentX -= buttonGap
			}

			// Dislike button
			currentX -= dislikeWidth
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, dislikeWidth, buttonHeight, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s" text-anchor="middle" dominant-baseline="middle">- %d</text>`,
				currentX+dislikeWidth/2, buttonY, textColor, comment.Dislikes))
			currentX -= buttonGap

			// Like button
			currentX -= likeWidth
			parts = append(parts, fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, likeWidth, buttonHeight, borderColor))
			parts = append(parts, fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="%s" text-anchor="middle" dominant-baseline="middle">+ %d</text>`,
				currentX+likeWidth/2, buttonY, textColor, comment.Likes))
		}
	}

	parts = append(parts, "</svg>")
	return strings.Join(parts, "\n")
}
