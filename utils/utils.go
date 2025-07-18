package utils

import (
	"fmt"
	"strings"
	"time"
)

// Parse "MM-YYYY" into time.Time
func ParseMonthYear(dateStr string) (time.Time, error) {
	// Split into month and year
	parts := strings.Split(dateStr, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid date format, expected MM-YYYY")
	}

	// Parse the month and year
	layout := "01-2006" // Go's reference time format (MM-YYYY)
	return time.Parse(layout, dateStr)
}


