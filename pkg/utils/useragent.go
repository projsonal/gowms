package utils

import (
	"regexp"
	"strings"
)

type DeviceType string

const (
	DeviceDesktop DeviceType = "desktop"
	DeviceMobile  DeviceType = "mobile"
	DeviceTablet  DeviceType = "tablet"
	DeviceUnknown DeviceType = "unknown"

	unknownLabel = "Unknown"
)

type DeviceInfo struct {
	Browser        string     `json:"browser"`
	BrowserVersion string     `json:"browser_version"`
	OS             string     `json:"os"`
	OSVersion      string     `json:"os_version"`
	DeviceType     DeviceType `json:"device_type"`
}

func (d DeviceInfo) String() string {
	return d.Browser + " " + d.BrowserVersion + " di " + d.OS + " " + d.OSVersion
}

type browserPattern struct {
	name    string
	pattern *regexp.Regexp
}

var browserPatterns = []browserPattern{
	{"Edge", regexp.MustCompile(`Edg(?:A|iOS)?/([\d.]+)`)},
	{"Opera", regexp.MustCompile(`OPR/([\d.]+)`)},
	{"Samsung Internet", regexp.MustCompile(`SamsungBrowser/([\d.]+)`)},
	{"Chrome", regexp.MustCompile(`Chrome/([\d.]+)`)},
	{"Firefox", regexp.MustCompile(`Firefox/([\d.]+)`)},
	{"Safari", regexp.MustCompile(`Version/([\d.]+).*Safari`)},
}

type osPattern struct {
	name       string
	pattern    *regexp.Regexp
	versionMap func(raw string) string
}

var osPatterns = []osPattern{
	{"Windows", regexp.MustCompile(`Windows NT ([\d.]+)`), windowsVersionName},
	{"Android", regexp.MustCompile(`Android ([\d.]+)`), passthrough},
	{"iOS", regexp.MustCompile(`iPhone OS ([\d_]+)`), iosVersionName},
	{"iPadOS", regexp.MustCompile(`iPad.*OS ([\d_]+)`), iosVersionName},
	{"macOS", regexp.MustCompile(`Mac OS X ([\d_.]+)`), macVersionName},
	{"Linux", regexp.MustCompile(`(Linux)`), passthrough},
}

func passthrough(raw string) string { return raw }

func windowsVersionName(raw string) string {
	switch raw {
	case "10.0":
		return "10/11"
	case "6.3":
		return "8.1"
	case "6.2":
		return "8"
	case "6.1":
		return "7"
	default:
		return raw
	}
}

func iosVersionName(raw string) string {
	return strings.ReplaceAll(raw, "_", ".")
}

func macVersionName(raw string) string {
	return strings.ReplaceAll(raw, "_", ".")
}

func ParseUserAgent(userAgent string) DeviceInfo {
	info := DeviceInfo{
		Browser:        unknownLabel,
		BrowserVersion: unknownLabel,
		OS:             unknownLabel,
		OSVersion:      unknownLabel,
		DeviceType:     DeviceUnknown,
	}
	if strings.TrimSpace(userAgent) == "" {
		return info
	}

	if name, version, ok := matchBrowser(userAgent); ok {
		info.Browser = name
		info.BrowserVersion = version
	}
	if name, version, ok := matchOS(userAgent); ok {
		info.OS = name
		info.OSVersion = version
	}
	info.DeviceType = detectDeviceType(userAgent)

	return info
}

func matchBrowser(userAgent string) (name, version string, ok bool) {
	for _, bp := range browserPatterns {
		if m := bp.pattern.FindStringSubmatch(userAgent); m != nil {
			return bp.name, m[1], true
		}
	}
	return "", "", false
}

func matchOS(userAgent string) (name, version string, ok bool) {
	for _, op := range osPatterns {
		m := op.pattern.FindStringSubmatch(userAgent)
		if m == nil {
			continue
		}
		raw := ""
		if len(m) > 1 {
			raw = m[1]
		}
		return op.name, op.versionMap(raw), true
	}
	return "", "", false
}

func detectDeviceType(userAgent string) DeviceType {
	lower := strings.ToLower(userAgent)
	switch {
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		return DeviceTablet
	case strings.Contains(lower, "mobile") || strings.Contains(lower, "iphone") || strings.Contains(lower, "android"):
		return DeviceMobile
	case strings.Contains(lower, "windows") || strings.Contains(lower, "macintosh") || strings.Contains(lower, "linux"):
		return DeviceDesktop
	default:
		return DeviceUnknown
	}
}
