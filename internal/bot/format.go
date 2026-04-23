package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
)

func formatDepartures(fromName, to string, now time.Time, window time.Duration, deps []domain.Departure) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s → %s (next %d min, now %s)\n\n",
		fromName, to, int(window.Minutes()), now.Format("15:04"))
	for _, d := range deps {
		b.WriteString(formatRow(d))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatEmpty(fromName, to string, now time.Time, window time.Duration) string {
	return fmt.Sprintf("No trains from %s toward '%s' in the next %d min (now %s).",
		fromName, to, int(window.Minutes()), now.Format("15:04"))
}

func formatRow(d domain.Departure) string {
	timeCol := d.ScheduledTime.Format("15:04")
	delayCol := formatDelay(d)
	platformCol := formatPlatform(d.Platform)
	category := strings.TrimSpace(d.TrainCategory)
	if category == "" {
		category = " "
	}
	return fmt.Sprintf("%s %-5s %-7s %-4s → %s", timeCol, delayCol, platformCol, category, d.Destination)
}

func formatDelay(d domain.Departure) string {
	if d.Status == domain.TrainStatusCancelled {
		return "canc"
	}
	if d.Delay > 0 {
		return fmt.Sprintf("+%d'", d.Delay)
	}
	if d.Delay < 0 {
		return fmt.Sprintf("%d'", d.Delay)
	}
	return ""
}

func formatPlatform(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "bin ?"
	}
	return "bin " + p
}
