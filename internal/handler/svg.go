package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/in-jun/github-profile-guestbook/internal/model"
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
		svgContent := generateLoginPromptSVG(username)
		c.Writer.Header().Set("Content-Type", "image/svg+xml")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.String(http.StatusOK, svgContent)
		return
	}

	rows, err := h.db.Query(`SELECT
		c.id,
		a.github_login,
		c.content,
		c.is_owner_liked,
		COUNT(DISTINCT l.id)  AS likes,
		COUNT(DISTINCT d.id)  AS dislikes
	FROM messages c
	JOIN users a         ON a.id = c.author_id
	JOIN users r         ON r.id = c.receiver_id
	LEFT JOIN likes l    ON l.message_id = c.id
	LEFT JOIN dislikes d ON d.message_id = c.id
	WHERE r.github_login = $1
	GROUP BY c.id, a.github_login, c.content, c.is_owner_liked
	ORDER BY
		c.is_owner_liked DESC,
		(COUNT(DISTINCT l.id) - COUNT(DISTINCT d.id)) DESC`, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer rows.Close()

	messages := make([]model.SvgMessageModel, 0)
	for rows.Next() {
		var cm model.SvgMessageModel
		if err := rows.Scan(&cm.ID, &cm.Author, &cm.Content, &cm.IsOwnerLiked, &cm.Likes, &cm.Dislikes); err != nil {
			continue
		}
		messages = append(messages, cm)
	}

	svgContent := generateMessageBox(username, messages)

	c.Writer.Header().Set("Content-Type", "image/svg+xml")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.String(http.StatusOK, svgContent)
}

