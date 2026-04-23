package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
)

func formatDepartures(line, stationName string, now time.Time, window time.Duration, deps []domain.Departure) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s at %s (next %d min, now %s)\n\n",
		line, stationName, int(window.Minutes()), now.Format("15:04"))
	for _, d := range deps {
		b.WriteString(formatRow(d))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatEmpty(line, stationName string, now time.Time, window time.Duration) string {
	return fmt.Sprintf("No %s departures from %s in the next %d min (now %s).",
		line, stationName, int(window.Minutes()), now.Format("15:04"))
}

func formatRow(d domain.Departure) string {
	timeCol := d.ScheduledTime.Format("15:04")
	delayCol := formatDelay(d)
	platformCol := formatPlatform(d.Platform)
	return fmt.Sprintf("%s %-5s %-7s → %s", timeCol, delayCol, platformCol, d.Destination)
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
