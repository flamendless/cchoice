package utils

import (
	"cchoice/internal/constants"
	"cchoice/internal/enums"
)

type StaffLogReference struct {
	Label  string
	URL    string
	NewTab bool
}

func ParseStaffLogSuccessID(result string) (encodedID string, ok bool) {
	matches := constants.ReStaffLogSuccessID.FindStringSubmatch(result)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func BuildStaffLogProductReference(slug, serial, status string) StaffLogReference {
	if enums.ParseProductStatusToEnum(status) == enums.PRODUCT_STATUS_ACTIVE {
		if slug == "" {
			return StaffLogReference{}
		}
		return StaffLogReference{
			Label:  "View Product in Shop",
			URL:    URLf("/product/%s", slug),
			NewTab: true,
		}
	}

	params := map[string]string{}
	if serial != "" {
		params["search_serial"] = serial
	}
	if productStatus := enums.ParseProductStatusToEnum(status); productStatus != enums.PRODUCT_STATUS_UNDEFINED {
		params["status"] = productStatus.String()
	}
	if len(params) == 0 {
		return StaffLogReference{}
	}

	return StaffLogReference{
		Label:  "View Product In Manage",
		URL:    URLWithParams("/admin/superuser/products", params),
		NewTab: false,
	}
}
