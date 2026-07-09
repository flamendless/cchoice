package forms

type AdminTrackedLinkPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminTrackedLinkForm struct {
	Name           string `form:"name" validate:"required"`
	Slug           string `form:"slug" validate:"required"`
	DestinationURL string `form:"destination_url" validate:"required"`
	Source         string `form:"source"`
	Medium         string `form:"medium"`
	Campaign       string `form:"campaign"`
	Status         string `form:"status"`
}
