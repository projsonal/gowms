package geoip

import (
	"fmt"
	"time"
)

func describeTimezone(tzName string) string {
	if tzName == "" {
		return ""
	}

	location, err := time.LoadLocation(tzName)
	if err != nil {
		return ""
	}

	abbreviation, offset := time.Now().In(location).Zone()
	return fmt.Sprintf("%s, %s", abbreviation, formatUTCOffset(offset))
}

func formatUTCOffset(offsetSeconds int) string {
	sign := "+"
	offset := offsetSeconds
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	hours := offset / 3600
	minutes := (offset % 3600) / 60

	if minutes == 0 {
		return fmt.Sprintf("UTC%s%d", sign, hours)
	}
	return fmt.Sprintf("UTC%s%d:%02d", sign, hours, minutes)
}