func generateMessageBox(userName string, messages []model.SvgMessageModel) string {
	const (
		width                  = 800
		padding                = 24
		headerFontSize         = 24
		headerBaseline         = 18
		headerBottomGap        = 16
		sectionTitleFontSize   = 14
		sectionTitleBaseline   = 11
		sectionTitleTopGap     = 24
		sectionTitleBottomGap  = 24
		titleDescent           = 3
		messageBoxHeight       = 102
		messageBoxGap          = 16
		messageBoxPadding      = 12
		buttonHeight           = 24
		buttonGap              = 8
		likeWidth              = 50
		dislikeWidth           = 50
		starWidth              = 32
		emptyBoxHeight         = 80
		bottomPadding          = 24
	)

	headerTextY := padding + headerBaseline
	headerLineY := headerTextY + headerBottomGap
	sectionTitleY := headerLineY + sectionTitleTopGap + sectionTitleBaseline
	messagesStartY := sectionTitleY + titleDescent + sectionTitleBottomGap

	var totalHeight int
	if len(messages) == 0 {
		totalHeight = messagesStartY + emptyBoxHeight + bottomPadding
	} else {
		messagesHeight := len(messages)*messageBoxHeight + (len(messages)-1)*messageBoxGap
		totalHeight = messagesStartY + messagesHeight + bottomPadding
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, width, totalHeight))
	builder.WriteString(`<style>
		svg { --bg-color: white; --text-color: black; --border-color: #e0e0e0; --gray-color: #666666; }
		@media (prefers-color-scheme: dark) {
			svg { --bg-color: black; --text-color: white; --border-color: #333333; --gray-color: #999999; }
		}
		text { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
	</style>`)

	builder.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="var(--bg-color)"/>`, width, totalHeight))

	builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" font-weight="700" fill="var(--text-color)">%s</text>`,
		padding, headerTextY, headerFontSize, template.HTMLEscapeString(userName)))

	builder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--border-color)" stroke-width="1"/>`,
		padding, headerLineY, width-padding, headerLineY))

	builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" font-weight="700" fill="var(--gray-color)" letter-spacing="0.5">GUESTBOOK</text>`,
		padding, sectionTitleY, sectionTitleFontSize))

	if len(messages) == 0 {
		emptyY := messagesStartY
		builder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="var(--border-color)" stroke-width="1"/>`,
			padding, emptyY, width-padding*2, emptyBoxHeight))

		iconBaseline := emptyY + 15 + 18
		builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="24" font-weight="700" fill="var(--gray-color)" text-anchor="middle">—</text>`,
			width/2, iconBaseline))

		textBaseline := iconBaseline + 6 + 12 + 11
		builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="14" fill="var(--gray-color)" text-anchor="middle">No messages yet</text>`,
			width/2, textBaseline))
	} else {
		for i, message := range messages {
			messageY := messagesStartY + i*(messageBoxHeight+messageBoxGap)

			builder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="var(--border-color)" stroke-width="1"/>`,
				padding, messageY, width-padding*2, messageBoxHeight))

			textX := padding + 16
			textWidth := width - padding*2 - 32
			authorY := messageY + messageBoxPadding + 11
			builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="14" font-weight="700" fill="var(--text-color)">%s</text>`,
				textX, authorY, template.HTMLEscapeString(message.Author)))

			contentY := messageY + messageBoxPadding + 21 + 4
			contentHeight := 21
			builder.WriteString(fmt.Sprintf(`<foreignObject x="%d" y="%d" width="%d" height="%d">`,
				textX, contentY, textWidth, contentHeight))
			builder.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml" style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; height: 100%;">`)
			builder.WriteString(fmt.Sprintf(`<div style="margin: 0; padding: 0; box-sizing: border-box; font-size: 14px; color: var(--text-color); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; word-break: break-word;">%s</div>`,
				template.HTMLEscapeString(message.Content)))
			builder.WriteString(`</div>`)
			builder.WriteString(`</foreignObject>`)

			buttonY := contentY + 21 + 8 + buttonHeight/2
			currentX := textX

			builder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="var(--border-color)" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, likeWidth, buttonHeight))
			builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="var(--text-color)" text-anchor="middle" dominant-baseline="middle">+ %d</text>`,
				currentX+likeWidth/2, buttonY, message.Likes))
			currentX += likeWidth + buttonGap

			builder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="var(--border-color)" stroke-width="1"/>`,
				currentX, buttonY-buttonHeight/2, dislikeWidth, buttonHeight))
			builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="500" fill="var(--text-color)" text-anchor="middle" dominant-baseline="middle">- %d</text>`,
				currentX+dislikeWidth/2, buttonY, message.Dislikes))
			currentX += dislikeWidth + buttonGap

			if message.IsOwnerLiked {
				builder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="var(--border-color)" stroke-width="1"/>`,
					currentX, buttonY-buttonHeight/2, starWidth, buttonHeight))
				builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="var(--text-color)" text-anchor="middle" dominant-baseline="middle">★</text>`,
					currentX+starWidth/2, buttonY))
			}
		}
	}

	builder.WriteString("</svg>")
	return builder.String()
}

func generateLoginPromptSVG(userName string) string {
	const (
		width           = 800
		height          = 160
		padding         = 24
		usernameFontSize = 24
		usernameBaseline = 18
		messageFontSize  = 14
		messageBaseline  = 11
		lineGap          = 16
		textGap          = 24
	)

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, width, height))
	builder.WriteString(`<style>
		svg { --bg-color: white; --text-color: black; --border-color: #e0e0e0; --gray-color: #666666; }
		@media (prefers-color-scheme: dark) {
			svg { --bg-color: black; --text-color: white; --border-color: #333333; --gray-color: #999999; }
		}
		text { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
	</style>`)

	builder.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="var(--bg-color)"/>`, width, height))

	usernameY := padding + usernameBaseline
	builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" font-weight="700" fill="var(--text-color)">%s</text>`,
		padding, usernameY, usernameFontSize, template.HTMLEscapeString(userName)))

	lineY := usernameY + lineGap
	builder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--border-color)" stroke-width="1"/>`,
		padding, lineY, width-padding, lineY))

	messageY := lineY + textGap + messageBaseline
	builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="var(--gray-color)">This profile hasn't been claimed yet.</text>`,
		padding, messageY, messageFontSize))

	subMessageY := messageY + textGap
	builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="%d" fill="var(--gray-color)">Visit github-profile-guestbook.injun.dev/%s to claim this profile.</text>`,
		padding, subMessageY, messageFontSize, template.HTMLEscapeString(userName)))

	builder.WriteString("</svg>")
	return builder.String()
}
