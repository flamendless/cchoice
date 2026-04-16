package components

import (
	"slices"

	"cchoice/cmd/web/components/svg"
	"cchoice/cmd/web/models"
	"cchoice/internal/enums"
)

var baseStaffCards = []models.StaffCard{
	{Link: "/admin/staff/attendance", Title: "Attendance", Description: "View your attendance", Icon: svg.Clock("text-cchoice")},
	{Link: "/admin/staff/time-off", Title: "Time Off", Description: "Request a time off", Icon: svg.Box("text-cchoice")},
	{Link: "/admin/profile", Title: "Profile", Description: "View and manage your profile", Icon: svg.User("text-cchoice")},
}

var staffCardsWithRoles = []struct {
	Card        models.StaffCard
	AllowedRole enums.StaffRole
}{
	{Card: models.StaffCard{Link: "/admin/superuser/products/create", Title: "Create Product", Description: "Create a product", Icon: svg.Box("text-cchoice")}, AllowedRole: enums.STAFF_ROLE_CREATE_PRODUCT},
	{Card: models.StaffCard{Link: "/admin/cpoints/generate", Title: "Generate C-Points", Description: "Generate C-Points for a customer", Icon: svg.Lightning("text-cchoice")}, AllowedRole: enums.STAFF_ROLE_CREATE_CPOINTS},
	{Card: models.StaffCard{Link: "/admin/holidays", Title: "Holidays", Description: "Manage Philippines holidays", Icon: svg.Calendar("text-cchoice")}, AllowedRole: enums.STAFF_ROLE_MANAGE_HOLIDAYS},
	{Card: models.StaffCard{Link: "/admin/promos", Title: "Manage Promos", Description: "Manage promos", Icon: svg.Box("text-cchoice")}, AllowedRole: enums.STAFF_ROLE_MANAGE_PROMOS},
}

func filterStaffCardsByRole(roles []enums.StaffRole) []models.StaffCard {
	result := make([]models.StaffCard, 0, len(staffCardsWithRoles))
	for _, card := range staffCardsWithRoles {
		if slices.Contains(roles, card.AllowedRole) {
			result = append(result, card.Card)
		}
	}
	return result
}

func getAllStaffCards(roles []enums.StaffRole) []models.StaffCard {
	return slices.Concat(baseStaffCards, filterStaffCardsByRole(roles))
}
