package utils

import (
	"cchoice/internal/constants"
)

type StaffLogReference struct {
	Label string
	URL   string
}

func ParseStaffLogSuccessID(result string) (encodedID string, ok bool) {
	matches := constants.ReStaffLogSuccessID.FindStringSubmatch(result)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func BuildStaffLogReference(module, action, productSlug string) StaffLogReference {
	if module != constants.ModuleProducts {
		return StaffLogReference{}
	}
	if action != constants.ActionCreate && action != constants.ActionUpdate {
		return StaffLogReference{}
	}
	if productSlug == "" {
		return StaffLogReference{}
	}
	return StaffLogReference{
		Label: "View Product",
		URL:   URLf("/product/%s", productSlug),
	}
}
