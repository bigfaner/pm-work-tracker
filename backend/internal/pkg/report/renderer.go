package report

import (
	"fmt"
	"strings"
	"time"

	"pm-work-tracker/backend/internal/dto"
)

// RenderMarkdown converts a ReportPreview into a Markdown document.
func RenderMarkdown(preview *dto.ReportPreview, weekStart time.Time) []byte {
	var buf strings.Builder

	// Header: # 周报 YYYY-WXX
	_, isoWeek := weekStart.ISOWeek()
	fmt.Fprintf(&buf, "# 周报 %d-W%02d\n\n", weekStart.Year(), isoWeek)

	for _, section := range preview.Sections {
		// Main item section
		mainTitle := section.MainItem.Title
		if section.MainItem.IsKeyItem {
			mainTitle = "[重点] " + mainTitle
		}
		fmt.Fprintf(&buf, "## %s\n\n", mainTitle)

		for _, sub := range section.SubItems {
			fmt.Fprintf(&buf, "### %s\n\n", sub.Title)
			fmt.Fprintf(&buf, "完成度: %.0f%%\n\n", sub.Completion)

			if len(sub.Achievements) > 0 {
				buf.WriteString("**成果:**\n\n")
				for _, a := range sub.Achievements {
					fmt.Fprintf(&buf, "- %s\n", a)
				}
				buf.WriteString("\n")
			}

			if len(sub.Blockers) > 0 {
				buf.WriteString("**卡点:**\n\n")
				for _, b := range sub.Blockers {
					fmt.Fprintf(&buf, "- %s\n", b)
				}
				buf.WriteString("\n")
			}
		}
	}

	return []byte(buf.String())
}
