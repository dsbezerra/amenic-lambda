package timeutil

import (
	"fmt"
	"log"
	"time"
)

// Now ...
func Now() time.Time {
	return time.Now().In(time.UTC)
}

// StartOfDay returns the time for the start of the day.
func StartOfDay() time.Time {
	return FloorToDay(nil, nil)
}

// StartOfDayForTime returns the start of the day time for the given date.
func StartOfDayForTime(t *time.Time) time.Time {
	return FloorToDay(t, nil)
}

// IsToday checks whether a given date is the current day of execution or not.
func IsToday(t *time.Time) bool {
	if t == nil {
		return false
	}

	ts := StartOfDay()         // Today start
	ds := StartOfDayForTime(t) // Date start

	return ts.Equal(ds)
}

// FloorToDay round down the given time to start of the day.
func FloorToDay(t *time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc, _ = time.LoadLocation("America/Sao_Paulo")
	}

	if t == nil {
		tt := time.Now().In(loc)
		t = &tt
	}

	y, m, d := t.Date()
	rounded := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return rounded
}

// TimeToSimpleDateString converts a time type to a YYYY-MM-DD date string
func TimeToSimpleDateString(t *time.Time) string {
	if t == nil {
		return ""
	}

	y, m, d := t.Year(), t.Month(), t.Day()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
}

// FriendlyFormat converts a given time to a more friendly format such as DD/MM/YYYY (HH:MM)
func FriendlyFormat(t *time.Time, withTime bool) string {
	if t == nil {
		return ""
	}

	date := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	y, m, d, h, mn := date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute()
	if withTime {
		return fmt.Sprintf("%02d/%02d/%d %02d:%02d", d, m, y, h, mn)
	}

	return fmt.Sprintf("%02d/%02d/%d", d, m, y)
}

// TimeTrack ...
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
