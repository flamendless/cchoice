package shop

import "cchoice/internal/utils"

const NoBrandHighlight = "__NONE__"

func brandSidePanelListURL(selectedBrandLabel string) string {
	switch {
	case selectedBrandLabel == NoBrandHighlight:
		return utils.URLWithParams("/brands/side-panel/list", map[string]string{
			"selected_brand": NoBrandHighlight,
		})
	case selectedBrandLabel != "":
		return utils.URLWithParams("/brands/side-panel/list", map[string]string{
			"selected_brand": selectedBrandLabel,
		})
	default:
		return utils.URL("/brands/side-panel/list")
	}
}

func resolveSidePanelHighlight(selectedBrandLabel string) string {
	if selectedBrandLabel == NoBrandHighlight {
		return ""
	}
	return selectedBrandLabel
}
