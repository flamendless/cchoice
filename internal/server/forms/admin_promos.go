package forms

type AdminPromoPath struct {
	ID string `param:"id" validate:"required"`
}

type AdminPromoForm struct {
	Title         string `form:"title" validate:"required"`
	Description   string `form:"description" validate:"required"`
	MediaURL      string `form:"media_url"`
	StartDate     string `form:"start_date" validate:"required"`
	EndDate       string `form:"end_date" validate:"required"`
	Type          string `form:"type" validate:"required"`
	Status        string `form:"status"`
	BannerOnly    string `form:"banner_only"`
	Priority      int64  `form:"priority" validate:"required"`
	LinkType      string `form:"link_type"`
	TrackedLinkID string `form:"tracked_link_id"`
	LinkURL       string `form:"link_url"`
}
