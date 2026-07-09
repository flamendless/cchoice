package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"github.com/go-ini/ini"
	"go.uber.org/zap"
)

func MarshalThemeConfiguration(configuration map[string]string, configType enums.ThemeConfigType) (string, error) {
	switch configType {
	case enums.THEME_CONFIG_TYPE_JSON:
		b, err := json.Marshal(configuration)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case enums.THEME_CONFIG_TYPE_INI:
		file := ini.Empty()
		section, err := file.NewSection(ini.DefaultSection)
		if err != nil {
			return "", err
		}
		for k, v := range configuration {
			section.NewKey(k, v)
		}
		var buf bytes.Buffer
		if _, err := file.WriteTo(&buf); err != nil {
			return "", err
		}
		return buf.String(), nil
	default:
		return "", fmt.Errorf("unsupported theme configuration_type: %s", configType.String())
	}
}

func UnmarshalThemeConfiguration(raw string, configType enums.ThemeConfigType) (map[string]string, error) {
	switch configType {
	case enums.THEME_CONFIG_TYPE_JSON:
		result := map[string]string{}
		if raw == "" {
			return result, nil
		}
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return nil, err
		}
		return result, nil
	case enums.THEME_CONFIG_TYPE_INI:
		result := map[string]string{}
		if raw == "" {
			return result, nil
		}
		file, err := ini.Load([]byte(raw))
		if err != nil {
			return nil, err
		}
		for _, key := range file.Section(ini.DefaultSection).Keys() {
			result[key.Name()] = key.Value()
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported theme configuration_type: %s", configType.String())
	}
}

func sameConfig(incoming map[string]string, existing *Theme) bool {
	current, err := UnmarshalThemeConfiguration(existing.Configuration, existing.ConfigurationType)
	if err != nil {
		return false
	}
	if len(current) != len(incoming) {
		return false
	}
	for k, v := range incoming {
		if current[k] != v {
			return false
		}
	}
	return true
}

func (s *ThemeService) ReadColorFields(form url.Values) map[string]string {
	configuration := make(map[string]string, len(constants.ThemeColorKeys))
	for _, key := range constants.ThemeColorKeys {
		if v := form.Get("color_" + key); v != "" {
			configuration[key] = v
		}
	}
	return configuration
}

func (s *ThemeService) PreviewConfigurationFromQuery(ctx context.Context, query url.Values, logtag string) map[string]string {
	configuration := make(map[string]string, len(constants.ThemeColorKeys))
	if encodedTheme := query.Get("t"); encodedTheme != "" {
		decodedTheme, err := base64.RawURLEncoding.DecodeString(encodedTheme)
		if err == nil {
			if err := json.Unmarshal(decodedTheme, &configuration); err != nil {
				logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("message", "failed to parse theme preview payload"))
			}
		} else {
			logs.LogCtx(ctx).Warn(logtag, zap.Error(err), zap.String("message", "failed to decode theme preview payload"))
		}
	}
	if len(configuration) == 0 {
		configuration = s.ReadColorFields(query)
	}
	return configuration
}

func BuildThemeCSS(configuration map[string]string) string {
	colors := map[string]string{}
	for _, key := range constants.ThemeColorKeys {
		value := configuration[key]
		if !constants.ReThemeColor.MatchString(value) {
			continue
		}
		colors[key] = value
	}
	if len(colors) == 0 {
		return ""
	}

	get := func(key, fallback string) string {
		if value := colors[key]; value != "" {
			return value
		}
		return fallback
	}

	primary := get("primary", "#d9480f")
	primaryDark := get("primary-dark", primary)
	primaryEmphasis := get("primary-emphasis", primary)
	primaryHover := get("primary-hover", primaryDark)
	primaryMuted := get("primary-muted", primary)
	surface := get("surface", "#F7EFEA")
	accent := get("accent", primary)

	return fmt.Sprintf(`:root {
--color-primary: %[1]s;
--color-primary-dark: %[2]s;
--color-primary-emphasis: %[3]s;
--color-primary-hover: %[4]s;
--color-primary-muted: %[5]s;
--color-surface: %[6]s;
--color-accent: %[7]s;
}
.bg-primary, .file\:bg-primary::file-selector-button { background-color: %[1]s !important; }
.bg-primary-dark { background-color: %[2]s !important; }
.bg-primary-emphasis { background-color: %[3]s !important; }
.bg-primary-muted { background-color: %[5]s !important; }
.bg-surface { background-color: %[6]s !important; }
.bg-accent { background-color: %[7]s !important; }
.text-primary { color: %[1]s !important; }
.text-primary-dark { color: %[2]s !important; }
.border-primary { border-color: %[1]s !important; }
.border-primary-dark { border-color: %[2]s !important; }
.border-primary-muted { border-color: %[5]s !important; }
.stroke-primary { stroke: %[1]s !important; }
.fill-primary { fill: %[1]s !important; }
.accent-primary { accent-color: %[1]s !important; }
.focus\:border-primary:focus { border-color: %[1]s !important; }
.focus\:ring-primary:focus { --tw-ring-color: %[1]s !important; }
.peer-checked\:border-primary:is(:where(.peer):checked ~ *) { border-color: %[1]s !important; }
.custom-scrollbar::-webkit-scrollbar-track { background: %[6]s !important; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: %[3]s !important; border-color: %[6]s !important; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background-color: %[4]s !important; }
@media (hover: hover) {
.hover\:bg-primary:hover { background-color: %[1]s !important; }
.hover\:bg-primary-dark:hover, .hover\:file\:bg-primary-dark:hover::file-selector-button { background-color: %[2]s !important; }
.hover\:bg-primary-hover:hover { background-color: %[4]s !important; }
.hover\:bg-primary-muted:hover { background-color: %[5]s !important; }
.hover\:bg-surface:hover { background-color: %[6]s !important; }
.hover\:text-primary:hover { color: %[1]s !important; }
.hover\:text-primary-dark:hover { color: %[2]s !important; }
.hover\:border-primary:hover { border-color: %[1]s !important; }
.group-hover\:stroke-primary-dark:is(:where(.group):hover *) { stroke: %[2]s !important; }
}
`, primary, primaryDark, primaryEmphasis, primaryHover, primaryMuted, surface, accent)
}
