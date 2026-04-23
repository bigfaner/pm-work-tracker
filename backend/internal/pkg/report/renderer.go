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
	buf.WriteString(fmt.Sprintf("# 周报 %d-W%02d\n\n", weekStart.Year(), isoWeek))

	for _, section := range preview.Sections {
		// Main item section
		mainTitle := section.MainItem.Title
		if section.MainItem.IsKeyItem {
			mainTitle = "[重点] " + mainTitle
		}
		buf.WriteString(fmt.Sprintf("## %s\n\n", mainTitle))

		for _, sub := range section.SubItems {
			buf.WriteString(fmt.Sprintf("### %s\n\n", sub.Title))
			buf.WriteString(fmt.Sprintf("完成度: %.0f%%\n\n", sub.Completion))

			if len(sub.Achievements) > 0 {
				buf.WriteString("**成果:**\n\n")
				for _, a := range sub.Achievements {
					buf.WriteString(fmt.Sprintf("- %s\n", a))
				}
				buf.WriteString("\n")
			}

			if len(sub.Blockers) > 0 {
				buf.WriteString("**卡点:**\n\n")
				for _, b := range sub.Blockers {
					buf.WriteString(fmt.Sprintf("- %s\n", b))
				}
				buf.WriteString("\n")
			}
		}
	}

	return []byte(buf.String())
}
