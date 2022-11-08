package app

import (
	"fmt"
	"math"
	"time"
)

func fmtBytes(size int64) string {
	switch {
	case size > 1<<30:
		return fmt.Sprintf("%.1f GiB", float64(size)/(1<<30))
	case size > 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(size)/(1<<20))
	case size > 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(size)/(1<<10))
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}

func fmtTimeSince(t time.Time) string {
	n := time.Now()
	switch {
	// less than a minute
	case n.Sub(t) < time.Minute:
		return "just now"
		// less than an hour
	case n.Sub(t) < time.Hour:
		val := math.Round(n.Sub(t).Minutes())
		s := "s"
		if val == 1 {
			s = ""
		}
		return fmt.Sprintf("%.0f minute%s ago", val, s)
		// less than half a day
	case n.Sub(t) < 12*time.Hour:
		val := math.Round(n.Sub(t).Hours())
		s := "s"
		if val == 1 {
			s = ""
		}
		return fmt.Sprintf("%.0f hour%s ago", val, s)

	// today:
	case n.Year() == t.Year() && n.YearDay() == t.YearDay():
		return t.Format("3:04 PM")
		// yesterday
	case n.Year() == t.Year() && n.YearDay() == t.YearDay()+1:
		return "Yesterday"
		// this week
	case n.Year() == t.Year() && n.YearDay() <= t.YearDay()+7:
		return t.Format("Monday")

		// this year
	case n.Year() == t.Year():
		return t.Format("Jan 2")
	default:
		return t.Format("Jan 2, 2006")
	}
}
