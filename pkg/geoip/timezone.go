package geoip

import (
	"fmt"
	"strings"
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

var locationByTimezone = map[string]string{
	"Asia/Jakarta":      "Bandung, Indonesia",
	"Asia/Pontianak":    "Pontianak, Indonesia",
	"Asia/Makassar":     "Makassar, Indonesia",
	"Asia/Jayapura":     "Jayapura, Indonesia",
	"Asia/Singapore":    "Singapura, Singapura",
	"Asia/Kuala_Lumpur": "Kuala Lumpur, Malaysia",
}

func LocationFromTimezone(tzName string) string {
	if tzName == "" {
		return ""
	}
	if loc, ok := locationByTimezone[tzName]; ok {
		return loc
	}

	if _, err := time.LoadLocation(tzName); err != nil {
		return ""
	}
	parts := strings.Split(tzName, "/")
	city := strings.ReplaceAll(parts[len(parts)-1], "_", " ")
	return city
}
