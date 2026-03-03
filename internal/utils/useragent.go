package utils

import (
	"cchoice/internal/types"
	"fmt"

	"github.com/medama-io/go-useragent"
	"github.com/medama-io/go-useragent/agents"
)

var uaParser = useragent.NewParser()

func ParseUserAgent(userAgent string) types.UserAgentInfo {
	if userAgent == "" {
		return types.UserAgentInfo{}
	}

	ua := uaParser.Parse(userAgent)
	return types.UserAgentInfo{
		Browser:        string(ua.Browser()),
		BrowserVersion: ua.BrowserVersion(),
		OS:             string(ua.OS()),
		Device:         string(ua.Device()),
	}
}

func FormatUserAgentDevice(info types.UserAgentInfo) string {
	if info.Browser == "" {
		return ""
	}

	device := formatDevice(info.Device)

	if info.BrowserVersion != "" {
		return fmt.Sprintf("%s %s on %s (%s)", info.Browser, info.BrowserVersion, info.OS, device)
	}

	return fmt.Sprintf("%s on %s (%s)", info.Browser, info.OS, device)
}

func formatDevice(device string) string {
	switch agents.Device(device) {
	case agents.DeviceDesktop:
		return "Desktop"
	case agents.DeviceMobile:
		return "Mobile"
	case agents.DeviceTablet:
		return "Tablet"
	case agents.DeviceTV:
		return "TV"
	case agents.DeviceBot:
		return "Bot"
	default:
		return device
	}
}
